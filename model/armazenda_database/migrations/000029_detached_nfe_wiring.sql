-- 000029_detached_nfe_wiring.sql
--
-- Persistence for the detached NF-e (NF-e Avulsa) emission wiring
-- (docs/sdd/detached_nfe_input_wiring): keeps the stored row coherent
-- with the emitted XML for the builder capabilities added by
-- nfe_builder_replay_gaps. No reader consumes them yet — they are the
-- future inputs of the detached-draft rebuild (worker TODO).
--
-- Per-emission item discounts ride the existing items_json JSONB
-- (DetachedInvoiceItem.Discount, absent key = zero for old rows), so no
-- schema change is needed for discounts here.

ALTER TABLE detached_nfe_invoice
    ADD COLUMN IF NOT EXISTS finnfe TEXT NOT NULL DEFAULT '1';
ALTER TABLE detached_nfe_invoice
    ADD COLUMN IF NOT EXISTS nf_refs JSONB;
ALTER TABLE detached_nfe_invoice
    ADD COLUMN IF NOT EXISTS entrega JSONB;
