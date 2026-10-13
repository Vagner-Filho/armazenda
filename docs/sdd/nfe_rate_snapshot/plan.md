# plan:nfe_rate_snapshot

Implementation plan for `sdd.md` in this folder.

**Tree precondition:** `docs/sdd/nfe_dead_schema_cleanup/` is Done
(2026-10-09). Build on the post-cleanup tree — it already contains migration
`000027_drop_dead_nfe_columns.sql`, `UpdateInvoiceSent` /
`UpdateDetachedInvoiceSent` wired at the five send sites (`service.go:345`,
`:474`, `worker.go:272`, `detached_service.go:341`, `:1030`), and the
double-gated DB test `model/nfe_model/invoice_lifecycle_model_test.go`. Line
numbers refer to that tree as of 2026-10-11; re-locate by symbol name.

## Test Strategy

Behaviors to cover, where they live, and why some checks are DB-gated:

1. **Resolution + validation invariants (pure functions — core regression
   guard).** New `service/nfe_service/rates_test.go`:
   - `TestMergeRates_SnapshotIsIdempotent` — `MergeRates(merged, cfg) ==
     merged` when all five rates are non-nil.
   - `TestMergeRates_MissingAxesSurfaceInValidation` — with no user rates and
     an unconfigured farm, the merged struct has nil pointers and
     `validateResolvedRates` reports exactly those axes (never a guessed
     value).
   - `TestValidateResolvedRates_SimpleNacionalICMSOptional` — SN regime does
     not require ICMS; Regime Normal does.
   - `TestValidateResolvedRates_UserZeroIsExempt` — a user-typed zero is not
     reported as missing (the `TaxRates` contract, `invoice.go:150-155`).
   - `TestValidateResolvedRates_RequiredInBothRegimes` — missing
     PIS/COFINS/IBS/CBS reported for both regimes; messages in pt-br naming
     each axis.
   - `TestMergeEffectiveOverrides_*` — the codes helper (R5): user code >
     farm default > builder default per field; ICMS field resolves CST vs
     CSOSN per regime; resolved values always non-empty. Helper is pure — no
     fixtures.
2. **Persist-the-snapshot behavior (AC4/AC8).** DB-gated model tests, per the
   cleanup's proven pattern (double-gated: full `DB_*` env + `RUN_DB_TESTS=1`;
   **one `go test -run` invocation per test** — `InitNFeModel` is a singleton,
   two DB tests in one process reuse the first test's closed pool). Either
   extend `invoice_lifecycle_model_test.go` or add
   `invoice_tax_snapshot_model_test.go`: persist via the real
   `CreateInvoiceWithEmission` / `CreateDetachedInvoice` with a snapshot
   (5 rates + 6 resolved codes), read back through the public getters
   (`GetInvoiceByAccessKey` / `GetDetachedInvoiceByAccessKey`), assert all
   five rate columns non-null and the six code columns populated with the
   resolved values; also assert a **no-override emission** freezes the
   farm-config values (the fallback arm no longer exists).
   `NFE_SEFAZ_GUARD=block` must be set for any app-adjacent run (guard rules).
3. **Single-source structure (AC2/AC5).** Grep checklist — exactly one
   `MergeRates` call per prepare function, none inside
   `buildItems`/`buildDetachedItems`; code resolution only in
   `MergeEffectiveOverrides`; `pickRate`/`hasAnyUserRate` gone; no fallback
   rate literals left in production files.
4. **No schema change** — the six code columns and the five rate columns
   already exist on their respective tables; no migration in this SDD.
5. **Environment needs:** e2e DB at `localhost:5433` (`test/e2e`:
   `bun run db:start`, `bun run db:seed`), `psql` for any manual inspection,
   `DB_HOST/DB_USER/DB_PASS/DB_NAME/DB_PORT` (no defaults in code).
   **No UI-emission smoke is planned** — see the sdd.md note: the fixture's
   empty `certificate_data` (`test/e2e/tests/detached_cfop_selector.spec.js:17-19`)
   makes signing fail before persist, and seeding cert bytes is prohibited.

Pre-implementation: baseline suites in Phase 0, note pre-existing failures,
then set `Status: WIP`.

