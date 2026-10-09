# plan:nfe_dead_schema_cleanup

Implementation plan for `sdd.md` in this folder. Line numbers refer to the
codebase state at planning time (2026-10-09); re-locate by symbol name if they
drift.

## Test Strategy

Behaviors to cover, where they live, and why some checks are manual:

1. **Column drops** (AC1/AC2) — schema-level; no automated migration test
   harness exists in this repo. Verify by manual SQL smoke on the e2e DB
   (environment below). The drop statements are trivial guarded metadata
   changes; the migration system (`RunMigrations()`, embedded SQL, tracked in
   `schema_migrations`) is exercised on every app startup including e2e setup.
2. **Field/function removal** (AC3/AC4) — compile-time checks; `go build ./...`
   catches dangling references to deleted struct fields/functions; the Go suite
   proves behavior of the survivors unchanged.
3. **`sent_at` stamping** (AC5/AC6) — five call sites. Automated unit coverage
   is not feasible with the current seam: `NFeService` methods call the
   concrete `nfe_model.GetNFeModel()` singleton (the hand-written-mock pattern
   covers only narrow interfaces like `detachedResponseModel`). Verify by the
   Phase-3 grep checklist plus the guard-blocked manual smoke (stamp survives
   the refused send on a `draft` row). Making NFeService model-injectable is
   explicitly out of scope.
4. **Detached `dh_cont`/`x_just`** (AC7) — requires the SVC path, which needs
   SEFAZ (prohibited offline). Focused code inspection on the changed
   signature/call sites + the full Go suite.
5. **E2E** (AC8) — `make test-e2e`; guards both invoice list scans (which lose
   two SELECT columns each) and the detached pages/fixture.

Environment needs: e2e test DB at `localhost:5433` (`test/e2e`:
`bun run db:start`, `bun run db:seed` — seeded fixture user for the smoke run;
fixtures intentionally define **no** `nfe_farm_config`, and every app start in
this plan runs with `NFE_SEFAZ_GUARD=block`). App env vars must include
`DB_HOST/DB_USER/DB_PASS/DB_NAME/DB_PORT` (no defaults in code). PostgreSQL
client (`psql`) needed for the smoke SQL.

Pre-implementation: run the baseline suites and note pre-existing failures
before updating `Status: WIP` (see Phase 0).

## Phase 0 — Baseline (before any code change)

- `go build ./...`
- `go test ./model/nfe_model/ ./service/nfe_service/...`
- Note results and any pre-existing failures in `sdd.md` → set `Status: WIP`.

## Phase 1 — Migration: drop the dead columns

**File:** `model/armazenda_database/migrations/000027_drop_dead_nfe_columns.sql`
(next free number after 000026 — re-verify with `ls` at implementation time).

```sql
-- Guarded DO-block, pattern of 000022_detached_nfe_item_cfop.sql:
DO $$
BEGIN
    -- drop nfe_invoice.icms_value
    -- drop nfe_invoice.svc_endpoint_used
    -- drop detached_nfe_invoice.svc_endpoint_used
END $$;
```

- Columns NOT dropped: `sent_at` (stays, now populated), `dh_cont`, `x_just`
  (stays, now populated for detached SVC).
- Do **not** edit `000002_add_nfe_tables.sql` (migrations are append-only
  history; a fresh DB creates `icms_value` at 000002 and drops it at 000027).
- Add a pre-drop count query as a comment in the migration for operators
  (AC2 evidence lives in the smoke run, since the app applies the migration
  on startup).

## Phase 2 — Model layer: cleanup + signature change

**`model/nfe_model/model.go`**

- Remove `i.icms_value,` and `i.svc_endpoint_used,` (single line each) from the
  three SELECT column lists: `GetInvoiceByDeparture` (~:444),
  `GetInvoicesByFarm` (~:651), `GetInvoiceByAccessKey` (~:688).
- `scanInvoice` (~:722-769): drop the `icmsValue` local + scan slot +
  `inv.ICMSValue` assignment; drop the `svcEndpoint` local + scan slot +
  `inv.SVCEndpointUsed` assignment. Keep `SentAt` scanning (column stays).
- `Invoice` struct: remove `ICMSValue` (~:216) and `SVCEndpointUsed` (~:234).
- `UpdateInvoiceSent` (~:606): add idempotency guard —
  `WHERE id = $1 AND sent_at IS NULL`.
