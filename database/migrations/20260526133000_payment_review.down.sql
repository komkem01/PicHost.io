-- +goose Down
-- +goose StatementBegin
ALTER TABLE payment_transactions
    DROP CONSTRAINT IF EXISTS fk_payment_transactions_reviewed_by;

DROP INDEX IF EXISTS idx_payment_transactions_reviewed_at;
DROP INDEX IF EXISTS idx_payment_transactions_reviewed_by;

ALTER TABLE payment_transactions
    DROP COLUMN IF EXISTS reviewed_at,
    DROP COLUMN IF EXISTS reviewed_by,
    DROP COLUMN IF EXISTS review_reason;
-- +goose StatementEnd
