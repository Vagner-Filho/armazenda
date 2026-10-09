# sdd:nfe_dead_schema_cleanup

## Status
Done

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
| AC5 | Each of the five send call sites stamps exactly the row it is about to send, via `UpdateInvoiceSent` / `UpdateDetachedInvoiceSent`; model functions guard with `WHERE sent_at IS NULL` | Code inspection `grep` checklist in plan.md Phase 3 (PASS) **plus** DB-gated model tests proving the stamp persists and the guard holds (`invoice_lifecycle_model_test.go`). Automated unit coverage of the service call sites remains impractical: services call the concrete `*nfe_model.NFeModel` singleton with no injection seam. Dependency injection on NFeService is explicitly out of scope |
| AC6 | Smoke parity: a guard-blocked detached emission attempt persists the row with `status='draft'` **and** `sent_at IS NOT NULL` (stamp happens before the refused send) | UI smoke **not executable offline**: the e2e fixture intentionally carries empty certificate bytes (`test/e2e/tests/detached_cfop_selector.spec.js:17-19`) and AGENTS.md prohibits seeding certificate data, so signing fails before any row is created. Behavior confirmed by best-available checks: DB-gated test proves the stamp persists on a `draft` row and is idempotent; the AC5 checklist proves stamp-before-send ordering; guard tests/e2e prove sends are refused offline (PASS) |
| AC7 | A hypothetical detached SVC contingency call writes `dh_cont`/`x_just` on the new row | SEFAZ/SVC unavailable (prohibited). Covered by a DB-gated model test calling `CreateDetachedInvoice` with contingency parameters (tpEmis=6, dhCont, xJust) and asserting the persisted row, plus focused code inspection of the signature/call sites + AC3 suite (PASS) |
| AC8 | E2E suite unaffected (the `detached_nfe_rascunho.spec.js` fixture INSERT names no dropped column) | `make test-e2e` (PASS) |

## Verification Results

### Baseline (before any code change)

- `go build ./...` — PASS
- `go test ./model/nfe_model/ ./service/nfe_service/...` — PASS (no pre-existing failures)

### Final results (2026-10-09)

- `go build ./...` — PASS
- `make test-go` — PASS (exit 0; no FAIL lines)
- `make test-js` — PASS (exit 0; no JS changed, run as regression guard)
- `gofmt -l model/nfe_model/ service/nfe_service/` — clean

**AC1 — dropped columns absent (PASS).** Booting the new binary against the pre-000027 database logged `applied migration 000027: drop dead nfe columns`; then:

```
SELECT table_name, column_name FROM information_schema.columns
WHERE column_name IN ('icms_value','svc_endpoint_used');
-- 0 rows
```

**AC2 — no data lost (PASS).** Before the migration ran (app built at HEAD, migrations 1..26 applied; `schema_migrations` max=26), the three columns existed and every row was NULL:

```
SELECT COUNT(*) FROM nfe_invoice WHERE icms_value IS NOT NULL;            -- 0
SELECT COUNT(*) FROM nfe_invoice WHERE svc_endpoint_used IS NOT NULL;     -- 0
SELECT COUNT(*) FROM detached_nfe_invoice WHERE svc_endpoint_used IS NOT NULL; -- 0
```

Corroborating history check: `git grep "icms_value\s*=\|svc_endpoint_used\s*=" $(git rev-list --all) -- '*.go'` returns nothing — no Go code ever wrote these columns.

**AC3 — build + Go suite (PASS).** See commands above.

**AC4 — dead functions/fields gone (PASS).** `grep -rn "CreateInvoiceDeprecated\|GetInvoiceTaxRates\|GetDetachedInvoiceTaxRates\|ICMSValue\|SVCEndpointUsed\|svc_endpoint_used\|i\.icms_value" --include="*.go" .` → no matches. Remaining textual matches are only the SDD/plan docs, the append-only history migrations (000002/000005/000021) and the new 000027 migration itself.

**AC5 — five stamp sites + guards (PASS).** Phase-3 checklist:

```
grep -n "UpdateInvoiceSent(" service/nfe_service/*.go          → service.go:345, service.go:474, worker.go:272   (3 sites: normal, SVC, worker rebuild)
grep -n "UpdateDetachedInvoiceSent(" service/nfe_service/*.go  → detached_service.go:341, :1030                (2 sites: normal, SVC)
grep -n "AND sent_at IS NULL" model/nfe_model/*.go             → model.go:582, detached_model.go:275           (both guards)
```

