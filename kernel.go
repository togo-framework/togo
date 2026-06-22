// Package togo is the microkernel of the togo framework. The kernel is
// deliberately thin: configuration, a hook/event bus, a plugin loader+registry,
// a database pool, and server bootstrap. Every capability — REST, GraphQL, auth,
// dashboard, resources — ships as a Plugin installed by the CLI and discovered
// here.
package togo

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/togo-framework/togo/cache"
	"github.com/togo-framework/togo/orm"
	"github.com/togo-framework/togo/queue"
	"github.com/togo-framework/togo/storage"
)

// Kernel is the shared runtime handed to every plugin and used by the app's
// entrypoint to mount REST/GraphQL and serve.
type Kernel struct {
	Config  *Config
	Router  chi.Router
	Hooks   *Hooks
	Log     *slog.Logger
	Cache   cache.Cache
	Queue   queue.Queue
	Storage storage.Storage

	db      *sql.DB
	plugins []Plugin
	booted  bool
}

// New constructs a kernel: loads config, logger, router (with recovery +
// request-logging middleware) and hook bus, and seeds the plugin list from
// auto-discovery (blank-imported plugin packages).
func New() *Kernel {
	log := newLogger()
	k := &Kernel{
		Config:  LoadConfig(),
		Router:  chi.NewMux(),
		Hooks:   newHooks(),
		Log:     log,
		Cache:   cache.NewMemory(),
		Queue:   queue.NewMemory(func(err error) { log.Error("queue job failed", "err", err) }),
		Storage: storage.NewFS(env("STORAGE_DIR", "storage")),
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

// SQL returns a lazily-opened database/sql handle using Config.DBDriver (default
// "sqlite"). Driver registration is the app's responsibility (blank-import the
// driver, or a DB provider plugin) — SQLite is core; Postgres/MySQL/etc. are
// provider plugins that register their driver and set DB_DRIVER.
func (k *Kernel) SQL(ctx context.Context) (*sql.DB, error) {
	if k.db != nil {
		return k.db, nil
	}
	if k.Config.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}
	d, err := sql.Open(k.Config.DBDriver, k.Config.DatabaseURL)
	if err != nil {
		return nil, err
	}
	if err := d.PingContext(ctx); err != nil {
		return nil, err
	}
	k.db = d
	return d, nil
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

// Dialect returns the ORM dialect for the configured driver.
func (k *Kernel) Dialect() orm.Dialect { return orm.DialectFor(k.Config.DBDriver) }

// Close releases the DB handle.
func (k *Kernel) Close() {
	if k.db != nil {
		_ = k.db.Close()
	}
}
