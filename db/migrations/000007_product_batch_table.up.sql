CREATE TABLE IF NOT EXISTS product_batch (
    id integer generated always as identity unique,
    product_id integer not null references product(id) on delete CASCADE,
    price_id integer not null references product_price(id),
    receiving_batch_id integer not null references receiving_batch(id),
    quantity integer not null,
    cost_price integer not null,
    expiry_date date,
    batch_no varchar,
    created_at timestamp default current_timestamp
);