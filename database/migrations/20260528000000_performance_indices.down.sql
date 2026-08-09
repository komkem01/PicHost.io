--bun:split

DROP INDEX IF EXISTS idx_storages_short_code;
DROP INDEX IF EXISTS idx_images_user_id;
DROP INDEX IF EXISTS idx_images_storage_id;
DROP INDEX IF EXISTS idx_images_expires_at;
DROP INDEX IF EXISTS idx_payment_transactions_user_id;
DROP INDEX IF EXISTS idx_payment_transactions_checkout_ref;
DROP INDEX IF EXISTS idx_audit_logs_user_id;
