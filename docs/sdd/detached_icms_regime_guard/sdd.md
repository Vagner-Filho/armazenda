# sdd:detached_icms_regime_guard

## Status
Backlog

## Goal
Make the detached NF-e flow's ICMS code validation **regime-aware**: today
the service branch that consumes the `icmsCST` emission field depends on the
farm's `tax_regime`, but the local validation does not — a Simples-Nacional
farm silently copies any regime-normal CST into `<CSOSN>` and ships
schema-invalid XML to SEFAZ, which rejects it with a 215 family error.
This SDD closes that gap locally for both emission flows, and completes the
detached form's UX so a Simples farm operator sees CSOSN options instead of
CSTs.

Incident that surfaced this (2026-10-14, first homolog replay): the
operation's farm config had `Regime Tributário = Simples Nacional`; the G1
payload posted `icmsCST=00`; the emitted XML carried
`<ICMSSN102><orig>0</orig><CSOSN>00</CSOSN>…` and SEFAZ returned
`cvc-type.3.1.3: The value '00' of element 'CSOSN' is not valid.`
(A valid CSOSN is 101/102/103/201/202/203/300/400/500/900. Homolog replay of
the G1–G7 corpus was separately unblocked by switching the farm config to
Lucro Real — the corpus emitter is a regime-normal producer, CRT 3.)

## Background

- `buildDetachedItems` (`service/nfe_service/detached_service.go`, ~line 724)
  branches on `defaults.TaxRegime(farmConfig.TaxRegime)`:
  - Simples Nacional → `icmsCSOSN = codeOr(codes.ICMSCST, defaults.CSOSNSemPermissaoCredito /* "102" */)`
  - otherwise → `icmsCST = codeOr(codes.ICMSCST, defaults.CSTTributadaIntegral /* "00" */)`
  The **same** `icmsCST` emission field feeds either slot — for a Simples
  farm every regime-normal CST value becomes an invalid `<CSOSN>`.
- The departure flow has the identical branch (`service/nfe_service/service.go`,
  ~line 828) feeding `buildItems` — same latent issue for a Simples farm.
- `validateDetachedCSTs` (`service/nfe_service/detached_wiring.go`) validates
  `codes.ICMSCST` against `validDetachedICMSCSTs = {00,20,40,41,50,51,90}`
  **regardless of regime** — `00` passes as a "valid CST" and is then
  reinterpreted as a CSOSN downstream. PIS/COFINS validation there is
  regime-independent and correct as-is.
- Builder (`pkg/nfe/xml/builder.go` `buildICMS`) is sound: CSOSN branch
  switches on the CSOSN set ({101, 201, 202/203, 500, 900 → fixed groups,
  default → ICMSSN102}); normal branch switches on CST. No builder change
  belongs to this SDD.
- The chain an invalid code follows today: emission form `icmsCST` →
  `MergeEffectiveOverrides` (override > farm default > regime default) →
  `validateDetachedCSTs` (regime-blind, passes) → `buildDetachedItems`
  copies the value into the CSOSN slot → builder emits `<CSOSN>00</CSOSN>`
  → SEFAZ 215.
- The default CSOSN when the field is left empty is `102`
  (`CSOSNSemPermissaoCredito`); the default CST for normal regimes is `00`.
- `CRT()` mapping (defaults.TaxRegime): Simples → `1`,
  Lucro Real/Presumido → `3` — switching the farm regime to Lucro Real
  reproduces the corpus emitter's stamp.

## Requirements

### Functional
1. **Regime-aware ICMS guard (shared helper)** — extend
   `validateDetachedCSTs` (or a successor, e.g.
   `validateTaxCodesForRegime(codes *entity.InvoiceOverrides, regime defaults.TaxRegime) []string`)
   so the ICMS branch depends on the resolved regime:
   - **Simples Nacional**: the value (override *or* farm default) must be a
     CSOSN in `{101, 102, 103, 201, 202, 203, 300, 400, 500, 900}`. A
     regime-normal CST (e.g. `00`, `20`, `51`) is rejected with a pt-br
     warning naming the valid set **and** pointing to the fix location:
     `Produtor no Simples Nacional: informe um CSOSN válido para ICMS (101, 102, 103, 201, 202, 203, 300, 400, 500 ou 900). Verifique o campo CST ICMS do formulário ou o padrão ICMS nas Configurações → NF-e.`
   - **Lucro Real / Lucro Presumido**: value must be in the existing CST
     whitelist `{00, 20, 40, 41, 50, 51, 90}` (unchanged); a CSOSN-style
     code (e.g. `102`, `101`) is rejected with a pt-br message pointing at
     the CST options.
   - Empty values stay allowed (the resolution defaults fill in the right
     slot per regime: `102` for Simples, `00` for normal) — same rule as
     today.
   - The helper is regime-aware from a `defaults.TaxRegime` parameter;
     `defaults.TaxRegimeSimplesNacional` is the branch discriminator, any
     other value follows the regime-normal path.
