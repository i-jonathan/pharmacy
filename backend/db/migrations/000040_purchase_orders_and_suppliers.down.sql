DROP INDEX IF EXISTS product_batch_purchase_order_item_idx;
ALTER TABLE product_batch DROP COLUMN IF EXISTS purchase_order_item_id;

DROP TABLE IF EXISTS purchase_order_item;
DROP TABLE IF EXISTS purchase_order;

DROP INDEX IF EXISTS receiving_batch_supplier_idx;
ALTER TABLE receiving_batch DROP COLUMN IF EXISTS supplier_id;

DROP TABLE IF EXISTS supplier;