## Phase 0 — Baseline (post-cleanup tree)

- `go build ./...`
- `go test ./service/nfe_service/... ./model/nfe_model/`
- Record results (incl. any pre-existing failures) in `sdd.md`; set
  `Status: WIP`.

## Phase 1 — Pure layer: two-source resolution, validation, snapshot struct

**`service/nfe_service/service.go`**

1. Define the snapshot struct next to the resolution helpers:

```go
// emissionTaxSnapshot carries the effective tax configuration of one
// emission: the merged rates the item XML used and the resolved tax codes.
// It is what gets persisted so rebuilds reproduce the emission exactly.
type emissionTaxSnapshot struct {
    Rates entity.TaxRates
    Codes entity.InvoiceOverrides // only the six tax fields resolved; user fields copied through
}
```

2. Rework `MergeRates` (`:946-959`): per axis
   `resolveRate(override *decimal.Decimal, farm decimal.Decimal) *decimal.Decimal`
   — override non-nil wins (including zero = explicit exempt); farm non-zero
   wins; otherwise nil. **Delete `pickRate` and the five hardcoded fallback
   arguments.** All five returned pointers may now be nil. Rewrite the
   function doc comment (drop the "last-resort default" language).
3. Implement `validateResolvedRates(rates entity.TaxRates, regime
   defaults.TaxRegime) []string` — per R10/R11: ICMS required only for
   `defaults.TaxRegimeNormal`; PIS/COFINS/IBS/CBS always; nil pointer =
   missing; pt-br message per missing axis naming it and the two places to
   fix (emission form or farm config).
4. Implement `MergeEffectiveOverrides(user *entity.InvoiceOverrides,
   farmConfig *entity_public.FarmConfig, regime defaults.TaxRegime)
   entity.InvoiceOverrides` — copies user fields as-is, resolves the six tax
   fields with the exact precedence/constants the builders use today
   (`CSTTributadaIntegral` / `CSOSNSemPermissaoCredito` per regime; PIS/COFINS
   "01"; `IBSCBSCSTTributadaIntegral`; `CClassTribDefault`). Pure; no model
   access.
5. Unit tests from Test Strategy item 1 — write them now; they are the
   regression guard for every later phase. `gofmt` as you go.

## Phase 2 — Departure flow: validate, return, persist the snapshot

**`service/nfe_service/service.go`**

1. `prepareInvoiceBuildData` (`:46`): add a 6th return value
   `emissionTaxSnapshot`; at `:181` compute
   `rates := MergeRates(userRates, farmNFeConfig)` then
   `ratesErrs := validateResolvedRates(rates, defaults.TaxRegime(farmNFeConfig.TaxRegime))`
   and `codes := MergeEffectiveOverrides(overrides, farmNFeConfig, ...)` —
   **validation fires here**, returning the pt-br warning toast (existing
   config-incomplete toast shape) before `buildItems` runs (which derefs the
   rate pointers at `:795-799`) and before any number allocation downstream.
