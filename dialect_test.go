package togo

import "testing"

func TestDialectReturning(t *testing.T) {
	cases := map[string]bool{
		"sqlite": false, "sqlite3": false, "pgx": false, "postgres": false,
		"postgresql": false, "mysql": true, "unknown": false,
	}
	for driver, want := range cases {
		if got := DialectFor(driver).NoReturning; got != want {
			t.Errorf("DialectFor(%q).NoReturning = %v, want %v", driver, got, want)
		}
	}
}
