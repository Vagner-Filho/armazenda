-- Add IBS and CBS columns to NF-e tables to comply with the 2026 indirect tax
-- reform (EC 132/2023, NT 2025.002-RTC / MOC 7.0). New groups <IBSCBS> (per-item)
-- and <IBSCBSTot> (per-NF-e totals) become mandatory in the NF-e layout starting
-- August 2026. Rates default to the 2026 symbolic values (CBS 0.9%, IBS 0.1%)
-- and can be overridden per farm.

-- IMPORTANT decimal-rate convention: the XML schema expects the rate as a
-- decimal (e.g. 0.001 for 0.1%). `parsePercentRateOrNil` in
-- router/nfe_router/router.go divides the form-submitted percentage string by
-- 100 to produce the decimal rate stored in these columns. `defaults.IBSRate2026`
-- and `defaults.CBSRate2026` in pkg/nfe/defaults/reform.go also use this
-- convention. The previous defaults (0.1000/0.9000) were 100x too large and
-- caused SEFAZ rejection 1026 ("Alíquota do IBS da UF inválida") in 2026.

ALTER TABLE nfe_farm_config
    ADD COLUMN IF NOT EXISTS cbs_rate     NUMERIC(7,4) NOT NULL DEFAULT 0.0090,
    ADD COLUMN IF NOT EXISTS ibs_rate     NUMERIC(7,4) NOT NULL DEFAULT 0.0010,
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

-- Data migration: correct any rows that were populated with the previous wrong
-- defaults. Matches the exact previous default values so user-customized rates
-- are left untouched.
UPDATE nfe_farm_config SET ibs_rate = 0.0010 WHERE ibs_rate = 0.1000;
UPDATE nfe_farm_config SET cbs_rate = 0.0090 WHERE cbs_rate = 0.9000;