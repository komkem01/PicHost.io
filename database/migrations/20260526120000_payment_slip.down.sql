SET statement_timeout = 0;

--bun:split

ALTER TABLE payment_transactions
    DROP COLUMN IF EXISTS slip_storage_id;
