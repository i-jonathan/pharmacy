# Pharmacy POS — Project Overview

## What It Is

This is a **pharmacy point-of-sale (POS) and inventory management system** — a full-stack web application designed for day-to-day operations of a physical pharmacy store. Prices are in **Nigerian Naira** (stored as kobo / integer cents).

---

## Tech Stack

| Layer | Technology |
|---|---|
| **Backend** | Go 1.24, `net/http` (stdlib), Gorilla sessions/CSRF |
| **Database** | PostgreSQL (via `pgx` + `sqlx`), 31 sequential migrations |
| **Templating** | Go `html/template` (server-rendered HTML) |
| **Frontend interactivity** | Vue 3 (via Vite, built into `static/dist/` and embedded) |
| **Styling** | TailwindCSS (compiled to `tailwind.css`) |
| **Auth** | Argon2id password hashing, cookie sessions (`gorilla/sessions`) |
| **Infrastructure** | Docker, Docker Compose (Go app + Postgres), Makefile |

---

## Architecture

The backend follows a **clean/layered architecture**:

```
main.go                       ← wires everything, embeds templates + static assets
├── config/                   ← env-based config (cleanenv + godotenv)
├── model/                    ← domain models (Product, Sale, User, StockTaking…)
├── repository/               ← database access (SQL queries via sqlx)
│   └── store.go              ← initialises the DB connection pool
├── service/                  ← business logic layer
│   └── interfaces.go         ← defines service contracts
├── adapter/http/
│   ├── router/               ← route registration (per-domain: user, inventory, sales, stock-taking)
│   ├── controller/           ← HTTP handlers that call services & render templates
│   ├── middleware/            ← auth, logging, CSRF
│   └── helper/               ← HTTP helper utilities
├── template/                 ← Go HTML templates (server-rendered pages)
│   └── static/               ← CSS, JS, Vite-built Vue bundles
├── internal/                 ← internal types (request/response DTOs)
└── db/migrations/            ← 31 Postgres migrations (golang-migrate)
```

The **frontend** directory is a Vue 3 + Vite project. It builds into `backend/template/static/dist/`, and the Go server embeds & serves those assets. The Vite manifest is parsed at startup so templates can reference hashed asset filenames via `ViteAsset`.

---

## Core Features

### 1. Sales / POS Receipt
- Product search (prefix-based) with prices displayed
- Add items to cart with quantity & pricing options
- Apply discounts, accept multiple payment methods
- Auto-generate receipt numbers
- **Print receipt** functionality
- **Hold sale** — park an in-progress sale and recall it later
- Navigation warning when cart is non-empty

### 2. Sales History
- Browse completed sales with filters
- Daily totals
- Drill into individual receipts

### 3. Returns
- Accept returns against previous sales
- Refund tracking with notes

### 4. Inventory Management
- Product catalog with categories, manufacturers, barcodes
- Multiple price options per product (e.g. tablet vs. pack)
- Default price tracking
- Reorder level tracking
- Expiry date tracking (earliest expiry per product)
- Stock movement ledger (receives, sales, returns)

### 5. Receiving / Supplier Deliveries
- Record incoming product batches by supplier
- Batch numbers and expiry dates
- Cost price capture per batch

### 6. Stock Taking
- Create named stock-taking sessions
- Snapshot current system quantities
- Record physical counts (dispensary + store)
- Track who last updated each count
- Complete/cancel sessions
- Role-based permissions for stock-taking

### 7. Users & Auth
- User registration with username/password validation
- Argon2id password hashing
- Session-based authentication with middleware
- **Role-based access control** (admin, cashier roles with granular permissions)

---

## Database Schema Highlights

31 migrations covering:
- `users`, `roles`, `permissions`, `role_permissions`
- `categories`, `products`, `product_prices`
- `receiving_batches`, `product_batches`
- `stock_movements` (ledger)
- `sales`, `sale_items`, `sale_payments`
- `returns`, `return_items`
- `held_transactions`
- `stock_takings`, `stock_taking_items`
- An `inventory_view` (materialised/view) for aggregated stock + earliest expiry

---

## Running the App

```bash
# Start Postgres
docker compose up -d data

# Run migrations
make migrate-up DB_URL="postgres://..."

# Build frontend assets
make assets    # runs Tailwind + Vite build

# Start server
make run       # or: make dev (assets + run)
```

The server listens on **port 8000**.

---

## Outstanding TODO Items

- [ ] Modal on sales receipt after save when change > 0
- [ ] Stock taking (partially implemented — UI & backend services exist)
