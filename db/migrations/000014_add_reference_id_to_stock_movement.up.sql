ALTER TABLE stock_movement
ADD COLUMN reference_id integer;

UPDATE stock_movement
SET reference_id = batch_id;

ALTER TABLE stock_movement
ALTER COLUMN reference_id SET NOT NULL;
