-- PurchaseOrder/Draft lets a user write purchase order drafts and submit them for approval.
-- PurchaseOrder/Approve is for managers: create orders directly, edit any draft, approve or
-- reject, and close or move approved orders.
INSERT INTO permissions (resource, action)
SELECT 'PurchaseOrder', 'Draft'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'PurchaseOrder' AND action = 'Draft'
);

INSERT INTO permissions (resource, action)
SELECT 'PurchaseOrder', 'Approve'
WHERE NOT EXISTS (
    SELECT 1 FROM permissions WHERE resource = 'PurchaseOrder' AND action = 'Approve'
);

-- Cashiers draft; managers and admins approve.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON p.resource = 'PurchaseOrder'
 AND ((r.name = 'Cashier' AND p.action = 'Draft')
   OR (r.name IN ('Manager', 'Admin') AND p.action = 'Approve'))
WHERE NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
