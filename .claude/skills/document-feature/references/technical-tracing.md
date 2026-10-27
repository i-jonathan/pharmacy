# Trace a feature for the technical doc

Load before writing or refreshing any `technical.md`.

Trace every user action from the guide end to end, in the order a request travels, then the way back. Run independent searches in parallel.

## For each action

| Layer | Where to look | Record |
|---|---|---|
| Screen | `frontend/src/next/views/`, `components/`, `composables/`, `store.js` | Component and function that starts the action; client-side validation; what is cached or kept in `localStorage` |
| Request | the `fetch` / `csrfFetch` call | Method, path, payload shape (field names and money units), idempotency key, version field |
| Route | `backend/main.go` mount + `adapter/http/router/<area>.go` | Full path after prefix stripping; middleware chain: `AuthMiddleware`, `AddPermissionsToContext`, `RequirePermissions(...)`, CSRF (main stack), or none |
| Controller | `adapter/http/controller/<area>.go` | Parsing, where the user id comes from (session context, never the body), error mapping |
| Service | `service/<area>.go` and helpers | Transaction boundaries, `HasPermission` checks, validations, idempotency handling, version checks, row locks, audit entries |
| Repository | `repository/<area>.go`, queries in `repository/query.go` | Queries run, tables read and written, locking (`FOR UPDATE`) |
| Data | `backend/db/migrations/` | Tables and columns touched, constraints and triggers that enforce rules, views (e.g. `inventory_view`) |
| Side effects | services and `adapter/websocket/` | `stock_movement` types written, held transactions created or cleared, WebSocket broadcasts, files generated (PDF/CSV) |
| Response | controller + view | Shape returned and how the screen uses it |

## Across the feature

- **Permissions**: every key involved, which roles get it by default (seeding migrations), where it is checked (route, service, UI). State plainly any action with no server-side check.
- **Data model**: tables owned by the feature and their relationships; which are append-only; money columns and their unit.
- **Invariants**: what must always hold (stock never double-counted, receipts never exceed order quantities) and what enforces each (constraint, lock, version, transaction).
- **Concurrency and retries**: what happens with two users at once, a double click, a retry after a lost response.
- **Configuration and constants**: hard-coded limits and windows, with their location.
- **Tests**: unit and integration tests covering the feature (`backend/service/*_test.go`, `internal/testdb`), and what isn't covered.
- **Recent history**: `git log --oneline -10 -- <feature paths>` and the related cards in `DONE.md`.
- **Known gaps**: anything the trace found missing or risky, each linked to a board card or offered as one.

Never write a path, function, table or permission from memory. Grep for it, and use the name the code uses.
