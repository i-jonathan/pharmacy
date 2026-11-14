# T-118 Attach a customer to a sale

Card: T-118 · Branch: feat/T-118-attach-customer-to-sale

## Goal

A cashier can find a customer by phone (or name) at the till, or add one on the spot, and attach them
to the sale. The sale remembers who bought, Sales History shows it, and the receipt and every
reprint carry the customer's name. This is what T-119 (history), T-041 (credit) and T-084 (loyalty)
build on.

## What I found

- `code` The till has a free-text box with "Walk-in Customer" as its default (`CartPanel.vue`
  Customer block) and a "+ New" button that does nothing. The text is kept in `usePos.js` `customer`,
  stored with a held sale (`holdCart`), printed on the receipt (`printReceipt`, `receipt.js` prints
  `r.customer`, so every walk-in receipt says "Walk-in Customer"), and **never sent to the server**:
  `completeSale` builds a payload with no customer, and `types.Sale`, `model.Sale` and the `sales`
  table have no customer field (`createSaleQuery`, migration 11). The card is right.
- `code` `CreateSale` (`service/sales.go`) is idempotent by key: `CreateSaleTx` is
  `INSERT … ON CONFLICT (idempotency_key) DO NOTHING`; when nothing is inserted the first save's
  receipt number is returned and nothing else runs. The customer check has to run only for a new
  sale, or a retry of a saved sale could fail because the customer was deactivated in between.
- `code` Sales History is built by `FetchSalesHistory` from `fetchSalesQuery` (no customer column),
  and its route already carries the permissions (`AddPermissionsToContext` in `router/sales.go`), so
  `HasPermission` works there. Anyone signed in can read Sales History today.
- `code` Reprints are built in the browser from the history row (`SalesHistory.vue` `receiptFor`);
  the server only records the reprint. So a reprint shows whatever the history row carries.
- `code` Customers exist (T-040): `GET /customers/?q=` searches by name or phone and hides deactivated
  customers by default, `POST /customers/` adds one (409 `phone_taken` names the owner);
  `customer:view` and `customer:manage` are held by Cashier, Manager and Admin. `customer.id` is
  `BIGSERIAL`.
- `code` The completed-sale call throws a fixed "Failed to complete sale" for any refusal
  (`usePos.js` `completeSale`), so a server message such as "customer deactivated" would never reach
  the cashier.
- `code` Pattern to copy for FK errors: `mapForeignKeyError` / `ErrInvalidReference`
  (`repository/purchase_orders.go`). Sale mocks embed `repository.PharmacyRepository`
  (`service/idempotency_test.go`), so a nil customer adds no repository calls to existing tests.
