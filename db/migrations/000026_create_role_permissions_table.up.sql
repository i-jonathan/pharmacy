CREATE TABLE role_permissions (
    id integer generated always as identity unique,
    role_id integer references roles(id),
    permission_id integer references permissions(id)
)
