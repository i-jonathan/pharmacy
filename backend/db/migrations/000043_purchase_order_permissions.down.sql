DELETE FROM role_permissions rp
USING permissions p
WHERE rp.permission_id = p.id
  AND p.resource = 'PurchaseOrder' AND p.action IN ('Draft', 'Approve');

DELETE FROM permissions WHERE resource = 'PurchaseOrder' AND action IN ('Draft', 'Approve');
