# T-123 Allow receiving more than ordered against a purchase order

Card: T-123 · Branch: fix/T-123-receive-over-ordered

## Goal
When a supplier delivers more than was ordered (bonus stock, pack rounding), the cashier or manager can receive the whole delivery against the order instead of splitting it into an order receipt and a separate "no order" receipt. There is no upper limit; the screen asks for confirmation when a line is over.

## What I found
- Verified on the card: the screen blocks it — `ReceiveItems.vue:785` (per row; red qty + "only N still outstanding on PO-…").
- Verified: the service blocks it — `purchase_orders.go:768` (per order line, rows of the same line are summed first by `groupPurchaseOrderReceipts`, so a split delivery is checked as a whole).
- Verified: the SQL guard `AND received_quantity + $1 <= quantity` — `repository/query.go:620`. `ErrPurchaseOrderItemNotReceivable` (`repository/purchase_orders.go:19`) is also returned for wrong order / wrong product, so it stays.
- Verified: the table check `CHECK (received_quantity <= quantity)` — `000040_purchase_orders_and_suppliers.up.sql:64`. Migration 40 has shipped, so a new migration (48) drops it.
- Verified: order status already treats a line at or over its ordered quantity as complete (`received_quantity < quantity` in `refreshPurchaseOrderStatusQuery`, `query.go:629`). No change.
- Verified: `outstandingPurchaseOrderItems` (`purchase_orders.go:580`) only counts `remaining > 0`, so Move outstanding and Close ignore over-received lines. No change.
- Verified: editing a line cannot go below what was received (`purchase_orders.go:202`, `PurchaseOrders.vue:1379`). Stays as is.
- `code` Risk: `received_quantity` is a 32-bit integer and the order quantity is capped at `MaxInt32` (`purchase_orders.go:25`). With the cap gone, `received + receipt` could overflow and surface as a 500. Needs a cap that is a 400.
- `code` Places that compute "outstanding" as `quantity - received_quantity` and would go negative, mis-sum, or hide a button:
  - `PurchaseOrders.vue:864` `totals()` (the "Move outstanding" button is `v-if="totals(o).outstanding"`, so an over-received line could cancel out another line's real shortfall and hide it) and `:866` `progress()` could exceed 100%
  - `PurchaseOrders.vue:554-556` detail table shows a negative number instead of "Done" when received > ordered
  - `PurchaseOrders.vue:1329` units-already-ordered per product (low-stock helper), would go negative
  - `PurchaseOrders.vue:1534` PDF "Outstanding" column
  - `ReceiveItems.vue:983` the "still has N units outstanding" toast
- `code` Audit entry (`purchase_orders.go:786-792`) already logs `quantity` and `outstanding_after`, which goes negative for an over-receipt. Test at `purchase_orders_test.go:921` asserts `outstanding_after:3`.
- `code` Pattern to copy for the confirmation: `confirmLargeOrder` in `PurchaseOrders.vue:1385` (`confirm({ title, message, lines, confirmLabel, cancelLabel })`).
- `code` Tests that change: `purchase_orders_test.go:380` ("more than outstanding is rejected, even when split over rows"). A DB-level check may also exist in `purchase_orders_integration_test.go` (will check while writing the tests).
- `memory`: commits and push are confirmed once, after implementation and review — not asked at plan approval.
- Wrong or stale on the card: nothing. The card was written from this research.
- Web research skipped: purely internal rule change; the behaviour was decided by the owner.

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| Overflow guard so `received + receipt` can't exceed 2,147,483,647 (400, not 500) | `code` | yes |
| Fix every "outstanding" calculation so an over-received line counts as 0 outstanding | `code` | yes |
| Show the extra on the order detail (e.g. "+20 extra") instead of "Done" | `code` | yes (small, same cell) |
| Confirm per order line (rows for the same line summed), not per row | `code` | yes |
| Audit entry records `over_received_by` per line | `code` | yes |
| Receiving extra against a line that is already fully received, or an order already `received`, is still not possible (the order isn't receivable once complete) | `code` | no; the extra from the same delivery is handled, later extras are "no order" receipts. Mention in the guide |
| Server-side "I confirmed" flag so API clients must acknowledge over-receipt | `model` (unverified) | no: the owner asked for a screen confirmation only |

## Decisions for you
None open. Settled: no upper limit; confirm on the screen.

## Design
**Migration `000048_purchase_order_over_receipt`**
- up: drop the check `received_quantity <= quantity` on `purchase_order_item`. The constraint is unnamed, so it is found by its definition in `pg_constraint` inside a `DO` block (the auto-generated name differs by Postgres version). `received_quantity >= 0` stays.
- down: re-add the check as a named constraint. It fails if any line is over-received, which is the honest outcome; the down note says to correct those lines first. (Silently clamping would rewrite what was received.)

**Repository** (`query.go`): `receivePurchaseOrderItemQuery` guard becomes `received_quantity::bigint + $1 <= 2147483647`, so overflow is a clean "not receivable" rather than a numeric error. Wrong order/product conditions stay.

**Service** (`purchase_orders.go`): the "only N still outstanding" check is replaced by "the total received for this line cannot exceed 2,147,483,647" (400). Audit details gain `over_received_by` (0 when within the order), computed as `max(0, received_after - quantity)`; `outstanding_after` becomes `max(0, …)`.

**Screen — receiving** (`ReceiveItems.vue`)
- Remove the per-row block in `validate()`.
- In `receiveItems()` after `validate()`: sum quantity per `purchase_order_item_id`, compare to that line's outstanding (`_poOutstanding`); if any line is over, `await confirm({ title: "More than was ordered", message: "…", lines: ["Paracetamol 500mg: 120 arriving, 100 outstanding (20 more)"], confirmLabel: "Receive anyway", cancelLabel: "Go back" })` and stop on cancel. Same look as `confirmLargeOrder`.
- Toast: sum `max(0, quantity - received)`.
- Row hint "N outstanding" stays; text for a row going over is neutral (no red).

**Screen — orders** (`PurchaseOrders.vue`): `totals()` sums per line `max(0, quantity - received)` for outstanding, `progress()` capped at 100; detail cell shows "Done" at exactly ordered, "+N extra" (green) over; units-already-ordered and the PDF Outstanding column use `max(0, …)`.

No API shape changes; no permission changes (`receive` is already gated).

## Steps
1. Migration 48 (+ down). Test: DB integration (`PHARMACY_TEST_DSN`) receives over the ordered quantity and the status becomes `received`; migration runs up and down on an empty database.
2. Repository guard + service check/audit. Test (TDD, `purchase_orders_test.go`): replace the "rejected" case with "over-receipt is accepted, split over rows too" and "past the 32-bit limit is a 400"; audit test for `over_received_by` and `outstanding_after` floored at 0.
3. `ReceiveItems.vue`: remove the block, add the grouped confirmation, fix the toast. Manual check.
4. `PurchaseOrders.vue`: outstanding calculations, progress, "+N extra", PDF. Manual check.
5. Docs: `technical.md` (the receiving row, the "never more than ordered" rule at line 133, `purchase_order_item` row at line 124, migration list), `guide.md` (receiving steps; one line on extras). Board note for `docs/features/README.md` not needed.
6. `make tailwind` only if a new utility class is used (the "+N extra" colour reuses existing classes). `make assets` before the manual check.

## Changes made in review (added after plan approval)
Found by `/code-review high` and fixed; none changes the approved behaviour.
- Editing an over-received order no longer forces its ordered quantity up to the excess: the floor is `min(received, ordered)` (`planPurchaseOrderEdit`, `quantityFloor` in the editor).
- The confirmation re-reads the order first (`refreshPurchaseOrderOutstanding`), because the server no longer refuses a double-receipt of a line; a line that already has all it asked for says so in the box.
- Clamped "outstanding" and "extra" figures moved into one helper, `frontend/src/next/lib/purchaseOrders.js`; `totals()` is single-pass.
- Migration 48 drops the check by its real name (`purchase_order_item_check`, verified on migration 47) with a by-definition fallback; the down migration restores that name. Both paths are tested.
- Deliberately unchanged: a line that is already complete cannot be reconnected by searching for its product (that extra is an ordinary receipt); the guide now says so.

## Testing
- Unit (`backend/service`): over-receipt accepted; split rows over; limit; audit fields; existing wrong-product/wrong-order/closed cases still pass.
- Integration (`PHARMACY_TEST_DSN`): receive more than ordered end to end, order becomes `received`, stock rises by the full amount, `received_quantity` > `quantity`; Move outstanding and Close ignore the over-received line.
- Manual in the running app (demo data): order with 2 lines, receive 120 of 100 → confirmation lists the line, Cancel returns to the receipt untouched, "Receive anyway" saves; detail shows "+20 extra"; totals/progress ≤ 100%; PDF shows 0 outstanding; a second line still short keeps "Move outstanding" visible.
- Not tested: concurrent over-receipts of the same line (the SQL guard only protects overflow now; `FOR UPDATE` on the order already serialises receipts).

## Deploy notes
Run migration 48 (`make migrate-up`). Rebuild assets and the binary (`make assets`). No permission changes; nobody needs to log in again. Rolling back after any over-receipt needs those lines corrected before migration 48's down will apply.

## Commits
Per the owner's saved preference: all proposed commits (logical groups: migration + repository + service; screens; docs + plan) and the push/PR are shown together after implementation and review, with one yes covering them.

## Follow-up cards
- None proposed. (Receiving extra against an already-complete order is deliberately left as an ordinary receipt.)

## Sources
None needed (internal rule change); all findings are from the repo.
