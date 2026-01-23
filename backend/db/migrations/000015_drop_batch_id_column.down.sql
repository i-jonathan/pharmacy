ALTER TABLE stock_movement
ADD COLUMN batch_id integer;

UPDATE stock_movement
SET batch_id = reference_id;
