# Board

Kanban board for the Pharmacy POS. Finished cards live in [DONE.md](DONE.md).

**In Progress** → **Up Next** → **Backlog** · **Done** → [DONE.md](DONE.md)

<!-- next-id: T-094 -->

<details>
<summary>How this board works</summary>

- Each `##` heading is a column; each `###` heading is a card. A card's position in its column is
  its rank: the top card is the next one to pick up.
- Move a card by moving its whole section (heading to the line before the next `###`).
- When a card is finished, move it to the top of the current month in [DONE.md](DONE.md), drop
  `priority` and `depends`, and add `done:` (YYYY-MM-DD) and `prs:`.
- New cards take the id in the `next-id` comment above; bump it afterwards. Ids are never reused.
- Keep In Progress small (one or two cards).

Card format (fields are `- key: value` lines directly under the heading, then a blank line, then
free Markdown; checklists use `- [ ]`):

```markdown
### T-000 Short imperative title

- priority: P0 | P1 | P2
- area: pos, inventory, purchasing, pharmacy, reports, admin, saas, security, ops
- depends: T-000, T-000   (optional)
- created: YYYY-MM-DD

What and why, with file or function references. Optional checklist:

- [ ] step
```

Priorities: **P0** blocks taking on a second customer · **P1** needed to compete · **P2** nice to have.

</details>

## In Progress

_Nothing in progress. Pull the top card from Up Next._

## Up Next

### T-017 Release purchase orders to production

- priority: P0
- area: ops
- created: 2026-10-26

PRs #28 and #29 are merged; production is at migration 39.

- [ ] Run migrations 40–43 against a copy of prod data and check the suppliers the backfill creates
- [ ] Check prod's roles are named Cashier, Manager and Admin (migration 43 grants by name, see T-092)
- [ ] Click through what #29's Playwright checks skipped: the admin category and permission delete dialogs, and the sales page "Discard and continue" path
- [ ] `make assets`, build the image, deploy, have every user log in again to pick up the new permissions

### T-018 Lock down public user registration

- priority: P0
- area: security
- created: 2026-10-26

`POST /user/register` is mounted without auth (`router/user.go`, `main.go`): anyone who can reach the server can create a Cashier account. Remove it or restrict to admins / invitations.

### T-019 Authenticate the /ws WebSocket

- priority: P0
- area: security
- created: 2026-10-26

`/ws` bypasses the middleware stack and `HandleWebSocket` never checks the session, so anyone can subscribe to stock-taking updates. Check the session (and tenant) before upgrading.

### T-020 Validate sale prices server-side

- priority: P0
- area: security
- created: 2026-10-26

`CreateSale` stores the client-sent `unit_price`, `discount`, `total` and `subtotal` as-is. Re-price from `product_price`, recompute totals, and enforce discount rules on the server.

### T-021 Fix money conversion rounding

- priority: P0
- area: security
- created: 2026-10-26

Amounts arrive as float naira and are converted with `int(x * 100)`, which truncates (e.g. 0.29 → 28 kobo). Send kobo integers from the client, or round explicitly.

### T-022 Prevent negative stock on sale

- priority: P0
- area: inventory
- created: 2026-10-26

`CreateSale` posts `OUT_SALE` without checking available stock. Block or require an override permission when a sale exceeds stock.

### T-023 Re-check permissions on each request

- priority: P0
- area: security
- created: 2026-10-26

Permissions are cached in the session cookie at login (`RequirePermissions` reads `session.Values["permissions"]`), so revoking a role or permission only takes effect after re-login or the 3am expiry. Load from DB (cached briefly) per request; add a way to end a user's sessions.

### T-024 Login rate limiting & cookie hardening

- priority: P0
- area: security
- created: 2026-10-26

No throttling or lockout on failed logins. Session cookie lacks `Secure`; review `SameSite`. Add HTTPS-only settings and security headers.

### T-025 CI pipeline

- priority: P0
- area: ops
- created: 2026-10-26

