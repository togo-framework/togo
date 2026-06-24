package togo

import (
	"strconv"
	"sync"
)

// Dialect captures the per-driver SQL differences the ORM needs. It lives in the
// kernel (not the orm plugin) so the kernel can expose Kernel.Dialect() without
// depending on the orm package — the orm plugin consumes togo.Dialect.
type Dialect struct {
	Placeholder func(n int) string // 1-based positional placeholder
	ILike       string             // case-insensitive LIKE operator
}

var (
	sqliteDialect   = Dialect{Placeholder: func(int) string { return "?" }, ILike: "LIKE"}
	postgresDialect = Dialect{Placeholder: func(n int) string { return "$" + strconv.Itoa(n) }, ILike: "ILIKE"}
	mysqlDialect    = Dialect{Placeholder: func(int) string { return "?" }, ILike: "LIKE"}
)

// Dialect registry. sqlite/postgres/mysql are built in (registered below); a DB
// driver plugin registers a dialect for any other driver from its init() via
// RegisterDialect — so the kernel stays driver-agnostic and drivers are plugins.
var (
	dialectMu sync.RWMutex
	dialects  = map[string]Dialect{}
)

func init() {
	RegisterDialect("sqlite", sqliteDialect)
	RegisterDialect("sqlite3", sqliteDialect)
	RegisterDialect("pgx", postgresDialect)
	RegisterDialect("postgres", postgresDialect)
	RegisterDialect("postgresql", postgresDialect)
	RegisterDialect("mysql", mysqlDialect)
}

// RegisterDialect registers the SQL dialect for a database/sql driver name. DB
// driver plugins call this from init() for drivers the kernel doesn't ship.
func RegisterDialect(driver string, d Dialect) {
	dialectMu.Lock()
	dialects[driver] = d
	dialectMu.Unlock()
}

// DialectFor returns the dialect registered for a database/sql driver name,
// falling back to the SQLite dialect for unknown drivers.
func DialectFor(driver string) Dialect {
	dialectMu.RLock()
	d, ok := dialects[driver]
	dialectMu.RUnlock()
	if ok {
		return d
	}
	return sqliteDialect
}
