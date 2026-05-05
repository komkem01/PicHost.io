CREATE TABLE audit_logs (
    id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID         REFERENCES users(id) ON DELETE SET NULL,
    action        VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id   UUID,
    ip_address    TEXT,
    user_agent    TEXT,
    metadata      JSONB,
    status        VARCHAR(20)  NOT NULL DEFAULT 'success',
    error_code    VARCHAR(100),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT current_timestamp
);

CREATE INDEX idx_audit_logs_user_id    ON audit_logs (user_id);
CREATE INDEX idx_audit_logs_action     ON audit_logs (action);
CREATE INDEX idx_audit_logs_status     ON audit_logs (status);
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at DESC);
CREATE INDEX idx_audit_logs_resource   ON audit_logs (resource_type, resource_id);

COMMENT ON TABLE  audit_logs              IS 'Immutable audit trail of every significant user and system action.';
COMMENT ON COLUMN audit_logs.user_id      IS 'User who performed the action; NULL for unauthenticated (guest) events.';
COMMENT ON COLUMN audit_logs.action       IS 'Dot-notation event name, e.g. auth.login, storage.upload, auth.token_refresh.';
COMMENT ON COLUMN audit_logs.resource_type IS 'Category of the affected resource: user, image, storage, session.';
COMMENT ON COLUMN audit_logs.resource_id  IS 'UUID of the primary affected resource.';
COMMENT ON COLUMN audit_logs.ip_address   IS 'Client IP address at the time of the event.';
COMMENT ON COLUMN audit_logs.user_agent   IS 'HTTP User-Agent string supplied by the client.';
COMMENT ON COLUMN audit_logs.metadata     IS 'Arbitrary JSON payload with event-specific detail (email, file_size, mime_type, etc.).';
COMMENT ON COLUMN audit_logs.status       IS 'Outcome of the action: success or failure.';
COMMENT ON COLUMN audit_logs.error_code   IS 'Application error code when status = failure.';
COMMENT ON COLUMN audit_logs.created_at   IS 'Wall-clock time when the event was recorded (immutable).';