GitHub Actions: `go test ./...` with Postgres (integration tests), `gofmt`/golines checks, `go vet`, frontend build, Docker build.

### T-026 Automated backups & restore

- priority: P0
- area: ops
- created: 2026-10-26

Scheduled Postgres backups with point-in-time recovery, off-site storage, tested restore procedure, per-tenant data export.

### T-027 Health check, graceful shutdown & observability

- priority: P0
- area: ops
- created: 2026-10-26

No `/healthz`, `server.ListenAndServe()` with no graceful shutdown or timeouts, logs only to a local `app.log`. Add health/readiness endpoints, structured logging to stdout, error tracking (e.g. Sentry), uptime monitoring and metrics.

### T-028 Runtime configuration

- priority: P0
- area: saas
- created: 2026-10-26

Config comes from `config/.env` embedded at build time (`//go:embed .env`), so secrets and settings are baked into the binary. Read from real environment variables or a secrets manager at runtime; one image for every environment.

### T-029 Multi-tenancy architecture

- priority: P0
- area: saas
- created: 2026-10-26

Every table is global today; one deployment = one store. Decide the isolation model (shared schema with `tenant_id` + row-level security, schema-per-tenant, or DB-per-tenant), then add tenant scoping to all queries, sessions and the WebSocket hub, and tenant resolution (subdomain or custom domain).

### T-030 Store settings & white-label branding

- priority: P0
- area: saas
- depends: T-029
- created: 2026-10-26

"Primocrest Pharmacy" is hardcoded in the receipt (`composables/usePos.js`), `Sidebar.vue`, `LoginPage.vue`, `NotFoundPage.vue` and the `<title>` of `login-v2.html`, `next-dashboard.html` and `error.html`. Add per-store settings and use them everywhere:

- [ ] Name, logo or icon (replacing the text-only brand mark)
- [ ] Receipt header and footer: address, phone
- [ ] Primary colour (brand tokens are in the Tailwind theme and `design-system/primocrest-pharmacy/MASTER.md`; status colours stay fixed so green, amber and red keep their meaning)
- [ ] Timezone (the dashboard uses server local time) and currency

Stored with the other settings from T-093.

### T-031 Bulk product import (CSV/Excel)

- priority: P0
- area: inventory
- created: 2026-10-26

Import products, prices, price options and opening stock with expiry from a spreadsheet, with validation preview and error report. The biggest onboarding blocker for a new store.

### T-032 Tenant signup & onboarding wizard

- priority: P0
- area: saas
- depends: T-029, T-031
- created: 2026-10-26

Self-serve signup that creates a tenant and its first admin, then a guided setup: store details, categories, products (import), users, opening stock.

### T-033 Subscription billing & plans

- priority: P0
- area: saas
- depends: T-029
- created: 2026-10-26

Plans, trials, recurring billing (Paystack/Flutterwave for NGN), dunning, plan limits (users, products, branches) and feature gating.

### T-034 Super-admin console

- priority: P0
- area: saas
- depends: T-029
- created: 2026-10-26

Platform operator view: tenants, plans and billing status, usage, suspend/reactivate, support impersonation (audited).

## Backlog

### T-035 Extend audit log to all sensitive actions

- priority: P1
- area: security
- created: 2026-10-26

Extend `audit_log` beyond purchase orders: sales and returns, stock movements and adjustments, stock taking, price and cost edits, receiving, user/role/permission changes. Shared `auditEntry` helper in the service layer, written in the same transaction. Admin screen to browse and filter.

### T-036 Capture batch/lot numbers on receiving

- priority: P1
- area: inventory
- created: 2026-10-26

`product_batch.batch_no` exists but is never written. Capture it on receiving and allow several batches (different expiry dates) per purchase order line. Prerequisite for batch-level stock / FEFO.

### T-037 Costs on purchase orders

- priority: P1
- area: purchasing
- created: 2026-10-26

Expected cost per line, order total, and comparison with the received cost.

### T-038 Supplier management screen

- priority: P1
- area: purchasing
- created: 2026-10-26

