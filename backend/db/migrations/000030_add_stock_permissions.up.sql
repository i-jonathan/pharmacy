INSERT INTO permissions (resource, action)
SELECT 'Stock', 'View'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'Stock' AND action = 'View'
);

INSERT INTO permissions (resource, action)
SELECT 'Stock', 'Complete'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'Stock' AND action = 'Complete'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON (p.resource = 'Stock' AND p.action IN ('View', 'Complete'))
WHERE r.name IN ('Admin', 'Manager')
AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
