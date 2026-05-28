-- Migration: Add NFe (Nota Fiscal Eletronica) tables
-- Layout: 4.00
-- State: Mato Grosso (MT)

-- ============================================================
-- 1. FARM NFe CONFIGURATION
-- ============================================================

CREATE TABLE IF NOT EXISTS nfe_farm_config (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    farm_id INTEGER NOT NULL UNIQUE REFERENCES farm(id) ON DELETE CASCADE,

    -- Digital certificate (A1)
    certificate_path TEXT NOT NULL,
    certificate_password_encrypted TEXT NOT NULL,

    -- Environment and numbering
    environment SMALLINT NOT NULL DEFAULT 2 CHECK (environment IN (1, 2)), -- 1=production, 2=homologation
    serie SMALLINT NOT NULL DEFAULT 1,
    next_number INTEGER NOT NULL DEFAULT 1,

    -- Emitter data
    tax_regime SMALLINT NOT NULL DEFAULT 1 CHECK (tax_regime IN (1, 2, 3)), -- 1=Simples Nacional, 2=Lucro Real, 3=Lucro Presumido
    emitter_type SMALLINT NOT NULL DEFAULT 1 CHECK (emitter_type IN (1, 2)), -- 1=CNPJ, 2=CPF (rural producer)
    cnpj_emitter TEXT,
    cpf_emitter TEXT,
    ie_emitter TEXT NOT NULL,

    -- Emitter address (copied from farm but may diverge)
    emitter_uf TEXT NOT NULL DEFAULT 'MT',

    -- Default transport settings
    default_mod_frete SMALLINT NOT NULL DEFAULT 1 CHECK (default_mod_frete IN (0, 1, 2, 3, 4, 9)),

    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================
-- 2. PRODUCT TO FISCAL CODE MAPPING
-- ============================================================

CREATE TABLE IF NOT EXISTS nfe_product_config (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    farm_id INTEGER NOT NULL REFERENCES farm(id) ON DELETE CASCADE,
    product_id SMALLINT NOT NULL REFERENCES product(id),

    ncm TEXT NOT NULL,
    default_cfop TEXT NOT NULL,
    default_cest TEXT,
    unit TEXT NOT NULL DEFAULT 'KG',
    description TEXT,

    -- Default taxation by regime
    default_icms_cst TEXT,
    default_pis_cst TEXT,
    default_cofins_cst TEXT,

    UNIQUE(farm_id, product_id)
);

-- Seed default configs for corn and soy per existing farm
INSERT INTO nfe_product_config (farm_id, product_id, ncm, default_cfop, unit, description, default_icms_cst, default_pis_cst, default_cofins_cst)
SELECT
    f.id,
    p.id,
    CASE p.name
        WHEN 'Milho' THEN '1005.90.00'
        WHEN 'Soja' THEN '1201.00.10'
        ELSE '0000.00.00'
    END,
    '5102',
    'KG',
    p.name,
    NULL,
    NULL,
    NULL
FROM farm f
CROSS JOIN product p
ON CONFLICT DO NOTHING;

-- ============================================================
-- 3. ATOMIC NUMBERING CONTROL
-- ============================================================

CREATE TABLE IF NOT EXISTS nfe_numbering (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    farm_id INTEGER NOT NULL REFERENCES farm(id) ON DELETE CASCADE,
    serie SMALLINT NOT NULL,
    last_number INTEGER NOT NULL DEFAULT 0,
    UNIQUE(farm_id, serie)
);

CREATE OR REPLACE FUNCTION nfe_allocate_number(p_farm_id INTEGER, p_serie SMALLINT)
RETURNS INTEGER AS $$
DECLARE
    v_number INTEGER;
BEGIN
    INSERT INTO nfe_numbering (farm_id, serie, last_number)
    VALUES (p_farm_id, p_serie, 1)
    ON CONFLICT (farm_id, serie)
    DO UPDATE SET last_number = nfe_numbering.last_number + 1
    RETURNING nfe_numbering.last_number INTO v_number;

    RETURN v_number;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- 4. INVOICE TRACKING
-- ============================================================

CREATE TABLE IF NOT EXISTS nfe_invoice (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    departure_id INTEGER NOT NULL REFERENCES departure(id) ON DELETE RESTRICT,

    -- Access key (44 digits)
    access_key TEXT UNIQUE NOT NULL,

    -- Numbering
    serie SMALLINT NOT NULL,
    number INTEGER NOT NULL,

    -- Lifecycle status
    status TEXT NOT NULL DEFAULT 'draft'
        CHECK (status IN ('draft', 'signed', 'pending', 'processing', 'authorized', 'denied', 'cancelled')),

    -- Operation data
    cfop TEXT NOT NULL,
    ncm TEXT NOT NULL,
    quantity_kg NUMERIC(12, 3) NOT NULL,
    unit_price NUMERIC(15, 4) NOT NULL,
    total_value NUMERIC(15, 2) NOT NULL,
    icms_value NUMERIC(15, 2),

    -- XMLs
    xml_signed TEXT,
    xml_authorized TEXT,
    xml_cancel_event TEXT,

    -- SEFAZ responses
    protocol TEXT,
    sefaz_status_code TEXT,
    sefaz_motive TEXT,

    -- Rejection or cancellation reason
    rejection_reason TEXT,
    cancellation_reason TEXT,

    -- Retry and queue
    retry_count INTEGER NOT NULL DEFAULT 0,
    last_retry_at TIMESTAMP WITHOUT TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    signed_at TIMESTAMP WITHOUT TIME ZONE,
    sent_at TIMESTAMP WITHOUT TIME ZONE,
    authorized_at TIMESTAMP WITHOUT TIME ZONE,
    cancelled_at TIMESTAMP WITHOUT TIME ZONE
);

-- Unique constraint per farm: serie + number must be unique within the same farm
-- We derive farm from departure, so we use a partial unique index approach
-- or enforce at application level. For simplicity, we add a composite unique
-- that includes farm_id via a function index.
CREATE UNIQUE INDEX IF NOT EXISTS idx_nfe_invoice_farm_serie_number
ON nfe_invoice (serie, number, (SELECT farm FROM departure WHERE id = departure_id));
