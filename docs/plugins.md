# Plugins

togo is a **microkernel**: every capability is a plugin. The kernel boots plugins
in ascending `Priority` (0–100) and discovers them automatically.

## Install

```bash
togo install togo-framework/plugin-auth-supabase
togo plugin:list
```

`togo install owner/repo` reads the plugin's `togo.plugin.yaml`, runs `go get`,
regenerates `internal/plugins/plugins.gen.go` (blank imports → auto-discovery),
records it in `togo.yaml`, and tidies modules. On the next `togo serve` the kernel
registers and boots it.

## Authoring a plugin

A plugin implements `togo.Plugin` and registers itself in `init()`:

```go
package myplugin

import (
	"context"
	"github.com/togo-framework/togo"
)

type Plugin struct{}

func init() { togo.Register(&Plugin{}) }

func (*Plugin) Name() string     { return "my-plugin" }
func (*Plugin) Priority() int    { return 50 }
func (p *Plugin) Register(k *togo.Kernel) error { return nil }       // bind services, middleware
func (p *Plugin) Boot(ctx context.Context, k *togo.Kernel) error {  // mount routes, run migrations
	k.Router.Get("/api/hello", func(w http.ResponseWriter, r *http.Request) { /* ... */ })
	return nil
}
```

Add a `togo.plugin.yaml` manifest (name, priority, backend package, env, migrations).
Start from [`plugin-template`](https://github.com/togo-framework/plugin-template).

## Kernel services

Plugins receive `*togo.Kernel`: `Config`, `Router` (chi), `Hooks` (priority event
bus), and `DB(ctx)` (pgx pool from `DATABASE_URL`).
