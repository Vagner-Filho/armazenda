-- Add IBS and CBS columns to NF-e tables to comply with the 2026 indirect tax
-- reform (EC 132/2023, NT 2025.002-RTC / MOC 7.0). New groups <IBSCBS> (per-item)
-- and <IBSCBSTot> (per-NF-e totals) become mandatory in the NF-e layout starting
-- August 2026. Rates default to the 2026 symbolic values (CBS 0.9%, IBS 0.1%)
-- and can be overridden per farm.

ALTER TABLE nfe_farm_config
    ADD COLUMN IF NOT EXISTS cbs_rate     NUMERIC(7,4) NOT NULL DEFAULT 0.9000,
    ADD COLUMN IF NOT EXISTS ibs_rate     NUMERIC(7,4) NOT NULL DEFAULT 0.1000,
    ADD COLUMN IF NOT EXISTS cbs_cst      VARCHAR(3)   NOT NULL DEFAULT '000',
    ADD COLUMN IF NOT EXISTS ibs_cst      VARCHAR(3)   NOT NULL DEFAULT '000',
    ADD COLUMN IF NOT EXISTS c_class_trib VARCHAR(10)  NOT NULL DEFAULT '000001';

ALTER TABLE nfe_invoice_tax_rates
    ADD COLUMN IF NOT EXISTS cbs_rate NUMERIC(7,4),
    ADD COLUMN IF NOT EXISTS ibs_rate NUMERIC(7,4);

ALTER TABLE nfe_invoice
    ADD COLUMN IF NOT EXISTS cbs_value NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS ibs_value NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS cbs_cst   VARCHAR(3)    NOT NULL DEFAULT '000',
    ADD COLUMN IF NOT EXISTS ibs_cst   VARCHAR(3)    NOT NULL DEFAULT '000',
    ADD COLUMN IF NOT EXISTS c_class_trib VARCHAR(10) NOT NULL DEFAULT '000001';