- `web` Other POS products attach customers the same way: search by phone or name, an "add new"
  when nothing is found (saved and attached in one step), and a default walk-in when none is chosen
  ([Loyverse](https://help.loyverse.com/help/customer-identification),
  [ConnectPOS](https://www.connectpos.com/knowledgebase/adding-customers-guest-existing-new),
  [Cin7 Core](https://help.core.cin7.com/hc/en-us/articles/10553585082383-Add-a-customer-to-a-POS-sale),
  [Retailcloud](https://retailcloud.freshdesk.com/support/solutions/articles/12000005905-add-customer-to-a-sale)).
  I found nothing on what a receipt should print for the customer.
- `web` Privacy regulators advise not asking shoppers for personal details the receipt doesn't need,
  and keeping any such request optional ([Irish DPC e-receipts guidance](https://www.dataprotection.ie/en/dpc-guidance/guidance-retailers-issuing-e-receipts),
  [CHOICE](https://www.choice.com.au/shopping/consumer-rights-and-advice/your-rights/articles/receipt-refusal)).
  Retail-side, not Nigerian; I am not giving legal advice. The design keeps the customer optional
  and prints only the name.
- `memory` Commits and push are asked once, after implementation and review.

## Making it complete

| Gap | Source | In scope? |
|---|---|---|
| `sales.customer_id` (nullable FK, index) | card | yes |
| Till: search by phone or name, pick, remove | card | yes |
| Till: add a new customer when none is found, attached at once | `web` (every POS does) | yes (uses `POST /customers/`) |
| Customer saved with the sale and returned by Sales History | card | yes |
| Name on the receipt and on reprints | card | yes |
| Held sales keep the customer | `code` (held payload has one today) | yes |
| Server refuses an unknown or deactivated customer, and a user who may not see customers | `code` (permissions) | yes |
| Refusal reaches the cashier in words | `code` (fixed message today) | yes |
| Customer hidden from users without `customer:view` (history and POS) | `code` (privacy) | yes |
| Docs, tests, demo data | `code` | yes |
| Change or remove the customer on a saved sale | `model` | follow-up |
| "What's new" slide for customers | `code` (needs the release tag, owner's call) | follow-up |
| Filter Sales History by customer, a customer's own history | card T-119 | T-119 |
| Phone or notes on the receipt, customer shown on held sales | `model` | no |

## Decisions for you

1. **Receipt text.** Recommend: an attached customer prints `Customer: <name>`; a sale with no
   customer prints nothing. Today every walk-in receipt prints "Walk-in Customer", which says
   nothing. Alternative: keep printing "Walk-in Customer" for walk-ins.
2. **Who may attach.** Recommend `customer:view` (the cashier must be able to see customers to pick
   one; the server checks it). A user with neither customer permission sees no customer control and
   sells as walk-in. `customer:manage` is needed for "add new" at the till.
3. **A customer deactivated after being chosen** (for example on a held sale resumed later): the sale
   is refused with "this customer has been deactivated; remove them from the sale or reactivate
   them", and the cashier removes the customer and completes the sale. Alternative: let the sale
   through. Recommend refusing, consistent with "no longer found at the till". A retry of a sale that
   already saved is unaffected.
4. **Who sees the customer on Sales History and reprints.** Recommend only users with
   `customer:view`; others get the same history without it. Alternative: everyone who can open Sales
   History.
5. **Free-text names on held sales from before this change** are dropped when resumed (the text was
   only ever printed on the receipt). Held sales made after this change keep the chosen customer.

## Design

**Migration `000049_sales_customer`** (down drops the index and column)
- `ALTER TABLE sales ADD COLUMN customer_id BIGINT REFERENCES customer(id)`, and a partial index
  `sales_customer_idx ON sales (customer_id) WHERE customer_id IS NOT NULL` (T-119 and credit need it).
  Existing sales stay NULL. No new permission.

**Backend**
- `types.Sale.CustomerID int` (`customer_id`, 0 = none); `model.Sale.CustomerID *int`;
  `createSaleQuery` takes it; `CreateSaleTx` maps a foreign-key violation to `ErrInvalidReference`.
- `fetchSalesQuery` also selects `customer_id`. New `BulkFetchCustomerNames(ctx, ids)` and
  `GetCustomerTx(ctx, tx, id)` on the customer repository.
- `CreateSale`: a negative id is a 400; a non-zero id needs `customer:view` (403); the sale is
  inserted with it (unknown id → 400 "customer was not found"); only if the sale is new (not a key
  retry) the customer is read and a deactivated one is refused with 409 and `code:
  "customer_inactive"`, rolling the sale back.
- `FetchSalesHistory`: `SaleResponse.Customer *{id, name}`, filled only when the caller has
  `customer:view`, from one bulk lookup.
- No audit entry: sales are not audited today and attaching a customer changes no stock or money.

**Frontend**
- `usePos.js`: `customer` becomes `null` or `{id, name, phone}`; the sale payload sends `customer_id`;
  a held sale stores the object (an old text value is ignored on resume); `completeSale` passes the
  server's message on for 400, 403 and 409 instead of the fixed text.
- New `components/pos/CustomerPicker.vue` replaces the text box and the dead "+ New": shows "Walk-in
  customer" with **Add customer**, or the chosen customer with change and remove; search box (250 ms
  debounce, `GET /customers/?q=`), results with name and phone, **Add "<text>" as a new customer**
  when nothing matches (name and phone, `POST /customers/`; on `phone_taken` it shows who has the
  number and lists them to pick); Escape closes; hidden entirely without `customer:view`.
- `receipt.js`: `Customer: <name>` only when a name is given. `SalesHistory.vue`: customer name
  beside the cashier in the sale details and on the reprint (`receiptFor`).
- New Tailwind classes need `make tailwind` and the regenerated `tailwind.css` committed.
- Demo seed: attach customers to a few demo sales.

## Steps

1. Migration 49 and its migration test (column, FK, index, rollback). `PHARMACY_TEST_DSN`.
2. Types, model, queries, repository (`CreateSaleTx`, bulk names, `GetCustomerTx`).
3. `CreateSale` and `FetchSalesHistory` with tests first: no customer unchanged; permission refusal;
   unknown customer 400; deactivated customer 409 and nothing saved; key retry skips the check;
   history carries the name only with `customer:view`. A real-database test of a sale with a customer.
4. Demo seed and its test.
5. Till: `usePos.js`, `CustomerPicker.vue`, `CartPanel.vue`, `PosView.vue`, `receipt.js`;
   `SalesHistory.vue`. `make assets`.
6. Run the app on a demo database and use it as cashier, manager and a user with neither permission.
7. Docs: `docs/features/customers/` (guide and technical, drop "nothing reads the list yet"),
   `CLAUDE.md`, plan Changes-after-review if any.

## Testing

Unit (mock repository) and real-database integration tests as listed; router tests are unchanged
because no route is added. Manual: search by `0803…`, `+234 803…` and name; add a new customer at the
till; duplicate phone at the till; walk-in sale; deactivate a customer then complete a held sale;
reprint shows the name; a user without `customer:view` sees no picker and no customer in history.
Not tested: browser automation (none in the repo).

## Deploy notes

Run migration 49 (additive, nullable; 48 is T-123's). No new permission. Run `make assets` before building the
image. Printed receipts for walk-ins stop saying "Walk-in Customer" (Decision 1).

## Commits

Per memory, one yes after implementation and review. Proposed groups: (1) migration and tests,
(2) repository and service with tests and demo data, (3) till, receipt and Sales History,
(4) docs.

## Follow-up cards

To add on the board after approval: **Change or remove the customer on a saved sale** (P2, audited,
needs a reason and `customer:manage`); **What's new: customers slide** (P2, needs the release tag
and pictures, `docs/whats-new/README.md`).

## Sources

- [Loyverse: customer identification by phone](https://help.loyverse.com/help/customer-identification): search by phone, attach to the ticket
- [ConnectPOS: guest, existing, new customer](https://www.connectpos.com/knowledgebase/adding-customers-guest-existing-new): create-and-attach in one step
- [Cin7 Core: add a customer to a POS sale](https://help.core.cin7.com/hc/en-us/articles/10553585082383-Add-a-customer-to-a-POS-sale): walk-in default
- [Retailcloud: add customer to a sale](https://retailcloud.freshdesk.com/support/solutions/articles/12000005905-add-customer-to-a-sale): partial phone search
- [Irish DPC: e-receipts guidance](https://www.dataprotection.ie/en/dpc-guidance/guidance-retailers-issuing-e-receipts): keep personal details at checkout optional
