CREATE TABLE IF NOT EXISTS return_items (
    id integer generated always as identity unique,
    return_id integer references returns(id),
    sale_item_id integer references sales_item(id),
    quantity integer not null,
    created_at timestamptz default now()
)
