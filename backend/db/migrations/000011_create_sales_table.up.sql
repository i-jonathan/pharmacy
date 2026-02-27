CREATE TABLE IF NOT EXISTS sales (
    id integer generated always as identity unique,
    receipt_number varchar(50) unique not null,
    cashier_id integer references users(id),
    subtotal integer not null,
    discount integer default 0,
    total integer not null,
    status varchar(50) not null,
    created_at timestamp default current_timestamp
);
