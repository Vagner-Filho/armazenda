ALTER TABLE IF EXISTS nfe_invoice
    ADD COLUMN IF NOT EXISTS tp_emis SMALLINT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS dh_cont TIMESTAMP WITHOUT TIME ZONE,
    ADD COLUMN IF NOT EXISTS x_just TEXT,
    ADD COLUMN IF NOT EXISTS contingency_parent_id INTEGER REFERENCES nfe_invoice(id),
    ADD COLUMN IF NOT EXISTS svc_endpoint_used TEXT;

-- Widen the status check constraint to include the new 'superseded' status.
-- PostgreSQL requires dropping and re-adding the constraint; IF EXISTS avoids
-- failures when the constraint name differs or has already been dropped.
ALTER TABLE nfe_invoice DROP CONSTRAINT IF EXISTS nfe_invoice_status_check;
ALTER TABLE nfe_invoice ADD CONSTRAINT nfe_invoice_status_check
    CHECK (status IN ('draft', 'signed', 'pending', 'authorized', 'denied', 'cancelled', 'superseded'));
