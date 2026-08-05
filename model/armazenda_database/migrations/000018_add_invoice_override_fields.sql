-- Add per-emission override fields to nfe_invoice so draft retries and
-- SVC contingency rebuilds can reconstruct the exact same NF-e the user
-- originally configured in the emit modal.

ALTER TABLE nfe_invoice
    ADD COLUMN IF NOT EXISTS natureza_op TEXT,
    ADD COLUMN IF NOT EXISTS product_description TEXT,
    ADD COLUMN IF NOT EXISTS cest TEXT,
    ADD COLUMN IF NOT EXISTS unit TEXT,
    ADD COLUMN IF NOT EXISTS mod_frete INTEGER,
    ADD COLUMN IF NOT EXISTS icms_cst TEXT,
    ADD COLUMN IF NOT EXISTS pis_cst TEXT,
    ADD COLUMN IF NOT EXISTS cofins_cst TEXT,
    ADD COLUMN IF NOT EXISTS inf_cpl TEXT;
