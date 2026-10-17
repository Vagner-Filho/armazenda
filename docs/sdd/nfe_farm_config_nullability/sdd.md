# sdd:nfe_farm_config_nullability

## Status
Done

## Goal
Make farm NFe configuration saves partial and crash-free, and make the emission
requirements degrade gracefully:

- Saving `nfe_farm_config` with only certificate data (or no certificate data at
  all) must succeed. Today both crash with `SQLSTATE 23502` because
  `UpsertFarmConfig` explicitly inserts NULL into `cbs_cst` / `ibs_cst` /
  `c_class_trib` (NOT NULL since migration 000019) and `certificate_data`
  (NOT NULL since migration 000004) — the column defaults are dead letters
  because every INSERT lists those columns explicitly.
- A config row missing tax bits must degrade to the NT default constants at
  emission (already the behavior of `MergeEffectiveOverrides` / `codeOr`),
  never to a DB error.
- A farm config arriving without `environment` / valid `tax_regime` from a
  programmatic caller must be normalized / rejected with a clear pt-br
  conversation instead of violating the DB CHECK constraints (currently the
  router masks this by hardcoding `Environment = 2`).
- Emission with a config that has no certificate must fail with a clear pt-br
  warning naming the config screen, before the signer produces a cryptic error.

This SDD is independent from `docs/sdd/nfe_rate_defaults/` (rates never block
emission) — they were sliced apart so each can land and be reviewed alone.

## Requirements

### Functional

- **FR1 — Relax schema (migration `000031_relax_nfe_farm_config_nulls.sql`):**
  `ALTER TABLE nfe_farm_config` … `ALTER COLUMN … DROP NOT NULL` on
  `cbs_cst`, `ibs_cst`, `c_class_trib` and `certificate_data`. Aligns these
  columns with `default_icms_cst` / `default_pis_cst` / `default_cofins_cst`
  (already nullable, migration 000008). Ruin-safe re-runs: dropping NOT NULL on
  an already-nullable column is a no-op success. Decisions D1/D2
  (session review): relax the DB rather than default-fill in the model, so the
  config keeps the "explicitly set vs unset" distinction the service layer
  already understands.

- **FR2 — `UpsertFarmConfig` nil pass-through
  (`model/nfe_model/model.go`):** after the migration, nil
  `DefaultCBSCST` / `DefaultIBSCST` / `DefaultCClassTrib` and nil
  `CertificateData` persist as NULL. No code-side default filling.

- **FR3 — Environment normalization in the model
  (`model/nfe_model/model.go`):** `Environment == 0` → `2` (Homologação)
  before the INSERT. Reasoning: Go zero value means "not provided"; the DB
  default is 2; today `router/nfe_router/router.go::saveNFeConfig` hardcodes
  the same value, so the model becomes the single normalization point and the
  router hardcode is removed (FR4). Values 1/2 pass through untouched; other
  out-of-CHECK values keep the current DB-error behavior (out of scope).

- **FR4 — Router cleanup (`router/nfe_router/router.go::saveNFeConfig`):**
  remove the redundant `fConfig.Environment = 2` line. The config form always
  submits `environment=2` (the only enabled `<option>`), so behavior is
  unchanged.

- **FR5 — Tax-regime validation in the model:** before any DB write,
  `UpsertFarmConfig` rejects `TaxRegime ∉ {1,2,3}` with
  `*model_error.ModelError{IsServerErr: false, Message:
  "Regime tributário inválido: selecione Simples Nacional, Lucro Real ou Lucro
  Presumido"}`. The web handler's existing `IsServerErr == false` path surfaces
  it as a warning toast. Silent defaulting was rejected (D3 — session review):
  assigning a regime has fiscal consequences in the emitted XML (`CRT`), so an
  unidentified value is refused, not guessed.

- **FR6 — Emission certificate gate
  (`service/nfe_service/service.go::prepareInvoiceBuildData` and
  `service/nfe_service/detached_service.go::prepareDetachedBuildData`):**
  immediately after the respective farm-config-nil check, if
  `len(farmNFeConfig.CertificateData) == 0` return
  `entity_public.GetWarningToast("Certificado digital não configurado",
  "Envie o arquivo do certificado (.pfx, .p12) e a senha em NF-e em Configurações")`.
  Implemented as one shared unexported helper in `service/nfe_service` called
  by both paths (the detached/departure prepare functions are not
  model-injectable — tests are same-package pure-function tests).

- **FR7 — Read-path compatibility:** `GetFarmConfig` needs no changes — code
  columns already scan into `*string` (NULL → nil) and `certificate_data`
  scans into `[]byte` (NULL → nil slice). Emission XML is unchanged by these
  NULLs: `MergeEffectiveOverrides` and `codeOr` fall back to the identical NT
  constants (`'000'` IBS/CBS CST, `'000001'` cClassTrib) the UI placeholders
  suggest today.

