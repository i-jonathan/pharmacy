CREATE TABLE IF NOT EXISTS receiving_batch (
    id integer generated always as identity unique,
    supplier_name varchar(255) not null,
    received_by_id integer references users(id),
    note text,
    created_at timestamp default current_timestamp
)