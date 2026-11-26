-- Sale/Oversell: complete a sale that takes a product's stock below zero. Without it the sale is
-- refused; every oversell is audited and opens a stock flag. Managers and admins get it.
INSERT INTO permissions (resource, action)
SELECT 'Sale', 'Oversell'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'Sale' AND action = 'Oversell'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.resource = 'Sale' AND p.action = 'Oversell'
WHERE r.name IN ('Manager', 'Admin')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
