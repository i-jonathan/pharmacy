# Primocrest Pharmacy POS

A full-stack **pharmacy point-of-sale (POS) and inventory management system** for day-to-day operations of a physical pharmacy store. Prices are in **Nigerian Naira** (stored as kobo / integer cents).

---

## Tech Stack

| Layer              | Technology                                                              |
|--------------------|-------------------------------------------------------------------------|
| **Backend**        | Go 1.24, `net/http` (standard library), Gorilla sessions / CSRF / WebSocket |
| **Database**       | PostgreSQL (via `pgx` + `sqlx`), 43 sequential migrations               |
| **Templating**     | Go `html/template` (server-rendered HTML)                               |
| **Frontend**       | Vue 3 + vue-router + shadcn-vue, built with Vite into `static/dist/` and embedded |
| **Styling**        | TailwindCSS v4                                                          |
| **Auth**           | Argon2id password hashing, cookie sessions, role-based access control    |
| **Infrastructure** | Docker Compose, Makefile                                                |

---

## Architecture

The backend follows a **clean / layered architecture**:

```
main.go                       ← wires everything, embeds templates + static assets
├── config/                   ← env-based config (cleanenv + godotenv)
├── model/                    ← domain models (Product, Sale, User, StockTaking…)
├── repository/               ← database access (SQL queries via sqlx)
│   └── store.go              ← initialises the DB connection pool
├── service/                  ← business logic layer
│   └── interfaces.go         ← defines service contracts
├── adapter/http/
│   ├── router/               ← route registration (per domain)
│   ├── controller/           ← HTTP handlers → services → templates/JSON
│   ├── middleware/            ← auth, logging, permission injection
│   └── helper/               ← HTTP response utilities
├── adapter/websocket/        ← WebSocket hub for real-time stock-taking updates
├── template/                 ← Go HTML templates (SPA shell, login, error page)
│   └── static/               ← CSS, fonts, JS, Vite-built Vue bundles
├── internal/                 ← shared types (request/response DTOs) & constants
├── httperror/                ← structured HTTP error types
└── db/migrations/            ← 43 PostgreSQL migrations (golang-migrate)
```

### Frontend / Vue

The `frontend/` directory is a Vue 3 + Vite project. It builds into `backend/template/static/dist/`. The Go server embeds and serves those assets, parsing the Vite manifest at startup so templates can reference hashed filenames via the `ViteAsset` template function.

There are **two Vite entry points**:

| Entry Point          | Served by                                   | Purpose                                                  |
|----------------------|---------------------------------------------|----------------------------------------------------------|
| `src/next/main.js`   | `template/next-dashboard.html` at `/app/*`  | Main single-page app (vue-router): dashboard, POS, inventory, receiving, purchase orders, stock taking, admin |
| `src/login/main.js`  | `template/login-v2.html` at `/user/login`   | Login page                                               |

The SPA calls the JSON endpoints under `/api`, `/sales`, `/inventory`, `/admin` and `/stock-taking`. Assets are embedded in the Go binary, so rebuild (`make build-frontend`) and restart the server after frontend changes.

---

## Features

### 1. Sales / POS
- Product search matches the start of the name, barcode or manufacturer (case-insensitive)
- Cart with quantity and a selectable price option per item (e.g. tablet vs. pack)
- Discounts, and split payments across Cash, Card and Transfer
- Receipt numbers generated server-side on save; **print receipt** from the POS
- **Hold** a sale and resume it later from the Held Sales page
- Leaving the POS with a non-empty cart prompts you to hold or discard it (in-app navigation guard plus a browser `beforeunload` warning)
- Idempotency keys prevent duplicate submissions

### 2. Sales History & Returns
- Browse sales by date range and search; view item and payment details
- Period total shown only to users with `salestotal:view`
- **Return items** from a sale's detail view; returned items and refunds are listed on the original sale

### 3. Inventory Management
- Product catalog with categories, manufacturers, barcodes
- **Multiple price options** per product, one of them the default
- Reorder level monitoring (Low Stock page, with PDF export)
- Expiry tracking (earliest expiry per product); Expiring page lists items expiring within 90 days (PDF export)
- **CSV export** of current stock
- Stock movement ledger (receives, sales, returns, stock-taking adjustments)
- Edit products (prices, stock, details); requires `inventory:edit`

