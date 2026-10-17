# plan:nfe_rate_defaults — Implementation Plan

Implements the spec in `docs/sdd/nfe_rate_defaults/sdd.md`.
Tests are written alongside (and, for the resolution core, before) the code
that changes the behavior.

## Test strategy (before / during implementation)

Behaviors to cover and where:

| Behavior | Test location / layer | Needs |
|----------|----------------------|-------|
| Default waterfall for an unconfigured farm (AC1) | `service/nfe_service/rates_test.go` — same-package unit (`package nfe_service`), mirrors current fixture helpers (`snapshotFarmConfig`, `unconfiguredFarmConfig`) | none |
| Precedence matrix (AC2) + idempotency (AC3) | same file, rewriting `TestMergeRates_MissingAxesSurfaceInValidation` / `TestValidateResolvedRates_*` into positive-resolution tests | none |
| Gate removal (AC4) | compile + grep (no runtime test possible for an absence) | none |
| Effective-snapshot persistence with defaults (AC5) | `model/nfe_model/invoice_tax_snapshot_model_test.go` — DB round-trip, gated `RUN_DB_TESTS=1` (reuse `dbRoundTripPool`) | local e2e Postgres |
| tPag default + overrides (AC6) | `service/nfe_service/default_payment_test.go` + `departure_inputs_test.go` — pure unit | none |

Environment needs:

- Offline only (SEFAZ guard rules in AGENTS.md — no test may dial SEFAZ).
- AC5 needs the test DB started/seeded as in `nfe_farm_config_nullability`:
  `cd test/e2e && bun run db:start && bun run db:seed`, then run gated tests
  with `DB_*` + `RUN_DB_TESTS=1`.

Focused commands:

```bash
go test -run 'TestMergeRates|TestDeparturePayment' -v ./service/nfe_service/
go test -run 'TestInvoiceTaxRates' -v ./model/nfe_model/     # with RUN_DB_TESTS=1 + DB_* env
make test-go   # completion check
```

## Phase 1 — Resolution core (rates waterfall)

1. Rewrite the expectations first (regression-first):
   `service/nfe_service/rates_test.go`:
   - Add `TestMergeRates_UnconfiguredFarmResolvesDefaults` (AC1) — currently
     fails ("expected 5 missing-rate errors") until `resolveRate` changes.
   - Rewrite `TestValidateResolvedRates_*` and the
     missing-axes test into AC2 precedence tests (user zero wins / farm
     non-zero wins / user beats farm).
   - `resolvedSnapshot` helper drops its `validateResolvedRates` fatal →
     plain merge (fixtures keep passing).
2. Implement in `service/nfe_service/service.go`:
   - `resolveRate(override *decimal.Decimal, farm decimal.Decimal, fallback decimal.Decimal) *decimal.Decimal`
     (unexported; free to change signature) — `override` if non-nil, else
     non-zero `farm`, else `fallback`.
   - `MergeRates` supplies the axis fallbacks: `decimal.Zero` for
     ICMS/PIS/COFINS, `defaults.IBSRate2026` / `defaults.CBSRate2026` for
     IBS/CBS (FR1).
   - Delete `validateResolvedRates` + its `prepareInvoiceBuildData` block
     (FR2); refresh comments on `resolveRate`, `MergeRates`, `buildItems`
     (the "intentionally no hardcoded fallback" rationale becomes the corpus
     rationale — cite `docs/research_nfe_use_case/USE_CASES.md`).
3. `service/nfe_service/detached_service.go`: remove the
   `validateResolvedRates` block in `prepareDetachedBuildData` (the merge
   + codes snapshot stay).
4. `service/nfe_service/worker.go`: comment-only update in
   `processDraftInvoice` (legacy rows merge to defaults instead of being
   refused) (FR3).
5. Check for other impacted tests
   (`grep -rn "validateResolvedRates\|Configuração de taxas incompleta"`
   → only historical docs after this phase).

