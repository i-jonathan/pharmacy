# Customers: technical

## Purpose

A stored list of customers (name, phone, email, notes, consent to be contacted) with a screen to manage it, and the till attaches a customer to a sale (`sales.customer_id`), so Sales History and receipts know who bought. Staff guide: [guide.md](./guide.md)

## Overview

```mermaid
flowchart LR
    View[Customers.vue] -->|GET /customers/| List[ListCustomers]
    View -->|POST /customers/| Create[CreateCustomer]
    View -->|PUT /customers/:id| Update[UpdateCustomer]
    View -->|POST /customers/:id/deactivate| Active[SetCustomerActive]
    Create --> Customer[(customer)]
    Update --> Customer
    Active --> Customer
    List --> Customer
    Create -.->|same transaction| Audit[(audit_log)]
    Update -.->|same transaction| Audit
    Active -.->|same transaction| Audit
```

## Code map

| Layer | Location |
|---|---|
| Screens | `frontend/src/next/views/Customers.vue` (list, search, side panel form); `components/PagerBar.vue`; `composables/useConfirm.js`; at the till `components/pos/CustomerPicker.vue` (used by `CartPanel.vue`), `composables/usePos.js`; `views/SalesHistory.vue` (customer line and reprint); `lib/receipt.js`; `lib/phone.js` (`showPhone`) |
| Entry points | `components/Sidebar.vue` (Sales section, shown with `customer:view`); `router.js` route `/customers` |
| Routes | `backend/adapter/http/router/customers.go` `InitCustomerRouter`, mounted at `/customers/` in `main.go`; the Vite dev proxy lists `/customers` in `frontend/vite.config.js` |
| Controller | `backend/adapter/http/controller/customers.go` |
| Service | `backend/service/customers.go` (`CustomerService` in `interfaces.go`); sales: `CreateSale` and `FetchSalesHistory` in `service/sales.go`; the shared `runInTx` and `appendAudit` are in `service/helpers.go` |
| Phone rules | `backend/model/phone.go` `NormalizePhone` |
| Repository | `backend/repository/customers.go` (`CustomerRepository` in `repository.go`) |
| Types | `backend/internal/types/customer.go` |
| Migrations | `000047_customers` (table, indexes, `Customer/View` and `Customer/Manage` permissions and role grants); `000049_sales_customer` (`sales.customer_id`, index `sales_customer_idx`) |
| Demo data | `backend/internal/demoseed/demoseed.go` `customers` |

## Actions

### Add a customer

| Step | Where | Detail |
|---|---|---|
| Starts | `Customers.vue` `save()` | Requires a name; sends `{version: 0, name, phone, email, notes, contact_consent}`; `saving` blocks a second click |
| Request | `POST /customers/` | Body as above, size-limited (`decodeLimitedJSON`); no retry key |
| Gate | `AuthMiddleware` → `AddPermissionsToContext` → `RequirePermissions(all, customer:manage)` | Service re-checks `customer:manage` |
| Service | `CreateCustomer` | `tidyCustomerRequest` trims, validates (including refusing NUL characters, which Postgres rejects), normalizes the phone; in one transaction (`runInTx`): `requirePhoneFree`, `CreateCustomerTx`, audit `created` |
| Data | `customer`, `audit_log` | `created_by_id` is the signed-in user (session context, never the body); `consent_updated_at` is stamped by SQL (`NOW()`) when consent is given at creation |
| Response | `201 {id}` | The page closes the panel, shows a notice and reloads the list |

### Edit a customer

| Step | Where | Detail |
|---|---|---|
| Starts | `Customers.vue` `open()`, `save()` | A click anywhere on the row opens it; if the panel holds unsaved typing a "Discard your changes?" confirm comes first (also for Add customer and Close); the form is filled from the list row; the phone is shown as `0803 123 4567` and normalized again on save |
| Request | `PUT /customers/:id` | The path id wins over any `id` in the body; `version` is the one the page loaded |
| Gate | as above | `customer:manage` |
| Service | `UpdateCustomer` | `lockCustomer` reads the row `FOR UPDATE` and refuses a stale version (409); compares name, phone, email, notes, consent; nothing changed means no write, no version bump and no audit entry; a changed phone goes through `requirePhoneFree`; `UpdateCustomerTx` stamps `consent_updated_at` with the database's `NOW()` only when consent changes (the app never sets it); audit `updated` with `{changed: [...]}` and `contact_consent: {from, to}` when consent changed |
| Data | `customer` (`version = version + 1`, `updated_at`), `audit_log` | The audit entry holds field names, not the old or new phone, email or notes |
| Response | `204` | On a stale-version 409 the page fetches `GET /customers/:id`, takes its `version`, and keeps what the person typed, with a message saying that saving again replaces the other person's change |

