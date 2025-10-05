CREATE TABLE held_transactions (
    id integer generated always as identity primary key unique,
    type varchar(20) not null,
    reference text unique,
    payload jsonb not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
)
