package orm

import (
	"context"
	"testing"
)

// Malicious column/operator names must be rejected, not interpolated into SQL.
func TestWhereRejectsInjection(t *testing.T) {
	q := For[struct{}](nil, sqliteDialect, "users").
		Where("id; DROP TABLE users--", "=", 1)
	if _, err := q.Get(context.Background()); err == nil {
		t.Fatal("expected injection in column to be rejected")
	}
	q2 := For[struct{}](nil, sqliteDialect, "users").Where("id", "= 1 OR 1=1 --", 1)
	if _, err := q2.Get(context.Background()); err == nil {
		t.Fatal("expected injection in operator to be rejected")
	}
	q3 := For[struct{}](nil, sqliteDialect, "users").Order("id; DROP TABLE users")
	if _, err := q3.Get(context.Background()); err == nil {
		t.Fatal("expected injection in order to be rejected")
	}
}
