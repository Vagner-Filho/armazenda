-- farm_product: add unit_measurement so each farm-created product carries
-- the unit of measurement that should be used when it is selected as an
-- NF-e item. Defaults to KG for existing rows.
ALTER TABLE farm_product
    ADD COLUMN IF NOT EXISTS unit_measurement TEXT NOT NULL DEFAULT 'KG';