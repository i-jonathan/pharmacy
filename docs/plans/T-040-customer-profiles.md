# T-040 Customer profiles

Card: T-040 · Branch: feat/T-040-customer-profiles

## Goal

The shop can keep a list of its customers (name, phone, email, notes, whether they agreed to be
contacted) and manage it on a Customers page. Nothing changes at the till yet: attaching a customer
to a sale is T-118, history is T-119, import and merge are T-120. T-041 (credit) and T-084 (loyalty)
build on this table.

The card was five features; on 2026-11-12 it was split (this plan covers the first slice only).

## What I found

- `code` No customer exists in the backend or database. The only trace is a free-text name in the
  POS cart (`usePos.js:30`, `CartPanel.vue:27-36`), printed on the receipt (`lib/receipt.js:109`) and
  held with a held sale (`usePos.js:164`); the sale itself does not store it. The card is correct.
- `code` Pattern to copy, newest complete slice: stock adjustments. Migration
  (`000044_stock_adjustments`), `types/stock_adjustment.go`, `repository/stock_adjustments.go`,
  `service/stock_adjustments.go`, `controller/stock_adjustments.go`, routes in
  `router/inventory.go:48-68`, view `StockAdjustments.vue`, router test
  `router/stock_adjustments_test.go`.
- `code` Permissions: a `permissions` row (`Resource`, `Action`) becomes the key
  `lower(resource:action)` at login (`controller/user.go:74`). Migration 43 seeds rows by role
  name. The key goes in `internal/constant/permissions.go`, the route is gated with
  `middleware.RequirePermissions`, the service re-checks with `HasPermission`
  (`service/helpers.go:15`), the UI with `hasPermission`/`PermissionGate`. The admin Permissions
  screen lists rows from the table, so no admin code is needed.
- `code` Audit: `s.audit(...)` (`service/helpers.go:91`, `InsertAuditLogTx`) in the same transaction
  as the change. `inTx` wraps a transaction (`helpers.go:68`). These are methods on
  `inventoryService`; a new customer service needs its own copies or a shared helper.
- `code` Each area has its own router mounted in `main.go:75-110` behind
  `AuthMiddleware(AddPermissionsToContext(...))`. The Vite dev proxy lists each prefix
  (`vite.config.js:26-38`), so a new `/customers` prefix must be added there or `npm run dev`
  cannot reach it.
- `code` Optimistic concurrency: purchase orders send the `version` they loaded and get 409 if it
  changed; I will use the same for customer edits.
- `code` A hook (`.claude/hooks/check-feature-docs.py`, `docs/features/feature-paths.json`) blocks a
  PR touching a documented feature without doc changes. A new feature needs its own guide,
  technical doc and a `feature-paths.json` entry (document-feature skill).
