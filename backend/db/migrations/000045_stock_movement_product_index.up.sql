-- Stock is the sum of a product's movements; without an index every per-product sum read the
-- whole table.
CREATE INDEX IF NOT EXISTS stock_movement_product_idx ON stock_movement (product_id);
