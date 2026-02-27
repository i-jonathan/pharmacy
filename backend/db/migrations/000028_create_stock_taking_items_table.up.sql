create table if not exists stock_taking_item (
    id integer generated always as identity primary key,
    stock_taking_id integer not null references stock_taking(id) on delete CASCADE,
    product_id integer not null references product(id),
    snapshot_quantity integer not null,
    dispensary_count integer,
    store_count integer,
    notes text,
    last_updated_by_id integer references users(id),
    last_updated_at timestamptz,
    created_at timestamptz not null default now(),

    CONSTRAINT uq_stock_taking_item
        UNIQUE (stock_taking_id, product_id)
)
