-- 000022_detached_nfe_item_cfop.sql
--
-- CFOP moves from the detached invoice header to each item in items_json.
--
-- Order matters: each existing item keeps its own CFOP. Only items that lack a
-- CFOP are backfilled from the legacy invoice-level column. The header column
-- is dropped only after every item is guaranteed to carry a value.
--
-- Idempotent: the whole block is guarded by the presence of the legacy column.
-- Signed/authorized XML is never rewritten.

DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM information_schema.columns
        WHERE table_schema = 'public'
          AND table_name = 'detached_nfe_invoice'
          AND column_name = 'cfop'
    ) THEN
        -- Backfill only the items whose cfop is missing/empty, preserving
        -- array order through WITH ORDINALITY.
        UPDATE detached_nfe_invoice
        SET items_json = (
            SELECT jsonb_agg(
                CASE
                    WHEN COALESCE(item->>'cfop', '') = ''
                        THEN jsonb_set(item, '{cfop}', to_jsonb(detached_nfe_invoice.cfop))
                    ELSE item
                END
                ORDER BY item_ordinality
            )
            FROM jsonb_array_elements(items_json) WITH ORDINALITY AS elements(item, item_ordinality)
        )
        WHERE items_json IS NOT NULL
          AND jsonb_typeof(items_json) = 'array'
          AND EXISTS (
              SELECT 1
              FROM jsonb_array_elements(items_json) AS elements(item)
              WHERE COALESCE(item->>'cfop', '') = ''
          );

        ALTER TABLE detached_nfe_invoice DROP COLUMN cfop;
    END IF;
END $$;
