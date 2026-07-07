-- 1. Add ncm to product
ALTER TABLE product
    ADD COLUMN IF NOT EXISTS ncm TEXT NOT NULL DEFAULT '00000000';

-- 2. Backfill product.ncm from nfe_product_config for existing DBs.
--    On a fresh DB, nfe_product_config is never created (the seed in
--    database.go:seedProducts only inserts into `product`), so this UPDATE
--    is a no-op and step 3 below covers the well-known products.
UPDATE product p
SET ncm = COALESCE(
    (SELECT npc.ncm FROM nfe_product_config npc
     WHERE npc.product_id = p.id
     ORDER BY npc.farm_id LIMIT 1),
    p.ncm
)
WHERE p.ncm = '00000000';

-- 3. Backfill well-known products by name. This covers fresh DBs where
--    nfe_product_config never existed. The seed in database.go:seedProducts
--    inserts these two rows, so the names are guaranteed to match.
UPDATE product SET ncm = '10059000' WHERE name = 'Milho' AND ncm = '00000000';
UPDATE product SET ncm = '12010010' WHERE name = 'Soja'  AND ncm = '00000000';

-- 4. Add the 10 per-farm fields to nfe_farm_config
ALTER TABLE nfe_farm_config
    ADD COLUMN IF NOT EXISTS default_cfop          TEXT         NOT NULL DEFAULT '5101',
    ADD COLUMN IF NOT EXISTS default_cest          TEXT         NULL,
    ADD COLUMN IF NOT EXISTS default_unit          TEXT         NOT NULL DEFAULT 'KG',
    ADD COLUMN IF NOT EXISTS default_icms_cst      TEXT         NULL,
    ADD COLUMN IF NOT EXISTS default_pis_cst       TEXT         NULL,
    ADD COLUMN IF NOT EXISTS default_cofins_cst    TEXT         NULL,
    ADD COLUMN IF NOT EXISTS default_natureza_op   TEXT         NULL,
    ADD COLUMN IF NOT EXISTS icms_rate             DECIMAL(5,4) NULL,
    ADD COLUMN IF NOT EXISTS pis_rate              DECIMAL(5,4) NULL,
    ADD COLUMN IF NOT EXISTS cofins_rate           DECIMAL(5,4) NULL;

-- 5. Backfill per-farm fields from any nfe_product_config row of the same farm.
--    Safe because the seed produces identical values across products within
--    a farm. Any non-null value is acceptable (per the chosen backfill rule).
UPDATE nfe_farm_config fc SET
    default_cfop        = COALESCE(fc.default_cfop,        (SELECT npc.default_cfop       FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    default_cest        = COALESCE(fc.default_cest,        (SELECT npc.default_cest       FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    default_unit        = COALESCE(fc.default_unit,        (SELECT npc.unit               FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    default_icms_cst    = COALESCE(fc.default_icms_cst,    (SELECT npc.default_icms_cst   FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    default_pis_cst     = COALESCE(fc.default_pis_cst,     (SELECT npc.default_pis_cst    FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    default_cofins_cst  = COALESCE(fc.default_cofins_cst,  (SELECT npc.default_cofins_cst FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    default_natureza_op = COALESCE(fc.default_natureza_op, (SELECT npc.natureza_op        FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    icms_rate           = COALESCE(fc.icms_rate,           (SELECT npc.icms_rate          FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    pis_rate            = COALESCE(fc.pis_rate,            (SELECT npc.pis_rate           FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1)),
    cofins_rate         = COALESCE(fc.cofins_rate,         (SELECT npc.cofins_rate        FROM nfe_product_config npc WHERE npc.farm_id = fc.farm_id LIMIT 1));

-- 6. Drop nfe_product_config
DROP TABLE IF EXISTS nfe_product_config CASCADE;
