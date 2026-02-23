ALTER TABLE stock_taking
    DROP CONSTRAINT IF EXISTS fk_stock_taking_completed_by,
    DROP COLUMN IF EXISTS completed_by_id;
