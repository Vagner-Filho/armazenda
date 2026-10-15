-- 000030_departure_nfe_inputs.sql
-- Persist the resolved departure NF-e payment, billing, and transport inputs
-- so draft retries and SVC rebuilds reproduce the approved XML.

ALTER TABLE nfe_invoice
    ADD COLUMN IF NOT EXISTS emission_inputs JSONB NOT NULL DEFAULT '{}'::jsonb;
