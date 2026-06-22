# togo documentation

togo is an open-source, API-first Go framework with a Laravel-grade developer
experience: a `togo` CLI, code generators, an ORM, a microkernel where everything
is a plugin, GraphQL + REST/OpenAPI out of the box, a Next.js frontend, built-in
auth, realtime, queues, and first-class AI/MCP integration.

## Contents
- [Getting started](getting-started.md)
- [Configuration](guide/configuration.md)
- [The core flow](guide/core-flow.md) — model → migrate → controller → view
- [ORM](guide/orm.md)
- [Validation](guide/validation.md)
- [Authentication & authorization](guide/auth.md)
- [Plugins (microkernel)](plugins.md)
- [Realtime & events](guide/realtime.md)
- [Mail & notifications](guide/mail-notifications.md)
- [Cache, queue, storage, workers](guide/services.md)
- [Search](guide/search.md)
- [Testing](guide/testing.md)
- [CLI reference](guide/cli.md)
- [AI & MCP](ai.md)
- [Generators](generators.md)
- [Deployment](deploy.md)

## Philosophy
A tiny kernel (config, router, hooks, plugin lifecycle, service container) with
every capability shipped as its own repo and registered via a provider on
blank-import. Install features with `togo install owner/repo`; SQLite is the
default DB (no Docker), swappable from `.env`.
