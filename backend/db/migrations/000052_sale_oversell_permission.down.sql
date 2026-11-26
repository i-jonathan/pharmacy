DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource = 'Sale' AND action = 'Oversell'
);
DELETE FROM permissions WHERE resource = 'Sale' AND action = 'Oversell';
