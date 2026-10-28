# Done

Finished cards from [BOARD.md](BOARD.md), newest first, grouped by month.

## 2026-10

### T-018 Lock down public user registration

- feature: Security hardening
- done: 2026-10-28
- prs: #32

Removed public `POST /user/register`. Admins add users in Administration → Users → Add user (`POST /admin/api/users`, behind `admin:access`, audited). The first admin on a new install comes from the `create-admin` subcommand of the executable. Passwords need 10+ characters, must not be on the bundled NCSC common-password list and must not contain the username as a word; the admin reset uses the same rule and now stores the password as typed. Plan: `docs/plans/T-018-lock-down-registration.md`.

### T-087 Purchase order approval workflow

- feature: Purchasing & receiving
- done: 2026-10-27
- prs: #29

Cashiers write drafts and submit them; managers approve or reject (with a reason), and their own orders skip approval. States `draft` → `pending_approval` → `open` or `closed`. Permissions `purchaseorder:draft` and `purchaseorder:approve`. Drafts are visible only to their author and approvers, can't be received against, and their PDF is stamped DRAFT. Every transition checks the order version and writes an audit entry. Approver badge and dashboard banner. Also: shared `ConfirmDialog` and `SupplierInput`, order card redesign, PDF watermarks. Migrations 000042, 000043.

### T-039 Seed purchase order permissions

- feature: Permissions & audit
- done: 2026-10-27
- prs: #29

Migration 000043 seeds `PurchaseOrder`/`Draft` (Cashier) and `PurchaseOrder`/`Approve` (Manager, Admin), so there is no manual permission step at deploy.

### T-016 Database integration test harness

- feature: Reliability & operations
- done: 2026-10-27
- prs: #28

`internal/testdb` creates/drops a database per test and runs real migrations; enabled with `PHARMACY_TEST_DSN`. Covers migrations and the purchase order lifecycle.

### T-015 Audit log foundation

- feature: Permissions & audit
- done: 2026-10-27
- prs: #28

Append-only `audit_log` (a trigger blocks UPDATE/DELETE) written in the same transaction as the change: who, when, reason, before/after. Only purchase orders use it so far. Migration 000041.

### T-014 Purchase orders & suppliers

- feature: Purchasing & receiving
- done: 2026-10-27
- prs: #28

Build an order from search, low-stock items or items not yet in the product list; edit, export a PDF, close with a reason, or move outstanding quantities to another supplier. Receiving against an order pre-fills outstanding quantities and rejects over-receiving. `supplier` table (find-or-create, case-insensitive, backfilled from receipts). Edits carry the order version. Writes need `purchaseorder:create`. Migration 000040.

### T-013 Vue SPA / new UI

- feature: Reliability & operations
- done: 2026-10-02
- prs: #26, #27

Migrated from Go templates to a Vue 3 + shadcn-vue SPA served under `/app/`, embedded in the Go binary.

### T-012 Point of Sale

- feature: Point of sale
- done: 2026-10-02
- prs: #2, #26, #27

Prefix search on name/barcode/manufacturer, cart with price-option per line, discounts, split payments (Cash/Card/Transfer), server-generated receipt numbers, print receipt, idempotent submission, leave-with-cart guard.

## 2026-09

### T-011 Low stock & expiring items pages

- feature: Reports & insights
- done: 2026-09-29
- prs: #26

Dedicated Low Stock and Expiring pages with PDF export (jsPDF).

### T-010 Authentication & sessions

- feature: Security hardening
- done: 2026-09-29
- prs: #26

Login page (built before PRs were used), Argon2id password hashing, gorilla/sessions cookie sessions (expire at 3am), CSRF via gorilla/csrf with `GET /csrf-token` for the SPA.

## 2026-06

### T-009 Stock taking

- feature: Inventory control
- done: 2026-06-02
- prs: #9, #12, #18, #24, #25

Named stock-take sessions, live multi-user counting over WebSocket, dispensary + store counts, notes, expiry correction; completing (`stock:complete`) posts IN/OUT stock-taking adjustments.

## 2026-05

### T-008 Receiving (goods-in)

- feature: Purchasing & receiving
- done: 2026-05-09
- prs: #1, #22

Receive supplier deliveries with cost price and expiry per line; hold/resume a receiving; received-items history; idempotency keys prevent duplicate receipts.

### T-007 Role-based access control & admin panel

- feature: Permissions & audit
- done: 2026-05-09
- prs: #7, #23

Roles (Admin/Manager/Cashier), permissions and role-permission assignment. Admin panel (`admin:access`): users (change role, reset password), roles, permissions, categories.

## 2026-04

### T-006 Dashboard

- feature: Reports & insights
- done: 2026-04-08
- prs: #15, #17

KPIs with period-over-period trend, sales trend chart, category split, top sellers, expiring items, expiry by category, low stock, recent transactions; date filter.

### T-005 Product catalog

- feature: Inventory control
- done: 2026-04-05
- prs: #1, #13

Products with category, barcode, manufacturer, reorder level, cost price. Multiple price options per product (e.g. tablet vs pack, `quantity_per_unit`) with a default. Edit product incl. manual stock adjustment (`inventory:edit`).

## 2025-12

### T-004 Sales history

- feature: Point of sale
- done: 2025-12-16
- prs: #2, #8

Browse sales by date range and search, item and payment detail, net period total (minus refunds) gated by `salestotal:view`.

### T-003 Returns

- feature: Point of sale
- done: 2025-12-15
- prs: #5, #6

Return items against an original sale; stock goes back in (`IN_SALE_RETURN`); returns shown on the sale.

## 2025-10

### T-002 Stock export to CSV

- feature: Reports & insights
- done: 2025-10-28
- prs: #4

`GET /inventory/report/stock` CSV export.

Note: the new SPA no longer links to it.

### T-001 Hold & resume sales

- feature: Point of sale
- done: 2025-10-07
- prs: #3

Park an in-progress sale and resume or void it from the Held Sales page.
