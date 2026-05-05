CREATE OR REPLACE VIEW inventory_view AS
WITH stock_by_product AS (
    SELECT
        product_id,
        COALESCE(SUM(
            CASE
                WHEN movement_type LIKE 'IN%' THEN quantity
                WHEN movement_type LIKE 'OUT%' THEN -quantity
                ELSE 0
            END
        ), 0) AS stock
    FROM stock_movement
    GROUP BY product_id
),
expiry_by_product AS (
    SELECT
        product_id,
        MIN(expiry_date) AS earliest_expiry
    FROM product_batch
    GROUP BY product_id
)
SELECT
    p.id,
    p.name,
    p.category_id,
    c.name AS category,
    p.default_price_id,
    pp.selling_price AS default_price,
    p.reorder_level,
    p.manufacturer,
    e.earliest_expiry,
    COALESCE(s.stock, 0) AS stock
FROM product p
LEFT JOIN category c ON p.category_id = c.id
LEFT JOIN product_price pp ON p.default_price_id = pp.id
LEFT JOIN expiry_by_product e ON e.product_id = p.id
LEFT JOIN stock_by_product s ON s.product_id = p.id;
