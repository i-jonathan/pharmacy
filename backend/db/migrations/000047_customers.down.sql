DELETE FROM role_permissions rp
USING permissions p
WHERE rp.permission_id = p.id
  AND p.resource = 'Customer' AND p.action IN ('View', 'Manage');

DELETE FROM permissions WHERE resource = 'Customer' AND action IN ('View', 'Manage');

DROP TABLE IF EXISTS customer;
