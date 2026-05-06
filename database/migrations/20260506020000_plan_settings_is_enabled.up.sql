ALTER TABLE plan_settings
    ADD COLUMN IF NOT EXISTS is_enabled BOOLEAN NOT NULL DEFAULT TRUE;

UPDATE plan_settings
SET is_enabled = TRUE
WHERE is_enabled IS NULL;

COMMENT ON COLUMN plan_settings.is_enabled IS 'Whether this plan is currently enabled for use in admin-managed plan flows.';
