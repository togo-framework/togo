# CLI reference

| Command | Purpose |
|---------|---------|
| `togo new <app>` | scaffold a project (`--features`) |
| `togo dev` | run all services, hot reload, colored logs |
| `togo serve --host --port` | production-style server (deploy) |
| `togo make:model/controller/view/resource/action` | the core flow |
| `togo make:factory/seeder/query/migration/graphql/api/page` | single artifacts |
| `togo make:test / make:e2e` | tests (PHPUnit / Playwright) |
| `togo make:plugin <name>` | scaffold a plugin mini-app |
| `togo generate` | codegen pipeline |
| `togo migrate` / `togo seed` | database |
| `togo install <owner>/<repo>` | install a plugin |
| `togo format` / `togo lint` | code standards (Pint/PHPStan-style) |
| `togo agent "<desc>"` | agentic scaffolding via Claude Code + MCP |
| `togo mcp:install --role admin\|user` | publish the Claude ecosystem |
| `togo supabase up/down/status` | local Supabase stack |
| `togo deploy` | Terraform deploy |
| `togo upgrade` | self-update |
