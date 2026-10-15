# plan — nfe_cst_zero_values

## Problem summary

`buildDetachedItems` (and the departure-flow twin in `service.go`) fill item tax entity values unconditionally from farm default rates. For CSTs whose XML groups declare no `vBC`/rate/value (ICMS **40/41/50**; PIS/COFINS **04–09**), the XML layer correctly omits the values (`buildPIS`/`buildCOFINS`/ICMS 40-family branches), but `buildTotal` sums the entity values — so `ICMSTot` carries `vBC/vICMS/vPIS/vCOFINS` that no item declared. SEFAZ rejects with **531** (and would fire the PIS/COFINS totals rules next: the G4 XML already had `vPIS 6000.00 / vCOFINS 42000.00` with `PISNT/COFINSNT 08` items).

Fix at the **mapping layer**: zero the entity values exactly when the emitted group declares nothing. Totals logic stays untouched.

## Reference files

- `service/nfe_service/detached_service.go` — `buildDetachedItems`: ICMS branch (~724–734), PIS/COFINS (~787–798), entity wiring (777–809)
- `service/nfe_service/service.go` — departure-flow item imposto mapping (~900–930, lines 909/915/921/928)
- `pkg/nfe/xml/builder.go` — emit rules per CST: ICMS40/41/50 groups declare nothing (408–411); PISNT/COFINSNT 04–09 declare only CST (505–518, 524–537); `buildTotal` aggregates entity values (603–642) — **no changes here**
- `pkg/nfe/defaults/` — new predicates (see Phase 1)
- `service/nfe_service/test/mocks.go` — hand-written mock pattern for service tests

## Test strategy (before implementation)

Behaviors to cover (all offline; mocks only; no SEFAZ):

| # | Behavior | Layer / location | Fixture needs |
|---|----------|------------------|---------------|
| T1 | Suspended item (CST 50) → zeroed ICMS entity values | detached service test (table-driven 40/41/50) | farm config fixture with Lucro Real + ICMS/PIS/COFINS rates; one item input |
| T2 | NT items (CST 08; parameterize 04) → zeroed PIS/COFINS entity values | same test | same fixture |
| T3 | CST 00/90 unchanged (non-zero, as today) | same test (parameterized) | same fixture |
| T4 | Built XML with zeroed-entity item → `ICMSTot vBC/vICMS/vPIS/vCOFINS = 0.00`, `vProd/vNF` intact | builder test (`pkg/nfe/xml/builder_test.go`) | hand-built `entity.InvoiceInput` with one zeroed CST-50/08 item and non-zero vProd — pins the totals contract |
| T5 | Departure flow: same zeroing rule | `service/nfe_service` test (departure item mapping) | existing departure-service mocks/fixtures |
| T6 | Full suite stays green | `make test-go` | — |

Commands: focused `go test -run … ./pkg/nfe/xml/ ./service/nfe_service/` during dev; `make test-go` before completion.

Baseline first: run `go test ./pkg/nfe/xml/ ./service/nfe_service/` and note pre-existing failures before touching code.

## Phases

### Phase 0 — Baseline
1. Run focused suites; record any pre-existing failures in this file.

### Phase 1 — Predicates (single source of truth)
1. Add to `pkg/nfe/defaults` (near the CST sets, e.g. `agriculture.go`):
   - `ICMSDeclaresValues(cst string) bool` → true exactly for {"00","20","51","90"}.
   - `PISDeclaresValues(cst string) bool` / `COFINSDeclaresValues(cst string) bool` → true exactly for {"01","02"} (the `PISAliq`-routed whitelist; 04–09 are NT).
2. Unit tests for the predicates (table-driven).

### Phase 2 — Detached mapper (T1–T3)
1. In `buildDetachedItems`:
   - Lucro Real branch (~729–734): gate `vBC/vICMS/picms` fill on `defaults.ICMSDeclaresValues(icmsCST)`; otherwise zero. (CSOSN/Simples branch already zeroes — leave.)
   - PIS/COFINS (~787–798): when NOT `PISDeclaresValues` (resp. COFINS), set `VBC=PPIS=VPIS=0` / `VBC=PCOFINS=VCOFINS=0`.
2. Regression tests T1–T3 (implement together with the change; run focused suite).

### Phase 3 — Departure mapper (T5)
1. Apply the same gating in `service.go` item mapping (~900–930).
2. Regression test for the departure path.

### Phase 4 — Totals contract test (T4) + full suite (T6)
1. Builder test: zeroed CST-50/CST-08 item input → assert `ICMSTot` zeros with intact `vProd/vNF` (documents the invariant "totals = Σ declared").
2. Run `make test-go`.
3. `gofmt` on touched files (builder.go untouched).

### Phase 5 — Documentation
1. Update `sdd.md` Verification Results (commands + PASS/FAIL).
2. Note the homolog re-run (AC8) as pending-manual; hand the user the unchanged G4 payload.

## Risks / out of scope
- **CST 51 (deferimento)** unchanged: its group declares vBC/pICMS and is self-consistent today (G1 authorized). No deferral-specific totals semantics are attempted.
- **PISOutr family (03, 99)** not whitelisted/supported — untouched.
- **vICMSDeson / benefit descriptions** for CST 40 — out of scope.
- **IBS/CBS** values already item-derived and consistent — untouched.
- Existing homolog DB rows (id 26–29, denied) are historical artifacts; no data migration or backfill.
- SEFAZ homolog re-runs are human-driven (`NFE_SEFAZ_GUARD=off`) and NOT part of the automated verification; record outcomes in `sdd.md` when performed.
- No rollback risk beyond a revert: no schema change, no XML shape change.
