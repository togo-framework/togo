package togo

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestKernelSQL(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		dsn      string
		wantErr  error
		errText  string
		wantOpen int32 // live driver connections after SQL() + Close()
	}{
		{name: "missing DSN", driver: "togo-fake", dsn: "", errText: "DATABASE_URL is not set"},
		{name: "unknown driver", driver: "no-such-driver", dsn: "x", errText: `unknown driver "no-such-driver"`},
		{name: "open fails", driver: "togo-fake", dsn: "open-fail", wantErr: errFakeOpen},
		{name: "ping fails", driver: "togo-fake", dsn: "ping-fail", wantErr: errFakePing},
		{name: "ok", driver: "togo-fake", dsn: "ok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := testDriver.open.Load()
			k := &Kernel{Config: &Config{DBDriver: tt.driver, DatabaseURL: tt.dsn}}
			db, err := k.SQL(context.Background())
			switch {
			case tt.wantErr != nil:
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
			case tt.errText != "":
				if err == nil || !strings.Contains(err.Error(), tt.errText) {
					t.Fatalf("err = %v, want containing %q", err, tt.errText)
				}
			default:
				if err != nil || db == nil {
					t.Fatalf("SQL() = %v, %v; want handle", db, err)
				}
				again, _ := k.SQL(context.Background())
				if again != db {
					t.Fatal("second SQL() call did not reuse the cached handle")
				}
			}
			if err != nil && k.db != nil {
				t.Fatal("failed SQL() cached a handle")
			}
			k.Close()
			if got := testDriver.open.Load() - before; got != 0 {
				t.Fatalf("%d driver connection(s) leaked", got)
			}
		})
	}
}
