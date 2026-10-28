# Board

Kanban board for the Pharmacy POS. Finished cards live in [DONE.md](DONE.md).

**In Progress** → **Up Next** → **Backlog** · **Done** → [DONE.md](DONE.md)

<!-- next-id: T-107 -->

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
- feature: one of the features listed below
- depends: T-000, T-000   (optional)
- merged: T-000, T-000    (optional: cards folded into this one)
- created: YYYY-MM-DD

What and why, with file or function references. Optional checklist:

- [ ] step
```

Features: Security hardening, Permissions & audit, SaaS platform, Reliability & operations, Offline & devices, Point of sale, Customers & loyalty, Inventory control, Purchasing & receiving, Pharmacy compliance, Reports & insights, Notifications, Admin & settings, Finance & integrations.

Merged cards keep their ids in the survivor's `merged:` field; ids are never reused.

Priorities: **P0** blocks taking on a second customer · **P1** needed to compete · **P2** nice to have.

</details>

## In Progress

_Nothing in progress. Pull the top card from Up Next._

## Up Next

### T-017 Release purchase orders to production

- priority: P0
- feature: Reliability & operations
- created: 2026-10-26

PRs #28 and #29 are merged; per PR #28, production was at migration 39.

- [ ] Run migrations 40–43 against a copy of prod data and check the suppliers the backfill creates
- [ ] Check prod's roles are named Cashier, Manager and Admin (migration 43 grants by name, see T-099)
- [ ] Click through what #29's Playwright checks skipped: the admin category and permission delete dialogs, and the sales page "Discard and continue" path
- [ ] `make assets`, build the image, deploy, have every user log in again to pick up the new permissions

### T-019 Authenticate the /ws WebSocket

- priority: P0
- feature: Security hardening
- created: 2026-10-26

`/ws` bypasses the middleware stack and `HandleWebSocket` never checks the session, so anyone can subscribe to stock-taking updates. Check the session (and tenant) before upgrading.

### T-020 Validate sale prices server-side

- priority: P0
- feature: Security hardening
- created: 2026-10-26

`CreateSale` stores the client-sent `unit_price`, `discount`, `total` and `subtotal` as-is. Re-price from `product_price`, recompute totals, and enforce discount rules on the server.

### T-021 Fix money conversion rounding

- priority: P0
- feature: Security hardening
- created: 2026-10-26

Amounts arrive as float naira and are converted with `int(x * 100)`, which truncates (e.g. 0.29 → 28 kobo). Send kobo integers from the client, or round explicitly.

### T-022 Prevent negative stock on sale

- priority: P0
- feature: Security hardening
- created: 2026-10-26

`CreateSale` posts `OUT_SALE` without checking available stock. Block or require an override permission when a sale exceeds stock.

### T-023 Re-check permissions on each request

- priority: P0
- feature: Permissions & audit
- created: 2026-10-26

Permissions are cached in the session cookie at login (`RequirePermissions` reads `session.Values["permissions"]`), so revoking a role or permission only takes effect after re-login or the 3am expiry. Load from DB (cached briefly) per request; add a way to end a user's sessions.

### T-024 Login rate limiting & cookie hardening

- priority: P0
- feature: Security hardening
- created: 2026-10-26

No throttling or lockout on failed logins. Session cookie lacks `Secure`; review `SameSite`. Add HTTPS-only settings and security headers.

Also cap request body size on every JSON endpoint (one middleware or shared helper). T-018 added `decodeLimitedJSON` (16 KB, 413 when exceeded) to the admin create-user and reset-password handlers only.

### T-025 CI pipeline

- priority: P0
- feature: Reliability & operations
- created: 2026-10-26

GitHub Actions: `go test ./...` with Postgres (integration tests), `gofmt`/golines checks, `go vet`, frontend build, Docker build.

### T-026 Automated backups & restore

- priority: P0
- feature: Reliability & operations
- created: 2026-10-26

Nothing in the repo or the compose setup makes backups. Start with the single store, then grow into the hosted version.

- [ ] Daily `pg_dump` on the store machine, kept for a set number of days, compressed
- [ ] Copy each backup automatically to a fixed location set in config (shared folder, external drive, Google Drive / Dropbox / S3)
- [ ] Share a backup on demand by any available means (download from the admin screen, OS share sheet, email or WhatsApp the file)
- [ ] Last-backup status and a warning in the admin screen when a backup fails or is overdue
- [ ] Documented, tested restore
- [ ] Hosted: point-in-time recovery, off-site copies, per-tenant data export

### T-027 Health check, graceful shutdown & observability

- priority: P0
- feature: Reliability & operations
- created: 2026-10-26

No `/healthz`, `server.ListenAndServe()` with no graceful shutdown or timeouts, logs only to a local `app.log`. Add health/readiness endpoints, structured logging to stdout, error tracking (e.g. Sentry), uptime monitoring and metrics.

### T-028 Runtime configuration

- priority: P0
- feature: SaaS platform
- created: 2026-10-26

Config comes from `config/.env` embedded at build time (`//go:embed .env`), so secrets and settings are baked into the binary. Read from real environment variables or a secrets manager at runtime; one image for every environment.

