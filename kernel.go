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

)

// Kernel is the shared runtime handed to every plugin and used by the app's
// entrypoint to mount REST/GraphQL and serve.
type Kernel struct {
	Config  *Config
	Router  chi.Router
	Hooks   *Hooks
	Log     *slog.Logger
	Cache    Cache
	Queue    Queue
	Storage  Storage
	Realtime Broker
	I18n     Translator

	db       *sql.DB
	services map[string]any
	plugins  []Plugin
	mw       []func(http.Handler) http.Handler
	booted   bool
}

// UseMiddleware registers global HTTP middleware that wraps the whole router at
// serve time (outermost first-registered). Unlike chi's Router.Use — which a
// plugin can only call before ANY route is mounted — this can be called by any
// number of plugins at any provider priority: the middleware is applied as an
// outer wrapper in Handler(), after all providers have run. Prefer this over
// k.Router.Use for cross-cutting middleware (CORS, auth context, tracing).
func (k *Kernel) UseMiddleware(m ...func(http.Handler) http.Handler) {
	k.mw = append(k.mw, m...)
}

// Handler returns the kernel's HTTP handler: the router wrapped by every
// middleware registered via UseMiddleware (first-registered is outermost).
func (k *Kernel) Handler() http.Handler {
	var h http.Handler = k.Router
	for i := len(k.mw) - 1; i >= 0; i-- {
		h = k.mw[i](h)
	}
	return h
}

// Set stores an arbitrary service in the kernel container, so any plugin can
// inject capabilities the core doesn't know about (e.g. auth).
func (k *Kernel) Set(key string, v any) {
	if k.services == nil {
		k.services = map[string]any{}
	}
	k.services[key] = v
}

// Get retrieves a service registered via Set.
func (k *Kernel) Get(key string) (any, bool) {
	v, ok := k.services[key]
	return v, ok
}

// New constructs a kernel: loads config, logger, router (with recovery +
// request-logging middleware) and hook bus, and seeds the plugin list from
// auto-discovery (blank-imported plugin packages).
func New() *Kernel {
	// The kernel core is tiny: config, router, hooks. Every capability (log,
	// cache, queue, storage, realtime, i18n) is contributed by a Provider, so the
	// kernel is itself built over swappable plugins.
	k := &Kernel{
		Config:  LoadConfig(),
		Router:  chi.NewMux(),
		Hooks:    newHooks(),
		services: map[string]any{},
		plugins:  Discovered(),
	}
	k.Log = defaultLogger() // baseline; the log plugin (if installed) overrides it
	// Day-zero middleware MUST be registered before any provider mounts routes —
	// chi forbids Use() after routes exist (e.g. the auth plugin adds routes).
	k.Router.Use(k.recovery, k.requestLogger)
	// The Go backend only serves API/GraphQL/docs, so unmatched routes return JSON.
	k.Router.NotFound(jsonErrorHandler(http.StatusNotFound, "not found"))
	k.Router.MethodNotAllowed(jsonErrorHandler(http.StatusMethodNotAllowed, "method not allowed"))
	// Providers run last: some (auth) mount routes + their own middleware.
	k.applyProviders()
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
	// SQLite is single-writer: a multi-connection pool causes lock contention and
	// SQLITE_READONLY_DBMOVED under concurrent requests (e.g. an open SSE stream +
	// a write). Serialize to one connection. Server DBs keep the default pool.
	if isSQLite(k.Config.DBDriver) {
		d.SetMaxOpenConns(1)
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
	return http.ListenAndServe(k.Config.Addr, k.Handler())
}

// Dialect returns the ORM dialect for the configured driver.
func (k *Kernel) Dialect() Dialect { return DialectFor(k.Config.DBDriver) }

// isSQLite reports whether the driver is SQLite (single-writer).
func isSQLite(driver string) bool {
	return driver == "sqlite" || driver == "sqlite3"
}

// T translates a key in the configured locale (trans() equivalent).
func (k *Kernel) T(key string) string {
	if k.I18n == nil {
		return key
	}
	return k.I18n.T(k.Config.Locale, key)
}

// Close releases the DB handle.
func (k *Kernel) Close() {
	if k.db != nil {
		_ = k.db.Close()
	}
}
