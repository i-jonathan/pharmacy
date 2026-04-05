DELETE FROM permissions p
WHERE p.resource = 'Inventory'
    AND p.action = 'Edit'
    AND NOT EXISTS (
        SELECT 1 FROM role_permissions rp
        WHERE rp.permission_id = p.id
    )
;