Rename and merge duplicate suppliers, contact details (phone, email, address), payment terms.

### T-088 Block supplier changes on orders with receipts

- priority: P1
- area: purchasing
- created: 2026-10-27

Changing or clearing the supplier on an order that already has receipts is allowed silently, and the receipts stay with the old supplier. Block it and point to Move outstanding.

### T-093 Settings screen for hard-coded values

- priority: P1
- area: admin
- created: 2026-10-27

Hard-coded values should become settings that an admin edits on a settings screen. Store them in the database, read them once per session and serve them to the SPA, so the frontend and backend share one value. Store branding (T-030) uses the same store.

- [ ] Large purchase order confirmation: 100,000 units per line or more than 200 lines (`CONFIRM_QUANTITY`, `CONFIRM_LINES` in `views/PurchaseOrders.vue`)
- [ ] Near-expiry alert on Receive Items: under 183 days turns amber and needs "I confirm the near-expiry dates" (`expiryClass`, `expiryTitle`, `hasNearExpiry` in `views/ReceiveItems.vue`); past dates stay red
- [ ] Expiring-soon window on the dashboard and expiring list: 90 days (`INTERVAL '90 days'` in `repository/dashboard.go`); decide whether it shares a value with the Receive Items alert
- [ ] Approval badge poll interval: 60 seconds (`usePendingPurchaseOrders.js`)
- [ ] Completed purchase orders cap: 100 (`purchaseOrderHistoryMax`)

### T-040 Customer records & purchase history

- priority: P1
- area: pos
- created: 2026-10-26

Customer name is typed on the receipt but never stored. Add customers (name, phone), attach to sales, view purchase history, search at the POS.

### T-041 Credit sales & debtor balances

- priority: P1
- area: pos
- depends: T-040
- created: 2026-10-26

Sell on account, track customer balances, record part-payments, statements and aged debt. Very common in Nigerian retail.

### T-042 Void / cancel a completed sale

- priority: P1
- area: pos
- created: 2026-10-26

No way to void a mistaken sale. Add void with reason, permission and audit entry; reverse stock movements and payments.

### T-043 Record refund method on returns

- priority: P1
- area: pos
- created: 2026-10-26

Returns store `total_refunded` but not how the money went back (cash, transfer, store credit). Needed for cash reconciliation.

### T-044 Shifts, till sessions & end-of-day cash-up

- priority: P1
- area: pos
- created: 2026-10-26

Open a till with a float, attach sales to a shift, close with counted cash per method, over/short, Z-report.

### T-045 Tax / VAT support

- priority: P1
- area: pos
- created: 2026-10-26

Configurable tax rates (e.g. 7.5% VAT), tax-inclusive/exclusive pricing, exempt products, tax on receipts and in reports.

### T-046 Discount limits, price overrides & promotions

- priority: P1
- area: pos
- created: 2026-10-26

Max discount per role, price override requiring permission, promotions (percentage, buy-X-get-Y, time-bound).

### T-047 Offline-capable POS

- priority: P1
- area: pos
- created: 2026-10-26

The POS stops working when the network or server drops. PWA with a local product cache and a queue of sales synced later (idempotency keys already exist).

### T-048 Thermal printing & digital receipts

- priority: P1
- area: pos
- created: 2026-10-26

ESC/POS thermal printer support (58/80mm), reprint from sales history, receipts by WhatsApp/SMS/email.

### T-049 Batch-level stock & FEFO

- priority: P1
- area: inventory
- depends: T-036
- created: 2026-10-26

Stock is computed per product from `stock_movement`; sales don't deduct from a batch, and expiry is the product's earliest date. Track quantity per batch and deduct first-expiry-first-out, so you know how much of each batch is left and what actually expires. Depends on capturing batch numbers.

### T-050 Stock adjustments & write-offs with reason codes

- priority: P1
- area: inventory
- created: 2026-10-26

Adjustments only happen through product edit. Add a dedicated adjustment flow with reasons (expired, damaged, theft, correction), permission and audit, and new movement types for write-offs.

### T-051 Multi-location stock & transfers

