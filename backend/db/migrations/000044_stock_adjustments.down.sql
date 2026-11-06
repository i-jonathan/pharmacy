DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource = 'Stock' AND action IN ('Adjust', 'Flag')
);
DELETE FROM permissions WHERE resource = 'Stock' AND action IN ('Adjust', 'Flag');

-- The stock movements written by adjustments stay: they are part of the stock history and are
-- still counted. Only the records explaining them go.
DROP TABLE IF EXISTS stock_flag;
DROP TABLE IF EXISTS stock_adjustment_item;
DROP TABLE IF EXISTS stock_adjustment;