### 4. Receiving / Supplier Deliveries
- Record deliveries by supplier with quantity, cost price, selling price and expiry per item
- Optionally link received items to a purchase order item
- **Hold** partially completed receive forms for later
- Received-items history
- Idempotency keys prevent duplicate receives

### 5. Purchase Orders
- **Approval flow:** a user with the *draft* permission (for example a cashier) writes a **draft**. Managers can already see and edit it while it is being written. When the author is done they **submit it for approval**; from then on only a manager can change it. A manager edits if needed, then **approves** (the order becomes `open`) or **rejects** it with a reason (closed, kept on record). Managers who create an order skip the drafts: theirs starts open. Drafts and submitted orders can't be received against and have no PDF; the author can only read their own orders. Managers see a badge on *Purchase Orders* and a dashboard banner when orders are waiting
- Build an order from products you stock (search), from **low stock** (pick only the rows you want; items already on another open order are flagged and the suggested quantity accounts for them), or from items **not in your product list yet** (typed by name; an exact match to an existing product is added as that product)
- Each order has an optional **additional notes** field, and each item can carry its own note (brand, strength, pack size...). Notes show on the order, in the item dialog, on the receive screen and on the PDF, and an item's note travels with it when outstanding items are moved to another supplier
- Orders are listed with a coloured status (open blue, partial amber, received green, closed grey) and a progress bar; **View** opens a dialog with the order's items, filterable (all / outstanding / complete), searchable and paged, so large orders do not become one long list. Long item lists in the editor are filtered and paged the same way
- Every change is written to an append-only **audit log** in the same transaction (who, when, reason, and before/after values): creation, edits, receipts (with the stock batch and which order line it filled), closing and moving. The View dialog's **Activity** shows it; closing and moving require a reason. Edits carry the order's version, so saving a stale copy is refused instead of overwriting a newer receipt or edit
- Orders can be exported as a PDF for the supplier, edited while open or partial, and filtered into *Waiting for stock* and *Completed*
- **Partial supply:** click Receive on an order to open the receive screen with each line pre-filled to its outstanding quantity. Receive what arrived, drop what did not; the order becomes `partial` and the rest stays outstanding. Receiving more than is outstanding is rejected
- New items are registered (or linked to an existing product) on the receive screen, which links the product to the order line
- **Another supplier for the unfilled items:** *Other supplier* on a partial order closes it with what it received and creates a new open order, linked to the first, containing only the outstanding quantities for the supplier you choose. Receipts record their own supplier, so a delivery can also be recorded from a different supplier than the order was placed with (the receive screen flags this)
- Reducing a line's quantity to what has already arrived completes it; **Close order** stops waiting for whatever is still outstanding (status `closed`, stock already received is unaffected)
- Status is `draft`, `pending_approval`, `open`, `partial`, `received` or `closed`. Received quantities are updated in the same database transaction as the stock receipt, so they cannot drift apart

### 6. Stock Taking
- Create stock-taking sessions (only one can be in progress at a time)
- Auto-snapshot of current system quantities
- Record physical counts for **dispensary** and **store** separately
- Tracks who last updated each count
- **Real-time collaboration** via WebSocket; updates are broadcast to users viewing the same session
- Optionally update a product's expiry while counting
- Completing a session writes stock-movement entries for the differences; requires `stock:complete`
- Sessions can have status `in_progress`, `completed` or `cancelled`, but no endpoint currently sets `cancelled`

### 7. Dashboard
- **KPI cards**: Total Sales, Total Orders, Total Products, Low Stock Items, Expiring Soon (with a date filter)
- **Sales trend chart**, top-selling products, recent transactions, and a low-stock table (top 5)
- Low-stock and expiring cards link to their full pages