2. **Both flows gated** — call the regime-aware guard after
   `MergeEffectiveOverrides`:
   - detached: `prepareDetachedBuildData` (already the call site — passes
     `regime` now).
   - departure: `prepareInvoiceBuildData` gets the same check (small call
     addition) so a Simples farm cannot ship an invalid `<CSOSN>` from
     `nfe-emit-modal.html` either. The helper name stays
     flow-agnostic and lives in `detached_wiring.go`
     (rename acceptable — `tax_code_guard.go`-style file rename optional,
     documented in the plan).
3. **PIS/COFINS unchanged** — the existing whitelists
   (`{01, 02, 04…09}`; `03` excluded) apply to both regimes and keep their
   current pt-br messages.
4. **Detached form UX (CSOSN vs CST select)** — on the detached emission
   page, make the ICMS code select reflect the farm regime so an operator
   can no longer paste a CST into a CSOSN slot:
   - The page view gains a regime flag (e.g. `IsSimplesNacional bool`,
     derived from `farmNFeConfig.TaxRegime` in the router's view assembly).
   - While the flag is set, `nfe-emit.html` renders a **CSOSN select**
     (`icmsCST` form field keeps its name — the service consumes either
     flavor through the same field, avoiding parser/view churn): options
     `101 Sem permissão p/ crédito`, `102 Tributação SN sem permissão p/
     crédito`, `103 Isenção/faixa de receita`, `201 Sem permissão com
     cobrança`, `202 Com cobrança do ICMS`, `203 Isc. com cobrança`,
     `300 Imune`, `400 Não tributada`, `500 ICMS deferido/extemporâneo/
     expedição/ST`, `900 Outros` (selected prefill from
     `defaultICMSCST`), plus pt-br helper text noting the Simples-Nacional
     semantics.
   - When the farm is regime-normal, the existing CST select renders
     unchanged.
   - Labels pt-br; no parser/service changes for the field name.
5. **Rascunho parity** — `detached-profile` saved data carries the same
   `icmsCST` override; no model or profile parser changes (the field name
   is kept), so profiles saved under a previous regime simply get
   validated against the *current* regime at emission time and surface the
   pt-br message above when incompatible (documented behavior, no silent
   migration of stored values).

### Non-Functional
- All new server messages in pt-br; no builder changes (goldens and
  `realworld_test.go` must stay byte-identical — a diff audit is part of
  wrap-up).
- Offline unit tests only; no DB migrations; SEFAZ guard untouched.
- `gofmt` clean (no new flags beyond the documented pre-existing ones).
- The guard must never block a value the *builder* can validly emit for the
  given regime (whitelists ≡ builder switches: ICMS normal = CST branches;
  Simples = CSOSN groups; default-branch values stay whitelisted-in as
  documented).

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | Simples regime: CST-style ICMS codes (`00/20/40/41/50/51/90`, plus non-whitelisted samples like `10`) rejected with the pt-br message naming the CSOSN set and the config-fix pointer; valid CSOSN codes (all 10) and empty pass. | Guard table test (`go test -run 'TestValidateTaxCodes|TestValidateDetachedCSTs' ./service/nfe_service/`) |
| AC2 | Normal regime (Lucro Real/Presumido): CSOSN-style codes (`101…900`) rejected (pt-br); existing CST whitelist stays green (`00/20/40/41/50/51/90` pass). | Same table test, second regime row set. |
| AC3 | PIS/COFINS behavior unchanged under both regimes (whitelist `{01,02,04–09}`; `03` rejected; empty allowed). | Guard table test rows + existing tests updated for the new signature stay green. |
| AC4 | Both emission paths invoke the regime-aware guard after code resolution (detached + departure call sites present). | Structural review + focused run of the prepare-level tests that exercise the guard (`go test -run 'TestValidate...' ./service/nfe_service/`); the pre-existing `rates_test.go` merge pipeline stays green. |
| AC5 | Detached page UI: with a Simples farm config the form shows the CSOSN select (posted under `icmsCST`); with a normal regime the CST select — view flag computed in the router. | View-flag unit test where the page-view helper allows (`go test ./router/nfe_router/` — the detached page template tests pattern, e.g. `TestNFeEmitPageSelectsSavedRecipient` family) + structural template review; template-compile covered by the existing router tests. |
| AC6 | Real-world goldens + corpus inventory unchanged (builder untouched). | `go test -run '^TestRealWorld' ./pkg/nfe/xml/` (+ `RW_DEBUG=1` diff audit, same allowed set). |
| AC7 | No regressions repo-wide; gofmt adds no flags; pt-br wording review done. | `make test-go`; `gofmt -l <changed files>`. |

Manual checks: none beyond the template render tests above (all offline).
The operator-side fix for the live incident — switching the farm config to
Lucro Real — is already done/verified outside this SDD.

## Verification Results

| Command | Result | Notes |
|---------|--------|-------|
| `go test ./pkg/nfe/... ./service/nfe_service/ ./router/nfe_router/` (baseline, pre-implementation) | PASS | Green as of `resptec_system_defaults` wrap-up (all packages ok) — re-run at WIP flip. |
| *(remaining rows to be filled as each command is actually run)* | | |
