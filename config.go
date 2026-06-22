package togo

import "os"

// Config holds runtime configuration resolved from the environment, so
// connections/URLs/endpoints stay dynamic (togo convention: .env + hooks).
type Config struct {
	Addr        string
	DatabaseURL string
	GraphQLPath string
	RESTPath    string
	DocsPath    string
}

// LoadConfig reads configuration from environment variables with sane defaults.
func LoadConfig() *Config {
	return &Config{
		Addr:        env("ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
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
