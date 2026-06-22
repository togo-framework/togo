package togo

import (
	"sort"

	"github.com/togo-framework/togo/cache"
	"github.com/togo-framework/togo/i18n"
	"github.com/togo-framework/togo/queue"
	"github.com/togo-framework/togo/realtime"
	"github.com/togo-framework/togo/storage"
)

// Provider contributes a service to the kernel. The kernel core is tiny (config,
// router, hooks, plugin lifecycle); EVERY other capability — including the
// built-in log/cache/queue/storage/realtime/i18n — is registered as a Provider.
// Nothing is hard-coded in the core: providers run in ascending Priority, and a
// later provider for the same capability overrides an earlier one, so any plugin
// can inject or replace anything by calling RegisterProvider in its init().
type Provider interface {
	ProviderName() string
	ProviderPriority() int
	Provide(k *Kernel) error
}

// providerRegistry is the global, append-only provider list (defaults + plugins).
var providerRegistry []Provider

// RegisterProvider registers a service provider globally. Call it from a
// package init() (the built-in defaults do exactly this). Higher Priority runs
// later and wins.
func RegisterProvider(p Provider) { providerRegistry = append(providerRegistry, p) }

// providerFunc adapts a func to a Provider.
type providerFunc struct {
	name     string
	priority int
	fn       func(*Kernel) error
}

func (p providerFunc) ProviderName() string   { return p.name }
func (p providerFunc) ProviderPriority() int  { return p.priority }
func (p providerFunc) Provide(k *Kernel) error { return p.fn(k) }

// RegisterProviderFunc is a convenience for registering a provider from a func.
func RegisterProviderFunc(name string, priority int, fn func(*Kernel) error) {
	RegisterProvider(providerFunc{name: name, priority: priority, fn: fn})
}

// Priority bands for built-in providers (plugins pick their own).
const (
	PriorityCore    = 0  // log, config-derived
	PriorityService = 50 // cache, storage, realtime, i18n
	PriorityLate    = 90 // queue (depends on log)
)

// The built-in capabilities register themselves here — exactly like a plugin
// would. Remove/override any of them with RegisterProvider.
func init() {
	RegisterProviderFunc("log", PriorityCore, func(k *Kernel) error { k.Log = newLogger(); return nil })
	RegisterProviderFunc("i18n", PriorityService, func(k *Kernel) error {
		k.I18n = i18n.Load(k.Config.LocaleDir, k.Config.Locale)
		return nil
	})
	RegisterProviderFunc("cache", PriorityService, func(k *Kernel) error { k.Cache = cache.NewMemory(); return nil })
	RegisterProviderFunc("storage", PriorityService, func(k *Kernel) error { k.Storage = storage.NewFS(k.Config.StorageDir); return nil })
	RegisterProviderFunc("realtime", PriorityService, func(k *Kernel) error { k.Realtime = realtime.NewBroker(); return nil })
	RegisterProviderFunc("queue", PriorityLate, func(k *Kernel) error {
		k.Queue = queue.NewMemory(func(err error) { k.Log.Error("queue job failed", "err", err) })
		return nil
	})
}

// applyProviders runs every registered provider in ascending priority order.
func (k *Kernel) applyProviders() {
	ps := append([]Provider(nil), providerRegistry...)
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].ProviderPriority() < ps[j].ProviderPriority() })
	for _, p := range ps {
		if err := p.Provide(k); err != nil && k.Log != nil {
			k.Log.Error("provider failed", "provider", p.ProviderName(), "err", err)
		}
	}
}

// Providers returns the names of registered providers in run order (for debugging).
func Providers() []string {
	ps := append([]Provider(nil), providerRegistry...)
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].ProviderPriority() < ps[j].ProviderPriority() })
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.ProviderName()
	}
	return out
}
