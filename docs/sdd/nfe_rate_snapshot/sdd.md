# sdd:nfe_rate_snapshot

## Status
Done

## Goal

Make each emission's tax layer deterministic, complete, and honest:

**resolve (user > farm) → validate (never guess) → persist (snapshot).**

1. **Resolve:** rates resolve from user-typed form values and farm config only.
2. **Validate:** a rate the emission actually needs that resolves to nothing is
   a blocking validation error — the system never silently guesses a tax rate.
3. **Persist:** the effective rates and resolved tax codes of every emission
   are frozen in the invoice rows, so draft/SVC rebuilds and any DB consumer
   reproduce the exact emitted XML regardless of later farm-config changes.

**Root causes being fixed:**
- **Unsnapshotted effective rates:** `MergeRates(userRates, farmConfig)`
  (`service/nfe_service/service.go:946-959`) resolved user-typed → farm
  config → **code-borne fallback** (ICMS 0.17, PIS 0.0165, COFINS 0.076,
  IBS/CBS 2026), but persistence recorded only the **raw user-typed rates**,
  and only when the user typed at least one (`hasAnyUserRate` — call sites
  `service.go:307`, `worker.go:246`, `detached_service.go:300`). Emissions
  that typed nothing stored no rate row at all; partial rows stored NULL
  axes. When the system reads the NF-e back (worker draft/SVC rebuild,
  `worker.go:214-219`, or any future DB-based reader), unsnapshotted axes
  resolve against whatever the farm config — or the code fallback — says
  *that day*: a config edit or app upgrade silently re-rates a rebuilt XML.
- **Guessed rates:** the hardcoded fallback constants existed to make early
  emission testing convenient. Emission now works; guessing is a fiscal
  hazard (a zero-guessed or wrongly-guessed rate produces a fiscally wrong
  but schema-valid XML) and goes away.

**Design decisions (user-confirmed):**
1. Snapshot for every emission — not only when the user typed rates.
2. Hoist the detached merge into the single-source preparation step
   (`prepareDetachedBuildData`) instead of merging inside
   `buildDetachedItems`.
3. Flip the row semantics — `*_tax_rates` rows record the effective rates the
   XML used, not raw user input; `rate_source` provenance parked for later.
4. No nominal-value columns — amounts are derivable (snapshotted rate ×
   stored base, regime-aware) and the signed/authorized XML carries every
   value authoritatively.
5. Include the resolved tax codes — same unsnapshotted-attribute disease for
   CST/CSOSN resolution; both rows already have the columns (`icms_cst`,
   `pis_cst`, `cofins_cst`, `ibs_cst`, `cbs_cst`, `c_class_trib`).
6. Remove the hardcoded fallback values — resolution is user > farm only;
   anything unresolvable is a validation error, not a guess. The 2026
   IBS/CBS constants stay defined in `pkg/nfe/defaults/reform.go` as
   documented legal rates but leave the resolution chain.
7. Regime carve-out — ICMS rate required only for Regime Normal; Simples
   Nacional zeroes the ICMS amounts and the rate is unused there.

**Tree precondition:** `docs/sdd/nfe_dead_schema_cleanup/` is **Done**
(2026-10-09). The current tree already contains: migration
`000027_drop_dead_nfe_columns.sql` (applied), `UpdateInvoiceSent` /
`UpdateDetachedInvoiceSent` wired at the five send sites (`service.go:345`,
`:474`, `worker.go:272`, `detached_service.go:341`, `:1030`),
`CreateDetachedInvoice` with `dhCont` / `xJust` params, and the double-gated
DB test `model/nfe_model/invoice_lifecycle_model_test.go`. Line numbers refer
to that post-cleanup tree (2026-10-11); re-locate by symbol name.

## Requirements

### Functional

**Snapshot persistence (the original scope)**

- **R1** — `prepareInvoiceBuildData` (`service/nfe_service/service.go:46`)
  returns the merged `entity.TaxRates` **plus the resolved effective tax
  codes** as a single snapshot value (new local struct,
  `emissionTaxSnapshot{Rates entity.TaxRates; Codes entity.InvoiceOverrides}`),
  so callers persist exactly what `buildItems` consumed. `Codes` is a copy of
  the user overrides with the six tax fields resolved (R5).
