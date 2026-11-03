DROP INDEX IF EXISTS stock_adjustment_idempotency_key_unique;
CREATE UNIQUE INDEX stock_adjustment_idempotency_key_unique
    ON stock_adjustment (created_by_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;

ALTER TABLE stock_adjustment DROP COLUMN IF EXISTS request_hash;
