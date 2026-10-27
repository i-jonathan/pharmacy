# Board

Kanban board for the Pharmacy POS. Finished cards live in [DONE.md](DONE.md).

**In Progress** → **Up Next** → **Backlog** · **Done** → [DONE.md](DONE.md)

<!-- next-id: T-104 -->

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

There are no backups today. Start with the single store, then grow into the hosted version.

- [ ] Daily `pg_dump` on the store machine, kept for a set number of days, compressed
- [ ] Copy each backup automatically to a fixed location set in config (shared folder, external drive, Google Drive / Dropbox / S3)
- [ ] Share a backup on demand by any available means (download from the admin screen, OS share sheet, email or WhatsApp the file)
- [ ] Last-backup status and a warning in the admin screen when a backup fails or is overdue
- [ ] Documented, tested restore
- [ ] Hosted: point-in-time recovery, off-site copies, per-tenant data export

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

### T-047 Online-first with solid offline operation

- priority: P0
- area: pos
- created: 2026-10-26

Run against a hosted server, but losing the connection must never stall work. Today every screen stops when the network or server drops.

- [ ] Local cache of products, prices and stock, refreshed in the background
- [ ] Sales, returns, receiving and stock counts saved locally and queued while offline (idempotency keys already exist for sales and receiving)
- [ ] Sync as soon as the connection returns, with retry and a visible pending count
- [ ] Clear online/offline indicator; reprint and hold work offline
- [ ] Conflict rules: stock going negative across devices, prices changed while offline, a held sale resumed on two tills
- [ ] Decide the offline store (service worker + IndexedDB, or the desktop app in T-094)

## Backlog

### T-035 Audit log everything

- priority: P1
- area: security
- created: 2026-10-26

Every change writes an `audit_log` entry in the same transaction: who, when, what, before and after, and a reason where it matters. Only purchase orders do this today. Add a shared `auditEntry` helper in the service layer.

- [ ] Sales, voids, returns and refunds
- [ ] Stock movements: receiving, manual adjustments, stock-take completion
- [ ] Product, price and cost edits; categories, groups and merges
- [ ] Users, roles and permissions; logins, logouts and failed logins
- [ ] Settings changes and data exports
- [ ] Admin screen to browse, filter (user, entity, action, date) and export the log

### T-099 Granular permissions with default roles

- priority: P1
- area: security
- depends: T-023
- created: 2026-10-27

Everything is behind a permission. The system ships with default roles that have permissions assigned, and every permission has an explanation.

- [ ] Permission catalog in code: key, name, description, group (Sales, Inventory, Purchasing, Reports, Admin…)
- [ ] Every route and every UI action checks a permission. Today many are ungated: creating products (`POST /inventory/add-item`), receiving, sales, returns, stock-take create and count, and most read endpoints
- [ ] Default roles seeded by migration and kept in sync with the catalog (e.g. Owner, Manager, Pharmacist, Cashier, Storekeeper)
- [ ] Admin screen lists permissions grouped, with their explanations, and lets admins build custom roles from the catalog
- [ ] Remove the free-form create/delete permission endpoints; permissions are defined by code
- [ ] Works with per-request checks (T-023) and role-independent seeding (T-092)

### T-103 Notification centre

- priority: P1
- area: admin
- created: 2026-10-27

One place for everything the system needs to tell people.

- [ ] Stored in-app inbox with a bell and unread count, per user
- [ ] Events: purchase orders submitted, approved or rejected (T-091), low stock and expiry (T-078), backup failed (T-026), offline sync problems (T-047), what's new (T-095), insights (T-100)
- [ ] Channels: in-app, email, WhatsApp/SMS, push (desktop and mobile)
- [ ] Per-user and per-role preferences, and a daily digest option

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

### T-040 Customer management

- priority: P1
- area: pos
- created: 2026-10-26

The customer name is typed on the receipt but never stored.

- [ ] Customer profiles: name, phone, email, notes, consent to be contacted
- [ ] Quick lookup by phone at the POS and attach to the sale
- [ ] Purchase history and totals per customer
- [ ] Import, edit and merge duplicate customers
- [ ] Balances and credit (T-041), loyalty (T-084)

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

### T-101 Keyboard workflow for POS and Receive Items

- priority: P1
- area: pos
- created: 2026-10-27

Make the two busiest screens fully usable without a mouse.

- [ ] POS: focus search (`/` or F2), arrow keys + Enter to add, change quantity and price option, remove a line, discount, choose a payment method, save, save and print, hold
- [ ] Receive Items: Enter moves to the next field, a new row at the end, quick expiry entry (MM/YY), save
- [ ] Barcode scanner input works from anywhere on both screens
- [ ] A `?` overlay listing the shortcuts; the same keys mean the same thing on both screens

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

### T-050 Manual stock management

- priority: P1
- area: inventory
- created: 2026-10-26

Adjust stock directly, without starting a stock take. Today the only way is the stock field on the product edit form.

- [ ] Adjust one product or many at once, by setting a new count or adding/removing a quantity
- [ ] Reason codes: expired, damaged, theft/loss, correction, opening stock, internal use; new movement types for write-offs
- [ ] Set or correct expiry and batch while adjusting
- [ ] Own permission, plus approval above a configurable quantity or value
- [ ] Audit entry for every adjustment; route the product edit stock field through this flow, or remove it

### T-051 Multi-store & multi-branch support

- priority: P1
- area: inventory
- created: 2026-10-26

One business owns one or more stores, and each store has one or more branches. Dispensary and store room become locations inside a branch (today they exist only as stock-take counts).

