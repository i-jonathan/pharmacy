-- Sale/ChangeOldCustomer: change or remove the customer on a sale made on an earlier day.
-- Anyone with Customer/Manage may do it for a sale made today; only managers get this one.
INSERT INTO permissions (resource, action)
SELECT 'Sale', 'ChangeOldCustomer'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'Sale' AND action = 'ChangeOldCustomer'
);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.resource = 'Sale' AND p.action = 'ChangeOldCustomer'
WHERE r.name IN ('Manager', 'Admin')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
