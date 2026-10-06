-- 000023_detached_nfe_operation_profiles.sql
--
-- Rascunhos de NF-e: named, farm-scoped reusable starting points for a
-- detached NF-e emission (NF-e avulsa). A profile stores an optional
-- recipient reference, an ordered items_json snapshot with per-item CFOP and
-- exact unit prices, optional operation/freight defaults, and optional
-- invoice-level tax overrides.
--
-- No archive state: deletion is hard, mirroring the romaneio rascunhos.
-- The unique index on (farm_id, LOWER(name)) makes rascunho names
-- case-insensitively unique per farm and frees the name on hard delete.

CREATE TABLE IF NOT EXISTS detached_nfe_profile (
    id             INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    farm_id        INTEGER NOT NULL REFERENCES farm(id),
    name           TEXT    NOT NULL,
    recipient_id   INTEGER REFERENCES person(id),

    -- Operation defaults
    natureza_op    TEXT,
    mod_frete      SMALLINT,
    inf_cpl        TEXT,

    -- Ordered item snapshot (DetachedProfileItem as JSONB). Item values are
    -- stored as profile values, never live product lookups.
    items_json     JSONB   NOT NULL DEFAULT '[]'::jsonb,

    -- Optional invoice-level tax CST / cClassTrib overrides
    icms_cst       TEXT,
    pis_cst        TEXT,
    cofins_cst     TEXT,
    ibs_cst        VARCHAR(3),
    cbs_cst        VARCHAR(3),
    c_class_trib   VARCHAR(10),

    -- Optional invoice-level tax rate overrides. NULL means "inherit the farm
    -- configuration"; an explicit 0.0000 is preserved as zero.
    icms_rate      NUMERIC(5,4),
    pis_rate       NUMERIC(5,4),
    cofins_rate    NUMERIC(5,4),
    ibs_rate       NUMERIC(7,4),
    cbs_rate       NUMERIC(7,4),

    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT detached_nfe_profile_name_check
        CHECK (length(trim(name)) > 0),
    CONSTRAINT detached_nfe_profile_mod_frete_check
        CHECK (mod_frete IN (0,1,2,3,4,9))
);

CREATE UNIQUE INDEX IF NOT EXISTS detached_nfe_profile_farm_name_unique
    ON detached_nfe_profile (farm_id, LOWER(name));

CREATE INDEX IF NOT EXISTS detached_nfe_profile_farm_idx
    ON detached_nfe_profile (farm_id);
