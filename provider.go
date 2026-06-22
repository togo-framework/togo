package togo

import "sort"

// Provider contributes a service to the kernel. The kernel core is tiny (config,
// router, hooks, log, plugin lifecycle). Every OTHER capability — cache, queue,
// storage, realtime, i18n, db drivers — lives in its own provider package under
// togo/providers/* and self-registers in init(). A project opts into features by
// blank-importing those packages (chosen at `togo new`), so the core hard-codes
// nothing and any plugin can inject/override a provider globally.
type Provider interface {
	ProviderName() string
	ProviderPriority() int
	Provide(k *Kernel) error
}

// providerRegistry is the global, append-only provider list (defaults + plugins).
var providerRegistry []Provider

// RegisterProvider registers a service provider globally (call from init()).
// Higher Priority runs later and wins.
func RegisterProvider(p Provider) { providerRegistry = append(providerRegistry, p) }

type providerFunc struct {
	name     string
	priority int
	fn       func(*Kernel) error
}

func (p providerFunc) ProviderName() string    { return p.name }
func (p providerFunc) ProviderPriority() int   { return p.priority }
func (p providerFunc) Provide(k *Kernel) error { return p.fn(k) }

// RegisterProviderFunc registers a provider from a func.
func RegisterProviderFunc(name string, priority int, fn func(*Kernel) error) {
	RegisterProvider(providerFunc{name: name, priority: priority, fn: fn})
}

// Priority bands (providers pick their own; built-in feature packages use these).
const (
	PriorityCore    = 0  // log
	PriorityService = 50 // cache, storage, realtime, i18n
	PriorityLate    = 90 // queue (depends on log)
)

// Only the logger is a core default — it's foundational and always present.
// All feature providers live in togo/providers/* and register themselves when
// the project imports them.
func init() {
	RegisterProviderFunc("log", PriorityCore, func(k *Kernel) error { k.Log = newLogger(); return nil })
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

// Providers returns the names of registered providers in run order.
func Providers() []string {
	ps := append([]Provider(nil), providerRegistry...)
	sort.SliceStable(ps, func(i, j int) bool { return ps[i].ProviderPriority() < ps[j].ProviderPriority() })
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = p.ProviderName()
	}
	return out
}
