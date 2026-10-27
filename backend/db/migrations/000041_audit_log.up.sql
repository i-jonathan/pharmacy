-- General-purpose, append-only record of who changed what and when. details holds the
-- action-specific facts (quantities, before and after values, linked records).
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    actor_id INTEGER REFERENCES users(id),
    entity_type TEXT NOT NULL,
    entity_id BIGINT NOT NULL,
    action TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    details JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX audit_log_entity_idx ON audit_log(entity_type, entity_id, occurred_at);
CREATE INDEX audit_log_actor_idx ON audit_log(actor_id, occurred_at);

CREATE FUNCTION audit_log_append_only() RETURNS trigger AS $$
BEGIN
    RAISE EXCEPTION 'audit_log is append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_log_no_change
    BEFORE UPDATE OR DELETE ON audit_log
    FOR EACH ROW EXECUTE FUNCTION audit_log_append_only();