- [ ] Stock, batches and stock takes per branch and location
- [ ] Prices shared across branches, with optional branch overrides
- [ ] Users and permissions per branch, and a branch switcher
- [ ] Transfers between branches, with an in-transit state and receipt at the other end
- [ ] Reports per branch plus a consolidated view (T-080)

### T-052 Product history

- priority: P1
- area: inventory
- created: 2026-10-26

One screen per product that shows its whole history:

- [ ] Stock ledger: every receipt, sale, return, adjustment and stock take, with running balance, user and reference
- [ ] Purchase cost history per receipt and supplier (already stored on `product_batch`)
- [ ] Selling price history: prices are updated in place today (`UpdateProduct`), so start recording price changes with who and when
- [ ] Restock dates and quantities, and time between restocks
- [ ] Simple charts for cost, price and stock over time

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

### T-097 Product groups with a group reorder level

- priority: P1
- area: inventory
- created: 2026-10-27

Group interchangeable products, e.g. different brands of amoxicillin 500mg that all sell at ₦600. A group has its own reorder level and total stock, and low stock is judged on the group as well as on each product. Optional shared group price. Groups feed reorder suggestions (T-053) and substitution (T-074).

### T-098 Merge duplicate products, categories and groups

- priority: P1
- area: inventory
- created: 2026-10-27

Pick a surviving record and move everything onto it: stock movements, batches, prices, sales and return lines, purchase order lines, stock-take items. Then remove or archive the duplicate. Preview what moves, require a permission and write an audit entry. Suppliers merge in T-038; groups come from T-097.

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

### T-096 More useful sales and inventory metrics

- priority: P1
- area: reports
- created: 2026-10-27

At-a-glance numbers for the dashboard; the full reports are T-060, T-061 and T-062.

- [ ] Sales: gross margin %, average basket value and items per sale, sales by hour and weekday, returns rate, comparison with the same day last week
- [ ] Inventory: stock value at cost, days of cover per product or group, stock turnover, sell-through since receipt, value at risk of expiry, fast movers next to dead stock

### T-100 Business insights

- priority: P1
- area: reports
- created: 2026-10-27

Go beyond tables: have the system point out what matters. The detailed reports are T-060 to T-062; the metrics are T-096.

- [ ] "Runs out in N days at the current sales rate" for each product or group
- [ ] Margin changes, e.g. the last restock cost went up but the selling price didn't
- [ ] Sales up or down against the same weekday or period last time, by product and category
- [ ] Money tied up in slow movers, and stock value at risk of expiry in the next 30/60/90 days
- [ ] Weekly summary for owners (delivered through T-103)

### T-063 Export data

- priority: P1
- area: reports
- created: 2026-10-26

Get data out in a usable form.

- [ ] CSV/Excel export on every list and report (and relink the existing stock CSV export)
- [ ] Full store export: products, prices, stock and batches, customers, sales, purchases, suppliers, as a zip of CSVs, for accounting, migration or leaving
- [ ] Optional scheduled export to the backup location (T-026)
- [ ] Exports need a permission and are audited (T-035)

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

### T-102 Make the app snappy

- priority: P1
- area: ops
- created: 2026-10-27

General speed work, measured first so it goes where the time actually is.

- [ ] Measure: request timings in logs, Postgres slow query log, JS bundle size per route
- [ ] Stock balance table instead of summing `stock_movement` on every read (T-066)
- [ ] Indexes for the hot queries (product search, sales history, dashboard)
- [ ] Cache the product catalog in the browser so POS search is instant
- [ ] Optimistic updates and prefetching for the common screens; pagination on every long list
- [ ] Compression and long cache headers for the embedded static assets

### T-067 Idempotency key for creating a purchase order

- priority: P1
- area: purchasing
- created: 2026-10-26

The Save button is disabled while saving, but a retry after a lost response can still create a duplicate order.

### T-082 Responsive mobile & tablet support

- priority: P1
- area: saas
- created: 2026-10-26

Every screen works on phones and tablets, not just desktop.

- [ ] POS usable on a tablet; stock-take counting and receiving on a phone
- [ ] Dashboard, approvals and reports readable on a phone for owners
- [ ] Collapsible navigation, touch-sized controls, tables that become cards on small screens
- [ ] Known issue: purchase order card actions wrap below about 1280px

A native mobile app can come later through the PWA / app shell in T-094.

### T-084 Customizable loyalty program

- priority: P1
- area: pos
- depends: T-040, T-093
- created: 2026-10-26

Loyalty rules an admin sets up per store, instead of one fixed scheme.

- [ ] Earn rate (points per ₦ spent), with multipliers or exclusions by category or product (e.g. exclude prescription-only items)
- [ ] Redemption value, minimum to redeem, and points expiry
- [ ] Optional tiers with their own rates, and bonuses (birthday, first purchase, promotions)
- [ ] Points balance on the receipt and at the POS; redeem as a payment method
- [ ] Every earn, redeem and manual change audited

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
- depends: T-103
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
- depends: T-103
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

### T-094 Desktop app

- priority: P2
- area: saas
- created: 2026-10-27

Run as an installed app instead of a browser tab. Options: an installable PWA first, then a native shell (Wails suits the Go backend; Tauri or Electron also work). A desktop app also helps offline storage (T-047), direct thermal printing and local backups (T-026). Needs auto-update.

### T-095 What's new on login

- priority: P2
- area: admin
- created: 2026-10-27

After a release, show each user a short "What's new" panel once after they log in, then remember per user that they've seen it. Release notes live in the repo (they could be drawn from DONE.md) and can be targeted by role.

### T-083 Localisation & multi-currency

- priority: P2
- area: saas
- created: 2026-10-26

Currency and formatting per tenant (money is kobo-only today), translations.

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
