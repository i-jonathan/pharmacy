CREATE TABLE IF NOT EXISTS users (
    id integer generated always as identity unique,
    username varchar(50) unique not null,
    password text not null,
    created_at timestamp default current_timestamp
)