DROP INDEX IF EXISTS receiving_batch_idempotency_key_unique;

ALTER TABLE receiving_batch
DROP COLUMN IF EXISTS idempotency_key;

DROP INDEX IF EXISTS sales_idempotency_key_unique;

ALTER TABLE sales
DROP COLUMN IF EXISTS idempotency_key;
