--bun:split

DROP INDEX IF EXISTS idx_payment_transactions_paid_unactivated;

ALTER TABLE payment_transactions
    DROP COLUMN IF EXISTS activated_at;