- `code` No phone library in `backend/go.mod`.
- `web` Nigerian numbers: the leading 0 is the trunk prefix and is dropped after +234, so
  `0803 123 4567` is `+234 803 123 4567`; storing E.164 (`+234…`) is the usual advice
  ([sent.dm](https://www.sent.dm/resources/NG), [mycountrymobile](https://www.mycountrymobile.com/?p=675015)).
  Sources disagree on how many digits follow +234; the owner fixed the rule at 10 (Decision 3).
- `web` NDPA 2023: lawful bases include consent and legitimate interest; consent must be informed,
  specific, freely given and unambiguous, the controller carries the burden of proving it, and
  people may withdraw consent and ask for access, correction and erasure
  ([KPMG](https://assets.kpmg.com/content/dam/kpmg/ng/pdf/the-nigeria-data-protection-act-2023.pdf),
  [Mondaq](https://www.mondaq.com/nigeria/data-protection/1366934/the-nigerias-data-protection-act-2023-a-look-at-key-provisions)).
  Sources disagree on section numbers, and I found nothing retail-specific. I am **not** giving
  legal advice; the design only makes consent *recordable* and customers *editable/deactivatable*.
  Whether the shop needs a privacy notice or registration with the NDPC is for the owner (and a
  lawyer) to check.
- `web` Other POS products search customers by phone, name or email, flag duplicates by phone, and
  merge keeping one record; Flourish makes merge a separately permissioned action
  ([Retail Pro](https://www.retailpro.com/News/blog/?p=1675010),
  [Flourish](https://docs.flourishsoftware.com/article/jes4ebsmxe-correct-a-customer-record)).
  That belongs to T-120.
- `memory` Commits and push are asked once, after implementation and review, not at plan approval
  (overrides the commit-style question in the skill).

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| `customer` table with name, phone, email, notes, consent | card | yes |
| Phone stored in one normalized form, unique | `web` + card (lookup by phone) | yes |
| Record *when* consent was given/withdrawn and by whom | `web` (burden of proof) | yes (audit entry + `consent_updated_at`) |
| Deactivate instead of delete (sales will reference customers) | `model` | yes |
| Search by name or phone, paged | card | yes |
| Permissions to view and manage, seeded per role | `code` | yes |
| Audit of create, edit, deactivate, reactivate | `code` (project convention) | yes |
| Edit conflicts (two staff editing) → 409 | `code` | yes |
| Docs (guide, technical, feature-paths) and tests | `code` | yes |
| Attach to a sale, receipt name | card | T-118 |
| Purchase history and totals | card | T-119 |
| CSV import, duplicate merge | card | T-120 |
| Hard delete / erase on request | `web` (NDPA erasure right) | follow-up (needs the sale link first) |
| Date of birth, address, allergies / medical notes | `model` | no (sensitive health data; not asked for) |
| "What's new" slide | `code` | no, wait for T-118 when it is visible at the till |

## Decisions for you

1. **Is a phone number required?** Recommend **optional**: name is required, phone and email are
   not, so a customer who will not give a number can still be saved. Unique only when present.
2. **One customer per phone, or allow shared phones?** Families often share a number. Recommend
   **unique phone** (a partial unique index, deactivated customers keep their number reserved):
   POS lookup by phone then returns one person and T-120 merge has a clear duplicate rule. A second
   person on the same number goes in without a phone, or notes. The alternative (allow duplicates,
   lookup shows a list) is more flexible but makes every later card pick-a-person.
3. **How strict is phone checking?** *Decided by the owner (2026-11-12):* a Nigerian number has
   10 digits after 234, or 11 if the trunk 0 is written (`0803 123 4567`, `803 123 4567`,
   `234 803 123 4567`, `+234 0803 123 4567` all become `+2348031234567`). Anything else Nigerian is
   refused. A number that starts with `+` and another country code is kept as typed digits (8-15
   digits, the E.164 limit). A small hand-written normalizer, no dependency.
4. **Who may do what?** Recommend two permissions: `customer:view` (Cashier, Manager, Admin) and
   `customer:manage` (create, edit, deactivate; also Cashier, Manager, Admin, because T-118 needs
   cashiers to add a customer at the till). Every change is audited. Alternative: managers only for
   edit and deactivate, with cashiers limited to create at the till (adds a third key).
5. **Sidebar placement.** Recommend a "Customers" link in the Sales section, shown with
   `customer:view`.

## Design

**Migration `000047_customers`** (down drops the table and the permission rows)
- `customer`: `id BIGSERIAL`, `name TEXT NOT NULL CHECK (btrim(name) <> '')`, `phone TEXT`
  (normalized, `+234…`), `email TEXT NOT NULL DEFAULT ''`, `notes TEXT NOT NULL DEFAULT ''`,
  `contact_consent BOOLEAN NOT NULL DEFAULT FALSE`, `consent_updated_at TIMESTAMPTZ`,
  `active BOOLEAN NOT NULL DEFAULT TRUE`, `version INTEGER NOT NULL DEFAULT 1`,
  `created_by_id`, `created_at`, `updated_at`.
- `UNIQUE INDEX customer_phone_unique ON customer(phone) WHERE phone IS NOT NULL`; an index on
  `lower(name)` for search.
- Permissions `Customer/View`, `Customer/Manage` and role grants by name, in the same
  insert-if-missing style as migration 43/44.

**Backend**
- `internal/constant/permissions.go`: `ViewCustomersPermissionKey = "customer:view"`,
  `ManageCustomersPermissionKey = "customer:manage"`.
- `internal/types/customer.go`: `Customer`, `CustomerRequest` (id, version, name, phone, email,
  notes, contact_consent), `CustomerPage`.
- `model/phone.go` (or `service/phone.go`): `NormalizePhone(string) (string, error)` with table
  tests; empty input means "no phone".
- Repository (`repository/customers.go`, methods added to `PharmacyRepository`): insert, update by
  id+version, get, list/search with paging, set active.
- Service (`service/customers.go`, `CustomerService` in `interfaces.go`): validate (name 1–120
  chars, email shape, notes ≤ 1000, trimmed), normalize phone, `HasPermission` checks, create/edit
  in a transaction with an audit entry (`customer` / `created`, `updated`, `deactivated`,
  `reactivated`; details list the changed fields, and consent before→after; the audit log is
  append-only). A phone already taken returns 409 naming the existing customer (and whether
  deactivated). `consent_updated_at` changes only when consent changes. A transaction/audit helper
  is shared rather than copied (small refactor of `inTx`/`audit` off `inventoryService`).
- Controller `adapter/http/controller/customers.go` and router `router/customers.go`, mounted at
  `/customers/` in `main.go`:
  `GET /customers?q=&page=&inactive=` (view), `GET /customers/{id}` (view),
  `POST /customers` (manage), `PUT /customers/{id}` (manage), `POST /customers/{id}/deactivate`
  and `/reactivate` (manage). Bodies capped with `decodeLimitedJSON`.

**Frontend**
- Route `/customers` in `router.js`, `views/Customers.vue`: search box (name or phone, debounced),
  paged table (name, phone, email, consent, status), "Show deactivated" toggle, a side form/dialog
  to add and edit (consent checkbox with a plain label such as "Agreed to be contacted by the
  pharmacy"), deactivate/reactivate through `confirm()` from `useConfirm.js`, errors via
  `readError`, a retry-safe create (the unique phone already makes a double submit a 409 that the
  page explains). View-only users see no add/edit controls (`hasPermission`).
- `Sidebar.vue`: link in Sales, gated on `customer:view`. `vite.config.js`: proxy `/customers`.
- New Tailwind classes need `make tailwind` and the regenerated `tailwind.css` committed.
- Demo seed (`internal/demoseed`): grant the new keys to the demo roles and add a few customers
  so screenshots later have data.

**Not touched:** sales, POS, receipt, held sales (T-118).

## Steps

1. Migration 47 + permission constants; integration test that migrations run and the roles get the
   permissions, and the unique-phone index behaves. (`PHARMACY_TEST_DSN`)
2. `NormalizePhone` with table tests (formats, `+234 0…`, foreign, too short/long, empty).
3. Types, repository, service with mocks: validation, duplicate phone, version conflict, audit
   entries written, consent timestamp rule, permission refusals, deactivate/reactivate.
4. Controller, router, `main.go` wiring; router test that each route needs its permission (like
   `stock_adjustments_test.go`).
5. Customers page, route, sidebar, Vite proxy, demo seed; `make tailwind`/`make assets`.
6. Manual check in the running app with a cashier, a manager and a user with neither permission.
7. Docs: `docs/features/customers/{guide,technical}.md`, README table, `feature-paths.json`,
   `CLAUDE.md` (migration list, permissions gotcha), `AGENTS.md` permission keys.

## Testing

- Unit: `go test ./service/ ./adapter/... ./model/...` for the items above.
- Integration: `PHARMACY_TEST_DSN=… go test ./...` (migration, unique index, audit in one
  transaction, concurrent create of the same phone gives one row and one 409).
- Manual: add, edit, search by `0803…` and `+234803…` and name, shared-phone refusal message,
  deactivate then reactivate, version conflict from two tabs, read-only user.
- Not tested: browser automation (the repo has none); the Vue page is checked by hand.

## Deploy notes

Run migration 47 (additive). Existing Cashier, Manager and Admin roles get both permissions by
name (same assumption as T-099 for migration 43: roles must be named that way); people must sign
in again to receive the new permissions in their session. `make assets` before building the image.
No data is created automatically.

## Commits

Per memory: one yes after implementation and review. Proposed logical groups: (1) board + plan,
(2) migration, constants, phone normalizer, repository, service, tests, (3) controller, router,
wiring, (4) Customers page, sidebar, proxy, seed, tailwind.css, (5) docs.

## Changes after review (2026-11-12)

`/code-review high` found things that changed the build from the design above:
- Phone search anchors a number written from its start (`0803`, `+234 803`) to the start of the
  stored number instead of matching anywhere; digits from the middle still match anywhere.
- The consent time is stamped by the database in both create and update (the app clock is not used).
- On a stale-edit conflict the page keeps what the person typed and takes the new version, instead
  of replacing the form; Deactivate/Reactivate refresh the version the same way.
- The list moves to the last page when the current page empties; rows are locked while a save runs.
- No retry key on create (needs a migration the plan didn't cover): card T-122.

## Follow-up cards

Created in this PR: T-118 attach to a sale, T-119 history and totals, T-120 import and merge,
T-121 erase or anonymize a customer on request (P2, NDPA erasure/access request process once
sales reference customers), T-122 retry key for adding a customer.

## Sources

- [KPMG: NDPA 2023](https://assets.kpmg.com/content/dam/kpmg/ng/pdf/the-nigeria-data-protection-act-2023.pdf): lawful bases, consent, rights
- [Mondaq: NDPA key provisions](https://www.mondaq.com/nigeria/data-protection/1366934/the-nigerias-data-protection-act-2023-a-look-at-key-provisions): consent and controller's burden of proof
- [sent.dm: Nigeria numbers](https://www.sent.dm/resources/NG): drop trunk 0 after +234, E.164
- [mycountrymobile: Nigeria format](https://www.mycountrymobile.com/?p=675015): same, digit counts disagree
- [Retail Pro](https://www.retailpro.com/News/blog/?p=1675010): duplicate check by phone in a pharmacy POS
- [Flourish](https://docs.flourishsoftware.com/article/jes4ebsmxe-correct-a-customer-record): edit vs merge vs delete, merge permissioned
