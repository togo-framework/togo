package togo

import (
	"context"
	"net/http"
	"time"
)

// Service contracts. The kernel defines these interfaces; every implementation
// lives in its own plugin repo (github.com/togo-framework/{cache,queue,storage,
// realtime,i18n,...}) and registers a provider. The core bundles no service impls.

// Cache is the key/value cache contract.
type Cache interface {
	Get(key string) (any, bool)
	Set(key string, value any, ttl time.Duration)
	Delete(key string)
}

// QueueHandler processes a dispatched job payload.
type QueueHandler func(ctx context.Context, payload any) error

// Queue dispatches named jobs to registered handlers.
type Queue interface {
	Handle(name string, h QueueHandler)
	Dispatch(ctx context.Context, name string, payload any) error
}

// Storage is the blob storage contract.
type Storage interface {
	Put(path string, data []byte) error
	Get(path string) ([]byte, error)
	Delete(path string) error
	Path(path string) string
}

// Broker fans out server-push (realtime) messages to connected clients.
type Broker interface {
	Publish(event, data string)
	Handler() http.HandlerFunc
}

// Translator resolves a key for a locale (trans()).
type Translator interface {
	T(locale, key string) string
}
