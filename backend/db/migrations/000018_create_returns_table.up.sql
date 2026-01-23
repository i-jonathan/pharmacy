CREATE TABLE IF NOT EXISTS returns (
    id integer generated always as identity unique,
    sale_id integer not null references sales(id),
    cashier_id integer references users(id),
    total_refunded integer not null,
    notes text,
    created_at timestamptz default now()
)