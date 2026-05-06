CREATE TABLE IF NOT EXISTS plan_settings (
    plan_key TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    monthly_price_thb INTEGER NOT NULL DEFAULT 0,
    storage_limit_bytes BIGINT NOT NULL DEFAULT 0,
    image_limit INTEGER NOT NULL DEFAULT 0,
    max_upload_mb INTEGER NOT NULL DEFAULT 0,
    allow_private BOOLEAN NOT NULL DEFAULT FALSE,
    custom_domain BOOLEAN NOT NULL DEFAULT FALSE,
    api_access BOOLEAN NOT NULL DEFAULT FALSE,
    priority_support BOOLEAN NOT NULL DEFAULT FALSE,
    no_ads BOOLEAN NOT NULL DEFAULT FALSE,
    watermark_removal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO plan_settings (
    plan_key,
    display_name,
    monthly_price_thb,
    storage_limit_bytes,
    image_limit,
    max_upload_mb,
    allow_private,
    custom_domain,
    api_access,
    priority_support,
    no_ads,
    watermark_removal
)
VALUES
    ('free', 'Free', 0, 1073741824, 200, 10, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
    ('basic', 'Basic', 149, 21474836480, 5000, 20, TRUE, TRUE, FALSE, FALSE, TRUE, TRUE),
    ('pro', 'Pro', 399, 107374182400, 50000, 64, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE),
    ('enterprise', 'Enterprise', 1499, 536870912000, 500000, 256, TRUE, TRUE, TRUE, TRUE, TRUE, TRUE)
ON CONFLICT (plan_key) DO NOTHING;

COMMENT ON TABLE plan_settings IS 'System-wide configurable plan conditions used by admin and runtime quota checks.';
COMMENT ON COLUMN plan_settings.plan_key IS 'Normalized plan key in lowercase, e.g. free, basic, pro, enterprise.';
COMMENT ON COLUMN plan_settings.display_name IS 'Display name shown in admin panel and API responses.';
COMMENT ON COLUMN plan_settings.monthly_price_thb IS 'Monthly price in Thai Baht.';
COMMENT ON COLUMN plan_settings.storage_limit_bytes IS 'Total storage quota in bytes. 0 means no storage allowed.';
COMMENT ON COLUMN plan_settings.image_limit IS 'Maximum number of stored images. 0 means unlimited.';
COMMENT ON COLUMN plan_settings.max_upload_mb IS 'Maximum upload size per file in megabytes. 0 means upload disabled.';
COMMENT ON COLUMN plan_settings.allow_private IS 'Whether private images are allowed for this plan.';
COMMENT ON COLUMN plan_settings.custom_domain IS 'Whether custom domain feature is enabled.';
COMMENT ON COLUMN plan_settings.api_access IS 'Whether API access feature is enabled.';
COMMENT ON COLUMN plan_settings.priority_support IS 'Whether priority support feature is enabled.';
COMMENT ON COLUMN plan_settings.no_ads IS 'Whether ads are disabled for this plan.';
COMMENT ON COLUMN plan_settings.watermark_removal IS 'Whether watermark removal feature is enabled.';
COMMENT ON COLUMN plan_settings.created_at IS 'Record creation timestamp.';
COMMENT ON COLUMN plan_settings.updated_at IS 'Record update timestamp.';
