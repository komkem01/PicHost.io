SET statement_timeout = 0;

--bun:split

ALTER TABLE payment_transactions
    ADD COLUMN IF NOT EXISTS slip_storage_id varchar(255);

COMMENT ON COLUMN payment_transactions.slip_storage_id IS 'Storage service record ID of the payment slip image uploaded by the user for manual bank-transfer verification.';
