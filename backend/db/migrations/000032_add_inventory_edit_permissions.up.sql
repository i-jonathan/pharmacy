INSERT INTO permissions (resource, action)
SELECT 'Inventory', 'Edit'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'Inventory' AND action = 'Edit'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON (p.resource = 'Inventory' AND p.action = 'Edit')
WHERE r.name IN ('Admin', 'Manager')
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
