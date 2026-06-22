# Getting started

Install the CLI:

```bash
curl -fsSL https://raw.githubusercontent.com/togo-framework/cli/main/install.sh | sh   # Linux/macOS
irm https://raw.githubusercontent.com/togo-framework/cli/main/install.ps1 | iex        # Windows
# or:
go install github.com/togo-framework/cli/cmd/togo@latest
```

Create and run an app:

```bash
togo new myapp && cd myapp
togo make:resource Post title:string body:text:nullable published:bool
togo generate          # sqlc + gqlgen + atlas + OpenAPI
togo migrate
togo serve             # backend + frontend together (installs web deps first run)
```

- API → http://localhost:8080 (GraphQL `/graphql`, REST `/api`, docs `/docs`)
- Web → http://localhost:3000

## Prerequisites

- **Go 1.26+**
- **Node 18+** (frontend)
- **sqlc** (`brew install sqlc`) for typed queries
- **Atlas** (`brew install ariga/tap/atlas`) for migrations
- **Postgres** (or Supabase) — `docker compose up -d`

## Keep the CLI current

```bash
togo upgrade
```