### Non-Functional

- No new schema besides the `000031` migration; migration stays idempotent per
  the house convention (`ADD/DROP` no-ops on re-run).
- All user-facing messages stay pt-br; internal log lines stay English.
- All tests must stay offline — no test may reach SEFAZ (network guard rules in
  AGENTS.md). DB-gated model tests use only the local e2e test database
  (loopback).
- ON CONFLICT semantics change, accepted: saving the form with a blank
  "CST CBS/IBS Padrão" / "cClassTrib Padrão" now **clears** the stored value
  instead of crashing — the intended "(opcional) blank = unset" semantics,
  matching how `default_icms_cst` already behaves.

## Acceptance Criteria

| # | Acceptance criterion | Verification |
|---|----------------------|--------------|
| AC1 | Migration applied → `cbs_cst`, `ibs_cst`, `c_class_trib`, `certificate_data` are `is_nullable = YES` in `nfe_farm_config` | DB-gated test `TestMigration000031_FarmConfigColumnsNullable_DBRoundTrip` (`model/nfe_model/farm_config_nullability_model_test.go`, `RUN_DB_TESTS=1` against `information_schema.columns`) |
| AC2 | Cert-only upsert (only `FarmID` + `CertificateData` set, everything else Go zero values) succeeds; row stores `environment = 2`, NULL code columns, the cert bytes; `GetFarmConfig` returns nil `DefaultCBSCST` / `DefaultIBSCST` / `DefaultCClassTrib` and no 23502 | DB-gated `TestUpsertFarmConfig_CertOnly_DBRoundTrip` |
| AC3 | `TaxRegime = 0` (and any value outside 1–3) → `IsServerErr = false` `ModelError` with the FR5 pt-br message, and **no** `nfe_farm_config` row written | DB-gated `TestUpsertFarmConfig_RejectsInvalidTaxRegime_DBRoundTrip` |
| AC4 | Full config (codes, cert data, regime, serie, rates) round-trips unchanged through `UpsertFarmConfig` → `GetFarmConfig`, with environment 1 preserved (not coerced) | DB-gated `TestUpsertFarmConfig_FullRoundTrip_DBRoundTrip` |
| AC5 | Upsert with nil `CertificateData` succeeds (first save without certificate upload — the second latent crash from migration 000004) | DB-gated `TestUpsertFarmConfig_NullCertificateData_DBRoundTrip` (also asserts `GetFarmConfig` scans NULL `certificate_data` without error) |
| AC6 | Certification gate: an emission attempt with a cert-less (or config-missing cert) farm returns the FR6 warning toast on both paths | Pure-unit `TestNFeCertGate_Helper` (`service/nfe_service`, same-package style used by `default_payment_test.go`) + wiring reviewed; manual smoke documented as NOT RUN / justification in `plan.md` Phase 3 if skipped (prepare functions not model-injectable) |
| AC7 | Environment normalization lives only in the model (`fConfig.Environment = 2` removed from the router); a form save with `environment=2` still stores 2 | Covered by AC2/AC4 round-trips through the service path plus a grep-level review note in `Verification Results` (mechanical, no runtime difference) |

## Verification Results
- `go test -run '^TestNFeCertGate' -v ./service/nfe_service/` → PASS (TestNFeCertGate_Helper, 4 subtests)
- `DB_* + RUN_DB_TESTS=1 go test -run '^TestMigration000031_FarmConfigColumnsNullable_DBRoundTrip$' -v ./model/nfe_model/` → PASS
- `DB_* + RUN_DB_TESTS=1 go test -run '^TestUpsertFarmConfig_CertOnly_DBRoundTrip$' -v ./model/nfe_model/` → PASS
- `DB_* + RUN_DB_TESTS=1 go test -run '^TestUpsertFarmConfig_RejectsInvalidTaxRegime_DBRoundTrip$' -v ./model/nfe_model/` → PASS
- `DB_* + RUN_DB_TESTS=1 go test -run '^TestUpsertFarmConfig_FullRoundTrip_DBRoundTrip$' -v ./model/nfe_model/` → PASS
- `DB_* + RUN_DB_TESTS=1 go test -run '^TestUpsertFarmConfig_NullCertificateData_DBRoundTrip$' -v ./model/nfe_model/` → PASS
- `go build ./...` → PASS
- `go test ./...` → pre-existing failure `TestDeparturePaymentAndBillingDefaults` (unrelated, present on baseline) and DB-gated tests skipped without RUN_DB_TESTS
- Manual smoke of emission certificate gate: NOT RUN (prepare functions not model-injectable; helper logic unit-covered) — justification per plan.md Phase 3
