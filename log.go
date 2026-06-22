package togo

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// newLogger builds a slog logger from LOG_LEVEL (debug|info|warn|error) and
// LOG_FORMAT (json|text, default text).
func newLogger() *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	opts := &slog.HandlerOptions{Level: level}
	var h slog.Handler = slog.NewTextHandler(os.Stderr, opts)
	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "json" {
		h = slog.NewJSONHandler(os.Stderr, opts)
	}
	return slog.New(h)
}

// ReportError logs an error and fires the "error" hook so trackers (Sentry,
// GlitchTip, …) shipped as plugins can capture it. This is togo's central error
// reporting path — call it from anywhere with an error worth surfacing.
func (k *Kernel) ReportError(ctx context.Context, err error) {
	if err == nil {
		return
	}
	k.Log.Error("error", "err", err)
	_ = k.Hooks.Fire(ctx, "error", err)
}

// recovery turns panics into a logged 500 + error-hook, never crashing the server.
func (k *Kernel) recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				k.ReportError(r.Context(), &panicError{rec})
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(`{"error":"internal server error"}`))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// requestLogger logs one line per request (method, path, status, duration).
func (k *Kernel) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()
		next.ServeHTTP(ww, r)
		k.Log.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"dur", time.Since(start).String(),
		)
	})
}

// jsonErrorHandler returns an http.HandlerFunc that writes a JSON error body.
func jsonErrorHandler(status int, msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"status":` + itoa(status) + `,"error":"` + msg + `"}`))
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [4]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

type panicError struct{ v any }

func (e *panicError) Error() string { return "panic: " + sprint(e.v) }

func sprint(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	if e, ok := v.(error); ok {
		return e.Error()
	}
	return "unknown"
}
