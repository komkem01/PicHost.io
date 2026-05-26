-- +goose Up
-- +goose StatementBegin
ALTER TABLE payment_transactions
    ADD COLUMN IF NOT EXISTS review_reason text,
    ADD COLUMN IF NOT EXISTS reviewed_by uuid,
    ADD COLUMN IF NOT EXISTS reviewed_at timestamptz;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'fk_payment_transactions_reviewed_by'
    ) THEN
        ALTER TABLE payment_transactions
            ADD CONSTRAINT fk_payment_transactions_reviewed_by
            FOREIGN KEY (reviewed_by)
            REFERENCES users (id)
            ON DELETE SET NULL;
    END IF;
END;
$$;

CREATE INDEX IF NOT EXISTS idx_payment_transactions_reviewed_by ON payment_transactions (reviewed_by);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_reviewed_at ON payment_transactions (reviewed_at);

COMMENT ON TABLE payment_transactions IS 'Records user checkout attempts and payment confirmation states for plan upgrades.';
COMMENT ON COLUMN payment_transactions.review_reason IS 'Admin review note or rejection reason for manual payment verification.';
COMMENT ON COLUMN payment_transactions.reviewed_by IS 'Admin user id who reviewed and decided the payment status.';
COMMENT ON COLUMN payment_transactions.reviewed_at IS 'Timestamp when an admin reviewed and finalized the payment status.';
-- +goose StatementEnd
