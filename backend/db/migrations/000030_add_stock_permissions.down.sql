DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions
    WHERE resource = 'Stock' AND action IN ('View', 'Complete')
)
AND role_id IN (
    SELECT id FROM roles WHERE name IN ('Admin', 'Manager')
);

DELETE FROM permissions
WHERE resource = 'Stock' AND action IN ('View', 'Complete');