Run: `go test ./service/nfe_service/` — AC1–AC4 green; `make test-js` not
applicable.

## Phase 2 — Persistence + worker implications

1. `model/nfe_model/model.go::insertInvoiceTaxRates` — doc comment only (FR4);
   the guarded skip stays as a defensive no-op.
2. Gated test extension (AC5) in
   `model/nfe_model/invoice_tax_snapshot_model_test.go`:
   `TestInvoiceTaxRates_DefaultResolvedRatesPersist_DBRoundTrip` —
   `CreateInvoiceWithEmission` with nil-rate arms / a `FarmConfig` with zero
   rates → row exists with `icms_rate=0.0000`, `pis_rate=0.0000`,
   `cofins_rate=0.0000`, `cbs_rate=0.0090`, `ibs_rate=0.0010`.
3. Run gated suite (test DB per environment needs above).

## Phase 3 — tPag default (bundled companion, FR5)

1. `service/nfe_service/departure_inputs.go::departurePayment`: default
   `tpag := "99"`; set `XPag = "Outros"` on the detail when the caller did
   not provide one (the 90 special case must still yield `vPag = 0.00` and
   no `xPag`).
2. Unit tests (AC6):
   - `TestDeparturePayment_DefaultIs99Outros` in `departure_inputs_test.go`
     (tPag 99, xPag "Outros", vPag = total);
   - explicit-driver cases in `default_payment_test.go` (01 with money,
     90 with zero, 99 with total) — update the comment that today calls 90
     "the departure-flow shape".
3. `router/nfe_router/router.go::getNFeEmitModal`: `"DefaultTPag": "90"` →
   `"99"` (mechanical template-data change; covered by the unit contract).
4. Run `go test ./service/nfe_service/`.

## Phase 4 — Docs + completion

1. AGENTS.md: rewrite the rate-policy sentence in the tax-reform section
   (waterfall user > farm > official 2026 constants / zero; no gate;
   corpus citation) (AC7).
2. `docs/sdd/nfe_rate_snapshot/sdd.md`: add a short superseded note on the
   R10 gate pointing here (AC7).
3. Completion checks: `make test-go`; gated suite green; record exact
   commands and PASS/FAIL/NOT RUN in `docs/sdd/nfe_rate_defaults/sdd.md`
   → Verification Results; set `Status: Done` only when all ACs are met.

## Risks / out of scope

- **Scope guard (user decision):** CST defaults (ICMS `'00'`, PIS/COFINS
  `'01'`), natOp texts, serie hints, CFOP→CST preselect are explicitly out —
  "CST inputs shouldn't matter … It is only about the rates."
- **Corpus footnote (documented, not acted on):** the real G1 flow pairs the
  zero `pICMS` with **CST 51 (diferido)** + `infCpl` text — a zero-config
  emission that keeps the builder's CST 00 default is schema-valid and
  totals-coherent (the homolog-proven 531 rule is the hard gate) but is
  semantically "tributada integralmente at 0 %". Farms that want the MT grain
  deferral set "CST ICMS Padrão = 51" in the config form; the treatment stays
  expressible via CST + `infCpl`, never via computed values.
- **Fiscal semantics:** defaulting to zero ICMS/PIS/COFINS is the farm's
  fiscal choice to opt into by leaving the config empty; typing rates per
  emission and saving farm defaults remain the supported paths for real
  rates.
- **IBS/CBS fallback constants** are regulator policy, not guesses (Ato
  Conjunto RFB/CGIBS; present in 115/115 corpus files); if SEFAZ later changes
  the symbolic phase values, `defaults.IBSRate2026` / `CBSRate2026` are the
  single place to update.
- **tPag risk:** a paying document generated by another flow that reused
  `defaultPayment(tpag, total)` with the old 90 default changes shape
  (`vPag` becomes the total). Detached NF-e payment flows are separate
  (`DetachedPaymentInput`) and are not touched.
