SET statement_timeout = 0;

--bun:split

ALTER TABLE "users" ADD COLUMN IF NOT EXISTS "email_verified_at" timestamptz;

--bun:split

CREATE TABLE IF NOT EXISTS email_verification_tokens (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    token_hash varchar(64) NOT NULL,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    PRIMARY KEY (id),
    CONSTRAINT fk_email_verification_tokens_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT uq_email_verification_tokens_token_hash UNIQUE (token_hash)
);

CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_token_hash ON email_verification_tokens (token_hash);
CREATE INDEX IF NOT EXISTS idx_email_verification_tokens_user_id ON email_verification_tokens (user_id);

COMMENT ON COLUMN users.email_verified_at IS 'Timestamp when user email was verified.';
COMMENT ON TABLE email_verification_tokens IS 'Table storing SHA-256 hashed email verification tokens.';
