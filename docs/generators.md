# Generators

`togo make:resource <Name> <field:type ...>` is the flagship generator. It emits
**per-resource fragments** across six targets plus regenerated registries, all
driven by `togo.resources.yaml` (the source of truth):

| File | Owner |
|---|---|
| `internal/models/<r>.go` | template (edit freely) |
| `internal/db/schema/<r>.sql`, `queries/<r>.sql` | template → sqlc |
| `db/atlas/schema/<r>.hcl` | template → Atlas |
| `internal/graph/schema/<r>.graphqls` | template → gqlgen resolvers |
| `internal/rest/<r>_handler.go` | template (Huma) |
| `web/lib/api/<r>.ts`, `web/app/<plural>/page.tsx`, hook | template (Next.js) |
| `internal/rest/registry.gen.go` | **regenerated** — do not edit |

## Field types

`string, text, int, bool, float, decimal, uuid, time, date, json`. Mark a field
optional with `:nullable` (or a quoted `?`):

```bash
togo make:resource Article title:string summary:text:nullable views:int
```

## Flags & other generators

- `--dry-run` preview, `--force` overwrite existing fragments.
- `make:model`, `make:query`, `make:graphql`, `make:api`, `make:migration`, `make:seeder`, `make:page`.
- `togo stub:publish` copies stubs to `./.togo/stubs` for per-project customization.

## `togo generate`

Runs the codegen pipeline: **sqlc → gqlgen → atlas diff → OpenAPI export**. Each
step is resilient (a missing tool warns and the pipeline continues).
