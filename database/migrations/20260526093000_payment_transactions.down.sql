SET statement_timeout = 0;

--bun:split

DROP TABLE IF EXISTS payment_transactions CASCADE;

--bun:split

DROP FUNCTION IF EXISTS set_payment_transactions_updated_at();

--bun:split

DROP TYPE IF EXISTS payment_status;
