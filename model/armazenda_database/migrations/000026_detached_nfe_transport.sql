-- 000026_detached_nfe_transport.sql
--
-- Transport document (MOC 7.0, Grupo X) of detached NF-e invoices, normalized
-- in its own 1:1 table referencing detached_nfe_invoice. The field detail is
-- stored as a single JSONB document (same serialize pattern as
-- detached_nfe_invoice.items_json), grouped under the Grupo X sub-blocks:
--   transportadora (X03): CNPJ/CPF, xNome, IE, xEnder, xMun, UF
--   veiculo (X18): placa, UF, RNTC (ANTT)
--   vol (X26): qVol, espécie, marca, pesoL, pesoB
--
-- mod_frete stays on detached_nfe_invoice; this table is written on emission
-- only when any transport field was informed.

CREATE TABLE IF NOT EXISTS detached_invoice_transportation (
    id          INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    invoice_id  INTEGER NOT NULL UNIQUE REFERENCES detached_nfe_invoice(id),
    transport   JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
