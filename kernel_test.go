package togo

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKernelServices(t *testing.T) {
	var k Kernel // zero value: Set must initialise the container lazily
	if _, ok := k.Get("auth"); ok {
		t.Fatal("Get on empty kernel reported a service")
	}
	k.Set("auth", "v1")
	k.Set("auth", "v2")
	k.Set("mailer", nil)
	tests := []struct {
		key    string
		want   any
		wantOK bool
	}{
		{"auth", "v2", true},
		{"mailer", nil, true}, // a nil service is still registered
		{"missing", nil, false},
	}
	for _, tt := range tests {
		if got, ok := k.Get(tt.key); got != tt.want || ok != tt.wantOK {
			t.Errorf("Get(%q) = %v, %v; want %v, %v", tt.key, got, ok, tt.want, tt.wantOK)
		}
	}
}

type mapTranslator map[string]string

func (m mapTranslator) T(locale, key string) string {
	if v, ok := m[locale+"."+key]; ok {
		return v
	}
	return key
}

func TestKernelT(t *testing.T) {
	tr := mapTranslator{"ar.hello": "مرحبا", "en.hello": "Hello"}
	tests := []struct {
		name   string
		i18n   Translator
		locale string
		key    string
		want   string
	}{
		{"no translator returns key", nil, "en", "hello", "hello"},
		{"uses configured locale", tr, "ar", "hello", "مرحبا"},
		{"other locale", tr, "en", "hello", "Hello"},
		{"missing key", tr, "en", "bye", "bye"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := &Kernel{Config: &Config{Locale: tt.locale}, I18n: tt.i18n}
			if got := k.T(tt.key); got != tt.want {
				t.Fatalf("T(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

func TestKernelDialect(t *testing.T) {
	tests := []struct {
		driver      string
		placeholder string
		ilike       string
		sqlite      bool
	}{
		{"sqlite", "?", "LIKE", true},
		{"sqlite3", "?", "LIKE", true},
		{"pgx", "$3", "ILIKE", false},
		{"postgres", "$3", "ILIKE", false},
		{"mysql", "?", "LIKE", false},
		{"cockroach", "?", "LIKE", false}, // unknown: falls back to SQLite dialect
	}
	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			k := &Kernel{Config: &Config{DBDriver: tt.driver}}
			d := k.Dialect()
			if got := d.Placeholder(3); got != tt.placeholder {
				t.Errorf("Placeholder(3) = %q, want %q", got, tt.placeholder)
			}
			if d.ILike != tt.ilike {
				t.Errorf("ILike = %q, want %q", d.ILike, tt.ilike)
			}
			if got := isSQLite(tt.driver); got != tt.sqlite {
				t.Errorf("isSQLite = %v, want %v", got, tt.sqlite)
			}
		})
	}
}

func TestRegisterDialect(t *testing.T) {
	t.Cleanup(func() {
		dialectMu.Lock()
		delete(dialects, "togo-test-db")
		dialectMu.Unlock()
	})
	RegisterDialect("togo-test-db", Dialect{Placeholder: func(int) string { return ":p" }, ILike: "ILIKE"})
	if d := DialectFor("togo-test-db"); d.ILike != "ILIKE" || d.Placeholder(1) != ":p" {
		t.Fatalf("registered dialect not returned: %+v", d)
	}
}

// SQLite is single-writer, so its pool is capped at one connection; other
// drivers keep database/sql's unlimited default.
func TestKernelSQLPoolLimit(t *testing.T) {
	for _, tt := range []struct {
		driver  string
		maxOpen int
	}{{"sqlite", 1}, {"togo-fake", 0}} {
		t.Run(tt.driver, func(t *testing.T) {
			k := &Kernel{Config: &Config{DBDriver: tt.driver, DatabaseURL: "ok"}}
			db, err := k.SQL(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			defer k.Close()
			if got := db.Stats().MaxOpenConnections; got != tt.maxOpen {
				t.Errorf("MaxOpenConnections = %d, want %d", got, tt.maxOpen)
			}
		})
	}
}

func TestKernelCloseWithoutDB(*testing.T) {
	(&Kernel{}).Close() // must not panic
}

func TestReportError(t *testing.T) {
	var buf bytes.Buffer
	k := newTestKernel(&buf)
	var got []any
	k.Hooks.On("error", 0, func(_ context.Context, p any) error { got = append(got, p); return nil })

	k.ReportError(context.Background(), nil)
	if len(got) != 0 || buf.Len() != 0 {
		t.Fatalf("nil error was reported: hook=%v log=%q", got, buf.String())
	}

	boom := errors.New("boom")
	k.ReportError(context.Background(), boom)
	if len(got) != 1 || got[0] != boom {
		t.Fatalf("error hook payload = %v, want [boom]", got)
	}
	if !strings.Contains(buf.String(), "err=boom") {
		t.Fatalf("log = %q, want err=boom", buf.String())
	}
}

func TestRequestLogger(t *testing.T) {
	var buf bytes.Buffer
	k := newTestKernel(&buf)
	h := k.requestLogger(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/posts", nil))

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", rec.Code)
	}
	for _, want := range []string{"msg=request", "method=POST", "path=/api/posts", "status=201", "dur="} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("log %q missing %q", buf.String(), want)
		}
	}
}

// The kernel built by New answers unmatched routes with JSON errors and keeps
// recovery in front of every route.
func TestNewRouterDefaults(t *testing.T) {
	isolateRegistries(t)
	k := New()
	k.Log = newTestKernel(nil).Log
	k.Router.Get("/ok", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	k.Router.Get("/panic", func(http.ResponseWriter, *http.Request) { panic("boom") })

	tests := []struct {
		method, path string
		status       int
		body         string
	}{
		{http.MethodGet, "/ok", 200, "ok"},
		{http.MethodGet, "/missing", 404, `{"status":404,"error":"not found"}`},
		{http.MethodPost, "/ok", 405, `{"status":405,"error":"method not allowed"}`},
		{http.MethodGet, "/panic", 500, `{"error":"internal server error"}`},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			k.Handler().ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			if rec.Code != tt.status || rec.Body.String() != tt.body {
				t.Fatalf("got %d %q, want %d %q", rec.Code, rec.Body.String(), tt.status, tt.body)
			}
		})
	}
}

func TestItoa(t *testing.T) {
	for _, tt := range []struct {
		n    int
		want string
	}{{0, "0"}, {7, "7"}, {404, "404"}, {500, "500"}, {9999, "9999"}} {
		if got := itoa(tt.n); got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestServe(t *testing.T) {
	t.Run("boot error is returned before listening", func(t *testing.T) {
		var log []string
		k := newTestKernel(nil)
		k.Config.Addr = "127.0.0.1:0"
		k.Use(&stubPlugin{name: "bad", registerErr: errors.New("nope"), log: &log})
		if err := k.Serve(context.Background()); err == nil || err.Error() != "register bad: nope" {
			t.Fatalf("Serve err = %v", err)
		}
	})
	t.Run("listen error is returned", func(t *testing.T) {
		k := newTestKernel(nil)
		k.Config.Addr = "not-a-valid-address"
		if err := k.Serve(context.Background()); err == nil {
			t.Fatal("Serve on an invalid address returned nil")
		}
	})
}
