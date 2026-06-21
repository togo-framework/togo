// Package togo is the microkernel of the togo framework. The kernel is
// deliberately thin: configuration, a hook/event bus, a plugin loader+registry,
// a database driver registry, and server bootstrap. Every capability — REST,
// GraphQL, auth, dashboard, resources — ships as a Plugin installed by the CLI.
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

// Kernel is the shared runtime handed to every plugin. Concrete service fields
// (Config, Hooks, Router, DB, Resources, GraphQL, REST) are added as the
// framework phases land.
type Kernel struct {
	plugins []Plugin
}

// Use registers a plugin with the kernel. Plugins are sorted and booted by the
// runtime; auto-discovery wires first-party and installed plugins here.
func (k *Kernel) Use(p Plugin) *Kernel {
	k.plugins = append(k.plugins, p)
	return k
}

// Plugins returns the registered plugins.
func (k *Kernel) Plugins() []Plugin { return k.plugins }

// New constructs an empty kernel.
func New() *Kernel { return &Kernel{} }
