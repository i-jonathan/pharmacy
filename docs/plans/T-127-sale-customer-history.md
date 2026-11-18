# T-127 Show who changed a sale's customer, and when, in Sales History

Card: T-127 · Branch: feat/T-127-sale-customer-history

## Goal
When a manager opens a sale in Sales History they can see every time its customer was changed or removed: who did it, when, from whom to whom, and why. Today this is only readable from `audit_log` in the database, so a disputed correction cannot be checked from the app.

## What I found
- Verified on the card: `ChangeSaleCustomer` writes an audit entry (entity `sale`, action `customer_changed`, reason, actor, time, `from_customer_id`, `to_customer_id`, `receipt_number`) in the same transaction as the change. `code` `backend/service/sales.go:521-532`. A repeat of an already-made change writes nothing (`sales.go:497`), so there are no duplicate entries.
- Verified on the card: no screen reads it. The only reader of `audit_log` is `FetchAuditLog`, used for purchase orders (`backend/repository/purchase_orders.go:297`); it is per-entity and not filtered by action. `code`
- The `reason` column holds `reason` or `reason: note`, with `reason` one of `picked_by_mistake`, `customer_asked`, `other` (`sales.go:512-515`, `saleCustomerChangeReasons`). `code`
- Sale audit rows are not only customer changes: `receipt_reprinted` is also written under entity `sale` (`constant.go:73`, `sales.go:626`). So the query must filter on action. `code`
- `audit_log` is indexed on `(entity_type, entity_id, occurred_at)` (`000041_audit_log.up.sql:14`), so reading the entries for one page of sales is cheap. `code`
- Sales History already builds each page in bulk and resolves customer names only for users with `customer:view` (`customerNamesFor`, `sales.go:571`, `FetchCustomerNames`, `repository/customers.go:99`). `FetchSalesHistory` runs in one transaction. `code`
- The detail panel is in `frontend/src/next/views/SalesHistory.vue:173-185` (customer line and Change button); the response type is `SaleResponse` in `backend/internal/types/sales.go`. `code`
- Customers are deactivated, never deleted, so every id in an old audit entry still resolves to a name. `code` (CLAUDE.md, `customers` docs)
- Nothing on the card is wrong or stale. No web research: this is an internal read-only screen with no new rules or libraries.

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| Change list in the sale details: when, by whom, from → to, reason and note | card | yes |
| Hidden from users without `customer:view` (names and ids are personal data, and the sale's own customer is already hidden from them) | `code` `sales.go:571`, CLAUDE.md | yes |
| Walk-in shown as "Walk-in" when from/to is 0 | `code` `sales.go:508-510` | yes |
| No extra query per sale (one query per page) | `code` pattern of `FetchSalesHistory` | yes |
| Older entries: all `customer_changed` rows come from T-124, so every row has the details fields | `code` | yes (parse defensively anyway) |
| Show receipt reprints in the same list | `code` `receipt_reprinted` exists | follow-up |
| Show the changes in the CSV export of Sales History | `code` `SalesHistory.vue:936` exports the customer name | follow-up (only if asked) |

## Decisions for you
1. **Embed in the history response vs load when a sale is opened.** Embed (recommended): one extra query per page, no loading state, no new route, no new permission rule to keep in step. The cost is a little more JSON per page for sales that were changed (rare). Loading on open needs a new route and a spinner for no real saving.
2. **Names shown are today's names.** An audit entry stores ids only (by design, see the customers docs), so a customer renamed since the change shows the new name. Recommended: accept it; storing names in the audit entry would put personal data in an append-only table.
3. **What's new slideshow:** no entry (small P2 addition). Say if you want one.

## Design
- **No migration, no new permission.** Read-only.
- `types.SaleCustomerChange` (JSON): `at` (RFC3339), `by` (username, may be empty), `from` and `to` (`*SaleCustomer`, null = walk-in), `reason` (code), `note`.
- `SaleResponse.CustomerChanges []SaleCustomerChange` (`customer_changes`, always an array, empty when none or when the caller lacks `customer:view`).
- Repository: `FetchSaleCustomerChanges(ctx, tx, saleIDs []int)` returns raw rows `{sale_id, occurred_at, actor, reason, details}` for `entity_type='sale' AND action='customer_changed' AND entity_id = ANY($1)`, ordered newest first (`occurred_at DESC, id DESC`), joined to `users` for the name. Added to the `PharmacyRepository` interface.
- Service (`FetchSalesHistory`): when the caller has `customer:view` and there are sales, fetch the rows, decode `details`, collect from/to ids, resolve names with `FetchCustomerNames` (merge with the sale customers so it is one lookup), split the stored reason into code and note (only when the part before `: ` is a known reason code; otherwise the whole text is the note), attach to each sale. A row whose details cannot be decoded is skipped and logged, not a failure of the page.
- Frontend: a "Customer changes" block in the detail panel under the customer line, shown only when `customer_changes.length > 0`. Each row: `Alice → Walk-in`, by `username`, date/time (`formatDate`), reason label (reuse `CUSTOMER_CHANGE_REASONS` labels), note if any. After a successful change in the same session, prepend the new entry locally is not possible (the response has no actor/time), so re-fetch the page that is open and keep the detail open (check how the dialog already refreshes the row; if it only patches the row, add the entry from the current user's name and the client time). Decide at step 4, noted in the PR.
- Tailwind: only classes already used in this view, to avoid a `make tailwind` rebuild; if a new one is needed, run `make tailwind`.

## Steps
1. Types + repository query + interface; integration test in `backend/service/sale_customer_*integration_test.go` style (real SQL): two changes on one sale, one on another, a `receipt_reprinted` entry present, ordering newest first, actor name, sale with no changes returns nothing. (`PHARMACY_TEST_DSN`)
2. Service: attach `customer_changes` in `FetchSalesHistory`; unit tests with the hand-written mock: user without `customer:view` gets none and the repository is not called; walk-in from/to become null; reason with note is split; unknown reason prefix stays whole as note; undecodable details skipped; names resolved in one lookup.
3. Frontend: render the block in the detail panel; update the local state after a change.
4. Docs: Customers `guide.md` (what the manager sees, who sees it) and `technical.md`; CLAUDE.md Customers paragraph (one sentence: history shown via `customer_changes`, `customer:view` only).
5. Review checklist, manual check in the running app (change a customer twice, view as a cashier without `customer:view`).

## Testing
- Unit (service mocks) and integration (SQL) as above; `go test ./...` with `PHARMACY_TEST_DSN`; `make check-go-format`, `make check-go-lines`.
- Manual: demo database, change a customer on a sale as a manager, open the sale details; check a user with no `customer:view` sees nothing and no ids/names are in the JSON.
- Not tested: Vue (no JS test tooling in the repo); checked by hand.

## Deploy notes
No migration. `make assets` and a binary restart (assets are embedded). No users need to log in again.

## Commits
Style chosen by the owner: in logical groups. Planned commit points: (1) types, repository query, service and their tests (steps 1-2); (2) frontend (step 3); (3) docs and plan (step 4). Each is shown to the owner for a yes before it is made.

## Follow-up cards
- P2, Customers & loyalty: Show receipt reprints (who, when) in Sales History sale details. Proposed, not yet added.

## Sources
- None outside the repo (internal screen).
