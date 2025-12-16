CREATE TABLE roles (
    id integer generated always as identity unique,
    name varchar(50) unique not null,
    created_at timestamptz default now()
)
