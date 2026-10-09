-- 000027_drop_dead_nfe_columns.sql
--
-- Drops NF-e invoice columns that never had a writer in any code path:
--
--   nfe_invoice.icms_value                  (000002) — ICMS totals were never persisted
--   nfe_invoice.svc_endpoint_used           (000005) — endpoint is derived per send, never stored
--   detached_nfe_invoice.svc_endpoint_used  (000021) — same
--
-- Migrations are append-only history: fresh databases still create these
-- columns at 000002/000005/000021 and drop them here.
--
-- Pre-drop evidence query for operators (each count must be 0 — nothing ever
-- wrote these columns, so no data can be lost):
--   SELECT COUNT(*) FROM nfe_invoice WHERE icms_value IS NOT NULL;
--   SELECT COUNT(*) FROM nfe_invoice WHERE svc_endpoint_used IS NOT NULL;
--   SELECT COUNT(*) FROM detached_nfe_invoice WHERE svc_endpoint_used IS NOT NULL;
--
-- Idempotent: each drop is guarded by the presence of its column
-- (000022_detached_nfe_item_cfop.sql pattern).

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'nfe_invoice'
          AND column_name = 'icms_value'
    ) THEN
        ALTER TABLE nfe_invoice DROP COLUMN icms_value;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'nfe_invoice'
          AND column_name = 'svc_endpoint_used'
    ) THEN
        ALTER TABLE nfe_invoice DROP COLUMN svc_endpoint_used;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'detached_nfe_invoice'
          AND column_name = 'svc_endpoint_used'
    ) THEN
        ALTER TABLE detached_nfe_invoice DROP COLUMN svc_endpoint_used;
    END IF;
END $$;
