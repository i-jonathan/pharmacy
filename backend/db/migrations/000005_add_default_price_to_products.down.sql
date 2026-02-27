ALTER TABLE product
DROP CONSTRAINT IF EXISTS fk_default_price_id;

ALTER TABLE product
DROP COLUMN IF EXISTS default_price_id;
