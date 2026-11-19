# T-020 Validate sale prices server-side

Card: T-020 · Branch: fix/T-020-server-side-sale-prices

## Goal
A sale is priced by the server, not by the till's browser. Whatever a client sends for price, line total, subtotal, discount or total, the saved sale uses the price on file, and a request whose numbers do not match is refused. This closes a hole in money integrity: today anyone who can reach `POST /sales/` can sell any product for ₦1, or for a negative amount, and the same stored total is what Returns later refunds.

## What I found
- Verified on the card: `CreateSale` stores the client's numbers as sent: `Discount`, `Subtotal`, `Total` (`backend/service/sales.go:63-67`) and per line `UnitPrice`, `Discount`, `TotalPrice` (`sales.go:129-136`), all from `types.Sale` / `types.SaleItem` JSON (`backend/internal/types/sales.go:14-39`) (`code`).
- The server already loads the selected price row per line, but only for `QuantityPerUnit` (`sales.go:138-143`); `selling_price` (kobo) is ignored (`code`, `model/inventory.go:46-52`).
- Not on the card, found in the code (`code`):
  - `price_id` is never checked against `product_id`. `FetchPriceByID` is `SELECT * FROM product_price WHERE id = $1` (`repository/query.go:272`), so a line can name product A with product B's price row and so a different price and a different `quantity_per_unit` (stock taken off).
  - An unknown `price_id` is a 500 ("failed to fetch product price"), not a 400 (`sales.go:138-142`).
  - `quantity` is not checked: zero or negative is saved, and a negative quantity posts a negative `OUT_SALE`, which adds stock (`sales.go:128-150`). There is no check that a sale has items.
  - Money is converted with `int(x * 100)` on floats, which truncates: `19.99 * 100` is `1998.9999…`, saved as 1998 (`sales.go:65-67,133-135,164`).
  - Payments (`Amount`) are saved unchecked, negative included, and need not add up to the total (`sales.go:159-166`).
  - `ReturnItems` refunds from the stored `sales_item.total_price` (`sales.go:906-909`), so an altered total is also an inflated refund. Fixing the sale fixes this.
- The till (`frontend/src/next/composables/usePos.js:226-244`) sends `unit_price`, `discount` and `total = price*qty - discount` from the cart, so an honest till will match the server except when a price changed after the product was added to the cart or the held sale was parked (`code`). The line discount input is free-typed (`CartPanel.vue:100-107`); nothing stops a negative or larger-than-the-line discount (`code`).
- A retry of a saved sale (same idempotency key) returns before items are processed (`sales.go:87-103`), so re-pricing after that point leaves retries unaffected: a retry after a price change is still answered with the first save (`code`).
- No discount rule or permission exists anywhere in the backend (`code`: no `discount` handling outside the sale code, repository queries, demo seed and types). Role-based discount limits, price overrides and promotions are card T-046 (Backlog, P1) (`code`, board).
- Existing tests: `customerSaleRepo.FetchPriceByID` returns `{QuantityPerUnit: 1}` with no product or price, and sales use `UnitPrice: 100`-style numbers, so mocks and fixtures need real prices (`code`, `service/sale_customer_test.go:81`, `idempotency_test.go:74`, integration tests under `service/*_integration_test.go`).
- Wrong or stale on the card: nothing wrong. "enforce discount rules" has no rules to enforce yet; the plan below sets the minimum sane ones and leaves role limits to T-046.
- Existing patterns to copy: typed refusals with `Reason` codes (`httperror.Conflict` with `refusal.Reason = "customer_inactive"`, `sales.go:117-122`); `badRequest(...)` helper; tx-scoped reads.
- Web research skipped: an internal integrity fix with no outside rule or library in play. (`model`, unverified: POS systems generally price on the server and treat client totals as a checksum.)

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| Unit price from `product_price.selling_price`, line total and sale subtotal/discount/total recomputed in kobo | `code` | yes |
| `price_id` must belong to the line's `product_id`; unknown id is a 400 | `code` | yes |
| Quantity must be a positive whole number; at least one item | `code` | yes |
| Line discount: whole kobo, ≥ 0, ≤ the line's gross (price × quantity) | `code` | yes |
| Client totals/price that disagree with the server are refused, not silently replaced | decision below | yes |
| Rounding with `math.Round`, not truncation | `code` | yes |
| Payments: reject negative or non-finite amounts | `code` | yes (small, same function) |
| Payments must cover the total; change handling | `model` unverified: depends on whether the till sends tendered cash or capped amounts | follow-up |
| Max discount per role, price override permission, promotions | `code`, board T-046 | follow-up (exists) |
| Till refreshes cart prices on a price change instead of remove-and-re-add | `model` | follow-up |
| Block sales above stock | `code`, board T-022 | follow-up (exists) |

## Decisions for you
Decision 1 was answered: **A, refuse with 409**. Commit style: logical groups.

1. **A price changed since the item went in the cart** (or a held sale resumes with old prices).
   - A. Refuse with 409 `price_changed` naming the product and the new price; the cashier removes and re-adds the line. The customer was quoted the cart price, and the receipt already shows it, so nothing is charged that they did not see. **Recommended.**
   - B. Silently charge the server price. The printed receipt, built from the cart, could then disagree with the record.
