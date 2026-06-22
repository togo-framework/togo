package togo

import "strconv"

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

// DialectFor returns the dialect for a database/sql driver name.
func DialectFor(driver string) Dialect {
	switch driver {
	case "pgx", "postgres", "postgresql":
		return postgresDialect
	case "mysql":
		return mysqlDialect
	default:
		return sqliteDialect
	}
}
