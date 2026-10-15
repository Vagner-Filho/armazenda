# plan:detached_icms_regime_guard

Implementation plan for making the ICMS code validation regime-aware and
surfacing CSOSN options for Simples-Nacional farms on the detached emission
page. Spec: `docs/sdd/detached_icms_regime_guard/sdd.md`.

## Test strategy (defined before implementation)

- **Layer**: pure unit tests in `service/nfe_service` (guard table over
  regime × ICMS/PIS/COFINS codes) + router page-view/template tests for the
  select swap; builder untouched (goldens as regression net).
- **Focused commands during development**: `go test -run <pattern>
  ./service/nfe_service/` per phase; `go test ./router/nfe_router/` for the
  view/template phase.
- **Before completion**: `make test-go` (AC7) + `gofmt -l <changed files>`
  + the goldens diff audit (AC6).
- **Behaviors to cover** (AC1–AC5):
  - Simples regime × ICMS codes: `{101,102,103,201,202,203,300,400,500,900}`
    pass; `{00,20,40,41,50,51,90,10,70}` reject with the pt-br message.
  - Lucro Real/Presumido × ICMS codes: `{00,20,40,41,50,51,90}` pass;
    `{101,102,903,900…}` reject (pt-br); `03`-style non-CST rejects.
  - PIS/COFINS rows under both regimes (whitelist changeless).
  - Empty ICMS/PIS/COFINS → allowed under both regimes.
  - Call-site presence (structural) for `prepareDetachedBuildData` and
    `prepareInvoiceBuildData`.
  - View flag: page view for a Simples farm config sets
    `IsSimplesNacional` (template swap covered by structural review +
    the template-compile tests that already run in `router/nfe_router`).
- **Environment**: none; offline; guard untouched; no fixture changes.

## Phase 1 — regime-aware guard (AC1–AC4)
1. `service/nfe_service/detached_wiring.go`:
   - Extend/rename `validateDetachedCSTs` →
     `validateTaxCodesForRegime(codes *entity.InvoiceOverrides, regime defaults.TaxRegime) []string`
     (keep the old name as a thin wrapper only while tests/usages migrate —
     remove it once call sites are updated; single rename preferred).
   - ICMS branch: Simples → CSOSN whitelist
     `validDetachedCSOSN = {101,102,201 x…}` (full set incl. 300/400); else
     keep `validDetachedICMSCSTs`. CSOSN value keys for pt-br message
     `Produtor no Simples Nacional: informe um CSOSN válido…`.
   - PIS/COFINS branches unchanged (regime-independent).
2. Call sites:
   - `prepareDetachedBuildData` (existing spot, passes `regime`).
   - `prepareInvoiceBuildData` (`service/nfe_service/service.go`, after its
     `MergeEffectiveOverrides`) — same guard, same pt-br toast shape as its
     other validation warnings (`Dados do emitente/destinatário…` style,
     e.g. toast title `Configuração de tributação inválida`).
3. Tests (`detached_wiring_test.go`): regime × code table (per Test
   strategy) replacing/adjusting the current `TestValidateDetachedCSTs`
   rows; naming per test `TestValidateTaxCodesForRegime_<case>`.

## Phase 2 — detached page select swap (AC5)
1. Router: the detached page view assembly
   (`detached_router.go` `getDetachedNFePage`/build view) computes
   `IsSimplesNacional = nfe_service.TaxRegimeIsSimples(farmNFeConfig.TaxRegime)`
   — add a tiny exported helper in `defaults` or reuse direct comparison
   (`defaults.TaxRegime(farmNFeConfig.TaxRegime) ==
   defaults.TaxRegimeSimplesNacional`).
2. `templates/pages/nfe-emit.html`: conditional ICMS block —
   Simples → CSOSN select (10 options, prefilled from `DefaultICMSCST`,
   helper text pt-br); else the current CST select. The form field name
   stays `icmsCST` (no parser churn). Preview fragment hidden input
   unchanged.
   - The page template compile is exercised by `router/nfe_router` tests;
     keep the markup inside the existing `{{ if .HasNFEFarmConfig }}` block.
3. Test: view-flag via the existing page-view unit seam (if the view data
   is exposed as a struct — extend `detachedPageView` with the bool;
   assert in a small test) — else structural (record).

## Phase 3 — wrap-up (AC6–AC7)
1. `go test -run '^TestRealWorld' ./pkg/nfe/xml/` + `RW_DEBUG=1` audit.
2. `make test-go`; `gofmt -l <changed files>` — expect only the pre-existing
   `pkg/nfe/xml/builder.go` item.
3. Fill SDD Verification Results; Status → Done when ACs met.

## Risks / notes
- **Locked-in farm defaults**: a Simples farm whose config default is a
  CST (e.g. `00` migrated from a previous regime) will be REJECTED at
  every emission until the operator fixes `CST ICMS padrão` in
  Configurações → NF-e — that is the desired loud behavior; the message
  points at exactly that field.
- **Saved rascunho/profiles**: profiles store the `icmsCST` override;
  incompatible stored values surface the guard message at the next
  emission (documented, no silent migration).
- **Single field, two flavors**: `icmsCST` keeps carrying the CSOSN under
  Simples — parser, rascunho and preview round-trip stay stable; only the
  UI options change per regime.
- **Departure modal untouched UI-wise**: `nfe-emit-modal.html` keeps the
  CST input for now (regime guard still protects it server-side); a UI
  parity pass there is a possible later tweak, not this SDD.
- **Builder untouched**: no `buildICMS` change — the goldens registry
  (`realworld_test.go`) expects zero diff-set movement; assert at wrap-up.

## Out of scope
- CBSCST/IBSCST/cClassTrib validation (separate hardening if ever needed —
  currently free-form per NT and not corpus-validated).
- Auto-imposing the farm regime on the emission form (operator could still
  post raw values; the guard makes it loud but the JSON is not locked).
- `idCSRT`/`CSRT` support (`resptec_system_defaults` follow-up).
- Renaming the form field / snapshot columns to `icmsCode` (churn vs
  benefit; renaming the helper function is fine).
