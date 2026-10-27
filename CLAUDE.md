# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

`AGENTS.md` in the repo root has the quick-start and a fuller facts list; read it too. This file covers what that one doesn't.

## Commands

```sh
docker compose up -d data                      # local Postgres
make migrate-up DB_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
make assets                                    # tailwind (Go templates) + vite build (Vue)
make run                                       # cd backend && go run .  → :8000
make dev                                       # assets + run
cd frontend && npm run dev                     # Vite dev server, proxies /api to :8000
```

- Migrations: `make create-migration MIGRATION_NAME=foo`, `make rollback-migration DB_URL=...`, `make clean-dirty-migration VERSION=N DB_URL=...`. Latest migrations: `000040_purchase_orders_and_suppliers` (`supplier`, `purchase_order`, `purchase_order_item`, `receiving_batch.supplier_id`, `product_batch.purchase_order_item_id`) and `000041_audit_log` (append-only `audit_log`, written in the same transaction as the change it describes; so far only purchase orders use it).
- Go formatting: `make check-go-format` / `make format-go`; line wrapping at 100 cols via `make check-go-lines` / `make format-go-lines` (needs `golines`; config in `.golangci.yml`).
- Database integration tests (`backend/db`, `backend/service/purchase_orders_integration_test.go`) are skipped unless `PHARMACY_TEST_DSN` points at a Postgres where the user can create databases, e.g. `PHARMACY_TEST_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" go test ./...`. Each test creates and drops its own database via `internal/testdb`, running the real migrations. Other tests use mocks.
- Tests: unit tests live in `backend/service`. Run all: `cd backend && go test ./service/ -v`; single: `go test ./service/ -run TestName -v`. Tests use hand-written mocks that embed `repository.PharmacyRepository` and override specific methods.
- No JS lint/test tooling and no CI.

## Architecture

Request flow: `main.go` → `adapter/http/router/*` → `controller/*` → `service/*` → `repository/*` (sqlx, raw SQL) → Postgres. Service contracts live in `service/interfaces.go`; the repository is one large `PharmacyRepository` interface in `repository/repository.go` with implementations split per domain file (e.g. `purchase_orders.go`). Adding a feature typically touches: migration, `internal/types`, repository interface + impl, service interface + impl, controller, router, then the frontend view.

Services own transactions (`BeginTx`/`CommitTx`/`RollbackTx`). Sales and receiving are idempotent via client-supplied keys, normalized before storage.

### Frontend

Almost everything is one Vue SPA (`frontend/src/next/`): Vue 3 + vue-router + shadcn-vue (reka-ui), served by Go under `/app/` (`router/app.go` serves the shell `next-dashboard.html` for any path and strips the `/app` prefix). Views in `next/views`, components in `next/components`, state in `store.js` and `composables/`. It calls JSON endpoints on `/api`, `/sales`, `/inventory`, `/admin`, `/stock-taking`. Go templates are now only `next-dashboard.html`, `login-v2.html` and `error.html`; Tailwind CSS for them is built by `make tailwind`. Vite has just two entries: `next` and `login` (`AGENTS.md` still says 6; that is stale).

Vite builds to `backend/template/static/dist/` (base `/static/dist/`); Go embeds it via `embed.FS` and resolves hashed filenames from the Vite manifest through the `ViteAsset` template function. **A frontend change is not visible from the Go server until `make build-frontend` is run and the binary restarted** (assets are embedded).

### Gotchas

- Money is integer kobo everywhere in backend and DB; convert only at display.
- `/ws` (live stock-taking) is handled before the middleware stack, so it bypasses auth/logging middleware.
- CSRF is enforced (gorilla/csrf in the `main.go` middleware stack); the SPA gets tokens from `GET /csrf-token`, and mutating requests need one. `/ws` is exempt.
- Config comes from the embedded `config/.env` (overridden at build time), not runtime environment.
- Suppliers are rows in `supplier`; services resolve a typed name with `resolveSupplier` (find-or-create, case-insensitive) inside the same transaction as the receipt or order. Screens still send names.
- Permission keys are in `internal/constant/permissions.go`; check them in both controller/middleware and the frontend when gating UI.
- Design tokens/guidelines for the UI are in `design-system/primocrest-pharmacy/MASTER.md`.
- `README.md` and `AGENTS.md` are partly stale (migration counts, Vue entry points); trust the code and `vite.config.js`.
- `backend/template/static/dist/` is not tracked in git, and the `Dockerfile` only copies `backend/`, so run `make assets` before `docker compose build server`.
