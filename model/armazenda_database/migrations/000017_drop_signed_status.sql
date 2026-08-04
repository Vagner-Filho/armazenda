-- The 'signed' status was a transient state from the original async-send
-- design (build+sign → persist as 'signed' → worker sends). The architecture
-- now sends synchronously, so 'signed' only survived a crash window between
-- persisting the XML and sending — and rows stuck there were never retried
-- (the worker only picks up 'pending' and 'draft').
--
-- Rescue any orphaned 'signed' rows by making them drafts so the retry
-- worker can pick them up (≤ 24h window), then drop the status from the
-- check constraint.
UPDATE nfe_invoice SET status = 'draft' WHERE status = 'signed';

ALTER TABLE nfe_invoice DROP CONSTRAINT IF EXISTS nfe_invoice_status_check;
ALTER TABLE nfe_invoice ADD CONSTRAINT nfe_invoice_status_check
    CHECK (status IN ('draft', 'pending', 'authorized', 'denied', 'cancelled', 'superseded'));
