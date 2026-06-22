package togo

import (
	"github.com/togo-framework/togo/cache"
	"github.com/togo-framework/togo/i18n"
	"github.com/togo-framework/togo/queue"
	"github.com/togo-framework/togo/realtime"
	"github.com/togo-framework/togo/storage"
)

// Provider contributes a service to the kernel. The kernel core is tiny (config,
// router, hooks, plugin lifecycle); every other capability — log, cache, queue,
// storage, realtime, i18n — is a Provider. A plugin can RegisterProvider to
// replace a default (e.g. a Redis cache) without touching the kernel.
type Provider interface {
	ProviderName() string
	Provide(k *Kernel) error
}

// providerRegistry holds plugin-registered providers (override defaults).
var providerRegistry []Provider

// RegisterProvider registers a service provider. Later registrations override
// earlier ones for the same capability.
func RegisterProvider(p Provider) { providerRegistry = append(providerRegistry, p) }

// providerFunc adapts a func to a Provider.
type providerFunc struct {
	name string
	fn   func(*Kernel) error
}

func (p providerFunc) ProviderName() string  { return p.name }
func (p providerFunc) Provide(k *Kernel) error { return p.fn(k) }

// defaultProviders are the built-in capabilities, applied in dependency order
// (log first; queue needs log). Each is overridable via RegisterProvider.
func defaultProviders() []Provider {
	return []Provider{
		providerFunc{"log", func(k *Kernel) error { k.Log = newLogger(); return nil }},
		providerFunc{"i18n", func(k *Kernel) error { k.I18n = i18n.Load(k.Config.LocaleDir, k.Config.Locale); return nil }},
		providerFunc{"cache", func(k *Kernel) error { k.Cache = cache.NewMemory(); return nil }},
		providerFunc{"storage", func(k *Kernel) error { k.Storage = storage.NewFS(k.Config.StorageDir); return nil }},
		providerFunc{"realtime", func(k *Kernel) error { k.Realtime = realtime.NewBroker(); return nil }},
		providerFunc{"queue", func(k *Kernel) error {
			k.Queue = queue.NewMemory(func(err error) { k.Log.Error("queue job failed", "err", err) })
			return nil
		}},
	}
}

// applyProviders runs default then registered providers, so plugins override.
func (k *Kernel) applyProviders() {
	for _, p := range append(defaultProviders(), providerRegistry...) {
		if err := p.Provide(k); err != nil && k.Log != nil {
			k.Log.Error("provider failed", "provider", p.ProviderName(), "err", err)
		}
	}
}