2. **Client totals that do not match** (subtotal, discount, total, line total, unit price): with A they are checked against the server's numbers; a mismatch with no price change is a 400 "the sale total does not add up", meaning a bug or tampering. Under B they would be ignored. (Follows from decision 1.)
3. **Discount rule for now**: 0 ≤ discount ≤ line gross, any cashier, until T-046 adds limits. This keeps today's behaviour for honest use; say so if you want a per-sale cap now instead.

## Design
- **No migration, no new permission.**
- **API:** `POST /sales/` body unchanged (the till still sends prices and totals; the server now treats them as a checksum). New refusals:
  - 400 for: no items, quantity ≤ 0, discount negative/non-finite/more than the line, a `price_id` that is unknown or belongs to another product, a negative or non-finite payment, client totals that do not match.
  - 409 `code: "price_changed"` when the client's `unit_price` differs from `product_price.selling_price`; the message names the product, the old and the new price.
- **Service (`service/sales.go`, `CreateSale`):** after the idempotency short-circuit and the customer check, price each line in the same transaction: fetch the price row, check `ProductID`, compute `gross = selling_price × quantity`, convert the client's discount to kobo with `math.Round` and check it, set `UnitPrice`, `Discount`, `TotalPrice = gross − discount`; sum the lines for sale `Subtotal`, `Discount`, `Total`; compare the client's numbers (converted with the same rounding) and refuse on mismatch. The sale row is inserted before items today (`CreateSaleTx`, needed for the idempotency key), so its subtotal, discount and total are filled in once the lines are priced: either move the line pricing before `CreateSaleTx` (validation only needs reads) or update the row afterwards. I will move pricing and validation before the insert, so a bad request never touches the sale table (the key lookup for a retry stays first, so retries are not re-priced).
- A small pure helper `priceSale(items, lookup) (…)` in `service` so the maths is unit tested without a repository.
- **Repository:** reuse `FetchPriceByID` (add a tx-scoped variant only if the tests show reads outside the transaction matter; a price edited mid-sale is the same race as today).
- **Audit:** none; sales are not audited today (T-035 covers that).
- **UI (added in review):** `usePos.js` `updateDiscount` rounds a line discount to whole kobo, because the server now refuses fractions of a kobo (a typed 10.555 would otherwise make the line total and the sale discount round differently). Needs `make build-frontend`; no new Tailwind classes. Otherwise no UI change: `usePos.js` already shows the server's message for 400/409 (`usePos.js:261-269`); the 409 message tells the cashier to remove and re-add the line. I will check `PosView.vue` displays the message and adjust only if it does not.
- **Idempotency:** a retry of a saved sale returns the first save before any of this runs, as now.

## Steps
1. Tests first (unit, `service/sales_pricing_test.go`): server price replaces a lowered client price (409), line total/subtotal/total recomputed from price × quantity − discount, cents round not truncate (e.g. ₦19.99), wrong-product `price_id`, unknown `price_id` (400), zero/negative quantity, no items, negative discount, discount above the line, client total mismatch, negative payment; plus a retry of a saved sale is not re-priced. Confirm they fail.
2. Implement `priceSale` and wire it into `CreateSale` before the sale insert. Fix mocks and fixtures in the existing sale tests (`sale_customer_test.go`, `idempotency_test.go`, `sales_receipt_test.go`) to carry real prices. Tests green.
3. Integration test (`PHARMACY_TEST_DSN`): save a sale through the real repository with a tampered unit price and total and confirm nothing is written; an honest sale stores the server's numbers; a refund after a return matches the stored lines.
4. Docs: one line in `CLAUDE.md` Gotchas (sale prices are server-set; client numbers are a checksum), and the customers/sales feature docs if they describe the sale payload.

## Testing
- Unit: `cd backend && go test ./service/ -run 'Sale|Price' -v`, then `go test ./...`.
- Integration: `PHARMACY_TEST_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" go test ./...`.
- Manual in the running app: ring up a normal sale, a discounted sale and a held-then-resumed sale; change the price of an item in another tab and complete the old cart to see the `price_changed` message. Needs `make build-frontend` only if the UI changes.
- Not tested: concurrent price edits during a sale (same race as today).

## Deploy notes
No migration or permission. Restart the server. Tills with a page open from before the deploy keep working unless their prices are stale.

## Commits
Logical groups (chosen by the owner). Planned points, each confirmed in chat before committing:
1. Plan file, then unit tests plus the pricing logic and fixed mocks (steps 1–2).
2. Integration test (step 3).
3. Docs (step 4).

## Follow-up cards
- P1, Sales: Sale payments must cover the total (and how change is recorded); the till's payment payload needs checking first.
- P2, Sales: Till refreshes a cart line's prices when the server answers `price_changed`, so the cashier need not remove and re-add it.
- (Existing: T-046 discount limits and price overrides, T-022 negative stock.)

## Sources
- Code paths above; board cards T-046 and T-022. No outside sources used.
