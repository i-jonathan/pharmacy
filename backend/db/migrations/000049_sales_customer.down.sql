DROP INDEX IF EXISTS sales_customer_idx;
ALTER TABLE sales DROP COLUMN IF EXISTS customer_id;
