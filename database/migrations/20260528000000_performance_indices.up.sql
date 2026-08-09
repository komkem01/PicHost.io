--bun:split

CREATE INDEX IF NOT EXISTS idx_storages_short_code ON storages (short_code);
CREATE INDEX IF NOT EXISTS idx_images_user_id ON images (user_id);
CREATE INDEX IF NOT EXISTS idx_images_storage_id ON images (storage_id);
CREATE INDEX IF NOT EXISTS idx_images_expires_at ON images (expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_payment_transactions_user_id ON payment_transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_checkout_ref ON payment_transactions (checkout_reference);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs (user_id);
