ALTER TABLE farm
    ADD COLUMN IF NOT EXISTS stripe_customer_id TEXT,
    ADD COLUMN IF NOT EXISTS stripe_subscription_id TEXT,
    ADD COLUMN IF NOT EXISTS subscription_status TEXT DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS subscription_current_period_end TIMESTAMP WITHOUT TIME ZONE,
    ADD COLUMN IF NOT EXISTS owner_document TEXT NULL,
    ADD COLUMN IF NOT EXISTS owner_document_type INTEGER NULL;

CREATE TABLE IF NOT EXISTS pending_registration (
        id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	email TEXT NOT NULL,
	name TEXT NOT NULL,
	passwd TEXT NOT NULL,
	cpf VARCHAR(11) NOT NULL,
	inscricao_estadual TEXT NOT NULL,
        role TEXT NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'user')),
        stripe_checkout_session_id TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
	owner_document TEXT,
	owner_document_type SMALLINT NULL,
	additional_ies TEXT[]
);

CREATE TABLE IF NOT EXISTS owner_subscription (
        id SMALLINT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
	owner_document text NOT NULL,
	owner_document_type int2 NOT NULL,
	stripe_customer_id text NULL,
	stripe_subscription_id text NULL,
	subscription_status text DEFAULT 'pending'::text NULL,
	subscription_current_period_end timestamp NULL,
	quantity int4 DEFAULT 1 NOT NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	tier_key text NULL,
	CONSTRAINT owner_subscription_owner_document_owner_document_type_key UNIQUE (owner_document, owner_document_type)
);
