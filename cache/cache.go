// Package cache defines the cache contract. Implementations live in their own
// repos (e.g. github.com/togo-framework/cache) and register a provider.
package cache

import "time"

// Cache is the key/value cache contract.
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration)
	Delete(key string)
}
