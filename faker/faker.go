// Package faker generates realistic fake data for factories and seeders.
// Dependency-free; uses math/rand.
package faker

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var (
	firstNames = []string{"Ahmad", "Sara", "Omar", "Lina", "Youssef", "Nour", "Karim", "Maya", "Tariq", "Hana", "Ali", "Dina"}
	lastNames  = []string{"Hassan", "Said", "Khalil", "Mansour", "Farouk", "Aziz", "Saleh", "Nasser", "Rashed", "Amin"}
	words      = []string{"alpha", "system", "data", "cloud", "signal", "vector", "matrix", "engine", "module", "stream", "node", "graph", "token", "query", "schema", "vertex", "atlas", "kernel"}
	domains    = []string{"example.com", "togo.dev", "mail.test", "acme.io"}
)

func pick(s []string) string { return s[rand.Intn(len(s))] }

// FirstName returns a random first name.
func FirstName() string { return pick(firstNames) }

// LastName returns a random last name.
func LastName() string { return pick(lastNames) }

// Name returns a random full name.
func Name() string { return pick(firstNames) + " " + pick(lastNames) }

// Email returns a random email address.
func Email() string {
	return strings.ToLower(pick(firstNames)+"."+pick(lastNames)) + fmt.Sprintf("%d@", rand.Intn(1000)) + pick(domains)
}

// Word returns a single random word.
func Word() string { return pick(words) }

// Words returns n random words joined by spaces.
func Words(n int) string {
	out := make([]string, n)
	for i := range out {
		out[i] = pick(words)
	}
	return strings.Join(out, " ")
}

// Sentence returns a capitalized sentence.
func Sentence() string {
	s := Words(6 + rand.Intn(6))
	return strings.ToUpper(s[:1]) + s[1:] + "."
}

// Paragraph returns several sentences.
func Paragraph() string {
	n := 3 + rand.Intn(3)
	out := make([]string, n)
	for i := range out {
		out[i] = Sentence()
	}
	return strings.Join(out, " ")
}

// Int returns a random int in [min, max].
func Int(min, max int) int {
	if max <= min {
		return min
	}
	return min + rand.Intn(max-min+1)
}

// Bool returns a random boolean.
func Bool() bool { return rand.Intn(2) == 1 }

// Float returns a random float in [min, max].
func Float(min, max float64) float64 { return min + rand.Float64()*(max-min) }

// UUID returns a random v4-ish UUID string.
func UUID() string {
	b := make([]byte, 16)
	for i := range b {
		b[i] = byte(rand.Intn(256))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// Time returns a random recent time.
func Time() time.Time {
	return time.Now().Add(-time.Duration(rand.Intn(60*24)) * time.Hour)
}
