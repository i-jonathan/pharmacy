-- A retry key is used once, by one save. request_hash fingerprints what was sent under it, so the
-- service can tell a resend of the same save (answered with it) from anything else (refused).
ALTER TABLE stock_adjustment ADD COLUMN IF NOT EXISTS request_hash VARCHAR(64);

-- Migration 44 allowed one key per user, so two users may share a key. Keep it on the earliest
-- save; the later ones are complete adjustments and simply lose their (now unusable) key.
UPDATE stock_adjustment later SET idempotency_key = NULL
WHERE idempotency_key IS NOT NULL
    AND EXISTS (
        SELECT 1 FROM stock_adjustment earlier
        WHERE earlier.idempotency_key = later.idempotency_key AND earlier.id < later.id
    );

DROP INDEX IF EXISTS stock_adjustment_idempotency_key_unique;
CREATE UNIQUE INDEX stock_adjustment_idempotency_key_unique
    ON stock_adjustment (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