- **R2** — Departure flow persists the full snapshot:
  `BuildInvoiceFromDeparture` and `attemptSVCContingency` persist
  `&snapshot.Rates` and `&snapshot.Codes`. The worker's draft rebuild
  persists the snapshot **returned by `prepareInvoiceBuildData`** (not its
  own pre-merge at `worker.go:219`) — identical for complete snapshots
  (idempotence); exact for legacy partial rows.
- **R3** — Detached flow hoists the resolution: `prepareDetachedBuildData`
  computes `MergeRates(input.TaxRates, farmNFeConfig)` and the resolved codes
  once, carries both on `detachedBuildData` (`RatesSnapshot`,
  `CodesSnapshot`); `buildDetachedItems` receives the pre-merged struct (its
  internal `MergeRates` call and inline CST/CSOSN resolution are removed);
  `BuildDetachedInvoice` and `attemptDetachedSVCContingency` persist
  `&build.RatesSnapshot` and `&build.CodesSnapshot`.
- **R4** — Every emission path persists the full snapshot: `*_tax_rates` rows
  from this point on have all five rate columns non-null (icms, pis, cofins,
  ibs, cbs), **and** the invoice row's tax-code columns are populated with
  the resolved codes (existing columns — no schema change). The
  skip-if-all-nil guard in the model insert helpers stays as a defensive
  legacy-shape guard.
- **R5** — New pure helper `MergeEffectiveOverrides(user
  *entity.InvoiceOverrides, farmConfig *entity_public.FarmConfig, regime
  defaults.TaxRegime) entity.InvoiceOverrides`: copies the user's override
  fields as-is and resolves the six tax fields with the exact precedence and
  default constants the builders apply today (override > farm config >
  builder default: ICMS CST `defaults.CSTTributadaIntegral` vs CSOSN
  `defaults.CSOSNSemPermissaoCredito` per regime; PIS/COFINS "01"; IBS/CBS
  `defaults.IBSCBSCSTTributadaIntegral`; `cClassTrib`
  `defaults.CClassTribDefault`). Both builders are refactored to **consume**
  the pre-resolved codes; XML output is byte-identical for identical inputs.
- **R6** — `hasAnyUserRate` (`service.go:984-987` and its three call sites)
  is deleted.
- **R7** — Backward compatibility: legacy rows with partial rate rows / empty
  code columns keep the existing per-field nil/empty fallback chains in the
  worker (`worker.go:212-219`) and `invoiceOverridesFromRecord`; no data
  migration rewrites historical rows.
- **R8** — Rebuild consumption stays read-only: `invoiceOverridesFromRecord`
  maps persisted (now complete) rows back into overrides; builder
  re-resolution becomes a no-op on snapshotted invoices (resolved, non-empty
  pointers win over farm defaults).

**Fallback removal (new scope)**

