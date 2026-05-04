DROP INDEX IF EXISTS images_expires_at_idx;
ALTER TABLE images DROP COLUMN IF EXISTS expires_at;

DROP TRIGGER  IF EXISTS user_quotas_updated_at_trigger ON user_quotas;
DROP FUNCTION IF EXISTS update_user_quotas_updated_at();
DROP INDEX    IF EXISTS user_quotas_user_id_idx;
DROP TABLE    IF EXISTS user_quotas;
