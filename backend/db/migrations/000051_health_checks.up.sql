-- Whether the customer agreed to the pharmacy keeping their health readings, and when that last
-- changed. Separate from contact_consent, which is about being contacted. Who changed it is in the
-- audit log.
ALTER TABLE customer ADD COLUMN IF NOT EXISTS health_consent BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE customer ADD COLUMN IF NOT EXISTS health_consent_updated_at TIMESTAMPTZ;

-- One blood pressure and/or blood glucose reading taken for a customer. Walk-ins are never stored.
-- taken_at is set by the server so a reading can't be backdated. The ranges catch typos such as
-- 1200/80; they are not clinical limits and nothing here judges a number. Glucose is kept in the
-- unit it was measured in, never converted. version changes with every correction; who corrected
-- what, from and to, is in the audit log.
CREATE TABLE health_check (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL REFERENCES customer(id),
    taken_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    taken_by_id INTEGER NOT NULL REFERENCES users(id),
    systolic INTEGER,
    diastolic INTEGER,
    pulse INTEGER,
    glucose NUMERIC(4, 1),
    glucose_unit TEXT,
    glucose_context TEXT,
    note TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1,
    idempotency_key TEXT,
    request_hash VARCHAR(64),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Systolic and diastolic come together; a pulse comes off the same cuff.
    CONSTRAINT health_check_bp_pair CHECK ((systolic IS NULL) = (diastolic IS NULL)),
    CONSTRAINT health_check_pulse_with_bp CHECK (pulse IS NULL OR systolic IS NOT NULL),
    CONSTRAINT health_check_bp_order CHECK (systolic IS NULL OR systolic > diastolic),
    CONSTRAINT health_check_systolic_range CHECK (systolic BETWEEN 50 AND 300),
    CONSTRAINT health_check_diastolic_range CHECK (diastolic BETWEEN 30 AND 200),
    CONSTRAINT health_check_pulse_range CHECK (pulse BETWEEN 20 AND 250),
    -- Glucose, its unit and when it was taken come together.
    CONSTRAINT health_check_glucose_complete CHECK (
        (glucose IS NULL) = (glucose_unit IS NULL)
        AND (glucose IS NULL) = (glucose_context IS NULL)
    ),
    CONSTRAINT health_check_glucose_unit CHECK (glucose_unit IN ('mmol/L', 'mg/dL')),
    CONSTRAINT health_check_glucose_context CHECK (
        glucose_context IN ('fasting', 'after_meal', 'random')
    ),
    CONSTRAINT health_check_glucose_range CHECK (
        glucose IS NULL
        OR (glucose_unit = 'mmol/L' AND glucose BETWEEN 1.0 AND 50.0)
        OR (glucose_unit = 'mg/dL' AND glucose BETWEEN 20 AND 900)
    ),
    CONSTRAINT health_check_something_measured CHECK (
        systolic IS NOT NULL OR glucose IS NOT NULL
    ),
    CONSTRAINT health_check_note_length CHECK (char_length(note) <= 500)
);

-- A retry key is used once, by one save; request_hash fingerprints what was sent under it.
CREATE UNIQUE INDEX health_check_idempotency_key_unique
    ON health_check (idempotency_key)
    WHERE idempotency_key IS NOT NULL;
CREATE INDEX health_check_customer_idx ON health_check (customer_id, taken_at DESC, id DESC);

-- HealthCheck/Record: record a reading and print its slip. HealthCheck/View: see a customer's
-- readings and reprint one. HealthCheck/Correct: correct a saved reading.
INSERT INTO permissions (resource, action)
SELECT 'HealthCheck', a FROM unnest(ARRAY['Record', 'View', 'Correct']) a
WHERE NOT EXISTS (SELECT 1 FROM permissions WHERE resource = 'HealthCheck' AND action = a);

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.resource = 'HealthCheck'
WHERE (r.name IN ('Manager', 'Admin') OR (r.name = 'Cashier' AND p.action = 'Record'))
  AND NOT EXISTS (
    SELECT 1 FROM role_permissions rp
    WHERE rp.role_id = r.id AND rp.permission_id = p.id
);
