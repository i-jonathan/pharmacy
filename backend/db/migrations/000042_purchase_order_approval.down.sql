DROP INDEX IF EXISTS purchase_order_created_by_idx;

ALTER TABLE purchase_order DROP COLUMN IF EXISTS approved_at;
ALTER TABLE purchase_order DROP COLUMN IF EXISTS approved_by_id;

-- Orders that never got approved cannot exist under the old rules; close them.
UPDATE purchase_order SET status = 'closed' WHERE status IN ('draft', 'pending_approval');

ALTER TABLE purchase_order DROP CONSTRAINT purchase_order_status_check;
ALTER TABLE purchase_order ADD CONSTRAINT purchase_order_status_check
    CHECK (status IN ('open', 'partial', 'received', 'closed'));
