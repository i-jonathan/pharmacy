# T-022 Prevent negative stock on sale

Card: T-022 · Branch: fix/T-022-prevent-negative-stock-on-sale

## Goal
A sale can no longer take a product's stock below zero unless someone allowed to do it confirms it,
and that override is recorded. This stops stock figures from drifting into negatives nobody explains,
and closes the race where a sale and a stock adjustment commit at the same moment.

## What I found
- Verified on the card: `CreateSale` writes `OUT_SALE` movements with no stock check —
  `backend/service/sales.go:146-157` (`code`).
- Verified on the card: adjustments lock products with `LockProductStockTx` (`FOR NO KEY UPDATE` in
  id order, then the stock is read in a second statement) — `backend/repository/stock_adjustments.go:19-44`,
  used at `backend/service/stock_adjustments.go:244` (`code`).
- `FOR NO KEY UPDATE` does not block the `FOR KEY SHARE` that foreign keys from `sale_item` and
  `stock_movement` take, so sales that lock products don't hold up receipts or flags; they queue only
  behind adjustments, product edits and other sales of the same product (`code`, comment at
  `stock_adjustments.go:19`).
- Receiving already updates product rows in id order "as stock adjustments lock them"
  (`backend/service/inventory.go:230-236`), so locking in id order in sales keeps every flow in the
  same order (no deadlock). Expiry updates after receiving/stock taking are single autocommit
  statements (`repository/inventory.go:202`, `repository/stock_taking.go:103`) (`code`).
- Stock is in base units: a line takes `quantity × quantity_per_unit` (`sales.go:151`). One product
  can be on two lines (unit and pack prices), so the check sums per product (`code`).
- The till greys out and won't add a product whose loaded stock is 0
  (`frontend/src/next/components/pos/ProductPanel.vue:70-86`), but the quantity can be raised past
  stock (`composables/usePos.js:132`), and the loaded list can be stale (`code`).
- A retry of a saved sale is answered before pricing (`sales.go:51-79`); it must not be refused for
  stock either (`code`).
- Sales during an open stock taking are double-counted — that's T-111, not this card (`code`, board).
- Commit and push: one yes for everything after implementation and review; don't ask about commit
  style (`memory`).
