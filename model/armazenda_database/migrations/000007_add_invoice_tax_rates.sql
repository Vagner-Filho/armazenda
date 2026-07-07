CREATE TABLE IF NOT EXISTS nfe_invoice_tax_rates (
    invoice_id  INTEGER PRIMARY KEY REFERENCES nfe_invoice(id) ON DELETE CASCADE,
    icms_rate   DECIMAL(5,4) NULL,
    pis_rate    DECIMAL(5,4) NULL,
    cofins_rate DECIMAL(5,4) NULL,
    created_at  TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
