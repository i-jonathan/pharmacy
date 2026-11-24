DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE resource = 'HealthCheck');
DELETE FROM permissions WHERE resource = 'HealthCheck';

DROP TABLE IF EXISTS health_check;

ALTER TABLE customer DROP COLUMN IF EXISTS health_consent_updated_at;
ALTER TABLE customer DROP COLUMN IF EXISTS health_consent;
