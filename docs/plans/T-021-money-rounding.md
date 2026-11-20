# T-021 Fix money conversion rounding

Card: T-021 · Branch: fix/T-021-money-rounding

## Goal
Prices typed into the product, receiving and edit screens are stored a kobo short whenever the
float for the naira amount sits just under the whole kobo (19.99 → 1998 kobo, 0.29 → 28). Store
the kobo the person typed, everywhere money comes in as naira.

## What I found
- Verified on the card: naira floats are converted with `int(x * 100)`, which truncates — `code`:
  - `service/inventory.go:52`, `:64` (CreateProduct: cost, default selling price)
  - `service/inventory.go:250`, `:259`, `:260`, `:270`, `:277` (ReceiveProductSupply: batch cost,
    product cost + default selling price, new/edited price options)
  - `service/inventory.go:438`, `:469` (UpdateProduct: cost, every price option)
- Stale on the card: sales are already fixed. T-020 added `toKobo` (`service/sales_pricing.go:44`),
  which rounds and refuses a fraction of a kobo, non-finite numbers and amounts over `MaxInt32`;
  every sale amount goes through it. Only the inventory paths above remain — `code`.
- Purchase orders carry no money (`types.PurchaseOrderItem`, `internal/types/inventory.go:102`);
  stock adjustments copy `product.CostPriceKobo` from the database — `code`.
- Every other conversion is kobo → naira for display (`/ 100`), which is exact enough — `code`.
- Receiving pre-fills a product with no cost price at 70% of its selling price
  (`frontend/src/next/views/ReceiveItems.vue:594`), so a real fraction of a kobo can arrive
  (₦123.45 × 0.7 = ₦86.415) without anyone typing one. Refusing fractions, as sales do, would
  make that receipt fail — `code`.
- No validation on UpdateProduct prices or receiving price-option changes: a negative price or a
  huge number (overflows the `INTEGER` column → 500, or a garbage int from the float conversion)
  is accepted — `code`. CreateProduct and the receiving lines already refuse ≤ 0.
- No unit tests exist for `service/inventory.go` — `code`.
- Web research skipped: internal conversion fix; the rounding rule (round half away from zero,
  `math.Round`) is the one T-020 already uses.

## Making it complete
| Gap | Source | In scope? |
|---|---|---|
| Round instead of truncate at all 10 inventory sites | `code` | yes |
| One shared naira→kobo helper used by sales and inventory | `code` | yes |
| Refuse negative / non-finite / too-big prices on product edit and price options (400, not 500) | `code` | yes |
| Round the 70% cost pre-fill to 2 decimals in ReceiveItems so the screen shows what is saved | `code` | yes |
| Send kobo integers from every screen instead of naira floats | card | follow-up (not needed once the server rounds; large API change) |

## Decisions for you
1. **Round on the server, or switch the API to kobo integers?** Recommend rounding on the server
   with a shared helper (keeps every screen and the JSON unchanged; sales already work this way
   since T-020). Switching the API touches product, receiving, held receipts and the till.
2. **A price with a fraction of a kobo on inventory screens (₦86.415):** round to the nearest kobo
   (recommended — the receiving pre-fill produces these and nobody can see the third decimal), or
   refuse with 400 like sales. Sales keep refusing: there the till's number is only a check.

## Design
- `service/money.go`: `nairaToKobo(naira float64) (int, error)` — `math.Round(naira*100)`; 400 for
  NaN/Inf, negative, or over `math.MaxInt32`. Sales' `toKobo` keeps its fraction check and calls
  the shared range check so both refuse the same bad numbers.
- `CreateProduct`, `ReceiveProductSupply`, `UpdateProduct` convert every price up front, before
  the transaction, and return the helper's 400 naming the field (e.g. "product 2: selling price is
  not valid"). Existing > 0 checks stay where they are; zero stays allowed where it is today.
- No migration, permission, audit or JSON change.
- Frontend: `ReceiveItems.vue` rounds the 70% pre-fill to kobo (`Math.round(x * 100) / 100`).

### Added in review (small, in scope)
- The "> 0" checks in CreateProduct and receiving now test the rounded kobo
  (`positivePriceKobo`): ₦0.004 passed the old float check and was stored as ₦0. The refusal
  reads "… must be at least ₦0.01".
- One helper, `priceKobo(what, naira)`, with refusals that say why: "cannot be negative",
  "is not a number", "is too large (the most is ₦21,474,836.47)". An unnamed price option is
  called "price option N" (`priceOptionLabel`).
- Receive's New Product dialog shows the server's refusal text (`readError`) instead of
  "Failed to create product".

### Added at the owner's request (revision 2): the review's follow-ups
The owner asked to fix the proposed follow-ups on this branch instead of opening cards.