**AC6 — stamp precedes the refused send (PASS via substitute checks; original UI smoke NOT RUN).**
The planned UI smoke cannot execute with the current e2e fixture: it seeds `nfe_farm_config` with empty `certificate_data`, so `BuildDetachedInvoice` fails at `BuildAndSign` → `LoadCertificateFromBytes` (empty PKCS#12) before `CreateDetachedInvoice`/the stamp; the fixture's own comment documents "empty certificate bytes; do not add certificate data or call emission/cancellation paths" (`test/e2e/tests/detached_cfop_selector.spec.js:17-19`), and AGENTS.md prohibits seeding certificate bytes into a test/smoke DB. Best-available confirmation (all PASS):

- DB-gated `TestDetachedInvoice_ContingencyFieldsAndSentStamp_DBRoundTrip` (`RUN_DB_TESTS=1` + DB_* against the e2e DB): `sent_at` NULL before the stamp, set after `UpdateDetachedInvoiceSent`, and a sentinel value survives a second stamp (guard holds) — i.e. a refused send still leaves the stamp on the draft row;
- AC5 checklist: every stamp call immediately precedes its send call;
- e2e/app run with `NFE_SEFAZ_GUARD=block` proves the sends are refused offline (`pkg/nfe/sefaz/guard_test.go` + app boot).

Commands:
```
RUN_DB_TESTS=1 DB_HOST=... go test -v -count=1 -run '^TestDetachedInvoice_ContingencyFieldsAndSentStamp_DBRoundTrip$' ./model/nfe_model/  → PASS
RUN_DB_TESTS=1 DB_HOST=... go test -v -count=1 -run '^TestInvoice_SentStamp_DBRoundTrip$' ./model/nfe_model/                        → PASS
```

**AC7 — detached dh_cont/x_just persistence (PASS).** Same DB-gated test creates a detached invoice with `tpEmis=6`, `dhCont=time.Now()`, `xJust=<reason>` and asserts the persisted row (`tp_emis=6`, `dh_cont` not NULL, `x_just=reason`), and that a normal emission (`1, nil, ""`) stores `dh_cont NULL` / `x_just` empty. SEFAZ/SVC itself remains untouchable offline; call-site inspection confirms the SVC path passes `time.Now(), reason`.

**AC8 — E2E suite (PASS with baseline-parity note).** `make test-e2e`: 103 passed / 16 failed / 4 skipped. Baseline at HEAD (same environment): 105 passed / 14 failed / 4 skipped — the failures are pre-existing WebKit timeouts/`WebKit encountered an internal error` (plus one Firefox checkbox flake) under full parallel load, present on both trees; all chromium tests pass on both. The four delta tests pass in isolation on this branch:

```
cd test/e2e && bunx playwright test tests/detached_cfop_selector.spec.js tests/detached_nfe_transport.spec.js \
  --project=firefox --project=webkit -g "seletor de CFOP|renders the three sub-groups"   → 4 passed
```

The suite also exercised the fresh-DB path (create `icms_value`/`svc_endpoint_used` at 000002/000005/000021 → drop at 000027) with the app booting and listing detached invoices normally.

### Extra coverage added

`model/nfe_model/invoice_lifecycle_model_test.go` (double-gated: full DB_* set + `RUN_DB_TESTS=1`, skipped by `make test-go` without the e2e DB):

- stamps/guards for both tables (AC5/AC6);
- contingency vs normal `dh_cont`/`x_just` persistence (AC7);
- runtime scan parity for **every** edited SELECT list — `GetDetachedInvoiceByAccessKey`, `GetDetachedInvoicesByFarm`, `GetDraft/PendingDetachedInvoicesForRetry`, `GetInvoiceByAccessKey`, `GetInvoiceByDeparture`, `GetInvoicesByFarm`, `GetDraft/PendingInvoicesForRetry` — all PASS against the post-000027 schema, closing the plan's "pgx scan parity is runtime-only" risk.

### Notes

- All smoke runs used `NFE_SEFAZ_GUARD=block`; the e2e DB (docker, tmpfs, port 5433) was destroyed after verification; no `nfe_farm_config` certificate data was ever seeded (the fixture's empty-bytes row is the pre-existing test default).
- The manual-smoke deviation for AC6 is recorded in the AC table above; the criterion's behavior is confirmed, but not through a single UI emission (impossible without a certificate).