- Delete `CreateInvoiceDeprecated` (~:260-262) and `GetInvoiceTaxRates`
  (~:365-384-ish).

**`model/nfe_model/detached_model.go`**

- Remove `d.svc_endpoint_used,` from the four SELECTs (~:189, :227, :426, :469);
  drop the matching scan slot and `inv.SVCEndpointUsed = svcEndpoint`
  (~:347); remove the struct field (~:56).
- `CreateDetachedInvoice` (~:74-141): add `dh_cont, x_just` to the INSERT
  column list + `$n` placeholders + args; new signature params
  `dhCont interface{}, xJust string` immediately after `tpEmis int`
  (mirrors `CreateInvoiceWithEmission`'s `tpEmis, dh_cont, x_just` tail).
  Update its doc comment to state the contingency-only overlap with the
  departure flow.
- Add `UpdateDetachedInvoiceSent(id int) error` mirroring `UpdateInvoiceSent`
  (incl. the `sent_at IS NULL` guard).
- Delete `GetDetachedInvoiceTaxRates` (~:388-404).

**Tests in this phase:** `go test ./model/nfe_model/` after the edits
(model tests are pure-struct, they exercise `scanDetachedInvoice`'s struct
contract only indirectly; the real scan-vs-SQL parity guard is evidence-based
— see Phase 4). No new model unit tests are added: nothing here is testable
without a DB harness (noted in Test Strategy).

## Phase 3 — Service wiring: stamps + contingency fields

Stamper error handling pattern — mirror the tolerated transport persistence
(`CreateDetachedInvoiceTransport` call sites): log via
`model_error.GetLoggerModel().Log(...)` and continue.

**`service/nfe_service/service.go`**

1. Departure normal emission: between `UpdateInvoiceSignedXML(invoiceID, …)`
   (~:339) and `invService.SendToSefaz(...)` (~:344), insert
   `UpdateInvoiceSent(invoiceID)` (log on error, continue).
2. Departure SVC attempt (`attemptSVCContingency`): before the
   `SendToSefazWithEmission` (~:467), stamp `newInvoiceID` (created ~:449).

**`service/nfe_service/worker.go`**

3. Draft rebuild path: before `SendToSefazWithEmission` (~:271), stamp
   `newInvoiceID` (created ~:256). Worker code errors go through `return
   fmt.Errorf(...)` only for local failures; a stamp failure logs and
   continues (mirror `IncrementRetryCount` tolerance).

**`service/nfe_service/detached_service.go`**

4. `BuildDetachedInvoice`: before `SendToSefaz` (~:340), stamp `invoiceID`
   via `UpdateDetachedInvoiceSent`.
5. `attemptDetachedSVCContingency`:
   - before `SendToSefazWithEmission` (~:1023), stamp `newInvoiceID`;
   - in the `CreateDetachedInvoice` call (~:996-1003), insert `time.Now()`
     and the `reason` var (already computed at ~:970) after `int(tpEmis)`.
6. Normal-emission `CreateDetachedInvoice` call (~:310-315): insert `nil, ""`
   after `1` (tpEmis), matching the new signature.

**Phase-3 grep checklist (AC5 evidence — paste into Verification Results):**

```
grep -n "UpdateInvoiceSent(" service/nfe_service/          → exactly 3 sites (1, 2, 3)
grep -n "UpdateDetachedInvoiceSent(" service/nfe_service/  → exactly 2 sites (4, 5)
grep -n "AND sent_at IS NULL" model/nfe_model/             → both model functions
grep -n "dhCont" model/nfe_model/detached_model.go         → signature + INSERT + arg
```

Also update the comment at `detachedBuildData` / `CreateDetachedInvoice`
docstring if it claims persistence scope that changed.

## Phase 4 — Verification

Order matters for AC2: run the count **before** starting the app (the app
applies migrations on startup).

1. `go build ./...` — compile gate.
2. `go test ./model/nfe_model/ ./service/nfe_service/...` — focused suites.
3. `make test-go` — full Go suite (AC3).
4. `make test-e2e` — Docker; auto-starts DB + app + fixtures (AC8).
5. Manual smoke (AC1/AC2/AC6, `NFE_SEFAZ_GUARD=block` mandatory):

```bash
cd test/e2e && bun run db:start && bun run db:seed        # 5433, fixture user
psql "host=localhost port=5433 … " \
  -c "SELECT COUNT(*) FROM nfe_invoice WHERE icms_value IS NOT NULL"   # = 0 (AC2)
NFE_SEFAZ_GUARD=block DB_HOST=localhost DB_PORT=5433 DB_USER=… DB_PASS=… \
  DB_NAME=… air                                            # migrations run at boot
psql … -c "SELECT table_name, column_name FROM information_schema.columns
  WHERE column_name IN ('icms_value','svc_endpoint_used');"             # 0 rows (AC1)
# then, as the fixture user in the browser:
#   NF-e Avulsa → preencher destinatário/item → Emitir
#   → toast "SEFAZ e SVC indisponiveis" (guard refuses), row stays draft
psql … -c "SELECT status, sent_at, dh_cont, x_just FROM detached_nfe_invoice
  ORDER BY id DESC LIMIT 1"     # status='draft', sent_at IS NOT NULL (AC6)
```

6. **Stop the app process** after the smoke run (worker keeps ticking
   otherwise — AGENTS.md smoke rules).
7. Record every command + PASS/FAIL/NOT RUN in `sdd.md` → Verification
   Results; flip `Status: Done` only when all mapped ACs pass; keep WIP with
   notes otherwise.

## Risks & Notes

- **pgx scan parity is runtime-only.** Removing a column from a SELECT without
  its scan slot (or vice versa) compiles fine and explodes at query time. Three
  query families in `model.go` and four in `detached_model.go` — each column
  must be removed from the list **and** the scan in the same edit. The e2e
  suite exercises the list/access-key paths at runtime; the receipt/cancel
  flows exercise the rest.
- **Single-writer constraint**: only the binary built from this change may run
  against the DB during verification (single-binary app; `air` rebuilds, so
  stale processes are the hazard → Phase 4 step 6).
- **`sent_at` includes refused/failed attempts** (stamp precedes the attempt).
  Rows with `status='draft'` and `sent_at` set are expected; measurement AC is
  `authorized_at − sent_at` only compared for rows that reach `authorized`.
- **`sent_at IS NULL` guard** makes re-stamps impossible even if a send path is
  somehow re-entered (e.g., future worker draft re-send); supersede-created rows
  get their own stamp.
- **Out of scope (keep in mind, tracked from prior discussion):**
  - rate-snapshot drift (farm-config change between emission and draft
    rebuild produces a differently-rated XML) — candidate follow-up: persist
    merged rates at emission;
  - persisting computed ICMS/PIS/COFINS totals (the IBS/CBS precedent pattern);
  - any UI/report consuming `sent_at`;
  - making `NFeService`'s model dependency injectable to unit-test the stamping.

## Implementation notes (2026-10-09)

Deviations from the plan as written, and why:

1. **Manual UI smoke for AC6 is not executable offline.** The plan assumed the
   e2e fixture user could emit a detached NF-e, but the fixture intentionally
   carries empty `certificate_data` (`test/e2e/tests/detached_cfop_selector.spec.js:17-19`)
   and AGENTS.md prohibits seeding certificate bytes — signing fails before
   `CreateDetachedInvoice`, so no row would ever be created. The behavior was
   instead confirmed by a new DB-gated model test plus the AC5 call-site
   checklist; details in `sdd.md` → Verification Results.
2. **Added `model/nfe_model/invoice_lifecycle_model_test.go`** (double-gated by
   full DB_* + `RUN_DB_TESTS=1`, same pattern as
   `detached_transport_model_test.go`). It covers: `sent_at` stamping and the
   `sent_at IS NULL` guard on both tables; contingency vs normal
   `dh_cont`/`x_just` persistence; and runtime scan parity for every edited
   SELECT (`Get*ByAccessKey`, `Get*ByDeparture`, `Get*ByFarm`,
   `GetDraft/Pending*ForRetry`). Run each DB test in its own `go test -run`
   invocation: `InitNFeModel` is a singleton, so two DB tests sharing one
   package process would reuse the first test's (closed) pool.
3. **Smoke used a HEAD-binary-first sequence** to make AC2 meaningful: start
   the e2e DB fresh, boot a binary built at HEAD (migrations 1..26, columns
   exist), run the NULL-count queries, then boot the new binary and verify
   000027 applied and the columns are gone.
4. One pre-existing `gofmt` violation in `service/nfe_service/worker.go`
   (leading space on a line) was fixed while touching the file.
