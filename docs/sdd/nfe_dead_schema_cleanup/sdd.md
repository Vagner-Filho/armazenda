# sdd:nfe_dead_schema_cleanup

## Status
Backlog

## Goal

Eliminate dead NF-e invoice schema and dead model code that create false expectations, and make the invoice rows carry the two pieces of lifecycle data they already pretend to have:

1. **Dead columns with no writer and no consumer** — `nfe_invoice.icms_value` (present since 000002, never written by any code path), and `svc_endpoint_used` on **both** `nfe_invoice` (000005) and `detached_nfe_invoice` (000021) (never written anywhere; only scanned into struct fields nothing reads). Drop them and remove the matching Go struct fields/SELECT columns.
2. **`sent_at` — currently dead in both tables** (the departure-flow writer `UpdateInvoiceSent` has zero callers; the detached flow has no writer at all). Instead of dropping, **populate it**: stamp each invoice row at the moment its first transmission attempt starts, so `authorized_at − sent_at` measures SEFAZ send→authorization latency per invoice.
3. **`detached_nfe_invoice.dh_cont` / `x_just` — defined but never populated.** The detached SVC contingency path writes `<dhCont>`/`<xJust>` into the signed XML and sets `tp_emis = 6|7`, but the row keeps `NULL` (the departure flow writes these; the detached flow forgot). Persist them at contingency creation for row/XML parity.
4. **Dead model functions adjacent to this schema** — delete `CreateInvoiceDeprecated`, `GetInvoiceTaxRates`, `GetDetachedInvoiceTaxRates` (no callers).

Audited and confirmed ALIVE (out of scope): `contingency_parent_id`, `last_retry_at`, `signed_at`, `authorized_at`, `xml_cancel_event`, all override/CST columns, `nfe_farm_config` fields, and `nfe_invoice_tax_rates` / `detached_nfe_tax_rates`.

## Requirements

### Functional

- **R1** — Drop `nfe_invoice.icms_value`, `nfe_invoice.svc_endpoint_used`, `detached_nfe_invoice.svc_endpoint_used` via a new idempotent migration (next number after 000026), guarded per the 000022 `DO $$ … IF EXISTS` pattern. Existing migrations are not edited (append-only history; fresh DBs create-then-drop through the normal sequence).
- **R2** — Remove every Go reference to the dropped columns: `Invoice.ICMSValue` / `Invoice.SVCEndpointUsed` struct fields, `scanInvoice` scan slots, and the `icms_value` / `svc_endpoint_used` entries in the three `nfe_invoice` SELECTs (`GetInvoiceByDeparture`, `GetInvoicesByFarm`, `GetInvoiceByAccessKey`); `DetachedInvoice.SVCEndpointUsed` + scan slot + the `svc_endpoint_used` entries in the four `detached_nfe_invoice` SELECTs.
- **R3** — Keep `sent_at` on both tables and populate it. Stamp just before each of the **five** send sites:
  1. departure normal emission (`service/nfe_service/service.go`, before `SendToSefaz` after `UpdateInvoiceSignedXML`),
  2. departure SVC contingency attempt (`service/nfe_service/service.go`, before `SendToSefazWithEmission`),
  3. departure worker draft rebuild (`service/nfe_service/worker.go`, before `SendToSefazWithEmission`),
  4. detached normal emission (`service/nfe_service/detached_service.go`, before `SendToSefaz`),
  5. detached SVC contingency attempt (`service/nfe_service/detached_service.go`, before `SendToSefazWithEmission`).
  Semantics: **`sent_at` = timestamp immediately before the first transmission attempt of this row's signed XML.** It records attempts, including ones that fail or get refused (ambiguous SEFAZ timeouts still put the XML in flight), so the measurement is attempt→resolution. One stamp per row, never reset.
- **R4** — Model support: keep `UpdateInvoiceSent` (add an idempotency guard `AND sent_at IS NULL`) and add a mirrored `UpdateDetachedInvoiceSent`. Stamping failures are logged via the logger model and tolerated — they must never block or change the emission outcome.
- **R5** — `CreateDetachedInvoice` gains `dhCont interface{}` and `xJust string` parameters (placed after `tpEmis` to mirror `CreateInvoiceWithEmission`); the detached SVC contingency call passes `time.Now()` and the reason it already computes ("Indisponibilidade do ambiente de autorizacao da SEFAZ de origem"); the normal detached emission call passes `nil, ""` (matching departure behavior where `dh_cont`/`x_just` are written only for contingency).
- **R6** — Delete dead functions: `CreateInvoiceDeprecated`, `GetInvoiceTaxRates` (`model/nfe_model/model.go`), `GetDetachedInvoiceTaxRates` (`model/nfe_model/detached_model.go`). `UpdateInvoiceSent` is kept (now called).

### Non-Functional

- **N1** — Zero data loss: every dropped column is NULL on every row of every database (nothing ever wrote it). Verified with a count check on the target DB before the migration runs.
- **N2** — Stamping must not introduce any user-visible behavior change: same toasts, same statuses, same endpoints.
- **N3** — All work, including smoke runs, stays SEFAZ-offline: `NFE_SEFAZ_GUARD=block` for any running app, no `nfe_farm_config` seeded, no direct SEFAZ calls (AGENTS.md "SEFAZ Network Isolation").
- **N4** — No user-facing text changes; all messages already existing remain in pt-br as-is.

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | After migrations, `information_schema.columns` shows no `icms_value` on `nfe_invoice` and no `svc_endpoint_used` on `nfe_invoice` / `detached_nfe_invoice` | Manual SQL smoke on the e2e DB (localhost:5433). Justified: the repo has no migration test harness; corroborated by app boot + test suites |
| AC2 | No data lost: before the migration runs, `SELECT COUNT(*) FROM nfe_invoice WHERE icms_value IS NOT NULL` = 0 on the target DB | Same manual smoke, executed on the live DB **before** starting the app |
| AC3 | Code builds and existing suites stay green with the removed fields/functions | `go build ./...` + `make test-go` (PASS) |
| AC4 | No references remain to `CreateInvoiceDeprecated`, `GetInvoiceTaxRates`, `GetDetachedInvoiceTaxRates` | `grep` across repo returns only the deletion commit's absence (documented command in plan.md, manual check) |
| AC5 | Each of the five send call sites stamps exactly the row it is about to send, via `UpdateInvoiceSent` / `UpdateDetachedInvoiceSent`; model functions guard with `WHERE sent_at IS NULL` | Code inspection `grep` checklist in plan.md Phase 3. Automated unit coverage is impractical today: services call the concrete `*nfe_model.NFeModel` singleton with no injection seam (mock pattern exists only for the narrow response interface). Dependency injection on NFeService is explicitly out of scope |
| AC6 | Smoke parity: a guard-blocked detached emission attempt persists the row with `status='draft'` **and** `sent_at IS NOT NULL` (stamp happens before the refused send) | Manual smoke: app with `NFE_SEFAZ_GUARD=block` + e2e fixture user, emit via UI, SQL-verify the row. Justified: real SEFAZ flow is prohibited offline |
| AC7 | A hypothetical detached SVC contingency call writes `dh_cont`/`x_just` on the new row | Cannot be exercised without SEFAZ/SVC (prohibited). Accepted by focused code inspection of the `CreateDetachedInvoice` signature/call sites + AC3 suite |
| AC8 | E2E suite unaffected (the `detached_nfe_rascunho.spec.js` fixture INSERT names no dropped column) | `make test-e2e` (PASS) |

## Verification Results

<!-- To be filled during implementation: exact commands and PASS/FAIL/NOT RUN with reasons. -->