- Card ranking puts money and stock integrity first (`memory`).
- Common POS products (`model`, unverified): many allow selling into negative by default and offer a
  "prevent overselling" setting; pharmacies with unreliable stock counts tend to need an override so
  the queue isn't stalled by a wrong count. Web research skipped: this is internal stock integrity, and
  the design rests on the code above, not on outside sources.

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| Lock the sale's products (id order) and read stock inside the transaction | `code` | yes |
| Refuse a new sale that takes any product below zero: 409 `code: "insufficient_stock"`, naming each product and what's in stock | `code` | yes |
| Override permission for managers, with an audit entry for every oversell | card | yes (see D1) |
| Till: on `insufficient_stock`, reload the product list and show the message; offer "Sell anyway" to users with the override | `code` | yes |
| Till: warn on a cart line whose quantity is more than the loaded stock (doesn't block) | `code` | yes |
| Till: let users with the override add a product showing 0 stock | `code` | yes (see D1) |
| Sales and receipts during a stock take | board T-111 | follow-up (exists) |
| Raise a stock flag automatically on an oversell so someone recounts | owner | yes (D2, decided in chat) |
| Manager approval at a cashier's till (PIN) instead of a manager signing in | `model` | follow-up |

## Decisions for you

**D1. Block outright, or block with a manager override?**
- (a) **Recommended:** block by default; a new permission `sale:oversell` (Manager, Admin) lets that user
  complete the sale after a confirmation. Each oversell writes an audit entry. The shop's stock figures
  won't always be right; without an override a sale waits until someone does a stock adjustment.
- (b) Hard block, no override: cleanest numbers. Every wrong count holds up the queue until a manager
  adjusts stock (Stock adjustments page) and the cashier retries.

**D2. Raise a stock flag automatically when a sale oversells?**
Decided: yes, in this card (owner, after approval). Small: no migration, no new screen.

Owner's answers: D1 = (a) override with `sale:oversell`; D2 = yes.

## Design
- **Migration 000052_sale_oversell_permission** (only with D1a): seeds `Sale`/`Oversell`
  (`sale:oversell`) to Manager and Admin, same shape as 000050. Down migration removes the
  role_permissions rows and the permission.
- **Constant** `OversellPermissionKey = "sale:oversell"` in `internal/constant/permissions.go`.
- **Request**: `types.Sale` gains `allow_oversell bool` (absent = false).
- **Service** (`CreateSale`, new sale only, after pricing and before `CreateSaleTx`):
  1. Sum `quantity × quantity_per_unit` per product; lock with `LockProductStockTx` (sorted, distinct ids).
  2. Any product where stock − taken < 0 is short. None short → carry on.
  3. Short and (not `allow_oversell` or no `sale:oversell`) → 409 `code: "insufficient_stock"`, message
     like "Not enough stock: Paracetamol 500mg (3 in stock, selling 10)." Money and stock numbers are
     fine to repeat here (unlike health readings). A user who sent `allow_oversell` without the
     permission gets 403.
  4. Short and allowed → write the sale, then one audit entry `sale`/`oversold` with the sale id and,
     per short product, id, stock before and quantity sold. Same transaction.
  5. Then, per short product, open a stock flag raised by the selling user, note
     "Sold N with M in stock (receipt R-…)", no count. New repository method
     `OpenSaleStockFlagTx` uses `ON CONFLICT … DO NOTHING` on `stock_flag_one_open_per_user`: a
     unique violation would abort the sale's transaction, and if this user already has an open flag
     on the product one is already waiting. `stock_at_flag` is read after the sale's movements, so
     it shows the negative figure. It appears on the existing Stock flags page.
  - A key retry (`keyUsed`) skips all of this, as it skips pricing. If the sale insert comes back as
    a duplicate key (a twin request saved it), the existing path returns the saved sale.
- **Repository**: reuse `LockProductStockTx`; no new query.
- **Till** (`PosView.vue`, `ProductPanel.vue`, `CartPanel.vue`):
  - On `insufficient_stock`: refresh the product list and show the server's message. If the user has
    `sale:oversell`, open the shared confirm dialog ("Sell anyway?", the message as detail, tone
    warning); agreeing resends the same sale (same retry key) with `allow_oversell: true`.
  - A cart line whose `qty × quantity_per_unit` is more than the loaded stock shows "Only N in stock"
    under it. It's a hint; the server decides.
  - Users with `sale:oversell` can add a product showing 0 stock (still greyed, with the hint).
- **Held sales**: nothing changes; a resumed held sale is checked when it's completed.
- **Concurrency**: two sales of the same product now queue behind each other for the length of one
  sale's transaction (milliseconds). An adjustment and a sale can no longer both pass their checks
  on the same stock.

## Steps
1. Migration + permission constant (D1a). Test: migrations run in the integration test DB.
2. Service: stock check and refusal — unit tests with mocks: no shortage passes; one short line is 409
   `insufficient_stock`; two lines of the same product are summed; `allow_oversell` without the
   permission is 403; with the permission it saves, writes the audit entry and opens a flag per short
   product; a key retry is not checked.
3. Integration test (`PHARMACY_TEST_DSN`): a real sale is refused at stock 2 selling 3; succeeds at
   stock 3; oversell by a manager leaves stock −1, an `oversold` audit row and an open flag with
   `stock_at_flag` −1; a second oversell by the same manager doesn't fail on the existing flag.
4. Till: handle `insufficient_stock` (refresh, message, confirm + resend for the override).
5. Till: cart line stock hint; let override users add 0-stock products. `make assets`, check in the app.
6. CLAUDE.md gotcha + migration list; feature doc touch if one covers the till.

## Testing
- Unit: `go test ./service/ -run 'TestCreateSale' -v`.
- Integration: `PHARMACY_TEST_DSN=... go test ./service/ -run 'Stock|Sale' -v`.
- Manual: in the running app, as a cashier sell more than stock (refused with the message); as a
  manager sell anyway (confirm, sale saved, stock negative, audit row); a normal sale still works.
- Not tested: the two-transaction race directly (it relies on Postgres row locks, same as the
  adjustment's existing check, which isn't race-tested either).

## Deploy notes
Run migration 000052. Permissions are kept in the session (`middleware/auth.go`), so Managers and
Admins must sign in again to get `sale:oversell`; until then they are refused like a cashier.
`make assets` and restart.

## Commits
Owner's standing rule (memory): logical groups, all shown for one yes after implementation and review.
Planned: (1) migration + service + tests, (2) till changes + built CSS, (3) docs.

## Follow-up cards
- T-148 Manager approval at a cashier's till for stock overrides — P2, Point of sale (added).
- T-111 (existing) extended: stock taking completion should lock products too (review finding).

## Sources
None outside the repo (web research skipped, see What I found).
