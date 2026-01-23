CREATE TABLE IF NOT EXISTS category (
    id integer generated always as identity unique,
    name varchar(40) not null,
    created_at timestamp default current_timestamp
)