package togo

import "context"

// Plugin is the contract every togo capability implements. The runtime boots
// plugins in ascending Priority order (0–100), mirroring laravilt's ordered
// service-provider lifecycle.
type Plugin interface {
	// Name uniquely identifies the plugin (e.g. "rest-huma", "auth-supabase").
	Name() string
	// Priority controls boot order; lower boots first. Infrastructure plugins
	// (config, db) use low values; feature plugins use higher ones.
	Priority() int
	// Register binds services, config, and hooks. No I/O or route mounting here.
	Register(k *Kernel) error
	// Boot starts the plugin: mount routes, register schema, run migrations.
	Boot(ctx context.Context, k *Kernel) error
}

// registry holds plugins registered via Register for auto-discovery. A plugin
// package adds itself in an init() func; the app blank-imports the package and
// the kernel picks it up — no manual wiring.
var registry []Plugin

// Register adds a plugin to the global auto-discovery registry.
func Register(p Plugin) { registry = append(registry, p) }

// Discovered returns a copy of the auto-registered plugins.
func Discovered() []Plugin {
	out := make([]Plugin, len(registry))
	copy(out, registry)
	return out
}