### 8. Users & Auth
- No public sign-up. Admins add staff in **Administration → Users → Add user** (username, role, starting password); the first admin on a new install comes from the `create-admin` command (see Quick Start)
- Usernames are 3–15 letters, digits or underscores. Passwords need at least 10 characters (surrounding spaces don't count), must not be on the bundled common-password list (NCSC top 100k, 10+ character entries) and must not contain the username as a word (so "mary123456" is refused for "mary", but "nevada" is fine for "ada")
- Creating a user writes an audit entry; one created by `create-admin` has no actor
- **Argon2id** password hashing
- Session-based authentication (Gorilla sessions)
- **Role-based access control**: users have one role (`Admin`, `Manager` or `Cashier`, chosen when the user is added), and roles are granted permissions through `role_permissions`
- Permissions are stored as `resource`/`action` pairs and exposed as lowercase `resource:action` keys. They are computed at login and kept in the session, so a permission change takes effect after the user logs in again

### 9. Permissions

| Key                          | Description                                    |
|------------------------------|------------------------------------------------|
| `salestotal:view`            | View sales totals on the history page and sales figures/trend on the dashboard |
| `stock:view`                 | Show system quantities and variances on the stock-counting screen (checked in the UI only) |
| `stock:complete`             | Complete / finalise a stock-taking session     |
| `inventory:edit`             | Edit product details (prices, stock info)      |
| `admin:access`               | Access the `/admin/` APIs (users, roles, permissions, categories) |
| `purchaseorder:draft`        | Write purchase order drafts, edit your own draft and submit it for approval (you only see your own orders) |
| `purchaseorder:approve`      | Manager: create orders directly, edit any draft, approve or reject, and close or move approved orders. Receiving against an order is not gated |

Seeded by migrations: `stock:view`, `stock:complete` and `inventory:edit` go to `Admin` and `Manager`; `admin:access` goes to `Admin` only; `purchaseorder:draft` goes to `Cashier` and `purchaseorder:approve` to `Manager` and `Admin` (migration 43). Change who holds them in Administration → Permissions; users pick up a change the next time they log in.

---

## Database Schema

43 migrations covering:

- `users`, `roles`, `permissions`, `role_permissions`
- `categories`, `products`, `product_prices`
- `receiving_batches`, `product_batches`
- `stock_movements` (ledger with movement types)
- `sales`, `sale_items`, `sale_payments`
- `returns`, `return_items`
- `held_transactions`
- `stock_takings`, `stock_taking_items`
- `inventory_view` (aggregated stock + earliest expiry)
- `idempotency_keys`
- `supplier` (one row per supplier name, unique ignoring case and spaces; referenced by `receiving_batch.supplier_id` and `purchase_order.supplier_id`, which also keep the name as written on the document)
- `purchase_order`, `purchase_order_item`

### Movement Types

| Type                    | Direction | Description                  |
|-------------------------|-----------|------------------------------|
| `IN_PURCHASE`           | In        | Supplier delivery receive    |
| `IN_SALE_RETURN`        | In        | Returned from a sale         |
| `IN_STOCK_TAKING`       | In        | Positive stock-taking adjustment |
| `OUT_STOCK_TAKING`      | Out       | Negative stock-taking adjustment |
| `IN_MANUAL_ADJUSTMENT`  | In        | Manual stock addition        |
| `OUT_MANUAL_ADJUSTMENT` | Out       | Manual stock reduction       |
| `OUT_SALE`              | Out       | Sold to customer             |

---

## Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Node.js 18+ (for frontend builds)
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI (for `make migrate-up`)

### Setup

```bash
# 1. Clone and enter the project
git clone <repo-url> && cd pharmacy

# 2. Create the env files (there is no .env.example)
#    - backend/config/.env is embedded into the Go binary at build time
#      (rebuild/restart after changing it)
#    - backend/.env is read by docker compose for the Postgres container
#    Keys: POSTGRES_USER, POSTGRES_DB, POSTGRES_PASSWORD, DB_HOST, DB_PORT,
#    CSRF_KEY, SESSION_SECRET

# 3. Start PostgreSQL
docker compose up -d data

# 4. Run database migrations
make migrate-up DB_URL="postgres://user:pass@localhost:5432/pharmacy?sslmode=disable"

# 5. Install frontend dependencies & build assets
cd frontend && npm install && cd ..
make assets

# 6. Create the first admin (prompts for a password, or set PHARMACY_ADMIN_PASSWORD)
cd backend && go run . create-admin -username owner && cd ..
#    With a built executable: ./<executable> create-admin -username owner
#    In Docker:               docker compose exec server /main create-admin -username owner
#    Without a terminal (scripts, `docker compose exec -T`), PHARMACY_ADMIN_PASSWORD is required

# 7. Start the server
make run
```

The server listens on **port 8000**.

### Makefile targets

| Command                     | Description                                   |
|-----------------------------|-----------------------------------------------|
| `make create-migration`     | Create a new migration file                   |
| `make migrate-up`           | Run pending migrations                        |
| `make rollback-migration`   | Roll back the last migration                  |
| `make clean-dirty-migration`| Force-mark a migration version                |
| `make tailwind`             | Rebuild TailwindCSS output                    |
| `make build-frontend`       | Build Vue frontend assets                     |
| `make assets`               | Build all UI assets (Tailwind + Vue)          |
| `make run`                  | Start the Go server                           |
| `make dev`                  | Build assets then start the server            |
| `make check-go-lines`       | Show a dry-run diff for Go lines over 100 chars |
| `make format-go-lines`      | Apply `golines` formatting to Go files          |
| `make check-go-format`      | Fail if Go files need `gofmt`                    |
| `make format-go`            | Apply standard `gofmt` formatting                |

Install `golines` before using those targets:

```sh
go install github.com/golangci/golines@latest
```

Set `GO_LINE_MAX_LEN` to change the 100-character target. The check target shows
proposed changes without editing files; run `make format-go-lines` to apply them.

---

## Project Structure

```
pharmacy/
├── backend/
│   ├── adapter/
│   │   ├── http/
│   │   │   ├── controller/       # HTTP handlers (per domain)
│   │   │   ├── helper/           # JSON response helpers
│   │   │   ├── middleware/       # Auth, logging, permissions
│   │   │   └── router/           # Route registration
│   │   └── websocket/            # WebSocket hub & handlers
│   ├── config/                   # Environment config
│   ├── db/migrations/            # 43 SQL migrations
│   ├── httperror/                # Structured error types
│   ├── internal/
│   │   ├── constant/             # App constants & permission keys
│   │   └── types/                # Request/response DTOs
│   ├── model/                    # Domain models
│   ├── repository/               # Database access layer
│   ├── service/                  # Business logic
│   ├── template/                 # next-dashboard.html, login-v2.html, error.html
│   │   └── static/               # CSS, fonts, JS, Vite dist/ (untracked)
│   └── main.go                   # Application entry point
├── frontend/
│   └── src/
│       ├── next/                 # Main SPA (views, components, composables, router)
│       ├── login/                # Login page entry
│       └── components/ui/        # shadcn-vue primitives
├── design-system/                # UI guidelines (MASTER.md)
├── compose.yml                   # Docker Compose (Postgres + Go server)
├── Dockerfile                    # Multi-stage Go build
├── Makefile                      # Development automation
└── tailwind.config.js            # TailwindCSS configuration
```

---

## API Routes

### User (`/user/`)
| Method | Path          | Description            |
|--------|---------------|------------------------|
| GET    | `/login`      | Login page             |
| POST   | `/login`      | Authenticate user      |
| GET    | `/logout`     | End session            |

All routes except `/user/*` and static assets require an authenticated session.

### Dashboard and app (`/api/`, `/app/`)
| Method | Path                    | Description                          |
|--------|-------------------------|--------------------------------------|
| GET    | `/api/dashboard`        | Dashboard JSON data                  |
| GET    | `/api/permissions/me`   | Current user's permissions           |
| GET    | `/app/*`                | SPA shell (client-side routing)      |

### Sales (`/sales/`)
| Method | Path                | Description                |
|--------|---------------------|----------------------------|
| POST   | `/sales/`           | Create a new sale          |
| GET    | `/sales/filter`     | Filter sales (JSON)        |
| POST   | `/sales/hold`       | Hold a sale                |
| GET    | `/sales/api/held`   | Held sales (JSON)          |
| DELETE | `/sales/held/{ref}` | Delete a held sale         |
| POST   | `/sales/returns`    | Return items               |

### Inventory (`/inventory/`)
| Method | Path                                    | Description                   |
|--------|-----------------------------------------|-------------------------------|
| POST   | `/inventory/add-item`                   | Create product                |
| GET    | `/inventory/item-list`                  | Inventory JSON data           |
| GET    | `/inventory/search`                     | Product search                |
| GET    | `/inventory/top-selling`                | Top-selling products          |
| GET    | `/inventory/suppliers/search`           | Supplier search               |
| GET    | `/inventory/report/stock`               | Export stock CSV              |
| GET    | `/inventory/product/{id}`               | Product details (JSON)        |
| PUT    | `/inventory/product/{id}`               | Update product (requires `inventory:edit`) |
| POST   | `/inventory/receive-items`              | Submit supplier delivery      |
| POST   | `/inventory/receive-items/hold`         | Hold receiving form           |
| GET    | `/inventory/receive-items/held/api`     | Held receipts (JSON)          |
| DELETE | `/inventory/receive-items/held/{ref}`   | Delete held receipt           |
| GET    | `/inventory/received-items-history/api` | Received items history (JSON) |
| GET    | `/inventory/purchase-orders`            | List purchase orders (`?view=active` or `history`) |
| POST   | `/inventory/purchase-orders`            | Create or edit a purchase order (needs `purchaseorder:draft` or `:approve`; a drafter's order starts as a draft, a manager's as open) |
| GET    | `/inventory/purchase-orders/pending-count` | Orders waiting for approval and drafts in progress (approve) |
| POST   | `/inventory/purchase-orders/{id}/submit` | Author submits a draft for approval |
| POST   | `/inventory/purchase-orders/{id}/approve` | Approve a draft or submitted order (approve) |
| POST   | `/inventory/purchase-orders/{id}/reject` | Reject a draft or submitted order with a reason (approve) |
| GET    | `/inventory/purchase-orders/{id}`       | Purchase order details        |
| POST   | `/inventory/purchase-orders/{id}/close` | Close an open or partial order (approve) |
| POST   | `/inventory/purchase-orders/{id}/move-outstanding` | Move unfilled items to a new order for another supplier (approve) |

### Stock Taking (`/stock-taking/`)
| Method | Path                                       | Description                              |
|--------|--------------------------------------------|------------------------------------------|
| GET    | `/stock-taking/api/list`                   | List all sessions                        |
| POST   | `/stock-taking/api/create`                 | Create a new session                     |
| GET    | `/stock-taking/api/{id}`                   | Session data (JSON)                      |
| POST   | `/stock-taking/api/{id}`                   | Complete session (`stock:complete`)      |
| GET    | `/stock-taking/api/{id}/items`             | Session items (JSON)                     |
| POST   | `/stock-taking/api/{id}/item/{product_id}` | Update item count                        |

### Admin (`/admin/`, requires `admin:access`)
| Method | Path                                            | Description                  |
|--------|-------------------------------------------------|------------------------------|
| GET    | `/admin/api/modules`                            | List modules                 |
| GET/POST | `/admin/api/permissions`                      | List / create permissions    |
| DELETE | `/admin/api/permissions/{id}`                   | Delete a permission          |
| POST/DELETE | `/admin/api/permissions/{id}/role/{roleId}` | Assign / remove from a role |
| GET    | `/admin/api/roles`                              | List roles                   |
| GET/POST | `/admin/api/users`                            | List / add users             |
| POST   | `/admin/api/users/{id}/role`                    | Change a user's role         |
| POST   | `/admin/api/users/{id}/reset-password`          | Reset a user's password      |
| GET/POST | `/admin/api/categories`                       | List / create categories     |
| PUT/DELETE | `/admin/api/categories/{id}`                | Update / delete a category   |

### WebSocket
| Path | Description                                          |
|------|------------------------------------------------------|
| `/ws?stockTakingId=N` | Real-time stock-taking updates via WebSocket |

---

## Development

### Adding a migration

```bash
make create-migration MIGRATION_NAME=add_widget_table
```

### Running tests

```bash
cd backend && go test ./service/ -v
```

The database integration tests (migrations and the purchase order lifecycle) are skipped unless
`PHARMACY_TEST_DSN` points at a Postgres server where the user may create databases. Each test
creates its own throwaway database, runs the real migrations in it and drops it afterwards:

```bash
cd backend && PHARMACY_TEST_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable" go test ./...
```

### Building frontend assets separately

```bash
make tailwind          # Rebuild Tailwind only
make build-frontend    # Rebuild Vue only
```

---

## Outstanding Items

- [ ] Modal on sales receipt after save when change > 0
- [ ] Stock taking (partially implemented — backend services & frontend UIs exist but may need finishing touches)
- [ ] Purchase orders (in progress on `feat/purchase-orders`)

---

## License

Private — internal use.
