CREATE TABLE IF NOT EXISTS product (
    id integer generated always as identity unique,
    name varchar(255) not null,
    barcode varchar(255),
    unit varchar(50),
    category_id integer references category(id) not null,
    reorder_level integer default 5,
    manufacturer varchar(255),
    cost_price integer,
    created_at timestamp default current_timestamp
);
