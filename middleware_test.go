package togo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

// TestUseMiddlewareWraps verifies that UseMiddleware applies middleware as an
// outer wrapper (first-registered is outermost) regardless of route mounting —
// the property that lets any number of plugins register middleware without the
// chi "Use before routes" constraint.
func TestUseMiddlewareWraps(t *testing.T) {
	k := &Kernel{Router: chi.NewMux()}

	// Mount a route FIRST (this would forbid chi Router.Use afterwards).
	k.Router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("pong"))
	})

	var order []string
	k.UseMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "outer")
			w.Header().Set("X-Outer", "1")
			next.ServeHTTP(w, r)
		})
	})
	k.UseMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "inner")
			next.ServeHTTP(w, r)
		})
	})

	rec := httptest.NewRecorder()
	k.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/ping", nil))

	if rec.Body.String() != "pong" {
		t.Fatalf("body = %q, want pong", rec.Body.String())
	}
	if rec.Header().Get("X-Outer") != "1" {
		t.Fatal("outer middleware did not run")
	}
	if len(order) != 2 || order[0] != "outer" || order[1] != "inner" {
		t.Fatalf("middleware order = %v, want [outer inner]", order)
	}
}

func TestHandlerWithoutMiddlewareIsRouter(t *testing.T) {
	k := &Kernel{Router: chi.NewMux()}
	k.Router.Get("/x", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(204) })
	rec := httptest.NewRecorder()
	k.Handler().ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))
	if rec.Code != 204 {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}
