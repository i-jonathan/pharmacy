CREATE TABLE permissions (
   id integer generated always as identity unique,
   resource varchar(50) not null,
   action varchar(50) not null,
   created_at timestamptz default now(),
   UNIQUE(resource, action)
)
