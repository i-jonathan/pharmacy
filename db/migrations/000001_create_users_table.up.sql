CREATE TABLE IF NOT EXISTS users (
    id integer generated always as identity,
    username varchar(50) unique not null,
    password text not null
)