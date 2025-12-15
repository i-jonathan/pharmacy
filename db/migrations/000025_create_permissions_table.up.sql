CREATE TABLE permissions (
   id integer generated always as identity unique,
   resource varchar(50) not null,
   action varchar(50) not null,
   UNIQUE(resource, action)
)
