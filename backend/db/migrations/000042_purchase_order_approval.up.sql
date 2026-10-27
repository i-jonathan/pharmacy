-- An order can now be written as a draft by one person and approved by another.
--   draft             being written; its author can still change it
--   pending_approval  the author is done and has submitted it for a manager to approve
-- Approving turns either into 'open', the state orders started in before.
ALTER TABLE purchase_order DROP CONSTRAINT purchase_order_status_check;
ALTER TABLE purchase_order ADD CONSTRAINT purchase_order_status_check
    CHECK (status IN ('draft', 'pending_approval', 'open', 'partial', 'received', 'closed'));

ALTER TABLE purchase_order ADD COLUMN approved_by_id INTEGER REFERENCES users(id);
ALTER TABLE purchase_order ADD COLUMN approved_at TIMESTAMPTZ;

-- Every order that exists now was effectively approved when it was created: there was no
-- approval step. Recording that keeps "has this order ever been approved" answerable for all rows.
UPDATE purchase_order SET approved_by_id = created_by_id, approved_at = created_at;

-- People without approval rights only see their own orders.
CREATE INDEX purchase_order_created_by_idx ON purchase_order(created_by_id, created_at DESC);
