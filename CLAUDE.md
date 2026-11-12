# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

`AGENTS.md` in the repo root has the quick-start and a fuller facts list; read it too. This file covers what that one doesn't.

Planned and in-flight work is tracked on the Waypoint board (project `Primocrest`); see Task board below.

## Commands

```sh
docker compose up -d data                      # local Postgres
make migrate-up DB_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
make assets                                    # tailwind (Go templates) + vite build (Vue)
make run                                       # cd backend && go run .  → :8000
make dev                                       # assets + run
cd frontend && npm run dev                     # Vite dev server, proxies /api to :8000
```

- Migrations: `make create-migration MIGRATION_NAME=foo`, `make rollback-migration DB_URL=...`, `make clean-dirty-migration VERSION=N DB_URL=...`. Latest migrations: `000040_purchase_orders_and_suppliers` (`supplier`, `purchase_order`, `purchase_order_item`, `receiving_batch.supplier_id`, `product_batch.purchase_order_item_id`), `000041_audit_log` (append-only `audit_log`, written in the same transaction as the change it describes; purchase orders and user creation use it so far) `000042_purchase_order_approval` (draft / pending_approval statuses, approved_by_id/approved_at) and `000043_purchase_order_permissions` (seeds `PurchaseOrder`/`Draft` to Cashier and `PurchaseOrder`/`Approve` to Manager and Admin) and `000044_stock_adjustments` (`stock_adjustment`, `stock_adjustment_item`, `stock_flag`; seeds `stock:adjust` to Manager/Admin and `stock:flag` to Cashier/Manager/Admin) and `000045_stock_movement_product_index` (the first index on `stock_movement`, on `product_id`, which every per-product stock sum uses) and `000046_stock_adjustment_request_hash` (`stock_adjustment.request_hash`; the retry key becomes unique across users) and `000047_customers` (`customer`, with `customer:view` and `customer:manage` seeded to Cashier/Manager/Admin).
- Never edit a migration once it has been pushed or applied anywhere: `migrate up` won't re-run it, so the change silently never reaches that database. Add a new migration instead.
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
- `RequirePermissions` refuses an API call with a JSON 403 message (plain text for page loads), so screens can show why.
- `AuthMiddleware` answers a signed-out API call (anything but a GET that accepts `text/html`) with 401 JSON; only page loads are redirected to sign in and remembered as `next`. A redirect would be followed by `fetch` to a 200 login page and read as success.
- CSRF is enforced (gorilla/csrf in the `main.go` middleware stack); the SPA gets tokens from `GET /csrf-token`, and mutating requests need one. `/ws` is exempt.
- Config comes from the embedded `config/.env` (overridden at build time), not runtime environment.
- Suppliers are rows in `supplier`; services resolve a typed name with `resolveSupplier` (find-or-create, case-insensitive) inside the same transaction as the receipt or order. Screens still send names.
- Purchase orders have an approval flow: a user with `purchaseorder:draft` writes drafts, and drafts are shared: any drafter sees every order, edits any draft, submits any draft, and can pull a submitted order back to a draft (`WithdrawPurchaseOrder`, audited `withdrawn`), so two cashiers do not write the same order. A submitted order is edited only by a manager (or after pulling it back). `purchaseorder:approve` (managers) creates orders directly, edits any draft, approves/rejects, and closes/moves approved orders. Services enforce it with `HasPermission`; the routes are gated in `router/inventory.go`. Only a user with neither permission is limited to orders they started (`canWorkOnPurchaseOrders`). Drafts and submitted orders cannot be received against, and their PDF is stamped DRAFT (rejected drafts have none).
- There is no public sign-up. Users are added by admins (`POST /admin/api/users`, `AdminService.CreateUser`) or by the `create-admin` subcommand of the binary (`create_admin.go`), which runs before the server starts. Once an Admin exists it refuses; the hidden `--recover`/`-r` flag (deliberately left out of `-h` and the refusal message) adds one anyway for lock-out recovery and marks the audit entry `"recovery": true`. Every password goes through `model.User.ValidatePassword` (10+ characters ignoring surrounding spaces, not in the embedded `model/common_passwords.txt`, not containing the username as a word, see `containsAsWord`). An audit entry with no signed-in user must set `AuditEntry.NoActor`; ActorID 0 is otherwise refused by the database.
- Stock changes outside sales, receiving and stock taking go through stock adjustments (`service/stock_adjustments.go`): products are locked and their stock read inside the transaction, every change has a reason, an audit entry and `IN/OUT_MANUAL_ADJUSTMENT` movements referencing the adjustment. Product edit no longer takes a stock value. "Never below zero" applies only to changes that lower stock (sales can already have taken it below). A "set" line sends `expected_stock` and is refused (409) if stock moved since; the page sends an expiry-only fix as "add 0". A person can have one open flag per product (partial unique index `stock_flag_one_open_per_user`). A retry key is used once (unique index); the adjustment stores a fingerprint of the request (`request_hash`), and a resend under a used key is answered with the first save only when it is the same request from the same user, otherwise 409 with `code: "already_saved"`, on which the page ends the form and opens History. The page makes one key per form. Adjustments refuse products already counted in an in-progress stock taking, whose completion posts (count − snapshot) and would apply the change twice (T-111 covers sales and receipts). Older `IN/OUT_MANUAL_ADJUSTMENT` movements written by product edit have `reference_id` = the product id, not an adjustment id; only movements dated after migration 44 can be joined to `stock_adjustment`. `/inventory/search` results carry `stock: 0` (not computed); load `/inventory/product/{id}` when the real stock matters.
- Customers (`customer` table, `service/customers.go`, Customers page): a phone is stored normalized by `model.NormalizePhone` (`+234` plus 10 digits; the typed form may include the trunk 0, 11 digits) and is unique across all customers, deactivated ones included. Customers are deactivated, never deleted. Edits send the `version` they loaded and get 409 if it moved; a no-change edit writes nothing. Consent changes are audited and stamp `consent_updated_at`; the audit entry lists changed field names, not values. A phone clash is a 409 with `code: "phone_taken"`. Nothing reads customers yet: the till still sends a typed name (T-118). Search text with a letter matches names, otherwise phone digits. `runInTx` and `appendAudit` in `service/helpers.go` are the shared transaction and audit helpers.
- Permission keys are in `internal/constant/permissions.go`; check them in both controller/middleware and the frontend when gating UI.
- SPA error replies are JSON `{error, code?}`; read them with `readError(response, fallback)` from `lib/api.js` (returns `{message, code}`), and make retry keys with `newRetryKey()` there.
- Confirmations use one dialog: `await confirm({ title, message, detail, tone, run, alternate })` from `composables/useConfirm.js`, rendered once by `ConfirmDialog.vue` in `App.vue`. Do not use `window.confirm` or write a new modal for a yes/no question. `run` keeps the dialog open with a busy state and shows its error inline.
- "What's new" is a sidebar slideshow of the newest release, with pictures that are real screens taken on demo data. To announce a release, add an entry on top of `frontend/src/next/lib/whatsNew.js` and take new pictures; how, including the demo database and the `PHARMACY_PORT` override for running a second server, is in `docs/whats-new/README.md`.
- Design tokens/guidelines for the UI are in `design-system/primocrest-pharmacy/MASTER.md`.
- `README.md` and `AGENTS.md` are partly stale (migration counts, Vue entry points); trust the code and `vite.config.js`.
- The SPA's utility classes come from the tracked `backend/template/static/css/tailwind.css`, built by `make tailwind` from the Vue sources. A new Tailwind class in a `.vue` file shows no styling until `make tailwind` (or `make assets`) is run and the result committed; `npm run build` alone is not enough.
- `backend/template/static/dist/` is not tracked in git, and the `Dockerfile` only copies `backend/`, so run `make assets` before `docker compose build server`.

