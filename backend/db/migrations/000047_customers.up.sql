-- Customers of the pharmacy. phone is stored normalized (+234...), is empty for a customer who
-- gave none, and belongs to one customer for good: a deactivated customer keeps their number so
-- it is never given to someone else by mistake. Customers are deactivated, never deleted, because
-- sales will refer to them. version changes with every edit so two people editing the same
-- customer can't overwrite each other.
CREATE TABLE customer (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL CONSTRAINT customer_name_check CHECK (btrim(name) <> ''),
    phone TEXT,
    email TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    -- Whether the customer agreed to be contacted by the pharmacy, and when that last changed.
    -- Who changed it, and when, is in the audit log.
    contact_consent BOOLEAN NOT NULL DEFAULT FALSE,
    consent_updated_at TIMESTAMPTZ,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_by_id INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX customer_phone_unique ON customer (phone) WHERE phone IS NOT NULL;
CREATE INDEX customer_name_idx ON customer (lower(name));

-- Customer/View: see customers. Customer/Manage: add, edit, deactivate and reactivate them.
-- Cashiers get both because adding a customer at the till is their job.
INSERT INTO permissions (resource, action)
SELECT 'Customer', a FROM unnest(ARRAY['View', 'Manage']) a
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'Customer' AND action = a);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.resource = 'Customer'
WHERE r.name IN ('Cashier', 'Manager', 'Admin')
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
