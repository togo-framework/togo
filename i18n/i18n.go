// Package i18n provides JSON-keyed translations for the backend. Locale files
// live in a directory (lang/<locale>.json); T looks up a key with fallback.
package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Bundle holds loaded locales.
type Bundle struct {
	mu      sync.RWMutex
	locales map[string]map[string]string
	fallback string
}

// New returns an empty bundle with the given fallback locale.
func New(fallback string) *Bundle {
	return &Bundle{locales: map[string]map[string]string{}, fallback: fallback}
}

// Load reads every <locale>.json in dir into a bundle (missing dir is fine).
func Load(dir, fallback string) *Bundle {
	b := New(fallback)
	matches, _ := filepath.Glob(filepath.Join(dir, "*.json"))
	for _, f := range matches {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var m map[string]string
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		locale := strings.TrimSuffix(filepath.Base(f), ".json")
		b.locales[locale] = m
	}
	return b
}

// T translates key for locale, falling back to the fallback locale then the key.
func (b *Bundle) T(locale, key string) string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if m, ok := b.locales[locale]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	if m, ok := b.locales[b.fallback]; ok {
		if v, ok := m[key]; ok {
			return v
		}
	}
	return key
}
