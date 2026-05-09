DELETE FROM role_permissions rp
USING permissions p
WHERE rp.permission_id = p.id
  AND p.resource = 'Admin' AND p.action = 'Access';

DELETE FROM permissions WHERE resource = 'Admin' AND action = 'Access';
