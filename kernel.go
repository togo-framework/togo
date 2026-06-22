// Package togo is the microkernel of the togo framework. The kernel is
// deliberately thin: configuration, a hook/event bus, a plugin loader+registry,
// a database pool, and server bootstrap. Every capability — REST, GraphQL, auth,
// dashboard, resources — ships as a Plugin installed by the CLI and discovered
// here.
package togo

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Kernel is the shared runtime handed to every plugin and used by the app's
// entrypoint to mount REST/GraphQL and serve.
type Kernel struct {
	Config *Config
	Router chi.Router
	Hooks  *Hooks
	Log    *slog.Logger

	pool    *pgxpool.Pool
	plugins []Plugin
	booted  bool
}

// New constructs a kernel: loads config, logger, router (with recovery +
// request-logging middleware) and hook bus, and seeds the plugin list from
// auto-discovery (blank-imported plugin packages).
func New() *Kernel {
	k := &Kernel{
		Config:  LoadConfig(),
		Router:  chi.NewMux(),
		Hooks:   newHooks(),
		Log:     newLogger(),
		plugins: Discovered(),
	}
	// Day-zero error handling + logging, applied before any routes are mounted.
	k.Router.Use(k.recovery, k.requestLogger)
	// The Go backend only serves API/GraphQL/docs, so unmatched routes return JSON.
	k.Router.NotFound(jsonErrorHandler(http.StatusNotFound, "not found"))
	k.Router.MethodNotAllowed(jsonErrorHandler(http.StatusMethodNotAllowed, "method not allowed"))
	return k
}

// Use explicitly registers a plugin (in addition to auto-discovered ones).
func (k *Kernel) Use(p Plugin) *Kernel {
	k.plugins = append(k.plugins, p)
	return k
}

// Plugins returns the registered plugins sorted by boot priority.
func (k *Kernel) Plugins() []Plugin {
	ps := append([]Plugin(nil), k.plugins...)
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].Priority() < ps[j].Priority() })
	return ps
}

// DB returns a lazily-opened Postgres pool from Config.DatabaseURL. Any database
// driver is supported via the connection string; Postgres (pgx) is the default.
func (k *Kernel) DB(ctx context.Context) (*pgxpool.Pool, error) {
	if k.pool != nil {
		return k.pool, nil
	}
	if k.Config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(ctx, k.Config.DatabaseURL)
	if err != nil {
		return nil, err
	}
	k.pool = pool
	return pool, nil
}

// Boot runs Register then Boot for every plugin in priority order. Safe to call
// once; subsequent calls are no-ops.
func (k *Kernel) Boot(ctx context.Context) error {
	if k.booted {
		return nil
	}
	ps := k.Plugins()
	for _, p := range ps {
		if err := p.Register(k); err != nil {
			return fmt.Errorf("register %s: %w", p.Name(), err)
		}
	}
	for _, p := range ps {
		if err := p.Boot(ctx, k); err != nil {
			return fmt.Errorf("boot %s: %w", p.Name(), err)
		}
	}
	k.booted = true
	return nil
}

// Serve boots plugins and starts the HTTP server on Config.Addr.
func (k *Kernel) Serve(ctx context.Context) error {
	if err := k.Boot(ctx); err != nil {
		return err
	}
	return http.ListenAndServe(k.Config.Addr, k.Router)
}

// Close releases the DB pool.
func (k *Kernel) Close() {
	if k.pool != nil {
		k.pool.Close()
	}
}