2. Rewrite the `:178-180` comment ("the raw user input is returned separately
   so the caller can persist a partial-override row") → snapshot semantics.
3. `buildItems` (`:790`): stop resolving the six tax fields inline — receive
   the pre-resolved `codes` (new param) and an explicit `regime
   defaults.TaxRegime` (no longer read from farm config for tax decisions);
   consume `codes` fields. The emitted XML for identical validated inputs is
   byte-identical (N2).
4. `BuildInvoiceFromDeparture`: destructure the new return at `:262`; replace
   the raw-or-nil block `:306-309` with
   `ratesToPersist := &snap.Rates` / `codesToPersist := &snap.Codes`, passing
   both at the create call `:326-329`. The existing effective-infCpl
   injection (`:310-317`) already runs before persist — keep it;
   `MergeEffectiveOverrides` does not touch `InfCpl`.
5. `attemptSVCContingency`: the call at `:361` passes the snapshot through;
   update the parameter docs inside (`taxRates` → snapshot semantics) and the
   create call `:449-453` (pass `&snap.Rates` / `&snap.Codes`).
6. Preview caller `:638`: destructure the new return as `_` (validation in
   prepare covers the preview for free).
7. `model/nfe_model/model.go`: doc comment only — `CreateInvoiceWithEmission`
   params flip from user-provided rates/overrides to effective-rate/code
   snapshot; no signature/logic changes.

## Phase 3 — Worker: prepared snapshot + graceful unresolvable-rate handling

**`service/nfe_service/worker.go`**

1. `:227` — destructure the new 6th return of `prepareInvoiceBuildData`.
2. `:243-248` — replace the `hasAnyUserRate` branch with
   `ratesToPersist := &snap.Rates`, `codesToPersist := &snap.Codes` (not the
   `:219` pre-merge: identical for complete snapshots — idempotence — and
   exact for legacy partial rows).
3. **Unresolvable rates (R12):** the prepare-failure branch (`:228-230`,
   toast from `validateResolvedRates`) currently returns a plain error
   without touching the retry counter — extend it to
   `IncrementRetryCount(inv.ID)` first (mirror the SVC-unavailable branch at
   `:195`) so a permanently unresolvable legacy draft logs on a bounded,
   backoff-visible schedule and stays draft.
4. Update comments `:212-218` (legacy-rows-only note) and `:243-244`.

## Phase 4 — Detached flow: hoist resolution + validation

**`service/nfe_service/detached_service.go`**

1. `detachedBuildData` — new fields `RatesSnapshot entity.TaxRates`,
   `CodesSnapshot entity.InvoiceOverrides`.
2. `prepareDetachedBuildData` — compute
   `ratesSnapshot := MergeRates(input.TaxRates, farmNFeConfig)` and
   `codesSnapshot := MergeEffectiveOverrides(input.Overrides, farmNFeConfig, ...)`
   **immediately after the farm-config fetch and before recipient
   resolution** (so the validation toast fires before the emission path
   creates a person row for an inline recipient — no wasted side effects),
   then `validateResolvedRates` and attach both to the returned struct.
3. `buildDetachedItems` — signature becomes
   `(inputs, farmConfig, snapshot emissionTaxSnapshot, overrides *entity.InvoiceOverrides)`
   (or the two pre-resolved params, whichever reads cleaner); delete the
   internal `rates := MergeRates(userRates, farmConfig)` at `:562` and the
   inline CST/CSOSN/PIS/COFINS/IBS/CBS/cClassTrib resolution (`:602-660`
   region) — consume `snapshot.Codes` fields instead. Keep the product-level
   override application (ProductName/NCM/CEST/unit) untouched. The detached
   CST-resolution unit-test assertions move to the helper tests (Phase 1).
4. `BuildDetachedInvoice`: replace `:299-303` with the snapshot pair; pass
   them at the create call (`:314-316` region) and to the SVC attempt
   (`:354`).
5. `attemptDetachedSVCContingency`: receives and forwards the snapshot pair
   into `CreateDetachedInvoice`; doc comment updated.
6. `model/nfe_model/detached_model.go`: `CreateDetachedInvoice` doc comment —
   same flip as Phase 2 item 7 (signature unchanged; cleanup already added
   `dhCont`/`xJust`).

## Phase 5 — Dead code, test updates, docs

1. Delete `hasAnyUserRate` (`service.go:984-987`); `grep -rn
   "hasAnyUserRate\|pickRate" --include="*.go"` → 0 hits (AC5).
2. `grep -n "0\.17\|0\.0165\|0\.076" service/nfe_service/*.go` → only test
   fixtures remain; no production fallback literals.
3. Rewrite the `buildDetachedItems` test call sites to pass the pre-resolved
   snapshot — `nfe_service.MergeRates(entity.TaxRates{}, cfg)` output plus
   the codes helper output, qualified per each test file's package
   declaration (`nfe_service_test` vs `nfe_service`):
   - `service/nfe_service/detached_preview_test.go` ~`:224`, ~`:298`
   - `service/nfe_service/detached_profile_service_test.go` ~`:102`
   `TestBuildDetachedItems_DefaultsAndTaxTotals` (which relied on the
   internal merge of empty user rates) now supplies explicit rates on the
   farm config or via the snapshot and asserts the same values. Check for
   `buildItems` test call sites the same way; update any `MergeRates`-using
   tests found by `grep -n "MergeRates(" service/nfe_service/*_test.go`.
4. Add/adjust the DB-gated snapshot round-trip tests (Test Strategy item 2),
   following the cleanup's pattern notes (double-gated; one `-run` per
   invocation).
