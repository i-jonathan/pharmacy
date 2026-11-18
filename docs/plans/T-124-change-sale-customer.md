# T-124 Change or remove the customer on a saved sale

Card: T-124 · Branch: feat/T-124-change-sale-customer

## Goal
A user who picked the wrong customer at the till can fix it afterwards from Sales History: change the customer on a saved sale, or remove them (back to walk-in). Anyone with `customer:manage` may do it for a sale made **today**; only managers (and admins) may change an older sale. It needs a reason, is audited, and Sales History and reprinted receipts show the new customer.

## What I found
- Verified on the card: "can't be changed once the sale is saved": no route or service method updates `sales.customer_id` after insert; only `CreateSaleTx` writes it — `backend/repository/sales.go:23`, `backend/adapter/http/router/sales.go:16-22` (code)
- Verified: CreateSale's checks to repeat: negative id → 400 "customer was not found"; non-zero id needs `customer:view`; unknown customer (FK violation or no row) → 400; deactivated customer → 409 `customer_inactive` — `backend/service/sales.go:31-37,98-116` (code)
- Verified: Sales History returns `customer {id,name}` only for callers with `customer:view`; the reprint dialog builds the receipt from that row, so a refreshed row is all a reprint needs — `service/sales.go:372`, `frontend/src/next/views/SalesHistory.vue:749-760` (code)
- `customer:manage` is seeded to Cashier, Manager and Admin (same as `customer:view`), so it cannot tell a cashier from a manager. The "older sales: managers only" rule therefore needs a permission only managers have; the pattern is `Stock/Adjust` in migration 44 — `internal/constant/permissions.go:19-23`, `db/migrations/000044_stock_adjustments.up.sql:66-84` (code)
- "Today" is decided in SQL as `created_at::date = CURRENT_DATE`, the same basis Sales History's date filter already uses (`created_at::date`), so the two never disagree on which sales are today — `repository/query.go:186-192` (code). No Go-side timezone handling exists in the backend (code).
- `sales` has no `version` column, so a stale-page guard must come from the customer the page saw (`expected_customer_id`), with the sale row locked `FOR UPDATE` — `backend/db/migrations/*sales_customer*` (code)
- Audit: `appendAudit` writes entity/action/reason/details in the caller's transaction; `AuditEntitySale` exists and the reprint already audits against a sale; `audit_log.reason` is a plain text column — `service/helpers.go:113-128`, `internal/constant/constant.go:54,73` (code)
- Customer audit entries record field names, not personal values (CLAUDE.md). I will follow that: the entry holds the receipt number and the old/new customer **ids**, not names or phones (code)
- Reason UI: `ReasonPicker.vue` (preset select + detail, "Other" needs text) with a reasons list per use — reusable with a new list (code)
- Returns exist on sales (`ReturnItems`); they do not touch the customer, so changing the customer is independent of them (code)
- Wrong or stale on the card: nothing found.
- Relevant patterns: stock adjustments (reason + audit + 409 on stale), customers (409 `code`, no-op edit writes nothing), `GetCustomerTx` share lock.

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| Change a saved sale's customer to another active customer | card | yes |
| Remove the customer (back to walk-in) | card | yes |
| Reason required, "Other" needs a note | card | yes |
| Audit entry (before/after ids, reason, actor) in the same transaction | card | yes |
| Refuse deactivated and unknown customers when attaching (same checks as CreateSale); removing a deactivated customer is allowed | card | yes |
| Stale-page guard (`expected_customer_id`) so two people fixing the same sale don't overwrite blind | `code` pattern (stock adjustments) | yes |
| Identical request (sale already has that customer) is a no-op success, writes nothing: also makes a double-click or retry harmless | `code` pattern (customers) | yes |
| Sales History shows who changed it / when | `model` (unverified) | follow-up |
| Customer search in Sales History search box | `model` | follow-up |
| Change the customer of a sale that has returns | — | yes, allowed; returns don't depend on the customer |

## Decisions (settled by you)
1. Anyone with `customer:manage` can change a customer; cashiers only on sales made today; only managers can change older ones.
2. Audit holds ids only (default I proposed, not objected to).
3. Open for your quick yes: the new permission's key. I propose resource `Sale`, action `ChangeOldCustomer` (key `sale:changeoldcustomer`), seeded to Manager and Admin only.

