DROP TABLE IF EXISTS user_farm_access;

ALTER TABLE owner_subscription RENAME TO owner;

CREATE TABLE IF NOT EXISTS farm_owner_subscription (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    farm_id INTEGER UNIQUE NOT NULL REFERENCES farm(id) ON DELETE CASCADE,
    owner_id INTEGER NOT NULL REFERENCES owner(id) ON DELETE CASCADE,
    stripe_customer_id TEXT,
    stripe_subscription_id TEXT,
    subscription_status TEXT DEFAULT 'pending',
    subscription_current_period_end TIMESTAMP WITHOUT TIME ZONE,
    tier_key TEXT
);

INSERT INTO farm_owner_subscription (
    farm_id, owner_id,
    stripe_customer_id, stripe_subscription_id,
    subscription_status, subscription_current_period_end,
    tier_key
)
SELECT
    f.id,
    os.id,
    f.stripe_customer_id,
    f.stripe_subscription_id,
    COALESCE(NULLIF(f.subscription_status, 'pending'), 'active'),
    f.subscription_current_period_end,
    os.tier_key
FROM farm f
JOIN owner os
    ON os.owner_document = f.owner_document
    AND os.owner_document_type = f.owner_document_type
WHERE f.owner_document IS NOT NULL
ON CONFLICT (farm_id) DO UPDATE SET
    owner_id = EXCLUDED.owner_id,
    stripe_customer_id = EXCLUDED.stripe_customer_id,
    stripe_subscription_id = EXCLUDED.stripe_subscription_id,
    subscription_status = EXCLUDED.subscription_status,
    subscription_current_period_end = EXCLUDED.subscription_current_period_end,
    tier_key = EXCLUDED.tier_key;

ALTER TABLE owner
    DROP COLUMN IF EXISTS stripe_customer_id,
    DROP COLUMN IF EXISTS stripe_subscription_id,
    DROP COLUMN IF EXISTS subscription_status,
    DROP COLUMN IF EXISTS subscription_current_period_end,
    DROP COLUMN IF EXISTS quantity,
    DROP COLUMN IF EXISTS tier_key;

ALTER TABLE farm
    DROP COLUMN IF EXISTS stripe_customer_id,
    DROP COLUMN IF EXISTS stripe_subscription_id,
    DROP COLUMN IF EXISTS subscription_status,
    DROP COLUMN IF EXISTS subscription_current_period_end,
    DROP COLUMN IF EXISTS owner_document,
    DROP COLUMN IF EXISTS owner_document_type;
