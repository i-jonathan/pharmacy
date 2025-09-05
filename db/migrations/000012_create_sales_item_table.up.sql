CREATE TABLE IF NOT EXISTS sales_item (
    id integer generated always as identity unique,
    sale_id integer not null references sales(id) on delete CASCADE,
    product_id integer not null references product(id),
    quantity integer not null,
    unit_price integer not null,
    discount integer default 0,
    total_price integer not null,
    created_at timestamp default current_timestamp
);
