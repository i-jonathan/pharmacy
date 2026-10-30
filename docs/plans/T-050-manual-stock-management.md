# T-050 Manual stock management

Card: BOARD.md T-050 · Branch: feat/T-050-manual-stock-management

## Goal

Let managers correct stock without a full stock take: write off expired or damaged items, record losses, enter opening stock, fix a miscount. Each correction records why, who and when.

## What I found

- **Wrong on the card** (`code`): "Today the only way is the stock field on the product edit form." The new product edit form (`views/ProductList.vue` `saveProduct`) has no stock field and sends no `stock`, so **the app has no way to adjust stock except a stock take.** The card will be corrected.
- **The API still allows a silent adjustment** (`code`, `service/inventory.go` `UpdateProduct`):
  - `PUT /inventory/product/:id` still takes `stock`, for anyone with `inventory:edit`.
  - It writes an `IN/OUT_MANUAL_ADJUSTMENT` movement with no reason, no user and no audit entry, and its `reference_id` is the product id.
  - It reads the current stock outside the transaction.
- **How stock works** (`code`): stock is the sum of `stock_movement` rows (`IN%` adds, `OUT%` subtracts; migration 38's `inventory_view`), in base units (a pack of 10 counts as 10). `reference_id` points at the record that caused the movement (sale, receipt, stock take).
- **Expiry** (`code`): stock taking can correct `product.current_expiry` (`UpdateProductCurrentExpiry`). Stock isn't tracked per batch (T-049), so per-batch adjustments aren't possible yet.
- **Patterns to reuse** (`code`): purchase orders for audit entries in the same transaction, `HasPermission`, permission seeding by migration (000043), and the shared confirm dialog. Sales and receiving for idempotency keys.
- **Common practice** (`web`, vendor guidance; the specific numbers are examples, not standards):
  - fixed reason codes
  - "write-off" (lost, damaged, expired) kept apart from count corrections
  - an audit trail of who changed what and when
  - approval above a threshold, with the person adjusting kept separate from the approver
  - reviewing adjustments by reason over time
- **Pharmacy specifics** (`model`, unverified): controlled drugs need loss reporting and lot details. That waits for T-049 and T-059; no Nigerian rule was checked for this ticket.

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| Adjust several products in one go: set a counted quantity, or add or remove a quantity | card; `web` | yes |
| A required reason from a fixed list, plus a note (required for "Other") | card; `web` | yes |
| Record who, when, before and after, and the cost at the time (for write-off value later) | `web`; T-062 | yes |
| Correct the expiry date while adjusting (product level, like stock taking) | card; `code` | yes |
| A permission of its own, given to Manager and Admin | card | yes |
| Audit entry for every adjustment | card; `web` | yes |
| History screen: who adjusted what, why, with the change in units | `web` (review by reason) | yes |
| Remove the silent `stock` field from product edit | `code` | yes |
| Approval above a quantity or value threshold | card; `web` | follow-up (decision 1) |
| Per-batch adjustments and lot details | card; `model` | follow-up, after T-049 |
| Entering quantities in packs instead of units | `model` | follow-up |
| Write-off and shrinkage reports by reason | `web` | T-062 |

## Decisions for you

1. **Approval:** **decided by the owner: only managers adjust; cashiers flag products for checking.**
   - A cashier flags a product (what looks wrong, and optionally the count they saw).
   - Managers see open flags and either adjust (which resolves the flag) or dismiss it with a note.
   - No threshold approval; that stays a follow-up.
2. **Who can adjust:** **decided: Manager and Admin** (`stock:adjust`). Flagging (`stock:flag`) goes to Cashier, Manager and Admin.
3. **Reasons:** **decided:** Expired, Damaged, Theft or loss, Count correction, Opening stock, Internal use, Other (Other needs a note).
4. **The old `stock` field on product edit:** **decided: remove it.**

## Design

**Migration `000044_stock_adjustments`**, with a down migration:
- `stock_adjustment`: id, reason (checked against the list), note, `created_by_id`, `created_at`, and `idempotency_key` (unique when set). Append-only: a mistake is fixed with a new adjustment.
- `stock_adjustment_item`: adjustment id, product id, `quantity_before`, `quantity_change` (signed), `unit_cost_kobo` (the product's cost price at the time), `expiry_before`, `expiry_after`. A check requires a quantity change or an expiry change.
- Seeds `Stock`/`Adjust` for Manager and Admin, and `Stock`/`Flag` for Cashier, Manager and Admin, the same way 000043 does.
- Creates the `stock_flag` table (above).

**Flags (owner's addition):**
- **Table `stock_flag`:**
  - id, product id, note (required), optional `counted_quantity`
  - status: `open`, `resolved` or `dismissed`
  - `flagged_by_id`, `created_at`
  - `resolved_by_id`, `resolved_at`, `resolution_note`, and `stock_adjustment_id` when resolved by an adjustment
- **API:**
  - `POST /inventory/stock-flags` (`stock:flag`): `{product_id, note, counted_quantity?}`
  - `GET /inventory/stock-flags` (`stock:adjust`): open flags, newest first
  - `GET /inventory/stock-flags/count` (`stock:adjust`): for the sidebar badge
  - `POST /inventory/stock-flags/:id/dismiss` (`stock:adjust`): `{note}`
  - An adjustment can carry `flag_ids`; those flags are resolved, and linked to the adjustment, in the same transaction.
- **Audit:** entity `stock_flag`, actions `created`, `dismissed` and `resolved`.
- **UI:**
  - **Flag for check** in product details for `stock:flag`
  - a Flagged list on the Stock Adjustments page with Adjust and Dismiss
  - a sidebar badge for managers, polled every 60 s like purchase approvals

**Stock movements:** one `IN_MANUAL_ADJUSTMENT` or `OUT_MANUAL_ADJUSTMENT` per changed item, with `reference_id` set to the adjustment id. No new movement types: the reason lives on the adjustment, so reports can group by it.

**API:**
- `POST /inventory/stock-adjustments` takes `{reason, note, idempotency_key, items: [{product_id, mode: "set"|"add"|"remove", quantity, expiry?}]}` and returns `201 {id}`. A retry with the same key returns the same id.
- `GET /inventory/stock-adjustments?page=` lists adjustments, newest first: user, reason, note, and items with product name, before, change and after.
- Both need `stock:adjust`, at the route and in the service.

**Service `CreateStockAdjustment`:**
1. Validate:
   - the reason, and a note when the reason is Other
   - 1 to 200 items, no product twice
   - set needs a quantity of 0 or more; add and remove need more than 0
   - an expiry date that's valid if given
2. In one transaction:
   - lock the products' rows (in id order, so concurrent adjustments queue up)
   - read the current stock inside the transaction
   - work out each change; refuse anything that would leave stock below zero
   - insert the adjustment and its items, the movements, and the expiry updates
   - write the audit entry (entity `stock_adjustment`, action `created`)

**Product edit:** remove `Stock` from `UpdateProductRequest` and its code path.

**UI:**
- New **Inventory → Stock Adjustments** page, shown to users with `stock:adjust`, with two parts:
  - **History:** date, who, reason, products and net units.
  - **New adjustment:**
    - search to add products
    - for each row: current stock, Set to / Add / Remove, quantity, an optional new expiry, and a live "12 → 10 (−2)"
    - reason and note
    - the shared confirm dialog summarises the change before saving
- The product details panel gets an **Adjust stock** button that opens the form with that product already added.

**Offline and idempotency:** the client sends a key, so a retried save can't double the adjustment.

## Steps

1. **Migration** up and down, with permission seeding. Test: migration test (apply, roll back, re-apply).
2. **Repository:** insert an adjustment and its items, lock products, read stock in the transaction, update expiry, list adjustments. Test: integration.
3. **Service** `CreateStockAdjustment` and `ListStockAdjustments`. Unit tests first: each mode, below-zero refused, duplicates, reason and note rules, idempotent retry, audit entry, permission. Plus an integration test: the stock seen through `inventory_view` after an adjustment, and concurrent adjustments of the same product.
4. **Routes and controller.** Test: route tests (permission, status codes, body limit).
5. **Remove `stock` from product edit.** Test: an update with `stock` changes nothing.
6. **UI:** the Stock Adjustments page, the sidebar link, and Adjust stock in product details. Then `make assets`. Test: one-off Playwright check.
7. **Docs and board:** README (routes, permission) and `CLAUDE.md` (a note on adjustments); correct the card.

## Testing

- **Automated:** `go test ./...` with `PHARMACY_TEST_DSN`, and migration apply, roll back and re-apply on a throwaway database.
- **Browser:** a one-off Playwright run through the new page on a throwaway server. No JS test tooling is committed (T-076).

## Deploy notes

- Run migration 44.
- Everyone must sign in again to pick up `stock:adjust` and `stock:flag` (T-023).
- `make assets`.

## Commits

Proposed at ship time, one approval for all:

1. Migration 44: tables, constraints, permission seeding, migration test
2. Fix `RequirePermissions` falling through when no permission matches
3. Stock adjustment and flag service and API (types, repository, service, controller, routes, tests)
4. Remove the stock field from product edit
5. Stock Adjustments page, sidebar badge, product panel actions
6. Docs, plan and board

## Follow-up cards

- **Approval for large stock adjustments** above a quantity or value threshold, reviewed by someone other than the person adjusting. P1, Inventory control. Depends on T-093.
- **Adjust stock in packs as well as units.** P2, Inventory control.
- Note on T-049: per-batch adjustments and lot details.
- Note on T-062: write-off value by reason from `stock_adjustment_item.unit_cost_kobo`.

## Sources

- https://appmaster.io/blog/inventory-adjustment-log-reason-codes-audit-trail — reason codes, write-off kept apart from count adjustments, approval thresholds, keeping adjusting and approving separate, and an audit log of who counted and approved.
- https://www.hyperbots.com/glossary/inventory-adjustments-approval — approval thresholds by amount, category or reason.
- https://www.mangoapps.com/templates/sop/inventory-adjustment-journal-sop — requiring a reason code and evidence; reviewing adjustments monthly.
