CREATE TABLE IF NOT EXISTS stock_movement (
    id integer generated always as identity unique,
    product_id integer not null references product(id) on delete CASCADE,
    batch_id integer not null references product_batch(id) on delete CASCADE,
    movement_type varchar(10) not null,
    quantity integer not null,
    created_at timestamp default current_timestamp
)
