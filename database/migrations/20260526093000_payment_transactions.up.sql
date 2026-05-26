SET statement_timeout = 0;

--bun:split

DO $$
BEGIN
    CREATE TYPE payment_status AS ENUM ('pending', 'paid', 'failed', 'cancelled', 'expired');
EXCEPTION
    WHEN duplicate_object THEN NULL;
END$$;

--bun:split

CREATE TABLE IF NOT EXISTS payment_transactions (
    id uuid NOT NULL DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL,
    plan_key text NOT NULL,
    amount_thb integer NOT NULL,
    currency char(3) NOT NULL DEFAULT 'THB',
    status payment_status NOT NULL DEFAULT 'pending',
    provider varchar(50) NOT NULL DEFAULT 'manual',
    checkout_reference varchar(64) NOT NULL,
    provider_reference varchar(255),
    payment_url text,
    expires_at timestamptz NOT NULL,
    paid_at timestamptz,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT current_timestamp,
    updated_at timestamptz NOT NULL DEFAULT current_timestamp,
    PRIMARY KEY (id),
    CONSTRAINT fk_payment_transactions_user_id FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
    CONSTRAINT fk_payment_transactions_plan_key FOREIGN KEY (plan_key) REFERENCES plan_settings (plan_key) ON DELETE RESTRICT,
    CONSTRAINT uq_payment_transactions_checkout_reference UNIQUE (checkout_reference),
    CONSTRAINT uq_payment_transactions_provider_reference UNIQUE (provider_reference)
);

CREATE INDEX IF NOT EXISTS idx_payment_transactions_user_id ON payment_transactions (user_id);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_status ON payment_transactions (status);
CREATE INDEX IF NOT EXISTS idx_payment_transactions_expires_at ON payment_transactions (expires_at);

CREATE OR REPLACE FUNCTION set_payment_transactions_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at = current_timestamp;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_payment_transactions_set_updated_at ON payment_transactions;
CREATE TRIGGER trg_payment_transactions_set_updated_at
BEFORE UPDATE ON payment_transactions
FOR EACH ROW
EXECUTE FUNCTION set_payment_transactions_updated_at();

COMMENT ON TABLE payment_transactions IS 'Records user checkout attempts and payment confirmation states for plan upgrades.';
COMMENT ON COLUMN payment_transactions.id IS 'Unique transaction id.';
COMMENT ON COLUMN payment_transactions.user_id IS 'Owner user id who initiated checkout.';
COMMENT ON COLUMN payment_transactions.plan_key IS 'Target plan key selected during checkout.';
COMMENT ON COLUMN payment_transactions.amount_thb IS 'Amount expected to be paid in Thai Baht.';
COMMENT ON COLUMN payment_transactions.currency IS 'Payment currency ISO code. Default THB.';
COMMENT ON COLUMN payment_transactions.status IS 'Current payment lifecycle state.';
COMMENT ON COLUMN payment_transactions.provider IS 'Payment gateway or method identifier.';
COMMENT ON COLUMN payment_transactions.checkout_reference IS 'Unique checkout reference exposed to users and gateway.';
COMMENT ON COLUMN payment_transactions.provider_reference IS 'Gateway transaction reference after confirmation.';
COMMENT ON COLUMN payment_transactions.payment_url IS 'Gateway payment URL to redirect user for checkout.';
COMMENT ON COLUMN payment_transactions.expires_at IS 'Checkout expiration timestamp before auto-expired.';
COMMENT ON COLUMN payment_transactions.paid_at IS 'Timestamp when payment was confirmed as paid.';
COMMENT ON COLUMN payment_transactions.metadata IS 'Additional provider payload in JSON format.';
COMMENT ON COLUMN payment_transactions.created_at IS 'Record creation timestamp.';
COMMENT ON COLUMN payment_transactions.updated_at IS 'Record last update timestamp.';
