# Primocrest Pharmacy POS

A full-stack **pharmacy point-of-sale (POS) and inventory management system** for day-to-day operations of a physical pharmacy store. Prices are in **Nigerian Naira** (stored as kobo / integer cents).

---

## Tech Stack

| Layer              | Technology                                                              |
|--------------------|-------------------------------------------------------------------------|
| **Backend**        | Go 1.24, `net/http` (standard library), Gorilla sessions / CSRF / WebSocket |
| **Database**       | PostgreSQL (via `pgx` + `sqlx`), 34 sequential migrations               |
| **Templating**     | Go `html/template` (server-rendered HTML)                               |
| **Frontend**       | Vue 3 + Vite (built into `static/dist/` and embedded)                   |
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
├── template/                 ← Go HTML templates (server-rendered pages)
│   └── static/               ← CSS, JS, Vite-built Vue bundles
├── internal/                 ← shared types (request/response DTOs) & constants
├── httperror/                ← structured HTTP error types
└── db/migrations/            ← 34 PostgreSQL migrations (golang-migrate)
```

### Frontend / Vue

The `frontend/` directory is a Vue 3 + Vite project. It builds into `backend/template/static/dist/`. The Go server embeds and serves those assets, parsing the Vite manifest at startup so templates can reference hashed filenames via the `ViteAsset` template function.

There are **three Vue entry points**:

| Entry Point                          | Mount Point                        | Purpose                          |
|--------------------------------------|------------------------------------|----------------------------------|
| `src/main.js`                        | `<div id="app">` (dashboard.html)  | Dashboard with charts            |
| `src/stock-taking/dashboard/main.js` | `<main id="stock-taking-dashboard">` | Stock-taking session list       |
| `src/stock-taking/counting/main.js`  | `<main id="stock-taking-app">`     | Live stock counting with WebSocket |

---

## Features

### 1. Sales / POS Receipt
- Product search (prefix-based, scrollable results with keyboard navigation)
- Add items to cart with quantity & price-tier selection
- Apply discounts, accept multiple payment methods (Cash, Card, Transfer)
- Auto-generated receipt numbers
- **Print receipt** functionality
- **Hold / park** a sale and recall it later
- Navigation warning when cart is non-empty
- Idempotency keys to prevent duplicate submissions

### 2. Sales History
- Browse completed sales with date-range filters
- Daily / filtered totals (role-gated permission)
- Drill into individual receipts with item and payment details
- **Returns** displayed alongside original sale items

### 3. Returns
- Accept returns against previous sales
- Quantity validation (cannot exceed sold quantity)
- Stock automatically returned to inventory
- Refund tracking with notes

### 4. Inventory Management
- Product catalog with categories, manufacturers, barcodes
- **Multiple price options** per product (e.g. tablet vs. pack)
- Default price tracking
- Reorder level monitoring
- Expiry date tracking (earliest expiry per product shown in views)
- **CSV export** of current stock
- Stock movement ledger (receives, sales, returns, stock-taking adjustments)
- Edit products (price, stock info) — requires `inventory:edit` permission

### 5. Receiving / Supplier Deliveries
- Record incoming product batches by supplier
- Batch numbers and expiry dates
- Cost price capture per batch
- Selling price configuration per batch
- **Hold** partially-completed receive forms for later
- Idempotency keys to prevent duplicate receives

### 6. Stock Taking
- Create named stock-taking sessions (one at a time)
- Auto-snapshot current system quantities
- Record physical counts in **dispensary** and **store** separately
- Track who last updated each count
- **Real-time collaboration** via WebSocket — updates broadcast to all users viewing the same session
- Automatic stock movement entries on completion
- Update product expiry during counting
- Mark sessions as completed or cancelled
- **Complete stock taking** requires `stock:complete` permission

### 7. Dashboard
- **KPI cards**: Today's total sales, transaction count, inventory items, low-stock count
- **Sales trend chart** (7-day)
- **Category breakdown** (pie chart)
- Low-stock items list with copy-to-clipboard
- Expiring items list (critical ≤ 7 days, warning ≤ 30 days) with copy-to-clipboard
- Modal views for full lists

### 8. Users & Auth
- Registration with username / password validation
- **Argon2id** password hashing
- Session-based authentication (Gorilla sessions)
- **Role-based access control** with granular permissions

### 9. Permissions

| Key                          | Description                                    |
|------------------------------|------------------------------------------------|
| `salestotal:view`            | View daily sales totals on the history page    |
| `stock:complete`             | Complete / finalise a stock-taking session     |
| `inventory:edit`             | Edit product details (prices, stock info)      |

---

## Database Schema

34 migrations covering:

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

### Setup

```bash
# 1. Clone and enter the project
git clone <repo-url> && cd pharmacy

