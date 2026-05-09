INSERT INTO permissions (resource, action)
SELECT 'Admin', 'Access'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'Admin' AND action = 'Access'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'Admin'
  AND p.resource = 'Admin' AND p.action = 'Access'
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
  );