- **R9** — `MergeRates` loses its third source: per axis, resolution is
  **user override (non-nil pointer; a typed zero is explicitly exempt) →
  farm config (non-zero value) → unresolvable (nil pointer)**. `pickRate` and
  the hardcoded fallback arguments (`0.17`, `0.0165`, `0.076`,
  `defaults.IBSRate2026`, `defaults.CBSRate2026` as resolution defaults) are
  removed. The farm-config zero/NULL case remains "unconfigured" (existing
  semantics, `pickRate`'s `IsZero()` check).
- **R10** — New validation gate `validateResolvedRates(rates entity.TaxRates,
  regime defaults.TaxRegime) []string`, called in **both** prepare steps
  immediately after rate resolution and **before** item building / number
  allocation (emission *and* preview paths get it free — both call prepare).
  Missing (nil) axes that the regime actually uses produce a pt-br warning
  toast listing each missing rate and where to fix it (farm config or the
  emission form), e.g. "Alíquota ICMS não configurada: informe-a na emissão
  ou configure o valor padrão da fazenda".
- **R11** — Regime carve-out: ICMS is required only for
  `defaults.TaxRegimeNormal` — Simples Nacional zeroes `vBC/pICMS/vICMS` and
  the ICMS rate is never consumed by the XML. PIS, COFINS, IBS, CBS are
  required in both regimes.
- **R12** — Worker draft-rebuild with unresolvable rates: the prepare step's
  validation toast fails the rebuild before any XML is built; the worker
  logs it, increments the retry count (mirroring the SVC-unavailable branch
  so the log stays bounded), and leaves the invoice as draft. Pending
  status-polling needs no rates and is unaffected. Authorized invoices are
  untouched — their XML is authoritative.
- **R13** — `defaults.IBSRate2026` / `CBSRate2026` remain defined in
  `pkg/nfe/defaults/reform.go` with their test, as the documented legal 2026
  rates — but no longer participate in rate resolution. AGENTS.md's
  resolution description ("user > farm > 2026 fallback") is updated to
  "user > farm, missing = validation error".

### Non-Functional

- **N1** — Rebuild-safety invariants locked by pure unit tests:
  `MergeRates(fullSnapshot, cfg) == fullSnapshot`; partial snapshots surface
  exactly their missing axes through `validateResolvedRates` (not guesses);
  `MergeEffectiveOverrides` precedence per field.
- **N2** — Zero XML/DANFE output change for emissions that pass validation:
  same resolution results for the same inputs (code moved, not changed).
  **Intended user-visible change (documented, not a regression):** emissions
  that previously succeeded by silently guessing a fallback rate now fail
  validation with an explanatory toast until the rate is configured or typed.
- **N3** — All work stays SEFAZ-offline: `NFE_SEFAZ_GUARD=block` for any
  running app, no certificate data seeded, no direct SEFAZ calls (AGENTS.md
  "SEFAZ Network Isolation").
- **N4** — No change to forms, statuses, or endpoints; new failure mode only
  for genuinely unconfigured rates, in pt-br.

## Known residual (documented, accepted)

- **Regime switch after emission:** the resolved ICMS code is frozen (CST or
  CSOSN into `icms_cst`), but rebuilds re-derive the CST-vs-CSOSN branch from
  the farm config's *current* regime; a farm switching Simples↔Normal between
  emission and a draft-window rebuild could misread the frozen code. Narrow
  window (≤ 24 h draft), rare event; XML remains authoritative. Freezing the
  regime has no row column — out of scope.
- **Preview→emit staleness:** a farm-config change between the DANFE preview
  and the emission submit can make the preview stale; the emission snapshots
  what it actually used. Not addressed here.

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | Resolution invariants: complete snapshot → `MergeRates` returns it unchanged; partial snapshot → `validateResolvedRates` lists exactly the missing (nil) axes the regime uses; no fallback values ever substituted | `go test ./service/nfe_service/` — `rates_test.go`: `TestMergeRates_SnapshotIsIdempotent`, `TestMergeRates_MissingAxesSurfaceInValidation`, `TestValidateResolvedRates_*` |
| AC2 | Codes helper precedence: user code > farm default > builder default per field; ICMS field resolves CST vs CSOSN per regime; resolved values always non-empty; zero user rate stays exempt (not "missing") | `TestMergeEffectiveOverrides_*` set + `TestValidateResolvedRates_UserZeroIsExempt` |
| AC3 | Regime carve-out: SN regime with no ICMS rate validates clean; Regime Normal without ICMS rate is blocked; missing PIS/COFINS/IBS/CBS blocked in both regimes | `TestValidateResolvedRates_SimpleNacionalICMSOptional`, `TestValidateResolvedRates_*` unit tests |
| AC4 | Normal + SVC attempt + worker-rebuild paths persist the full snapshot (5 rates + 6 resolved code columns) for both flows | **DB-gated model tests** (`RUN_DB_TESTS=1` + `DB_*`, double-gated, one `-run` per invocation — `InitNFeModel` singleton) asserting read-back via public getters carries all five rate columns non-null and the six code columns populated; a no-override emission freezes the **farm-config values** (the fallback arm no longer exists); plus focused code inspection for SVC/worker sites |
| AC5 | `hasAnyUserRate` and `pickRate` fully removed; no hardcoded fallback rate remains in the resolution chain | `grep -rn "hasAnyUserRate\|pickRate" --include="*.go"` → 0 hits; `grep -n "0.17\|0.0165\|0.076" service/nfe_service/` → only test fixtures |
| AC6 | Existing suites green; `buildDetachedItems` (3 sites) and any `buildItems` test call sites updated for hoisted signatures; rate-dependent tests now supply explicit rates | `make test-go` (PASS), `make test-e2e` (PASS) |
| AC7 | Docs flipped: builder/persist/worker comments + AGENTS.md resolution description | Manual inspection grep checklist (plan Phase 6) |
| AC8 | Runtime scan parity retained for every SELECT feeding the invoice getters after the new persist behavior is exercised | The existing double-gated round-trip test extended with rates-row + code-columns assertions (or sibling file, same pattern) |
| AC9 | Worker draft-rebuild with unresolvable rates: no rebuild, log + retry increment, row stays draft; pending polling unaffected | Code inspection of the prepare-failure branch + `TestValidateResolvedRates` contract; justified manual check (worker path needs SEFAZ/SVC and a legacy no-rates row) |

> **Offline smoke reality (from the cleanup SDD):** UI-emission smoke is not
> executable — the e2e fixture seeds `nfe_farm_config` with **empty
> certificate bytes** (`test/e2e/tests/detached_cfop_selector.spec.js:17-19`),
> so signing fails before any persist, and AGENTS.md prohibits seeding
> certificate data. Model-level persistence is verified directly against the
> e2e DB via the DB-gated tests; SEFAZ-adjacent paths are accepted by
> inspection. A UI smoke is only possible out-of-band with a throwaway
> human-supplied certificate — not planned by default.

## Out of scope

- Nominal-value columns — derivable (snapshotted rate × stored base,
  regime-aware) and XML-authoritative (`ICMSTot.vICMS/vPIS/vCOFINS`,
  `IBSCBSTot.vCBS`, …)
- `rate_source` provenance column
- Freezing the tax regime per row (residual accepted above)
- Backfilling historical/legacy rows with computed snapshots
- NaturezaOp/NCM/CEST/unit/mod_frete override freezing (non-tax fields; same
  disease, separately parked)
- Config-form warnings for unconfigured rates (emission-time validation is
  the enforcement point)
- Any UI/report consuming the snapshots
- Preview→emit staleness

## Verification Results

### Baseline (Phase 0, post-cleanup tree, 2026-10-11)

- `go build ./...` — PASS
- `go test ./service/nfe_service/... ./model/nfe_model/` — PASS (no pre-existing failures)

### Final results (2026-10-11)

- `go build ./...` — PASS
- `make test-go` (`go test ./...`) — PASS, 0 FAIL lines
- `gofmt -l service/nfe_service/ model/nfe_model/` — clean
- `make test-e2e` — 104 passed / 15 failed / 4 skipped. All 15 failures are
  `[webkit]` under full parallel load; the baseline at HEAD (cleanup SDD,
  2026-10-09) was 105/14/4 with the same documented WebKit flake pattern.
  Isolated re-run of all 15 failed specs (`--project=webkit`, serial):
  **18 passed / 0 failed** — confirms pre-existing environment flake, not this
  change.
- `make test-js` — NOT RUN (no JavaScript changes; unit suites only; JS files
  untouched in this SDD).

**AC1 / AC2 / AC3 — resolution + validation invariants (PASS).**
`go test ./service/nfe_service/ -run 'TestMergeRates|TestValidateResolvedRates|TestMergeEffectiveOverrides'` — PASS:
- `TestMergeRates_SnapshotIsIdempotent` — complete snapshot re-merge returns it
  unchanged (user arm wins; farm arm wins with no user rates);
- `TestMergeRates_MissingAxesSurfaceInValidation` — unresolvable axes come back
  nil and validation lists exactly 5 errors for an unconfigured farm — no
  substituted values anywhere;
- `TestValidateResolvedRates_SimpleNacionalICMSOptional`,
  `TestValidateResolvedRates_UserZeroIsExempt` (typed zero stays exempt),
  `TestValidateResolvedRates_RequiredInBothRegimes` (regime carve-out);
- `TestMergeEffectiveOverrides_*` (precedence user > farm > builder default;
  regime-dependent ICMS code; typed-only IBSCST carried into the shared CST;
  user fields copied through; idempotent on a resolved struct).

**AC5 — dead helpers gone (PASS).**

```
grep -rn "hasAnyUserRate\|pickRate" --include="*.go" .   → 0 hits
grep -n "0\.17\|0\.0165\|0\.076" service/nfe_service/*.go (excluding _test.go)
    → only the resolveRate doc comment documenting the retired history
```

**AC2 — single-source structure (PASS).**

```
grep -n "MergeRates(" service/nfe_service/*.go (excluding tests)
    → exactly 3 call sites + def:
      service.go:184 (prepareInvoiceBuildData),
      detached_service.go:123 (prepareDetachedBuildData),
      worker.go:222 (row-rebuild pre-merge, then re-merged inside prepare —
      idempotent), service.go:968 (def)
grep "MergeEffectiveOverrides( | validateResolvedRates(" → exactly 2 prepare
      call sites each + defs (service.go:993/:1037)
```

**AC4 / AC8 — snapshot persistence round trips (PASS).** New
`model/nfe_model/invoice_tax_snapshot_model_test.go` (double-gated like the
cleanup's lifecycle tests; one `-run` per invocation):

```
RUN_DB_TESTS=1 DB_* NFE_SEFAZ_GUARD=block go test -v -count=1 \
  -run '^TestInvoice_TaxSnapshot_DBRoundTrip$' ./model/nfe_model/    → PASS
RUN_DB_TESTS=1 DB_* NFE_SEFAZ_GUARD=block go test -v -count=1 \
  -run '^TestDetachedInvoice_TaxSnapshot_DBRoundTrip$' ./model/nfe_model/ → PASS
```

- Departure flow: `CreateInvoiceWithEmission` persists the 5-column tax-rates
  row (all non-null, values = snapshot) and the 6 code columns
  (`icms_cst`…"c_class_trib"`), read back through `GetInvoiceByAccessKey`
  (scan parity over the edited SELECT list included).
- Detached flow: same assertions via `GetDetachedInvoiceByAccessKey`; plus the
  legacy shape (nil rates) persists **no** tax-rates row (skip-if-all-nil
  guard), with the NOT NULL code columns still requiring values — an explicit
  NULL bind bypasses the column DEFAULTs, so production paths must (and now
  always do) pass resolved codes.
- SVC attempt / worker pass-through sites verified by focused code inspection
  (SEFAZ/SVC prohibited offline).

**AC6 — suites (PASS).** `make test-go` (see above) incl. the 3 updated
`buildDetachedItems` test call sites (via `resolvedSnapshot`, which runs the
exact production preparation) and the extended legacy-variant in
`TestBuildDetachedItems_UserRatesAndOverrides`; `make test-e2e` as recorded
above.

**AC7 — docs flipped (PASS).** Builder/persist/worker comments now state the
snapshot semantics; `CreateInvoiceWithEmission` doc flipped; AGENTS.md Tax
Reform Gate line updated to "user > farm only, missing = validation error, no
fallback".

**AC9 — worker graceful skip (PASS via inspection; real-path NOT RUN).**
The prepare-failure branch in `processDraftInvoice` now increments
`retry_count` before returning (bounded backoff-visible logging), and the
validation contract (`TestValidateResolvedRates_*`) guarantees no XML is built
for unresolvable rates. Exercising the true rebuild path requires SEFAZ/SVC
plus a legacy no-rates row — prohibited offline; accepted as a justified
manual residual (same rationale as the cleanup SDD's SEFAZ-adjacent criteria).

### Notes

- Pre-existing `gofmt -l` violations exist in `pkg/nfe/defaults/agricultura*.go`,
  `pkg/nfe/defaults/reform.go`, `pkg/nfe/entity/invoice.go`,
  `pkg/nfe/service/code128c.go`, `pkg/nfe/xml/builder.go` — untouched by this
  SDD and left as-is (out of scope).
- All runs were SEFAZ-offline with `NFE_SEFAZ_GUARD=block`; the e2e DB
  (tmpfs, port 5433) migrations were applied at app boot and the DB containers
  were destroyed after verification (`bun run db:stop`); no
  `nfe_farm_config` certificate data was seeded and no app/air process was
  left running.
- Behavior change delivered (intended): farms relying on the retired silent
  fallbacks now hit a pt-br validation toast ("Configuração de taxas
  incompleta: …") at emission/preview time until the rate is configured or
  typed per emission; a typed zero remains a valid explicit exemption.