- priority: P1
- area: inventory
- created: 2026-10-26

Branches/locations (and dispensary vs store, today only in stock-taking counts) with stock per location and transfer documents between them.

### T-052 Product stock ledger (bin card)

- priority: P1
- area: inventory
- created: 2026-10-26

Per-product movement history: every receipt, sale, return, adjustment and stock-take with running balance, user and reference.

### T-053 Reorder suggestions to draft purchase order

- priority: P1
- area: purchasing
- created: 2026-10-26

Suggest reorder quantities from reorder level and sales velocity, and generate a draft purchase order per supplier.

### T-054 Returns to supplier

- priority: P1
- area: purchasing
- created: 2026-10-26

Send expired, damaged or excess stock back to a supplier with a document, stock movement and expected credit.

### T-055 Supplier invoices & payables

- priority: P1
- area: purchasing
- created: 2026-10-26

Record supplier invoices against receipts, amounts owed, payments made and balances per supplier.

### T-056 Product archive & bulk price update

- priority: P1
- area: inventory
- created: 2026-10-26

Deactivate products without losing history; update prices in bulk (by category or percentage).

### T-057 Pharmacy product attributes

- priority: P1
- area: pharmacy
- created: 2026-10-26

Generic name / active ingredient, strength, dosage form, pack size, NAFDAC number, classification (OTC / POM / controlled).

### T-058 Prescriptions & dispensing records

- priority: P1
- area: pharmacy
- depends: T-057
- created: 2026-10-26

Record prescriptions (prescriber, patient, items, directions), link to sales, pharmacist sign-off required for POM items.

### T-059 Controlled drugs register

- priority: P1
- area: pharmacy
- depends: T-057
- created: 2026-10-26

Running register of controlled substances: receipts, dispensing, balances, who did it, for inspections.

### T-060 Profit & margin reports

- priority: P1
- area: reports
- created: 2026-10-26

Gross profit by period, product and category using batch cost (already captured).

### T-061 Sales breakdown & end-of-day reports

- priority: P1
- area: reports
- created: 2026-10-26

Sales by cashier, product, category, payment method and hour; end-of-day summary.

### T-062 Stock valuation, dead stock & write-off reports

- priority: P1
- area: reports
- depends: T-050
- created: 2026-10-26

Stock value at cost and retail, slow and dead movers, value of expired / written-off stock.

### T-063 Export every report to CSV/Excel

- priority: P1
- area: reports
- created: 2026-10-26

Consistent CSV/Excel export across reports and lists (and relink the existing stock CSV export).

### T-064 Admin user management

- priority: P1
- area: admin
- created: 2026-10-26

Admins can't create or deactivate users from the UI, and users can't change their own password. Add invite/create, deactivate, change password, last login.

### T-065 Release & deployment pipeline

- priority: P1
- area: ops
- created: 2026-10-26

Staging environment, versioned releases, migrations run as part of deploy, pinned images (compose uses `postgres:latest`).

### T-066 Stock balance performance

- priority: P1
- area: ops
- created: 2026-10-26

`inventory_view` sums all of `stock_movement` on every read, which slows down with volume and tenants. Maintain a stock balance table (or materialised view) updated in the same transaction.

### T-067 Idempotency key for creating a purchase order

- priority: P1
- area: purchasing
- created: 2026-10-26

The Save button is disabled while saving, but a retry after a lost response can still create a duplicate order.

### T-068 Change-due modal after saving a sale

- priority: P2
- area: pos
- created: 2026-10-26

Show a modal with the change to give when change > 0 after saving a sale.

### T-069 Purchase order history search & filters

- priority: P2
- area: purchasing
- created: 2026-10-26

Search and filter by supplier, status and date, with pagination; the Completed tab is capped at the newest 100.

### T-070 Show on-order quantities

- priority: P2
- area: purchasing
- created: 2026-10-26

Show quantities on open purchase orders on the Products list and Low Stock page.

### T-071 Decide permission for receiving against a PO

- priority: P2
- area: purchasing
- created: 2026-10-26

