ALTER TABLE product
ADD COLUMN current_expiry DATE NULL;


UPDATE product p
SET current_expiry = sub.min_expiry
FROM (
    SELECT product_id, MIN(expiry_date) AS min_expiry
    FROM product_batch
    WHERE expiry_date IS NOT NULL
    GROUP BY product_id
) AS sub
WHERE p.id = sub.product_id;

CREATE INDEX idx_product_current_expiry ON product(current_expiry);