## Design
- **Migration `000050_sale_customer_old_sales`** (down removes the role links and the permission): inserts permission `Sale`/`ChangeOldCustomer`, granted to Manager and Admin only (same idempotent pattern as migration 44). Key constant `ChangeOldSaleCustomerPermissionKey = "sale:changeoldcustomer"`.
- **API:** `PUT /sales/{id}/customer`, body `{ "customer_id": 12 | 0, "expected_customer_id": 7 | 0, "reason": "wrong_customer", "note": "" }`. Route wrapped in `RequirePermissions(customer:manage)` in `router/sales.go`; the service re-checks with `HasPermission` (and `customer:view` for a non-zero target). Replies `{ "customer": {id,name} | null }`.
- **Reasons (server list, `stock_adjustment` style):** `picked_by_mistake`, `customer_asked`, `other` (needs a note, ≤250 chars). Stored in `audit_log.reason` as `reason` or `reason: note`.
- **Service** `SaleService.ChangeSaleCustomer(ctx, ChangeSaleCustomerRequest, userID)` in `service/sales.go`, inside `runInTx`:
  0. (before the transaction) Gate: needs `customer:manage`.
  1. Validate: id > 0, `customer_id` ≥ 0, `expected_customer_id` ≥ 0, reason known, note rule.
  2. `FetchSaleForUpdateTx` (lock the sale row; 404 if missing).
  2b. `is_today` comes back from the locked fetch (`created_at::date = CURRENT_DATE`). If the sale is not from today and the caller lacks `sale:changeoldcustomer`: 403 `code: "sale_too_old"` ("only a manager can change the customer on an earlier day's sale"). This runs before the stale check, so a cashier learns nothing about older sales' customers beyond what History already shows them.
  3. If current customer ≠ `expected_customer_id`, 409 `code: "customer_changed"` ("someone else changed this sale's customer; refresh and check").
  4. If current == target: return success, write nothing.
  5. If target > 0: `GetCustomerTx` → not found 400; deactivated 409 `customer_inactive` (same wording as CreateSale).
  6. `UpdateSaleCustomerTx`; `appendAudit(sale, id, "customer_changed", reason, {receipt_number, from_customer_id, to_customer_id})`.
  7. Return the new customer name (looked up in the same tx).
- **Repository:** `FetchSaleCustomerForUpdateTx(tx, saleID) (receipt string, customerID *int, isToday bool, error)` and `UpdateSaleCustomerTx(tx, saleID, *int)` with the existing FK mapping to `ErrInvalidReference`.
- **Constant:** `AuditActionSaleCustomerChanged = "customer_changed"`.
- **UI (Sales History detail panel):** next to the "Customer:" line (or a "Walk-in" line when none), a small "Change" button shown with `customer:manage`, and only for a sale made today unless the user has `sale:changeoldcustomer` (the server enforces it either way). Review change: the page does not use the browser's clock; `GET /sales/filter` returns `made_today` per sale, computed by the database with the same `created_at::date = CURRENT_DATE` test as the rule, so the button and the refusal cannot disagree near midnight or across time zones. It opens a dialog: reuse `CustomerPicker` for choosing (new prop `canCreate=false`, so no new-customer form here; reactivation stays as it is for managers), a "Remove customer" button, and `ReasonPicker` with the sale reasons. Save calls the endpoint; on success it updates the detail row in place (so a reprint shows the new name), on `customer_changed` it reloads the page's rows and says why. Uses `readError`; no `window.confirm`. Removing is confirmed by the dialog's own Save (no extra yes/no).
- **Docs:** update `docs/features/customers/{guide,technical}.md`, the CLAUDE.md customers bullet ("cannot be changed yet (T-124)" becomes how it works), and a What's-new entry only if you want one (not planned).

## Steps
1. Migration 50 + permission constant + audit constant + request types + repository methods (interface and impl) — test: repository covered by the integration test in step 3.
2. `ChangeSaleCustomer` service with unit tests on mocks: validation, cashier refused on an older sale, manager allowed, forbidden without manage, stale expected, no-op, inactive/unknown target, removal of a deactivated customer allowed, audit written with reason.
3. Integration test (`sale_customer_integration_test.go`; migration 50 runs for real): change, an older sale (backdated `created_at`) refused for a cashier and allowed for a manager, remove, stale 409, inactive refused, audit row, history shows new customer, same-request retry writes one audit row.
4. Controller + route (`PUT /sales/{id}/customer`, manage gate) + controller/route test for the 403.
5. Frontend: `CustomerPicker` `canCreate` prop, `SaleReasons` list, change dialog in `SalesHistory.vue`, refresh behaviour.
6. Docs (feature docs, CLAUDE.md bullet), and run `make tailwind` if new classes appear.

## Testing
- `go test ./service/ ./adapter/...`, then the integration suite with `PHARMACY_TEST_DSN`.
- Manual in the demo app: change, remove, deactivated target, stale page (two tabs), then reprint and check the receipt name; sign in as a user without `customer:manage` and confirm the button is hidden and the API answers 403.
- Not tested: a browser test (no JS test tooling in the repo).

## Deploy notes
Run `make migrate-up` (migration 50). Managers and admins must sign in again if permissions are cached in the session (to be checked in step 4; I'll say what I find). Run `make assets` and restart the binary (frontend is embedded). Update CLAUDE.md's migrations list with 50.

## Commits
Per your saved preference, I won't ask commit-by-commit: after implementation and review I'll show the planned commits (logical groups: backend, frontend, docs) and push/PR in one message for one yes.

## Follow-up cards (to add only if you approve)
- P2, Customers & loyalty: Show who changed a sale's customer and when in Sales History (from the audit log).
- P2, Customers & loyalty: Search Sales History by customer.

## Sources
- None from the web: internal feature, no regulation or library choice involved (skipped on purpose).