1. **No ₦0 prices** — `code`: `UpdateProduct` accepts cost and price options of 0; receiving
   price options accept 0 (`ReceiveItems.vue` "Add Option" starts at `price: 0`, and `validate()`
   never looks at options). The local database has 6 of 797 products with cost 0 or none and
   1 price option at ₦0, so such rows exist.
   - Server: every selling price (create, receive line, receive options, every edit option) uses
     `positivePriceKobo`. Cost on edit: see decision A.
   - Receiving page: `validate()` marks a price option under ₦0.01 (the row's options button
     turns red, the message says which option) so the cashier fixes it before sending.
   - Product edit page: checks the same before sending, with the field named in the toast.
2. **ProductList cleared box sends `""` → "invalid json"** — `code`: `ProductList.vue:517,522`
   send `editForm.cost_price` / `opt.selling_price` straight from `v-model.number`. Send
   `Number(x) || 0` and check on the page first, so a cleared box reads "Cost price must be at
   least ₦0.01", not "invalid json". Same for `quantity_per_unit` and `reorder_level`, which
   have the same `v-model.number` problem.
3. **New Product dialog closed mid-save** — `code`: Cancel and the backdrop close the dialog
   while `newProductSaving`, and `saveNewProduct` reads `newProduct` and
   `purchaseOrderProductTarget` after the `await`, by which time a second dialog may have
   replaced them. Take a copy of both before the request; disable Cancel and the backdrop while
   saving.
4. **Send kobo integers from the SPA** — left as a card (decision B).

#### Decisions for you (revision 2)
- **A. Cost price on product edit:** require at least ₦0.01, as create and receiving already do
  (recommended: one rule everywhere; the 6 products at ₦0 get a cost the next time someone
  edits them, and the message says so), or keep allowing ₦0 on edit only.
- **B. Kobo integers from the SPA:** leave it as a P2 card (recommended: with rounding on the
  server it fixes nothing that is wrong today, and it touches product, receiving, held receipts
  and the till), or do it here.

Owner's answers (revision 2 approved): A — require ₦0.01+ for cost on edit; B — leave kobo
integers as a P2 card.

#### Steps (revision 2)
7. Server: selling prices and options through `positivePriceKobo` (+ edit cost per A); tests
   for 0 on each path, written first.
8. ProductList: numbers sent as numbers; page-side checks with clear toasts.
9. ReceiveItems: option prices checked in `validate()`; New Product dialog takes a copy before
   the request and can't be closed while saving.
10. `make assets`; run the app and use both screens (create, edit with a cleared box, an
    option at 0, cancel mid-save with a slowed network).

#### Added in review round 3 (small, in scope)
- Receiving no longer lists a product's default price among its price options: the row's Sell
  box is the default and receiving writes it from there, so the listed copy was overwritten
  anyway, and a default stored at ₦0 would now block the receipt.
- The page checks live in `frontend/src/next/lib/money.js` (`amountOf`, `underAKobo`,
  `priceOptionLabel`), shared by both screens; receipt rows use the same "under ₦0.01" check.
- The options button's red mark clears as soon as the option is fixed; no second New Product
  dialog opens while one is saving.

## Steps
1. `nairaToKobo` + table test (19.99 → 1999, 0.29 → 29, 86.415 → 8642, 1999999.99 → 199999999, negative, NaN, Inf,
   over max) — `service/money_test.go`.
2. Use it in CreateProduct; mock-repo test that 19.99/0.29 reach `CreateProductTx` /
   `CreateProductPriceTx` as 1999/29 — `service/inventory_test.go`.
3. Use it in ReceiveProductSupply (batch cost, bulk price map, price options); test the kobo
   reaching the repo and a 400 for a negative price option.
4. Use it in UpdateProduct; test kobo reaching the repo and a 400 for a negative/huge price,
   with no transaction started.
5. Point `toKobo`'s range check at the shared helper; existing sales tests stay green.
6. Round the pre-fill in `ReceiveItems.vue`; `make build-frontend`.

## Testing
- `cd backend && go test ./service/ -v` (new and existing tests), `go vet`, `make check-go-format`,
  `make check-go-lines`.
- Integration tests with `PHARMACY_TEST_DSN` (receipts integration test covers receiving).
- Manual: create a product at ₦19.99, edit cost to ₦0.29, receive one with a 3-decimal cost;
  check the stored kobo in Postgres.
- Not tested: the Vue rounding has no JS test tooling; checked by hand.

## Deploy notes
`make assets` and restart (one-line Vue change). No migration. Prices already stored a kobo short
stay as they are; fixing them would need a guess at what was typed.

## Commits
Logical groups (owner, on approval). Decisions: round on the server; round fractions of a kobo
on inventory screens. Planned commits:
1. Shared helper + inventory service conversions + tests (steps 1–5).
2. Receiving pre-fill rounding in `ReceiveItems.vue` (step 6).
3. This plan.

## Follow-up cards
- Send money as kobo integers from the SPA (P2, Security hardening): removes float conversion at
  the API boundary entirely. Only if you want it.

## Sources
- `service/sales_pricing.go` (T-020) — the existing rounding rule and drift tolerance.
