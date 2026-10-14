package togo

import "testing"

func TestLoadConfig(t *testing.T) {
	const sqliteDSN = "file:./togo.db?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_time_format=sqlite"
	defaults := Config{Addr: ":8080", DBDriver: "sqlite", DatabaseURL: sqliteDSN, GraphQLPath: "/graphql",
		RESTPath: "/api", DocsPath: "/docs", Locale: "en", LocaleDir: "lang", StorageDir: "storage"}
	with := func(f func(*Config)) Config { c := defaults; f(&c); return c }

	tests := []struct {
		name string
		env  map[string]string
		want Config
	}{
		{name: "defaults use sqlite with a local file", want: defaults},
		{
			name: "every variable overrides its default",
			env: map[string]string{"ADDR": ":9000", "DB_DRIVER": "pgx", "DATABASE_URL": "postgres://db/app",
				"GRAPHQL_PATH": "/gql", "REST_PATH": "/v1", "DOCS_PATH": "/reference", "LOCALE": "ar",
				"LOCALE_DIR": "i18n", "STORAGE_DIR": "/var/data"},
			want: Config{Addr: ":9000", DBDriver: "pgx", DatabaseURL: "postgres://db/app", GraphQLPath: "/gql",
				RESTPath: "/v1", DocsPath: "/reference", Locale: "ar", LocaleDir: "i18n", StorageDir: "/var/data"},
		},
		{
			name: "non-sqlite driver gets no default DSN",
			env:  map[string]string{"DB_DRIVER": "mysql"},
			want: with(func(c *Config) { c.DBDriver, c.DatabaseURL = "mysql", "" }),
		},
		{
			name: "explicit sqlite DSN wins over the default",
			env:  map[string]string{"DATABASE_URL": "file::memory:"},
			want: with(func(c *Config) { c.DatabaseURL = "file::memory:" }),
		},
		{
			name: "empty variable falls back to the default",
			env:  map[string]string{"ADDR": "", "LOCALE": ""},
			want: defaults,
		},
	}
	vars := []string{"ADDR", "DB_DRIVER", "DATABASE_URL", "GRAPHQL_PATH", "REST_PATH", "DOCS_PATH", "LOCALE", "LOCALE_DIR", "STORAGE_DIR"}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, v := range vars {
				t.Setenv(v, tt.env[v])
			}
			if got := *LoadConfig(); got != tt.want {
				t.Fatalf("LoadConfig() =\n %+v\nwant\n %+v", got, tt.want)
			}
		})
	}
}
