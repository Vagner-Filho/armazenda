# plan:nfe_farm_config_nullability — Implementation Plan

Implements the spec in `docs/sdd/nfe_farm_config_nullability/sdd.md`.
House rules honored: sdd/plan before code; tests live in the same phase as the
code that fulfills them; regression test written before/with the fix.

## Test strategy (before implementation)

Behaviors to cover and where:

| Behavior | Test location / layer | Type | Needs |
|----------|----------------------|------|-------|
| `000031` makes 4 columns nullable | `model/nfe_model/farm_config_nullability_model_test.go` (same dir, `package nfe_model_test`) | DB round-trip | local e2e Postgres + `RUN_DB_TESTS=1` |
| cert-only / nil-cert / complete / regime-0 upserts | same gated file, one test per AC (AC2–AC5) | DB round-trip (`UpsertFarmConfig` → raw pool assertions → `GetFarmConfig`) | same DB, farm row seeding for the FK |
| certificate gate helper (return toast vs nil) | `service/nfe_service/cert_gate_test.go` — same-package pure test, mirroring `default_payment_test.go` style | unit | none |
| wiring of the gate into `prepareInvoiceBuildData` + `prepareDetachedBuildData` | code review only (prepare functions fetch the model singleton; not injectable) | manual | — |

Environment needs:

- Test DB: `cd test/e2e && bun run db:start && bun run db:seed`, then run the
  gated tests with the DB_* env plus `RUN_DB_TESTS=1` (same double-gating and
  wording as `dbRoundTripPool` in `invoice_lifecycle_model_test.go`).
- Pay attention to SEFAZ isolation rules (AGENTS.md): nothing here touches
  SEFAZ; do not seed `nfe_farm_config` with real certificate bytes into a DB a
  running app connects to for smoke purposes beyond what the tests clean up.

Focused commands:

```bash
go test -run 'TestMigration000031|TestUpsertFarmConfig' -v ./model/nfe_model/   # with RUN_DB_TESTS=1 + DB_* env
go test -run 'TestNFeCertGate' -v ./service/nfe_service/
make test-go   # completion check
```

## Phase 1 — Migration + gated schema/upsert tests

1. Write `model/nfe_model/farm_config_nullability_model_test.go` — gated on
   `RUN_DB_TESTS=1` + `DB_*` (reuse `dbRoundTripPool`).
   Tests:
   - `TestMigration000031_FarmConfigColumnsNullable_DBRoundTrip` (AC1):
     `information_schema.columns` asserts `is_nullable='YES'` for the 4
     columns.
   - `TestUpsertFarmConfig_CertOnly_DBRoundTrip` (AC2, **regression-first**):
     seed a farm row (FK target), call `InitNFeModel(pool)`-backed
     `UpsertFarmConfig` with only `FarmID` + `CertificateData` — must succeed
     (today it fails with 23502) — then assert via raw SQL + `GetFarmConfig`:
     `environment=2`, NULL `cbs_cst`/`ibs_cst`/`c_class_trib`, cert bytes
     stored.
2. Apply the migration by starting the app once against the test DB
   (`RunMigrations` on startup) or running the SQL directly on the test DB.
3. Run the gated tests; AC1+AC2 PASS, AC3–AC5 pending.

**Note:** this phase reproduces the reported crash as a failing test before
the migration exists (run it twice: pre- and post-migration).

## Phase 2 — Model normalization + regime validation

1. `model/nfe_model/model.go`:
   - Normalize `Environment == 0 → 2` at the top of `UpsertFarmConfig` (FR3).
   - Add the FR5 regime validation before any DB write (FR5 message wording
     exactly as in the sdd).
2. Tests in the same gated file:
   - `TestUpsertFarmConfig_RejectsInvalidTaxRegime_DBRoundTrip` (AC3) —
     regime 0 and 4; assert `IsServerErr == false`, the pt-br message, and
     `SELECT COUNT(*)` unchanged for the farm.
   - AC2 assertions for `environment = 2` live here too (coercion path).
3. Run: gated suite (AC3 green), plus unchanged suites
   `go test ./model/...`.

## Phase 3 — Emission certificate gate

1. `service/nfe_service/cert_gate.go` (new, small): unexported helper, e.g.
   `nfeCertGate(cfg *entity_public.FarmConfig) *entity_public.Toast` — nil
   when certificate data present, else the FR6 warning toast.
2. Wire into `prepareInvoiceBuildData` (service.go) right after the
   farm-config-nil check, and into `prepareDetachedBuildData`
   (detached_service.go) at the equivalent point (FR6).
3. Unit test `service/nfe_service/cert_gate_test.go` (AC6):
   - nil config / nil cert bytes / empty bytes → toast with both strings;
   - cert bytes present → nil.
4. Wiring review note: both call sites return exactly when the helper is
   non-nil. Manual smoke optional and only on SEFAZ-free paths with
   `NFE_SEFAZ_GUARD=block`; if skipped, record as NOT RUN with the reason
   (prepare functions not model-injectable; helper logic unit-covered).

## Phase 4 — Router cleanup + completion

1. `router/nfe_router/router.go::saveNFeConfig`: delete
   `fConfig.Environment = 2` (FR4).
2. `gofmt -l .` clean (or `gofmt -w` on touched files).
3. Completion checks: `make test-go`; gated suite run with the test DB;
   record exact commands and PASS/FAIL/NOT RUN in
   `docs/sdd/nfe_farm_config_nullability/sdd.md` → `Insert Verification
   Results`; set `Status: Done` only if all ACs verified.

## Risks / out of scope

- **Out of scope:** rate-column nullability (zero = not-configured stays the
  contract; rates-policy is `docs/sdd/nfe_rate_defaults/`), serie coercion
  (0 is legal for CNPJ emitters; the router already defaults `Serie==0 → 1`),
  removing the `environment`/`tax_regime`/`default_mod_frete` CHECK
  constraints, XML/danfe changes, e2e suite additions.
- **Known semantic shift:** partial saves now persist NULL codes; a later
  full-save consistently overwrites. Existing rows keep their values until a
  save touches them (no data migration).
- **Risk:** `certificate_password_encrypted` stays NOT NULL; an empty string
  is stored when no password is provided (today's behavior) — later emission
  fails at decryption with the existing signer error. Not worsened by this
  SDD; cert-gate toast (FR6) only reacts to missing certificate bytes.
