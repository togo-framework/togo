<h1 align="center">togo</h1>
<p align="center"><em>Go, the artisan way.</em></p>

**togo** is an open-source, API-first application framework for the **Go + sqlc +
Atlas + GraphQL/OpenAPI + Next.js** stack. It brings a Laravel-artisan-like
developer experience to Go: a powerful CLI, code generators, a plugin
marketplace, Supabase auth/dashboard by default, built-in Terraform deploys, and
first-class Claude Code / MCP integration.

This repository is the **microkernel** — the thin core every togo app builds on.
The kernel provides configuration, a priority-ordered hook/event bus, a plugin
loader + registry, a database driver registry, and server bootstrap. Everything
else (REST, GraphQL, auth, dashboard, resources) is a **plugin** installed by the
CLI.

## The togo organization

| Repo | Role |
|---|---|
| [`togo`](https://github.com/togo-framework/togo) | Microkernel (this repo) |
| [`cli`](https://github.com/togo-framework/cli) | The `togo` binary — generators, db, install, mcp, deploy |
| [`create-togo-app`](https://github.com/togo-framework/create-togo-app) | Project template rendered by `togo new` |
| [`plugin-template`](https://github.com/togo-framework/plugin-template) | Starter for togo plugins |
| [`togo-mcp`](https://github.com/togo-framework/togo-mcp) | MCP server exposing generators to AI agents |

## Quickstart

```bash
go install github.com/togo-framework/cli@latest   # installs the `togo` binary
togo new myapp && cd myapp
togo make:resource Post title:string body:text:nullable
togo generate          # sqlc + gqlgen + atlas + openapi
togo migrate && togo serve
```

## Principles

- **API-first** — every resource is exposed over GraphQL *and* REST/OpenAPI 3.1.
- **Everything is a plugin** — the core is a microkernel; capabilities are installed.
- **Generator-first** — `togo make:resource` scaffolds across six targets from one manifest.
- **Dynamic by default** — connections, URLs, endpoints come from `togo.yaml` + `.env` + hooks.
- **AI-native** — projects ship `.claude/` skills/agents + `.mcp.json` for Claude Code.

## License

MIT