### T-029 Multi-tenancy architecture

- priority: P0
- feature: SaaS platform
- created: 2026-10-26

Every table is global today; one deployment = one store. Decide the isolation model (shared schema with `tenant_id` + row-level security, schema-per-tenant, or DB-per-tenant), then add tenant scoping to all queries, sessions and the WebSocket hub, and tenant resolution (subdomain or custom domain).

### T-030 Store settings & white-label branding

- priority: P0
- feature: SaaS platform
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
- feature: SaaS platform
- created: 2026-10-26

Import products, prices, price options and opening stock with expiry from a spreadsheet, with validation preview and error report. The biggest onboarding blocker for a new store.

### T-032 Tenant signup & onboarding wizard

- priority: P0
- feature: SaaS platform
- depends: T-029, T-031
- created: 2026-10-26

Self-serve signup that creates a tenant and its first admin, then a guided setup: store details, categories, products (import), users, opening stock.

First-admin creation moves here from the `create-admin` command (T-018) once a new store is set up in the app.

### T-033 Subscription billing & plans

- priority: P0
- feature: SaaS platform
- depends: T-029
- created: 2026-10-26

Plans, trials, recurring billing (Paystack/Flutterwave for NGN), dunning, plan limits (users, products, branches) and feature gating.

### T-034 Super-admin console

- priority: P0
- feature: SaaS platform
- depends: T-029
- created: 2026-10-26

Platform operator view: tenants, plans and billing status, usage, suspend/reactivate, support impersonation (audited).

### T-047 Online-first with solid offline operation

- priority: P0
- feature: Offline & devices
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
- feature: Permissions & audit
- created: 2026-10-26

Every change writes an `audit_log` entry in the same transaction: who, when, what, before and after, and a reason where it matters. Only purchase orders do this today. Add a shared `auditEntry` helper in the service layer.

- [ ] Sales, voids, returns and refunds
- [ ] Stock movements: receiving, manual adjustments, stock-take completion
- [ ] Product, price and cost edits; categories, groups and merges
- [ ] Users, roles and permissions; logins, logouts and failed logins
- [ ] Settings changes and data exports
- [ ] Admin screen to browse, filter (user, entity, action, date) and export the log

Password resets by an admin are not audited yet (T-018 audits user creation only); include them here.

### T-099 Granular permissions with default roles

- priority: P1
- feature: Permissions & audit
- depends: T-023
- merged: T-086, T-071, T-092
- created: 2026-10-27

Everything is behind a permission. The system ships with default roles that have permissions assigned, and every permission has an explanation.

