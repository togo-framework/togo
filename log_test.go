package togo

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestKernel returns a kernel with a discarding logger and a fresh hook bus.
func newTestKernel(logOut io.Writer) *Kernel {
	if logOut == nil {
		logOut = io.Discard
	}
	return &Kernel{
		Config: &Config{},
		Hooks:  newHooks(),
		Log:    slog.New(slog.NewTextHandler(logOut, nil)),
	}
}

func TestRecovery(t *testing.T) {
	tests := []struct {
		name     string
		panicVal any
		wantMsg  string
	}{
		{name: "string", panicVal: "boom", wantMsg: "panic: boom"},
		{name: "error", panicVal: errors.New("bad state"), wantMsg: "panic: bad state"},
		{name: "int", panicVal: 42, wantMsg: "panic: 42"},
		{name: "struct", panicVal: struct{ ID int }{7}, wantMsg: "panic: {7}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := newTestKernel(nil)
			var reported error
			k.Hooks.On("error", 0, func(_ context.Context, p any) error {
				reported, _ = p.(error)
				return nil
			})
			h := k.recovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic(tt.panicVal) }))

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

			if rec.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want 500", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if got := rec.Body.String(); got != `{"error":"internal server error"}` {
				t.Errorf("body = %s", got)
			}
			if reported == nil || reported.Error() != tt.wantMsg {
				t.Fatalf("error hook got %v, want %q", reported, tt.wantMsg)
			}
		})
	}
}

func TestRecoveryPassesThrough(t *testing.T) {
	k := newTestKernel(nil)
	h := k.recovery(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusTeapot) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want 418", rec.Code)
	}
}
