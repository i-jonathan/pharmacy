CREATE TABLE IF NOT EXISTS product_price (
    id integer generated always as identity unique,
    product_id integer not null references product(id) on delete CASCADE,
    quantity_per_unit integer not null,
    selling_price integer not null,
    created_at timestamp default current_timestamp
);