## Task board

The board is a Waypoint project named `Primocrest` (a SQLite database in `~/.local/share/waypoint/`, outside the repo). The repo holds no board file, so run every command with `--project Primocrest` (from here, plain `waypoint ls` finds nothing). Use the CLI; never edit the `.db` file directly. Card changes are made on the board straight away, not in a pull request.

- See what to work on next: `waypoint ls --next --json --project Primocrest`
- Read cards: `waypoint ls --json --project Primocrest` (add `--column Done` or `--archived` for finished ones), `waypoint show <ID> --json --project Primocrest`
- Add a card: `waypoint add "Title" --group "Group" --priority P1 --depends T-1,T-2 --body "..." --project Primocrest` (the id is assigned for you)
- Move a card: `waypoint move <ID> "<Column>" --project Primocrest`
- Change a field: `waypoint set <ID> key=value --project Primocrest` (title, group, priority, depends, body, or an extra field such as `prs=#39`)
- Tick a checklist item: `waypoint check <ID> <N> --project Primocrest`
- Finish a card: `waypoint done <ID> --project Primocrest`, with `prs=#N` set when its pull request was opened

Groups, priorities and columns must already exist on the board: `waypoint ls --json --project Primocrest` shows the ones in use. Columns are Up Next, In Progress, Backlog and Done. Priorities: P0 blocks taking on a second customer, P1 needed to compete, P2 nice to have.

If your agent speaks MCP and has no shell, run `waypoint mcp --project Primocrest` as an MCP server instead; it offers the same board through its tools.