5. Docs: builder/persist/worker comments flipped to snapshot semantics; the
   `parsePercentRateOrNil` doc's "(0.17)" example may stay (it documents the
   *form parsing* convention, not resolution). Update **AGENTS.md** — the Tax
   Reform Gate line "resolves `user > farm > 2026 fallback`" becomes
   "resolves `user > farm`; missing rates are a validation error (no
   fallback)". Also N2's user-visible-change caveat lands in the commit
   message.
6. `defaults.IBSRate2026`/`CBSRate2026` and `reform_test.go` stay untouched
   (documented legal rates, R13).

## Phase 6 — Verification

1. `go build ./...` — compile gate.
2. `go test ./service/nfe_service/ ./model/nfe_model/` (focused; iterate with
   `-run TestMergeRates|TestValidateResolvedRates|TestMergeEffectiveOverrides|TestBuildDetachedItems`).
3. `make test-go` (AC6).
4. `make test-e2e` (AC6; compare against the cleanup's WebKit-flake baseline:
   105 passed / 14 failed / 4 skipped at HEAD; verify any delta tests in
   isolation per the cleanup notes).
5. DB-gated snapshot tests (AC4/AC8):

```bash
cd test/e2e && bun run db:start && bun run db:seed
NFE_SEFAZ_GUARD=block RUN_DB_TESTS=1 DB_HOST=localhost DB_PORT=5433 \
  DB_USER=… DB_PASS=… DB_NAME=… \
  go test -v -count=1 -run '^TestInvoice_TaxSnapshot_DBRoundTrip$' ./model/nfe_model/
NFE_SEFAZ_GUARD=block RUN_DB_TESTS=1 DB_HOST=localhost DB_PORT=5433 \
  DB_USER=… DB_PASS=… DB_NAME=… \
  go test -v -count=1 -run '^TestDetachedInvoice_TaxSnapshot_DBRoundTrip$' ./model/nfe_model/
# then destroy the DB container as the cleanup SDD did
```

6. **Stop the app process** after any app-adjacent run (worker hazard rule).
7. Record every command + PASS/FAIL/NOT RUN in `sdd.md` → Verification
   Results; flip `Status: Done` only when all mapped ACs pass; keep `WIP`
   with notes otherwise.

## Risks & Notes

- **Behavior change is intentional and user-visible:** farms relying on the
  old silent fallbacks (ICMS 17%, PIS 1.65%, COFINS 7.6%, 2026 IBS/CBS) now
  get a blocking pt-br validation toast until they configure the farm rates
  or type them per emission. The per-emission typed-zero escape hatch
  (explicit exemption) is preserved (AC2). This trade-off is the point of the
  change — record it in the delivery summary.
- **`snapshot.Codes` must be a full override struct**, not just codes: the
  create functions persist `NaturezaOp`/`ProductDesc`/`CEST`/`Unit`/
  `ModFrete`/`InfCpl` from the same overrides param — the helper copies user
  fields through so nothing becomes NULL.
- **Validation ordering matters twice:** departure — before `buildItems`
  derefs the rate pointers (`:795-799`); detached — before the emission path
  creates the inline recipient person row (R12/Phase 4 step 2).
- **Snapshot struct evolution:** if a 7th attribute later needs freezing, add
  it to `emissionTaxSnapshot` and the helper — never re-introduce inline
  resolution in the builders.
- **Regime residual (accepted):** resolved ICMS code frozen; rebuild
  re-derives the CST/CSOSN branch from the current regime — sdd.md documents
  it; the signed XML stays authoritative and untouched.
- **pgx scan parity is runtime-only** (cleanup lesson): the six code columns
  are already in every SELECT; the runtime scan-parity DB tests added by the
  cleanup must keep passing — do not touch SELECT lists in this SDD.
- **Out of scope:** nominal-value columns (derivable + XML-authoritative),
  `rate_source`, regime freezing, backfills, non-tax override freezing,
  preview staleness, config-form warnings, NFeService model injection.
