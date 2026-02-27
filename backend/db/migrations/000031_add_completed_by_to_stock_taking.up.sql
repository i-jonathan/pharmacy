ALTER TABLE stock_taking
    ADD COLUMN completed_by_id INTEGER,
    ADD CONSTRAINT fk_stock_taking_completed_by
    FOREIGN KEY (completed_by_id)
    REFERENCES users(id) ON DELETE SET NULL;
