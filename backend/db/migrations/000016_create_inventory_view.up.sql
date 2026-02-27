-- Create inventory_view to consolidate product, category, pricing, expiry, and stock

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
    MIN(pb.expiry_date) AS earliest_expiry,  -- soonest expiry
    COALESCE(SUM(
        CASE
            WHEN sm.movement_type LIKE 'IN%'  THEN sm.quantity
            WHEN sm.movement_type LIKE 'OUT%' THEN -sm.quantity
            ELSE 0
        END
    ), 0) AS stock
FROM product p
LEFT JOIN category c ON p.category_id = c.id
LEFT JOIN product_price pp ON p.default_price_id = pp.id
LEFT JOIN product_batch pb ON pb.product_id = p.id
LEFT JOIN stock_movement sm ON sm.product_id = p.id
GROUP BY
    p.id, p.name, p.category_id, c.name,
    p.default_price_id, pp.selling_price,
    p.reorder_level, p.manufacturer;
