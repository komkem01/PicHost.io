SET statement_timeout = 0;

--bun:split

ALTER TABLE payment_transactions
    ADD COLUMN IF NOT EXISTS activated_at timestamptz;

-- Backfill: a historical `paid` row is treated as already applied to the user's
-- plan, so it must never be picked up by the new activation flow after deploy.
UPDATE payment_transactions
    SET activated_at = COALESCE(paid_at, updated_at)
    WHERE status = 'paid' AND activated_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_payment_transactions_paid_unactivated
    ON payment_transactions (user_id)
    WHERE status = 'paid' AND activated_at IS NULL;

COMMENT ON COLUMN payment_transactions.activated_at IS 'Timestamp when the paid entitlement was actually applied to the user (plan/expiry updated). NULL while a paid-but-unverified checkout awaits email verification.';
