-- 000021_detached_nfe.sql
--
-- Detached NF-e: invoices emitted independently from any departure.
-- Separate table to keep concerns isolated from nfe_invoice.

CREATE TABLE IF NOT EXISTS detached_nfe_invoice (
    id          INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    farm_id     INTEGER NOT NULL REFERENCES farm(id),
    recipient_id INTEGER NOT NULL REFERENCES person(id),
    access_key  TEXT    NOT NULL UNIQUE,
    serie       SMALLINT NOT NULL,
    number      INTEGER NOT NULL,
    status      TEXT    NOT NULL DEFAULT 'draft',

    -- Invoice-level fiscal defaults (per-item overrides live in items_json)
    cfop        TEXT    NOT NULL,
    natureza_op TEXT,
    mod_frete   SMALLINT,

    -- Totals (aggregated from items)
    total_value NUMERIC(15,2) NOT NULL,
    ibs_value   NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    cbs_value   NUMERIC(15,2) NOT NULL DEFAULT 0.00,

    -- Line items (array of DetachedInvoiceItem as JSONB)
    items_json  JSONB   NOT NULL,

    -- XML storage
    xml_signed      TEXT,
    xml_authorized  TEXT,
    xml_cancel_event TEXT,

    -- SEFAZ response
    protocol            TEXT,
    sefaz_status_code   TEXT,
    sefaz_motive        TEXT,
    rejection_reason    TEXT,
    cancellation_reason TEXT,

    -- Contingency
    tp_emis               SMALLINT NOT NULL DEFAULT 1,
    dh_cont               TIMESTAMP,
    x_just                TEXT,
    contingency_parent_id INTEGER REFERENCES detached_nfe_invoice(id),
    svc_endpoint_used     TEXT,

    -- Invoice-level tax overrides (CSTs)
    icms_cst     TEXT,
    pis_cst      TEXT,
    cofins_cst   TEXT,
    ibs_cst      VARCHAR(3)  NOT NULL DEFAULT '000',
    cbs_cst      VARCHAR(3)  NOT NULL DEFAULT '000',
    c_class_trib VARCHAR(10) NOT NULL DEFAULT '000001',
    inf_cpl      TEXT,

    -- Retry / worker
    retry_count   INTEGER   NOT NULL DEFAULT 0,
    last_retry_at TIMESTAMP,

    -- Timestamps
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    signed_at     TIMESTAMP,
    sent_at       TIMESTAMP,
    authorized_at TIMESTAMP,
    cancelled_at  TIMESTAMP,

    CONSTRAINT detached_nfe_status_check
        CHECK (status IN ('draft','pending','authorized','denied','cancelled','superseded')),
    CONSTRAINT detached_nfe_mod_frete_check
        CHECK (mod_frete IN (0,1,2,3,4,9))
);

-- Tax rate overrides (mirrors nfe_invoice_tax_rates pattern)
CREATE TABLE IF NOT EXISTS detached_nfe_tax_rates (
    invoice_id  INTEGER PRIMARY KEY REFERENCES detached_nfe_invoice(id) ON DELETE CASCADE,
    icms_rate   NUMERIC(5,4),
    pis_rate    NUMERIC(5,4),
    cofins_rate NUMERIC(5,4),
    ibs_rate    NUMERIC(7,4),
    cbs_rate    NUMERIC(7,4),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
