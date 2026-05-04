ALTER TABLE sales
ADD COLUMN IF NOT EXISTS idempotency_key varchar(100);

CREATE UNIQUE INDEX IF NOT EXISTS sales_idempotency_key_unique
ON sales (idempotency_key)
WHERE idempotency_key IS NOT NULL;

ALTER TABLE receiving_batch
ADD COLUMN IF NOT EXISTS idempotency_key varchar(100);

CREATE UNIQUE INDEX IF NOT EXISTS receiving_batch_idempotency_key_unique
ON receiving_batch (idempotency_key)
WHERE idempotency_key IS NOT NULL;
