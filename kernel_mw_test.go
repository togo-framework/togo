package togo

import (
	"net/http"
	"testing"
)

// A provider that mounts routes + its own middleware (like auth) must not make
// New() panic ("Use after routes"). Regression for the chi ordering bug.
func TestNewWithRouteMountingProvider(t *testing.T) {
	RegisterProviderFunc("test-routes", PriorityLate, func(k *Kernel) error {
		k.Router.Use(func(next http.Handler) http.Handler { return next })
		k.Router.Get("/_test", func(w http.ResponseWriter, r *http.Request) {})
		return nil
	})
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("New() panicked: %v", r)
		}
	}()
	k := New()
	if k.Router == nil {
		t.Fatal("nil router")
	}
}
