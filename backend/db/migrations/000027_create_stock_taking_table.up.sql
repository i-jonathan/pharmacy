CREATE TABLE IF NOT EXISTS stock_taking (
    id integer generated always as identity primary key,
    name text not null,
    status varchar(20) not null
        check (status IN ('in_progress', 'completed', 'cancelled')),
    started_at timestamptz not null default now(),
    completed_at timestamptz,
    created_by_id integer references users(id),
    created_at timestamptz not null default now()
)
