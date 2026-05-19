# Pharmacy POS — AGENTS.md

## Quick start

```sh
docker compose up -d data
make migrate-up DB_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
make assets    # tailwind + vue build
make run       # or `make dev` to rebuild assets first
```

Server listens on `:8000`. Stop with `Ctrl+C`.

## Commands

| Make target | What it does |
|---|---|
| `make create-migration MIGRATION_NAME=foo` | Creates new sequential migration pair |
| `make migrate-up DB_URL="..."` | Runs pending migrations |
| `make rollback-migration DB_URL="..."` | Rolls back one migration |
| `make clean-dirty-migration VERSION=N DB_URL="..."` | Force-marks migration version |
| `make tailwind` | Builds Tailwind only |
| `make build-frontend` | Builds Vue only |
| `make assets` | tailwind + build-frontend in order |
| `make dev` | assets + run |
| `cd backend && go test ./service/ -v` | Run tests (only file: `idempotency_test.go`) |

No lint, typecheck, or CI pipeline configured.

## Repo structure

- **Go backend** — stdlib `net/http`, Gorilla sessions/CSRF/WebSocket, `pgx`+`sqlx` for DB, `html/template` for server-rendered pages
- **Vue 3 + Vite** — builds into `backend/template/static/dist/`, embedded via Go `embed.FS`
- **TailwindCSS v4** — PostCSS plugin, run via `make tailwind`
- **36 PostgreSQL migrations** via `golang-migrate` in `backend/db/migrations/`

Architecture: `main.go` → router → controller (HTTP handlers) → service (business logic) → repository (sqlx queries) → DB.

## Key facts

- **Prices are kobo** (int cents). Divide by 100 for display (see `formatPrice` in `main.go`).
- **No ORM**. Raw SQL via `sqlx`. Service layer owns transaction logic (`BeginTx`/`CommitTx`/`RollbackTx`).
- **Session-based auth**, not JWT. `gorilla/sessions` with Argon2id password hashing.
- **CSRF middleware is commented out** in `main.go:114`.
- **WebSocket endpoint `/ws` bypasses the main middleware stack** (handled before middleware wrapping).
- **Permission keys** in `internal/constant/permissions.go`: `salestotal:view`, `stock:complete`, `inventory:edit`, `admin:access`.
- **Config** loaded from embedded `config/.env` via `godotenv` + `cleanenv`. `DB_HOST` default is `localhost`. `.env` overrides at build time.
- **6 Vite entry points** (vite.config.js `rollupOptions.input`). Vue apps mount on specific `<main id="...">` elements in Go templates.
- **Frontend uses shadcn-vue** components (reka-ui based), `@/` alias points to `src/`.
- **Idempotency** is implemented for both sales and receiving — keys are whitespace-normalized before storage.
- **Vite dev server** (`cd frontend && npm run dev`) proxies `/api` to `localhost:8000`.

## Testing

Only `backend/service/idempotency_test.go` exists. Tests use manual mock structs that embed `repository.PharmacyRepository` and override specific methods. No integration tests, no test DB setup.

## Docker

`compose.yml` defines two services: `data` (Postgres) and `server` (Go app). Run `docker compose up -d data` for local dev.
