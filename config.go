package togo

import "os"

// Config holds runtime configuration resolved from the environment, so
// connections/URLs/endpoints stay dynamic (togo convention: .env + hooks).
type Config struct {
	Addr        string // HTTP listen address (ADDR, default ":8080")
	DBDriver    string // database/sql driver name (DB_DRIVER, default "sqlite"); providers register others
	DatabaseURL string // driver DSN (DATABASE_URL); defaults to ./togo.db for SQLite
	GraphQLPath string // GraphQL mount path (GRAPHQL_PATH, default "/graphql")
	RESTPath    string // REST API prefix (REST_PATH, default "/api")
	DocsPath    string // API docs path (DOCS_PATH, default "/docs")
	Locale      string // default locale for Kernel.T (LOCALE, default "en")
	LocaleDir   string // translation files directory (LOCALE_DIR, default "lang")
	StorageDir  string // local storage root (STORAGE_DIR, default "storage")
}

// LoadConfig reads configuration from environment variables with sane defaults.
// SQLite is the default driver — no external database needed to get started.
func LoadConfig() *Config {
	driver := env("DB_DRIVER", "sqlite")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" && driver == "sqlite" {
		dsn = "file:./togo.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_time_format=sqlite"
	}
	return &Config{
		Addr:        env("ADDR", ":8080"),
		DBDriver:    driver,
		DatabaseURL: dsn,
		GraphQLPath: env("GRAPHQL_PATH", "/graphql"),
		RESTPath:    env("REST_PATH", "/api"),
		DocsPath:    env("DOCS_PATH", "/docs"),
		Locale:      env("LOCALE", "en"),
		LocaleDir:   env("LOCALE_DIR", "lang"),
		StorageDir:  env("STORAGE_DIR", "storage"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
