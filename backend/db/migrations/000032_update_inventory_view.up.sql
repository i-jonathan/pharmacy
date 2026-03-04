CREATE OR REPLACE VIEW inventory_view AS
SELECT
    p.id,
    p.name,
    p.category_id,
    c.name AS category,
    p.default_price_id,
    pp.selling_price AS default_price,
    p.reorder_level,
    p.manufacturer,
    p.current_expiry AS earliest_expiry,
    COALESCE(sm.stock, 0) AS stock
FROM product p
LEFT JOIN category c ON p.category_id = c.id
LEFT JOIN product_price pp ON p.default_price_id = pp.id
LEFT JOIN (
    SELECT
        product_id,
        COALESCE(SUM(
            CASE
                WHEN movement_type LIKE 'IN%'  THEN quantity
                WHEN movement_type LIKE 'OUT%' THEN -quantity
                ELSE 0
            END
        ), 0) AS stock
    FROM stock_movement
    GROUP BY product_id
) sm ON sm.product_id = p.id;