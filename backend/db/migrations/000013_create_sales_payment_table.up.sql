CREATE TABLE IF NOT EXISTS sales_payment (
    id integer generated always as identity unique,
    sale_id integer not null references sales(id) on delete CASCADE,
    payment_method varchar(30) not null,
    amount integer not null,
    created_at timestamp default current_timestamp
);