# 2. Copy environment file
cp backend/.env.example backend/.env
# Edit backend/.env with your database credentials

# 3. Start PostgreSQL
docker compose up -d data

# 4. Run database migrations
make migrate-up DB_URL="postgres://user:pass@localhost:5432/pharmacy?sslmode=disable"

# 5. Install frontend dependencies & build assets
cd frontend && npm install && cd ..
make assets

# 6. Start the server
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
│   ├── db/migrations/            # 34 SQL migrations
│   ├── httperror/                # Structured error types
│   ├── internal/
│   │   ├── constant/             # App constants & permission keys
│   │   └── types/                # Request/response DTOs
│   ├── model/                    # Domain models
│   ├── repository/               # Database access layer
│   ├── service/                  # Business logic
│   ├── template/                 # Go HTML templates
│   │   ├── partials/             # nav.html, subnav.html
│   │   └── static/               # CSS, JS, Vite dist/
│   └── main.go                   # Application entry point
├── frontend/
│   └── src/
│       ├── components/           # Vue components (Dashboard, Charts)
│       ├── stock-taking/
│       │   ├── counting/         # Stock counting Vue app
│       │   └── dashboard/        # Stock-taking list Vue app
│       └── main.js               # Dashboard entry point
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
| GET    | `/register`   | Registration page      |
| POST   | `/register`   | Create user account    |
| GET    | `/logout`     | End session            |

### Dashboard (`/api/` and `/app/`)
| Method | Path                | Description              |
|--------|---------------------|--------------------------|
| GET    | `/api/dashboard`    | Dashboard JSON data      |
| GET    | `/app/dashboard`    | Dashboard HTML page      |

### Sales (`/sales/`)
| Method | Path               | Description                |
|--------|--------------------|----------------------------|
| GET    | `/sales/`          | Sales receipt page         |
| POST   | `/sales/`          | Create a new sale          |
| GET    | `/sales/history`   | Sales history page         |
| GET    | `/sales/filter`    | Filter sales (JSON)        |
| POST   | `/sales/hold`      | Hold a sale                |
| GET    | `/sales/held`      | Held receipts page         |
| DELETE | `/sales/held/{ref}`| Delete a held receipt      |
| POST   | `/sales/returns`   | Return items               |

### Inventory (`/inventory/`)
| Method | Path                                  | Description                   |
|--------|---------------------------------------|-------------------------------|
| POST   | `/inventory/add-item`                 | Create product                |
| GET    | `/inventory/items`                    | Inventory list page           |
| GET    | `/inventory/item-list`                | Inventory JSON data           |
| GET    | `/inventory/receive-items`            | Receive items page            |
| POST   | `/inventory/receive-items`            | Submit supplier delivery      |
| POST   | `/inventory/receive-items/hold`       | Hold receiving form           |
| GET    | `/inventory/receive-items/held`       | Held receipts page            |
| DELETE | `/inventory/receive-items/held/{ref}` | Delete held receipt           |
| GET    | `/inventory/search`                   | Product search                |
| GET    | `/inventory/suppliers/search`         | Supplier search               |
| GET    | `/inventory/report/stock`             | Export stock CSV               |
| GET    | `/inventory/product/{id}`             | Product details (JSON)        |
| PUT    | `/inventory/product/{id}`             | Update product (requires permission) |

### Stock Taking (`/stock-taking/`)
| Method | Path                                        | Description                    |
|--------|---------------------------------------------|--------------------------------|
| GET    | `/stock-taking/`                            | Dashboard page                 |
| POST   | `/stock-taking/api/create`                  | Create a new session           |
| GET    | `/stock-taking/api/list`                    | List all sessions              |
| GET    | `/stock-taking/{id}`                        | Stock counting page            |
| GET    | `/stock-taking/api/{id}`                    | Session data (JSON)            |
| POST   | `/stock-taking/api/{id}`                    | Complete session               |
| GET    | `/stock-taking/api/{id}/items`              | Session items (JSON)           |
| POST   | `/stock-taking/api/{id}/item/{product_id}` | Update item count              |

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

### Building frontend assets separately

```bash
make tailwind          # Rebuild Tailwind only
make build-frontend    # Rebuild Vue only
```

---

## Outstanding Items

- [ ] Modal on sales receipt after save when change > 0
- [ ] Stock taking (partially implemented — backend services & frontend UIs exist but may need finishing touches)

---

## License

Private — internal use.