package togo

import "os"

// Config holds runtime configuration resolved from the environment, so
// connections/URLs/endpoints stay dynamic (togo convention: .env + hooks).
type Config struct {
	Addr        string
	DBDriver    string // database/sql driver name (default "sqlite"); providers register others
	DatabaseURL string
	GraphQLPath string
	RESTPath    string
	DocsPath    string
}

// LoadConfig reads configuration from environment variables with sane defaults.
// SQLite is the default driver — no external database needed to get started.
func LoadConfig() *Config {
	driver := env("DB_DRIVER", "sqlite")
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" && driver == "sqlite" {
		dsn = "file:./togo.db?_pragma=foreign_keys(1)&_time_format=sqlite"
	}
	return &Config{
		Addr:        env("ADDR", ":8080"),
		DBDriver:    driver,
		DatabaseURL: dsn,
		GraphQLPath: env("GRAPHQL_PATH", "/graphql"),
		RESTPath:    env("REST_PATH", "/api"),
		DocsPath:    env("DOCS_PATH", "/docs"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
