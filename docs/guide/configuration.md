# Configuration

All config is dynamic — from `.env` and `togo.yaml`. Never hard-code URLs or
connections. See `.env.example` for every key. Highlights:

| Key | Default | Purpose |
|-----|---------|---------|
| `DB_DRIVER` / `DATABASE_URL` | sqlite / file | database (swap to `pgx` for Postgres via `db-supabase`) |
| `CACHE_DRIVER` | memory | memory \| file \| database \| redis |
| `SESSION_DRIVER` | cookie | cookie \| database \| file \| redis |
| `AUTH_SECRET` | — | required in production (>= 32 bytes) |
| `MAIL_DRIVER` | log | smtp \| log \| resend |
| `SEARCH_DRIVER` | paradedb | paradedb \| elasticsearch \| opensearch |
| `LOG_LEVEL` / `LOG_FORMAT` | info / text | logging |
| `WORKER_AUTOSTART` | true | background workers |

Config is read by the kernel (`togo.Config`) and exposed to plugins via the
service container (`k.Get`/`k.Set`).
