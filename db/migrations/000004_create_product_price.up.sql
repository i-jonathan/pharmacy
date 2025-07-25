CREATE TABLE IF NOT EXISTS product_price (
    id integer generated always as identity unique,
    product_id integer not null references product(id) on delete CASCADE,
    quantity_per_unit integer not null,
    selling_price integer not null,
    is_default boolean default false,
    created_at timestamp default current_timestamp
);

CREATE UNIQUE INDEX one_default_price_per_product
ON product_price(product_id)
WHERE is_default = true;