### Attach a customer to a sale

| Step | Where | Detail |
|---|---|---|
| Starts | `CustomerPicker.vue` (in `CartPanel.vue`) | Hidden without `customer:view` unless a customer is already on the sale. Search starts at 2 characters, 250 ms debounce, `GET /customers/?q=`; Enter takes the first result; "Add … as a new customer" needs `customer:manage` and `POST /customers/` (on `phone_taken` it shows the message and searches that phone) |
| State | `usePos.js` `customer` | `null` or `{id, name, phone}`; cleared by `clearCart`; a held sale stores only `{id}` (the held-sale list is readable by every signed-in user); `restoreHeld` ignores an older typed-text value, restores the customer only for users with `customer:view` (otherwise it is kept in `hiddenCustomer`, not shown or sold to, shown as a note in `CartPanel.vue`, and written back if the sale is held again), and refreshes the name and phone from `GET /customers/:id`. F5 and F6 call the handlers `PosView.vue` registers with `setShortcutActions`, the same ones the buttons use (one request at a time, the server's reason shown, nothing for an empty or unpaid cart) |
| Request | `POST /sales/` | `completeSale` adds `customer_id` (omitted for a walk-in); 400, 403 and 409 replies pass the server's message and `code` on; on `already_saved` `PosView.vue` clears the cart (the sale is saved) and shows the message. Enter in the picker takes a match only when it is the only one |
| Gate | `AuthMiddleware` → `AddPermissionsToContext` | Service: `customer_id > 0` needs `customer:view` (403); a negative id is 400 |
| Service | `CreateSale` | Inserts the sale with `customer_id` (an unknown id is the foreign key failing, mapped to `ErrInvalidReference`, 400). Only when the sale is new, `GetCustomerTx` must find an active customer, or 409 with `code: "customer_inactive"` and the transaction rolls back. A retry of a saved sale (idempotency key) returns before this check: the same customer (or still a walk-in) gets the first save, a different one gets 409 `code: "already_saved"` so the receipt can't disagree with the record |
| Data | `sales.customer_id` (the customer's row is read `FOR SHARE`, so a deactivation waits for, or is seen by, the sale) | NULL for walk-ins and every sale made before migration 49; no audit entry (sales are not audited) |
| Response | `{receipt_number}` | Unchanged |

### Show the customer on history and receipts

| Step | Where | Detail |
|---|---|---|
| Request | `GET /sales/filter` | `FetchSalesHistory` selects `customer_id` and, only for callers with `customer:view`, makes one `FetchCustomerNames` lookup for the page |
| Response | `SaleResponse.customer` | `{id, name}` or null; deactivated customers are still named; the name is the current one |
| Screens | `SalesHistory.vue` | Customer line in the sale details; `receiptFor` passes the name to `receipt.js`, which prints `Customer: <name>` (the phone is never printed). The POS receipt does the same from the picked customer |

### Deactivate and reactivate

| Step | Where | Detail |
|---|---|---|
| Starts | `Customers.vue` `toggleActive()` | `confirm()` dialog whose `run` makes the request and shows errors inside the dialog |
| Request | `POST /customers/:id/deactivate` or `/reactivate` | `{version}` |
| Service | `SetCustomerActive` | Same lock and version check; a request that would not change the state does nothing; otherwise `SetCustomerActiveTx` (bumps `version`) and audit `deactivated` / `reactivated` |
| Response | `204` | Panel closes, list reloads |

### List and search

| Step | Where | Detail |
|---|---|---|
| Starts | `Customers.vue` `load()` | Search box is debounced 250 ms; changing the search or "Show deactivated" returns to page 1; a counter drops replies that arrive after a newer request |
| Request | `GET /customers/?q=&page=&inactive=1` | `GET /customers/:id` reads one |
| Gate | `RequirePermissions(all, customer:view)` | Service re-checks |
| Service | `ListCustomers`, `customerFilter` | Text with a letter searches names. Otherwise it is a phone search: a number written from its start with a leading `0` or `+` is anchored with `PhonePrefix` `+234` (or `+` for another country) so `0803` matches only numbers that start `+234803`; any other digits, including ones that begin `234`, match anywhere in the number so the end of a number is always findable; no digits means no filter Pages of 25 (`customersPerPage`); a page below 1 or above `maxCustomerPage` becomes 1 |
| Data | `ListCustomers` in `repository/customers.go` | `lower(name) LIKE` with `%`, `_` and `\` escaped (`likeEscaper`), or `phone LIKE prefix || digits || '%'` / `'%' || digits || '%'`; ordered by name; count and page are separate queries; deactivated customers only with `inactive` |

## Permissions

| Key | Default roles | Checked in |
|---|---|---|
| `customer:view` | Cashier, Manager, Admin (migration 47, by role name) | Routes (GET), `requireCustomerViewer` in the service, sidebar link and page; `CreateSale` (attach a customer) and `FetchSalesHistory` (see the customer); the picker |
| `customer:manage` | Cashier, Manager, Admin | Routes (POST, PUT), `requireCustomerManager`, page controls; "add a new customer" in the picker |

Keys come from the `permissions` rows `Customer/View` and `Customer/Manage`, lowercased at login. People must sign in again to receive them. Constants are in `internal/constant/permissions.go`.

## Data model

| Table | Role in this feature | Notes |
|---|---|---|
| `customer` | One row per person | `phone` is stored as `+234…` or `+<country code>…`, NULL when none; unique index `customer_phone_unique` (`WHERE phone IS NOT NULL`) includes deactivated customers; `name` check (`customer_name_check`) refuses blank; `version` starts at 1; no deletes |
| `sales` | `customer_id BIGINT REFERENCES customer(id)`, nullable | The customer on a sale; the foreign key means a customer with sales can never be deleted (they are only deactivated) |
| `audit_log` | Append-only history (`entity_type = 'customer'`) | Actions `created`, `updated`, `deactivated`, `reactivated`; written in the same transaction |

## Invariants and how they're enforced

- A phone number belongs to one customer — partial unique index; the service checks first to name the owner, and the index decides a race.
- A stale copy never overwrites a newer edit — `version` compared under `FOR UPDATE`.
- Every change has an audit entry — written inside the same transaction as the change (`appendAudit`).
- Phones are in one form — `NormalizePhone` in the service before any query or write. Nigerian numbers are 10 digits after `+234` (11 with the written 0), other countries 8 to 15 digits after `+`.
- No customer is deleted — there is no delete route or repository method.

## Concurrency and retries

- Two people adding the same phone: one insert wins; the other gets 409 with `code: "phone_taken"` (`TestTwoRequestsForTheSamePhoneSaveOneCustomer`).
- Two people editing one customer: the second gets 409; the page keeps their typing, takes the new version, and lets them choose to replace or close. After the last customer on the last page is deactivated, the page moves to the new last page.
- Double click: the form's `saving` flag blocks it. A customer without a phone has no unique key, and the request has no retry key, so a retry after a lost response can add a second customer; the page says to check the list first — T-122.

## Configuration

| Value | Where | Effect |
|---|---|---|
| 25 | `service/customers.go` `customersPerPage` | Customers per page |
| 120 / 254 / 1000 / 100 | `maxCustomerName`, `maxCustomerEmail`, `maxCustomerNotes`, `maxCustomerQuery` | Name, email, notes and search lengths |
| 250 ms | `Customers.vue` search debounce | Delay before a search runs |

## Tests

- `model/phone_test.go`: every written form of a Nigerian number, refusals, foreign numbers.
- `service/customers_test.go`: mock repository; permissions, validation, duplicate phone, version conflict, audit entries, search filter, paging.
- `service/customers_integration_test.go` (needs `PHARMACY_TEST_DSN`): the real SQL, anchored and anywhere phone search, `LIKE` escaping, the database-stamped consent time, paging, concurrent creates, audit order.
- `db/migrations_test.go` `TestCustomerMigration` (permissions by role, unique and blank-name rules, rollback) and `TestSalesCustomerMigration` (nullable column, foreign key, index, rollback and re-apply).
- `service/sale_customer_test.go`: walk-in, stored customer, permission, negative and unknown ids, deactivated customer rolled back, retry skips the check, history names the customer only with `customer:view`. `service/sale_customer_integration_test.go`: the same against the real database.
- `adapter/http/router/customers_test.go`: each route's permission and what reaches the service.
- `internal/demoseed`: the demo shop includes customers, and five demo sales have one.
- Not covered: the Vue pages and the picker (the repo has no JS tests); they were checked by hand in a browser on demo data.

## Known gaps

- The customer on a saved sale cannot be changed or removed — T-124.
- A customer's own history and totals, and a Sales History filter by customer — T-119. Import and merge — T-120.
- Sales have no request fingerprint: a retry is compared on its customer only, so other differences (items, payments) still get the first save as before.
- No way to erase or anonymize a customer on request, or to export what is stored about them — T-121.
- The audit trail has no screen; it can only be read from the database.
- Creating a customer has no retry key (see Concurrency and retries) — T-122.

## History

- T-118: `sales.customer_id`, the customer picker at the till, the customer on Sales History, reprints and receipts. Plan: `docs/plans/T-118-attach-customer-to-sale.md`.
- T-040: the customer table, permissions, service, routes and Customers page. Split from the original card, which also covered the till link, history, import and merge. Plan: `docs/plans/T-040-customer-profiles.md`.
