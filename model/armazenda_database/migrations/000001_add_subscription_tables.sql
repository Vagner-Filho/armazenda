ALTER TABLE farm
    ADD COLUMN IF NOT EXISTS stripe_customer_id TEXT,
    ADD COLUMN IF NOT EXISTS stripe_subscription_id TEXT,
    ADD COLUMN IF NOT EXISTS subscription_status TEXT DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS subscription_current_period_end TIMESTAMP WITHOUT TIME ZONE;

CREATE TABLE IF NOT EXISTS pending_registration (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    passwd TEXT NOT NULL,
    cpf VARCHAR(11) NOT NULL,
    inscricao_estadual TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'user')),
    stripe_checkout_session_id TEXT,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
