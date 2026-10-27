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
| `make check-go-lines` | Shows a dry-run diff for Go lines over 100 characters |
| `make format-go-lines` | Applies `golines` wrapping to Go source |
| `make check-go-format` | Fails if any backend Go file needs `gofmt` |
| `make format-go` | Applies standard `gofmt` formatting |
| `cd backend && go test ./service/ -v` | Run unit tests (mock repositories) |
| `cd backend && PHARMACY_TEST_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" go test ./...` | Also run the database integration tests (migrations, purchase order lifecycle). Each test creates and drops its own database; without the variable they are skipped |

No lint, typecheck, or CI pipeline configured.

Install `golines` with `go install github.com/golangci/golines@latest` before
using the Go line-length targets. Set `GO_LINE_MAX_LEN` to override 100 columns.
Go formatting follows the standard `gofmt` rules; there is no `gofmt` config file.

## Repo structure

- **Go backend** — stdlib `net/http`, Gorilla sessions/CSRF/WebSocket, `pgx`+`sqlx` for DB, `html/template` for server-rendered pages
- **Vue 3 + vue-router + Vite** — builds into `backend/template/static/dist/`, embedded via Go `embed.FS`
- **TailwindCSS v4** — PostCSS plugin, run via `make tailwind`
- **41 PostgreSQL migrations** via `golang-migrate` in `backend/db/migrations/`

Architecture: `main.go` → router → controller (HTTP handlers) → service (business logic) → repository (sqlx queries) → DB.

## Key facts

- **Prices are kobo** (int cents). Divide by 100 for display (see `formatPrice` in `main.go`).
- **No ORM**. Raw SQL via `sqlx`. Service layer owns transaction logic (`BeginTx`/`CommitTx`/`RollbackTx`).
- **Session-based auth**, not JWT. `gorilla/sessions` with Argon2id password hashing.
- **CSRF** is enabled (gorilla/csrf via `middleware.CSRFMiddleware` in the `main.go` stack, keyed by `CSRF_KEY` or `SESSION_SECRET`). The SPA fetches a token from `GET /csrf-token` (`frontend/src/next/lib/csrf.js`).
- **WebSocket endpoint `/ws` bypasses the main middleware stack** (handled before middleware wrapping).
- **Permission keys** in `internal/constant/permissions.go`: `salestotal:view`, `stock:complete`, `inventory:edit`, `admin:access` (`/admin/` requires the last), `purchaseorder:create` (create/edit/close purchase orders).
- **Config** loaded from embedded `config/.env` via `godotenv` + `cleanenv`. `DB_HOST` default is `localhost`. `.env` overrides at build time.
- **2 Vite entry points** (vite.config.js `rollupOptions.input`): `next` (main SPA, served at `/app/*` via `next-dashboard.html`) and `login`. Go templates are only `next-dashboard.html`, `login-v2.html`, `error.html`.
- **Frontend assets are embedded** in the Go binary (`dist/` is untracked): run `make build-frontend` and restart after frontend changes, and before `docker compose build server`.
- **Frontend uses shadcn-vue** components (reka-ui based), `@/` alias points to `src/`.
- **Idempotency** is implemented for both sales and receiving — keys are whitespace-normalized before storage.
- **Vite dev server** (`cd frontend && npm run dev`) proxies `/api`, `/sales`, `/inventory`, `/admin`, `/stock-taking` to `localhost:8000`.

## Testing

Only `backend/service/idempotency_test.go` exists. Tests use manual mock structs that embed `repository.PharmacyRepository` and override specific methods. No integration tests, no test DB setup.

## Docker

`compose.yml` defines two services: `data` (Postgres) and `server` (Go app). Run `docker compose up -d data` for local dev.
