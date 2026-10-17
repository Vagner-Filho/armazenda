-- Relax NOT NULL on nfe_farm_config columns so partial saves (e.g. certificate
-- only, or no certificate yet) succeed. Aligns cbs_cst / ibs_cst / c_class_trib
-- with the already-nullable default_icms_cst / default_pis_cst /
-- default_cofins_cst (migration 000008), and certificate_data with the "no cert
-- uploaded yet" flow (migration 000004). Idempotent: dropping NOT NULL on an
-- already-nullable column is a no-op success.
ALTER TABLE nfe_farm_config
    ALTER COLUMN cbs_cst DROP NOT NULL,
    ALTER COLUMN ibs_cst DROP NOT NULL,
    ALTER COLUMN c_class_trib DROP NOT NULL,
    ALTER COLUMN certificate_data DROP NOT NULL;