// Package i18n defines the translator contract. Implementations live in their
// own repos (e.g. github.com/togo-framework/i18n) and register a provider.
package i18n

// Translator resolves a key for a locale (trans()).
type Translator interface {
	T(locale, key string) string
}
