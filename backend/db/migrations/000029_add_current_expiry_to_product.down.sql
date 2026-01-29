DROP INDEX IF EXISTS idx_product_current_expiry;

ALTER TABLE product
DROP COLUMN IF EXISTS current_expiry;
