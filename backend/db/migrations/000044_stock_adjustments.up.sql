-- Manual stock adjustments: a manager corrects stock outside sales, receiving and stock taking.
-- Each adjustment records why, who and when; its items record before and after. Rows are never
-- edited: a mistake is corrected with a new adjustment. Stock itself still comes from
-- stock_movement, whose reference_id points at the adjustment.
CREATE TABLE stock_adjustment (
    id BIGSERIAL PRIMARY KEY,
    reason TEXT NOT NULL CONSTRAINT stock_adjustment_reason_check CHECK (reason IN (
        'expired', 'damaged', 'theft_loss', 'count_correction', 'opening_stock', 'internal_use',
        'other'
    )),
    note TEXT NOT NULL DEFAULT '',
    created_by_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    idempotency_key VARCHAR(100)
);

-- A retry key belongs to the person who sent it: the same key from someone else is a different
-- adjustment, never a retry of theirs.
CREATE UNIQUE INDEX stock_adjustment_idempotency_key_unique
    ON stock_adjustment (created_by_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
CREATE INDEX stock_adjustment_created_idx ON stock_adjustment (created_at DESC);

-- quantity_before and quantity_change are in base units; unit_cost_kobo is the product's cost
-- price at the time, kept for write-off value reports.
CREATE TABLE stock_adjustment_item (
    id BIGSERIAL PRIMARY KEY,
    stock_adjustment_id BIGINT NOT NULL REFERENCES stock_adjustment(id),
    product_id INTEGER NOT NULL REFERENCES product(id),
    quantity_before INTEGER NOT NULL,
    quantity_change INTEGER NOT NULL,
    unit_cost_kobo INTEGER NOT NULL DEFAULT 0,
    expiry_before DATE,
    expiry_after DATE,
    CONSTRAINT stock_adjustment_item_changes_something
        CHECK (quantity_change <> 0 OR expiry_after IS NOT NULL),
    UNIQUE (stock_adjustment_id, product_id)
);

CREATE INDEX stock_adjustment_item_product_idx ON stock_adjustment_item (product_id);

-- A cashier flags a product for a manager to check; the manager adjusts (resolving the flag) or
-- dismisses it.
CREATE TABLE stock_flag (
    id BIGSERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES product(id),
    note TEXT NOT NULL CHECK (btrim(note) <> ''),
    counted_quantity INTEGER CHECK (counted_quantity >= 0),
    -- The system stock when the flag was raised, so the count can be read against what moved since.
    stock_at_flag INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'open'
        CONSTRAINT stock_flag_status_check CHECK (status IN ('open', 'resolved', 'dismissed')),
    flagged_by_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_by_id INTEGER REFERENCES users(id),
    resolved_at TIMESTAMPTZ,
    resolution_note TEXT NOT NULL DEFAULT '',
    stock_adjustment_id BIGINT REFERENCES stock_adjustment(id)
);

CREATE INDEX stock_flag_open_idx ON stock_flag (created_at DESC) WHERE status = 'open';

-- One open flag per person per product: a double click or a repeat before a manager has looked
-- doesn't stack up. Someone else can still flag the same product.
CREATE UNIQUE INDEX stock_flag_one_open_per_user ON stock_flag (product_id, flagged_by_id)
    WHERE status = 'open';

-- Stock/Adjust: correct stock (managers). Stock/Flag: flag a product for checking (everyone on
-- the floor).
INSERT INTO permissions (resource, action)
SELECT 'Stock', a FROM unnest(ARRAY['Adjust', 'Flag']) a
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'Stock' AND action = a);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p
  ON p.resource = 'Stock'
 AND ((r.name IN ('Manager', 'Admin') AND p.action IN ('Adjust', 'Flag'))
   OR (r.name = 'Cashier' AND p.action = 'Flag'))
WHERE NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
