ALTER TABLE pending_registration
    ADD COLUMN IF NOT EXISTS stripe_price_id TEXT;