Today anyone who can receive can receive against a purchase order. Decide whether it needs its own permission.

### T-089 Withdraw, send back and notify on approvals

- priority: P2
- area: purchasing
- created: 2026-10-27

Let a drafter withdraw a submitted order back to draft, let a manager send it back for changes instead of only rejecting, and notify the drafter when their order is approved or rejected (today they only see it on their list).

### T-090 Show rejected drafts as Rejected

- priority: P2
- area: purchasing
- created: 2026-10-27

A rejected draft shows as "Closed", with the reason only in the activity log. Show "Rejected" and the reason on the order itself.

### T-091 Real notifications for approvers

- priority: P2
- area: purchasing
- created: 2026-10-27

Replace the sidebar badge and dashboard banner, which poll once a minute, with a stored inbox and push or email. Could share delivery with T-078.

### T-072 E2E tests for purchase orders & receive-against-order

- priority: P2
- area: ops
- created: 2026-10-26

Committed browser/end-to-end tests for the purchase order screens, the approval flow and the receive-against-order flow. PR #29 ran Playwright checks by hand; none are in the repo.

### T-092 Role-independent permission seeding

- priority: P2
- area: ops
- created: 2026-10-27

Migration 000043 grants permissions by role name (Cashier, Manager, Admin). Warn, or grant to a configured role, when a deployment's roles are named differently. Matters once tenants can rename roles.

### T-073 Barcode label printing

- priority: P2
- area: inventory
- created: 2026-10-26

Generate and print barcode/shelf labels for products without manufacturer barcodes.

### T-074 Generic substitution search

- priority: P2
- area: pharmacy
- depends: T-057
- created: 2026-10-26

Search by active ingredient to offer equivalent products when the requested brand is out of stock.

### T-075 Cashier PIN switching & admin 2FA

- priority: P2
- area: security
- created: 2026-10-26

Fast PIN-based user switching on a shared till; TOTP 2FA for admin accounts.

### T-076 Frontend lint & tests

- priority: P2
- area: ops
- created: 2026-10-26

ESLint/Prettier, component tests and E2E coverage for POS, receiving and stock taking. There is no JS tooling today.

### T-077 API documentation (OpenAPI)

- priority: P2
- area: ops
- created: 2026-10-26

Document the JSON API; groundwork for a public API.

### T-078 Low-stock & expiry alerts and scheduled reports

- priority: P2
- area: reports
- created: 2026-10-26

Email/WhatsApp alerts for low stock and upcoming expiries; daily summary to owners.

### T-079 Expense tracking & accounting export

- priority: P2
- area: reports
- created: 2026-10-26

Record shop expenses; export sales, purchases and expenses to accounting tools or Excel.

### T-080 Multi-branch consolidated analytics

- priority: P2
- area: reports
- depends: T-051
- created: 2026-10-26

Owner view across branches: sales, stock, transfers, branch comparison.

### T-081 Public API, webhooks & integrations

- priority: P2
- area: saas
- depends: T-077
- created: 2026-10-26

Tenant API keys, webhooks (sale created, low stock), integrations with e-commerce and delivery platforms.

### T-082 Mobile owner app / responsive views

- priority: P2
- area: saas
- created: 2026-10-26

Owners check sales and stock from a phone; responsive layouts or a lightweight mobile app.

### T-083 Localisation & multi-currency

- priority: P2
- area: saas
- created: 2026-10-26

Currency and formatting per tenant (money is kobo-only today), translations.

### T-084 Customer loyalty program

- priority: P2
- area: pos
- depends: T-040
- created: 2026-10-26

Points or discounts for repeat customers. Depends on customer records.

### T-085 Shared purchase order lock-and-check helper

- priority: P2
- area: purchasing
- created: 2026-10-26

Edit, close, move and receive each lock the order and check version and status. Share one helper so the four copies can't drift apart.

### T-086 Permissions on purchase order read endpoints

- priority: P2
- area: security
- created: 2026-10-26

Decide whether list, get and history need a permission.
