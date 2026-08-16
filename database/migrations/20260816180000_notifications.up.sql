-- Notifications Table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NULL REFERENCES users(id) ON DELETE CASCADE,
    target_role VARCHAR(30) NOT NULL DEFAULT 'user', -- 'user', 'admin', 'all'
    type VARCHAR(50) NOT NULL, -- 'moderation', 'payment', 'security', 'storage', 'system', 'announcement'
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    link TEXT NULL,
    metadata JSONB NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications(user_id);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread ON notifications(user_id, is_read, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_target_role ON notifications(target_role, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications(created_at DESC);