- [ ] Permission catalog in code: key, name, description, group (Sales, Inventory, Purchasing, Reports, Admin…)
- [ ] Every route and every UI action checks a permission. Today many are ungated: creating products (`POST /inventory/add-item`), receiving, sales, returns, stock-take create and count, and most read endpoints
- [ ] Default roles seeded by migration and kept in sync with the catalog (e.g. Owner, Manager, Pharmacist, Cashier, Storekeeper)
- [ ] Admin screen lists permissions grouped, with their explanations, and lets admins build custom roles from the catalog
- [ ] Remove the free-form create/delete permission endpoints; permissions are defined by code
- [ ] Decide the purchase order read endpoints: list, get and history have no route permission (the service only hides drafts from people who aren't the author or an approver)
- [ ] Decide whether receiving against a purchase order needs its own permission (today anyone who can receive can)
- [ ] Seed grants without matching role names: migration 000043 grants by name (`r.name = 'Cashier'`, `IN ('Manager', 'Admin')`), which breaks when a deployment's roles are named differently or tenants rename them
- [ ] Works with per-request permission checks (T-023)

- [ ] Split user management from `admin:access`: today anyone holding it can add users with any role, including Admin, and change roles. Limit which roles a non-Admin can grant (raised in T-018 review)

### T-104 Password policy toward NIST guidance

- priority: P1
- feature: Security hardening
- created: 2026-10-28

T-018 set a 10-character minimum, a common-password list and a username check. NIST SP 800-63B rev 4 also asks for at least a 64-character maximum, no composition rules, and a 15-character minimum while login is password-only. Move to 15 when it's workable for the shop or once there's a second factor (T-075).

### T-103 Notification centre

- priority: P1
- feature: Notifications
- merged: T-091, T-078
- created: 2026-10-27

One place for everything the system needs to tell people. Today the only notification is the approver badge and dashboard banner, which poll every 60 seconds (`usePendingPurchaseOrders.js`).

- [ ] Stored in-app inbox with a bell and unread count, per user
- [ ] Events: purchase orders submitted, approved, sent back or rejected (T-089); low stock and upcoming expiry; backup failed (T-026); offline sync problems (T-047); what's new (T-095); insights (T-100)
- [ ] Channels: in-app, email, WhatsApp/SMS, push (desktop and mobile)
- [ ] Per-user and per-role preferences, and a daily digest option

### T-037 Costs on purchase orders

- priority: P1
- feature: Purchasing & receiving
- created: 2026-10-26

Expected cost per line, order total, and comparison with the received cost.

### T-038 Supplier management screen

- priority: P1
- feature: Purchasing & receiving
- created: 2026-10-26

Rename and merge duplicate suppliers, contact details (phone, email, address), payment terms.

### T-093 Settings screen for hard-coded values

- priority: P1
- feature: Admin & settings
- created: 2026-10-27

Hard-coded values should become settings that an admin edits on a settings screen. Store them in the database, read them once per session and serve them to the SPA, so the frontend and backend share one value. Store branding (T-030) uses the same store.

- [ ] Large purchase order confirmation: 100,000 units per line or more than 200 lines (`CONFIRM_QUANTITY`, `CONFIRM_LINES` in `views/PurchaseOrders.vue`)
- [ ] Near-expiry alert on Receive Items: under 183 days turns amber and needs "I confirm the near-expiry dates" (`expiryClass`, `expiryTitle`, `hasNearExpiry` in `views/ReceiveItems.vue`); past dates stay red
- [ ] Expiring-soon window on the dashboard and expiring list: 90 days (`INTERVAL '90 days'` in `repository/dashboard.go`); decide whether it shares a value with the Receive Items alert
- [ ] Approval badge poll interval: 60 seconds (`usePendingPurchaseOrders.js`)
- [ ] Completed purchase orders cap: 100 (`purchaseOrderHistoryMax`)

### T-040 Customer management

- priority: P1
- feature: Customers & loyalty
- created: 2026-10-26

The customer name is typed on the receipt but never stored.

- [ ] Customer profiles: name, phone, email, notes, consent to be contacted
- [ ] Quick lookup by phone at the POS and attach to the sale
- [ ] Purchase history and totals per customer
- [ ] Import, edit and merge duplicate customers
- [ ] Balances and credit (T-041), loyalty (T-084)

### T-041 Credit sales & debtor balances

- priority: P1
- feature: Customers & loyalty
- depends: T-040
- created: 2026-10-26

Sell on account, track customer balances, record part-payments, statements and aged debt. Very common in Nigerian retail.

### T-042 Void / cancel a completed sale

- priority: P1
- feature: Point of sale
- created: 2026-10-26

No way to void a mistaken sale. Add void with reason, permission and audit entry; reverse stock movements and payments.

### T-044 Shifts, cash-up & refund methods

- priority: P1
- feature: Point of sale
- merged: T-043
- created: 2026-10-26

Cash control for each till.

- [ ] Open a till session with an opening float; every sale, return and payout belongs to a session
- [ ] Record how each refund was paid out (cash, transfer, store credit); returns only store `total_refunded` today
- [ ] Close with counted cash per payment method, showing expected, counted and over/short
- [ ] End-of-day Z-report per till and per day

### T-046 Discount limits, price overrides & promotions

- priority: P1
- feature: Point of sale
- created: 2026-10-26

Max discount per role, price override requiring permission, promotions (percentage, buy-X-get-Y, time-bound).

### T-101 Keyboard workflow for POS and Receive Items

- priority: P1
- feature: Point of sale
- created: 2026-10-27

Make the two busiest screens fully usable without a mouse.

- [ ] POS: focus search (`/` or F2), arrow keys + Enter to add, change quantity and price option, remove a line, discount, choose a payment method, save, save and print, hold
- [ ] Receive Items: Enter moves to the next field, a new row at the end, quick expiry entry (MM/YY), save
- [ ] Barcode scanner input works from anywhere on both screens
- [ ] A `?` overlay listing the shortcuts; the same keys mean the same thing on both screens

### T-048 Thermal printing & digital receipts

- priority: P1
- feature: Point of sale
- created: 2026-10-26

ESC/POS thermal printer support (58/80mm), reprint from sales history, receipts by WhatsApp/SMS/email.

### T-049 Batch tracking & FEFO

- priority: P1
- feature: Inventory control
- merged: T-036
- created: 2026-10-26

Track stock per batch, not just per product. Today stock is one number per product summed from `stock_movement`, and expiry is the product's earliest date, so you can't tell how much of an expiring batch is actually left.

- [ ] Capture the batch/lot number on receiving: `product_batch.batch_no` exists but `createProductBatchQuery` never writes it
- [ ] Several batches (different expiry dates) on one purchase order line
- [ ] Quantity on hand per batch; sales, returns and adjustments deduct from a batch
- [ ] First-expiry-first-out by default at the POS, with an override
- [ ] Expiry lists and write-offs per batch
- [ ] Recall lookup: which sales (and customers, T-040) took batch X

### T-050 Manual stock management

- priority: P1
- feature: Inventory control
- created: 2026-10-26

Adjust stock directly, without starting a stock take. Today the only way is the stock field on the product edit form.

- [ ] Adjust one product or many at once, by setting a new count or adding/removing a quantity
- [ ] Reason codes: expired, damaged, theft/loss, correction, opening stock, internal use; new movement types for write-offs
- [ ] Set or correct expiry and batch while adjusting
- [ ] Own permission, plus approval above a configurable quantity or value
- [ ] Audit entry for every adjustment; route the product edit stock field through this flow, or remove it

### T-051 Multi-store & multi-branch support

- priority: P1
- feature: SaaS platform
- merged: T-080
- created: 2026-10-26

One business owns one or more stores, and each store has one or more branches. Dispensary and store room become locations inside a branch (today they exist only as stock-take counts).

- [ ] Stock, batches and stock takes per branch and location
- [ ] Prices shared across branches, with optional branch overrides
- [ ] Users and permissions per branch, and a branch switcher
- [ ] Transfers between branches, with an in-transit state and receipt at the other end
- [ ] Reports per branch, plus an owner view across branches: sales, stock, transfers and branch comparison

### T-052 Product history

- priority: P1
- feature: Inventory control
- created: 2026-10-26

One screen per product that shows its whole history:

- [ ] Stock ledger: every receipt, sale, return, adjustment and stock take, with running balance, user and reference
- [ ] Purchase cost history per receipt and supplier (already stored on `product_batch`)
- [ ] Selling price history: prices are updated in place today (`UpdateProduct`), so start recording price changes with who and when
- [ ] Restock dates and quantities, and time between restocks
- [ ] Simple charts for cost, price and stock over time

### T-053 Reorder planning & on-order quantities

- priority: P1
- feature: Purchasing & receiving
- merged: T-070
- created: 2026-10-26

Know what to buy and what's already coming.

- [ ] Show quantities on open purchase orders on the Products list and the Low Stock page (today they only appear in the purchase order low-stock picker as "already on order")
- [ ] Suggest reorder quantities from the reorder level (or the group's, T-097), sales rate and supplier lead time, minus what's on order
- [ ] Generate a draft purchase order per supplier from the suggestions

### T-054 Returns to supplier

- priority: P1
- feature: Purchasing & receiving
- created: 2026-10-26

Send expired, damaged or excess stock back to a supplier with a document, stock movement and expected credit.

### T-055 Supplier invoices & payables

- priority: P1
- feature: Purchasing & receiving
- created: 2026-10-26

Record supplier invoices against receipts, amounts owed, payments made and balances per supplier.

### T-056 Product archive & bulk price update

- priority: P1
- feature: Inventory control
- created: 2026-10-26

Deactivate products without losing history; update prices in bulk (by category or percentage).

### T-097 Product groups with a group reorder level

- priority: P1
- feature: Inventory control
- merged: T-074
- created: 2026-10-27

Group interchangeable products, e.g. different brands of amoxicillin 500mg that all sell at ₦600. A group has its own reorder level and total stock, and low stock is judged on the group as well as on each product. Optional shared group price. Groups feed reorder planning (T-053).

- [ ] Group with its own reorder level, total stock and optional shared price
- [ ] Low stock judged per group as well as per product
- [ ] Substitution at the POS: when a product is out of stock, suggest in-stock products from the same group or with the same active ingredient (T-057)

### T-098 Merge duplicate products, categories and groups

- priority: P1
- feature: Inventory control
- created: 2026-10-27

Pick a surviving record and move everything onto it: stock movements, batches, prices, sales and return lines, purchase order lines, stock-take items. Then remove or archive the duplicate. Preview what moves, require a permission and write an audit entry. Suppliers merge in T-038; groups come from T-097.

### T-057 Pharmacy product attributes

- priority: P1
- feature: Pharmacy compliance
- created: 2026-10-26

Generic name / active ingredient, strength, dosage form, pack size, NAFDAC number, classification (OTC / POM / controlled).

The POS search already looks for `generic_name` (`pos/ProductPanel.vue`), but the backend has no such field, so that part of the search never matches.

### T-058 Prescriptions & dispensing records

- priority: P1
- feature: Pharmacy compliance
- depends: T-057
- created: 2026-10-26

Record prescriptions (prescriber, patient, items, directions), link to sales, pharmacist sign-off required for POM items.

### T-059 Controlled drugs register

- priority: P1
- feature: Pharmacy compliance
- depends: T-057
- created: 2026-10-26

Running register of controlled substances: receipts, dispensing, balances, who did it, for inspections.

### T-060 Sales & profit reports

- priority: P1
- feature: Reports & insights
- merged: T-061
- created: 2026-10-26

- [ ] Gross profit and margin by period, product and category, using batch cost (already captured on `product_batch`)
- [ ] Sales by cashier, product, category, payment method and hour
- [ ] End-of-day summary

### T-062 Stock valuation, dead stock & write-off reports

- priority: P1
- feature: Reports & insights
- depends: T-050
- created: 2026-10-26

Stock value at cost and retail, slow and dead movers, value of expired / written-off stock.

### T-096 More useful sales and inventory metrics

- priority: P1
- feature: Reports & insights
- created: 2026-10-27

At-a-glance numbers for the dashboard; the full reports are T-060, T-060 and T-062.

- [ ] Sales: gross margin %, average basket value and items per sale, sales by hour and weekday, returns rate, comparison with the same day last week
- [ ] Inventory: stock value at cost, days of cover per product or group, stock turnover, sell-through since receipt, value at risk of expiry, fast movers next to dead stock

### T-100 Business insights

- priority: P1
- feature: Reports & insights
- created: 2026-10-27

Go beyond tables: have the system point out what matters. The detailed reports are T-060 to T-062; the metrics are T-096.

- [ ] "Runs out in N days at the current sales rate" for each product or group
- [ ] Margin changes, e.g. the last restock cost went up but the selling price didn't
- [ ] Sales up or down against the same weekday or period last time, by product and category
- [ ] Money tied up in slow movers, and stock value at risk of expiry in the next 30/60/90 days
- [ ] Daily and weekly summaries for owners (delivered through T-103)

### T-063 Export data

- priority: P1
- feature: Reports & insights
- created: 2026-10-26

Get data out in a usable form.

- [ ] CSV/Excel export on every list and report (and relink the existing stock CSV export)
- [ ] Full store export: products, prices, stock and batches, customers, sales, purchases, suppliers, as a zip of CSVs, for accounting, migration or leaving
- [ ] Optional scheduled export to the backup location (T-026)
- [ ] Exports need a permission and are audited (T-035)

### T-064 Admin user management

- priority: P1
- feature: Admin & settings
- created: 2026-10-26

Admins can add users (T-018), but can't deactivate them, and users can't change their own password. Add deactivate, change own password, last login, and optionally invite by link instead of a starting password.

### T-106 Temporary passwords: change on first sign-in

- priority: P1
- feature: Security hardening
- created: 2026-10-28

The password an admin sets with Add user or Reset Password is known to the admin, so make it temporary.

- [ ] A `must_change_password` flag on users (migration), set by Add user and admin resets, not by `create-admin`
- [ ] While it's set, every page and API call except change password and sign out sends the user to a "Set your password" screen, enforced on the server
- [ ] The screen asks for the temporary password, a new one and a confirmation. Same rules as everywhere (`ValidatePassword`), and it must differ from the temporary one. Saving clears the flag and writes an audit entry
- [ ] Administration → Users shows "Password not set yet"
- [ ] Optional: temporary passwords expire after a set number of days

Shares the change-password endpoint and screen with T-064's "users change their own password". Build them together. Raised during T-018.

### T-065 Release & deployment pipeline

- priority: P1
- feature: Reliability & operations
- created: 2026-10-26

Staging environment, versioned releases, migrations run as part of deploy, pinned images (compose uses `postgres:latest`).

### T-102 Make the app snappy

- priority: P1
- feature: Reliability & operations
- merged: T-066
- created: 2026-10-27

General speed work, measured first so it goes where the time actually is.

- [ ] Measure: request timings in logs, Postgres slow query log, JS bundle size per route
- [ ] Stock balance table: `inventory_view` (migration 38) sums all of `stock_movement` on every read, which slows down as data grows. Keep a balance per product (and later per batch and branch) updated in the same transaction
- [ ] Indexes for the hot queries (product search, sales history, dashboard)
- [ ] Cache the product catalog in the browser so POS search is instant
- [ ] Optimistic updates and prefetching for the common screens; pagination on every long list
- [ ] Compression and long cache headers for the embedded static assets

### T-067 Purchase order robustness

- priority: P1
- feature: Purchasing & receiving
- merged: T-085, T-088
- created: 2026-10-26

Small fixes that keep purchase orders correct under retries, races and edits.

- [ ] Idempotency key for creating an order: Save is disabled while saving, but a retry after a lost response can still create a duplicate
- [ ] One shared lock-and-check helper: edit, close, move and receive each call `GetPurchaseOrderForUpdateTx` and compare the version themselves in `service/purchase_orders.go`
- [ ] Block changing or clearing the supplier on an order that already has receipts (today it's allowed silently and the receipts keep the old supplier); point to Move outstanding instead

### T-082 Responsive mobile & tablet support

- priority: P1
- feature: Offline & devices
- created: 2026-10-26

Every screen works on phones and tablets, not just desktop.

- [ ] POS usable on a tablet; stock-take counting and receiving on a phone
- [ ] Dashboard, approvals and reports readable on a phone for owners
- [ ] Collapsible navigation, touch-sized controls, tables that become cards on small screens
- [ ] Known issue: purchase order card actions wrap below about 1280px

A native mobile app can come later through the PWA / app shell in T-094.

### T-084 Customizable loyalty program

- priority: P1
- feature: Customers & loyalty
- depends: T-040, T-093
- created: 2026-10-26

Loyalty rules an admin sets up per store, instead of one fixed scheme.

- [ ] Earn rate (points per ₦ spent), with multipliers or exclusions by category or product (e.g. exclude prescription-only items)
- [ ] Redemption value, minimum to redeem, and points expiry
- [ ] Optional tiers with their own rates, and bonuses (birthday, first purchase, promotions)
- [ ] Points balance on the receipt and at the POS; redeem as a payment method
- [ ] Every earn, redeem and manual change audited

### T-045 Tax / VAT support

- priority: P2
- feature: Point of sale
- created: 2026-10-26

Configurable tax rates for retail beyond medicines: tax-inclusive or exclusive pricing, exempt products, tax on receipts and in reports. Medical and pharmaceutical products are VAT-exempt in Nigeria, so for a pharmacy this mostly matters for non-drug items (cosmetics, toiletries, groceries) and for selling the system to other kinds of retailers. Check current rules with an accountant before building.

### T-068 Change-due modal after saving a sale

- priority: P2
- feature: Point of sale
- created: 2026-10-26

Show a modal with the change to give when change > 0 after saving a sale.

### T-069 Purchase order history search & filters

- priority: P2
- feature: Purchasing & receiving
- created: 2026-10-26

Search and filter by supplier, status and date, with pagination; the Completed tab is capped at the newest 100.

### T-089 Approval workflow refinements

- priority: P2
- feature: Purchasing & receiving
- merged: T-090
- created: 2026-10-27

- [ ] A drafter can withdraw a submitted order back to draft
- [ ] A manager can send an order back for changes instead of only rejecting it
- [ ] A rejected draft shows "Rejected" and its reason on the order itself, not "Closed" with the reason only in the activity log
- [ ] Tell the drafter when their order is approved, sent back or rejected (through T-103)

### T-073 Barcode label printing

- priority: P2
- feature: Inventory control
- created: 2026-10-26

Generate and print barcode/shelf labels for products without manufacturer barcodes.

### T-075 Cashier PIN switching & admin 2FA

- priority: P2
- feature: Security hardening
- created: 2026-10-26

Fast PIN-based user switching on a shared till; TOTP 2FA for admin accounts.

### T-105 Case-insensitive unique usernames

- priority: P2
- feature: Security hardening
- created: 2026-10-28

Usernames are unique only case-sensitively, so "Ada" and "ada" can both exist, and login is case-sensitive. Add a unique index on `lower(username)` (migration), check existing data for clashes first, and decide whether login should ignore case. Found in T-018's review.

### T-076 Frontend lint & tests

- priority: P2
- feature: Reliability & operations
- merged: T-072
- created: 2026-10-26

There is no JS tooling today.

- [ ] ESLint and Prettier
- [ ] Component tests for the composables and stores
- [ ] Committed end-to-end tests (e.g. Playwright) for POS, receiving, stock taking, and purchase orders including approval and receive-against-order. PR #29's Playwright checks were run by hand and aren't in the repo
- [ ] Run them in CI (T-025)

### T-077 API documentation (OpenAPI)

- priority: P2
- feature: Finance & integrations
- created: 2026-10-26

Document the JSON API; groundwork for a public API.

### T-079 Expense tracking & accounting export

- priority: P2
- feature: Finance & integrations
- created: 2026-10-26

Record shop expenses; export sales, purchases and expenses to accounting tools or Excel.

### T-081 Public API, webhooks & integrations

- priority: P2
- feature: Finance & integrations
- depends: T-077
- created: 2026-10-26

Tenant API keys, webhooks (sale created, low stock), integrations with e-commerce and delivery platforms.

### T-094 Desktop app

- priority: P2
- feature: Offline & devices
- created: 2026-10-27

Run as an installed app instead of a browser tab. Options: an installable PWA first, then a native shell (Wails suits the Go backend; Tauri or Electron also work). A desktop app also helps offline storage (T-047), direct thermal printing and local backups (T-026). Needs auto-update.

### T-095 What's new on login

- priority: P2
- feature: Admin & settings
- created: 2026-10-27

After a release, show each user a short "What's new" panel once after they log in, then remember per user that they've seen it. Release notes live in the repo (they could be drawn from DONE.md) and can be targeted by role.

### T-083 Localisation & multi-currency

- priority: P2
- feature: SaaS platform
- created: 2026-10-26

Currency and formatting per tenant (money is kobo-only today), translations.
