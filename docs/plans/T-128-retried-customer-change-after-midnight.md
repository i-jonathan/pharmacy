# T-128 Answer a retried customer change as saved after midnight

Card: T-128 · Branch: fix/T-128-retried-customer-change-after-midnight

## Goal
Two changes to the rule for correcting the customer on a saved sale:

1. A retry of a change that already saved is answered as saved, never refused with "only a manager can change an earlier day's sale", whatever the clock says.
2. "Old" means more than 24 hours old, not "made on an earlier calendar day". A sale saved at 23:59 can still be corrected by a cashier at 00:01.

## What I found
- Verified on the card: `ChangeSaleCustomer` applies the day rule (`sale_too_old`) before the unchanged check, so `unchanged` is only computed after the 403 — `backend/service/sales.go:470-482` (`code`).
- An unchanged request writes and audits nothing (`sales.go:510-512`), and its answer carries only the customer's id and name, which Sales History already shows to anyone with `customer:view`; `ChangeSaleCustomer` requires that permission (`sales.go:436`). Answering it for a cashier on an old sale leaks nothing (`code`).
- The 24-hour rule alone does not fix the card: the cutoff only moves. A change at 09:59 on a sale made at 10:00 the day before, retried at 10:01, still gets `sale_too_old`. So both changes are needed (`model`, follows from the code above).
- Today's rule is SQL `created_at::date = CURRENT_DATE AS is_today` in two queries: `fetchSaleCustomerForUpdateQuery` (`backend/repository/query.go:167`) and `fetchSalesQuery` (`query.go:192`, feeds `SaleResponse.made_today` and the Change button). `sales.created_at` is `timestamptz` (migration 36), so `created_at >= now() - interval '24 hours'` is exact (`code`).
- Sales History's own date filter (`created_at::date >= $1`) is a separate thing and stays as it is (`code`, `query.go:193-194`).
- The old plan (T-124) chose `CURRENT_DATE` so the rule matched the history date filter; that reason no longer applies once the rule is a rolling window, and the button and the server still agree because both use the same SQL test (`code`, `docs/plans/T-124-change-sale-customer.md:13`).
- Existing test `TestOnlyAManagerChangesTheCustomerOfAnEarlierDaysSale` (`service/sale_customer_change_test.go:297`) sends a stale-and-changing request on an old sale and expects `sale_too_old`; it stays valid.
- Places that say "today" or "earlier day" and need rewording: `internal/types/sales.go:71,75,128`, `model/sales.go:19-20`, `internal/constant/permissions.go:26`, `adapter/http/router/sales.go:23`, `service/sales.go:386,470-482`, `frontend/src/next/views/SalesHistory.vue:833`, `CLAUDE.md` (Customers bullet), `docs/features/customers/technical.md:79,82,91,151`, `docs/features/customers/guide.md:84,119,140,155` (`code`).

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| Unchanged check before the old-sale rule | `code` sales.go:470-482 | yes |
| Old sale = more than 24 hours old, in both SQL queries | `code` query.go:167,192 | yes |
| Refusal message, comments, docs say "earlier day" | `code` list above | yes |
| Rename the JSON field `made_today` | `model` | no: keep the wire name, comment that it means "within 24 hours"; renaming touches the frontend for no gain |

## Decisions for you
- Go field name: rename `IsToday` to `IsRecent` (recommended, since "today" would be false) and keep the JSON key `made_today`. Say if you want the JSON key renamed too.

## Design
- **Repository:** both queries use `created_at >= now() - interval '24 hours' AS is_recent`; `db:"is_recent"` on `types.SaleCustomerState` and `model.Sale`. No migration.
- **Service:** in `ChangeSaleCustomer`, compute `unchanged` first. Apply the old-sale rule only when `!unchanged` (permission `sale:changeoldcustomer` unchanged); then the stale check. Refusal text: "only a manager can change the customer on a sale more than 24 hours old". Code stays `sale_too_old`.
- **Response:** `SaleResponse.made_today` is now true for a sale within the last 24 hours. Name kept; comment updated.
- **UI:** no behaviour change; comment in `SalesHistory.vue` reworded. Not rebuilding assets.
- **Docs:** `CLAUDE.md` and the two customer feature docs reworded to the 24-hour rule and the retry behaviour.
- Permissions, audit entries, idempotency: unchanged.

## Steps
1. Tests first in `backend/service/sale_customer_change_test.go` (rename `IsToday` to `IsRecent` there too): a cashier repeats a change on an old sale already on that customer and gets it back with no write or audit; the same for a repeated removal on an old walk-in sale; a cashier making a real change on an old sale still gets `sale_too_old`, including when also stale. Confirm the first two fail.
2. Reorder the checks in `service/sales.go`; reword the message and comments. Tests go green.
3. Switch both SQL queries to the 24-hour test; rename the field in `types`, `model`, `service` and the comments listed above.
4. Update `CLAUDE.md`, `docs/features/customers/technical.md` and `guide.md`.

## Testing
- Unit: `cd backend && go test ./service/ -run 'ChangeSaleCustomer|Customer' -v`, then `go test ./...`.
- Integration (`PHARMACY_TEST_DSN`): run `go test ./...` with it set so the sale-customer integration tests exercise the new SQL. If none covers `is_recent`, add one: a sale backdated to 23 hours ago is recent, one backdated 25 hours ago is not.
- Manual: not planned; the rule is SQL plus service logic.

## Deploy notes
None: no migration, no new permission. Restart the server for the new binary. No frontend asset rebuild (comment-only change).

## Commits
One commit for the whole ticket (plan file, tests, fix, docs), after review. The owner confirms the commit and the push separately.

## Follow-up cards
None.

## Sources
None needed (internal logic).
