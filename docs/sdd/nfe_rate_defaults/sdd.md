# sdd:nfe_rate_defaults

## Status
Done

## Goal
Replace the "rates required" gate with corpus-derived automatic fallbacks, so
an emission without any typed or saved rate never blocks:

Today `MergeRates` (user > farm) leaves a rate `nil` when neither the emission
form nor a non-zero farm default provides it, and `validateResolvedRates`
blocks the emission with "Configuração de taxas incompleta". The user must
configure rates before the first emission — the session's original complaint
("the system prevents emitting when the config is not fully set").

The 115 authorized production NF-e in `docs/research_nfe_use_case/USE_CASES.md`
(the real emitter used by the farm) prove the production norms:

- **IBS 0.1 % and CBS 0.9 %** (`pIBSUF 0.1000`, `pCBS 0.9000`) appear in 115/115
  files — the regulator-fixed 2026 symbolic rates (Ato Conjunto RFB/CGIBS),
  already encoded as `defaults.IBSRate2026` / `CBSRate2026`.
- **ICMS `pICMS 0.0000`** — a zero rate, authorized 107× — with the tax
  treatment expressed via the CST choice plus `infCpl` text, not computed
  values.
- **PIS/COFINS rates are never declared** (CST 08 `PISNT`/`COFINSNT` single
  child).

So the unresolvable-rate rule was never physically necessary for the real
flows. The rates waterfall becomes **user > farm config** and falls back to
zero (ICMS/PIS/COFINS, corpus-proven) or the official 2026 constants
(IBS/CBS, corpus-proven).

Deliberately out of this SDD (user decision, session review): tax-code CST
defaults, natureza op texts, serie hints, CFOP→CST preselect — CST inputs are
untouched.

## Requirements

### Functional

- **FR1 — Resolution waterfall (`service/nfe_service/service.go`):**
  `resolveRate` gains a final fallback per axis and never returns `nil`:
  the user-typed rate wins (> explicit zero = exemption, unchanged);
  else the farm-config rate when non-zero (unchanged);
  else the axis fallback —
  - IBS → `defaults.IBSRate2026` (`0.001`)
  - CBS → `defaults.CBSRate2026` (`0.009`)
  - ICMS, PIS, COFINS → `decimal.Zero`

  `MergeRates`'s exported signature stays `(overrides entity.TaxRates,
  farmConfig *entity_public.FarmConfig) entity.TaxRates` — the worker and
  detached callers do not change.

- **FR2 — Rate gate deleted:** `validateResolvedRates` and its two callers
  go away — the "Configuração de taxas incompleta" block in
  `prepareInvoiceBuildData` (service.go) and in `prepareDetachedBuildData`
  (detached_service.go). `buildItems` keeps its pre-resolved-rates contract;
  its comments change from "pre-validated" to "pre-resolved, from the
  waterfall" and the `resolveRate` doc comment ("intentionally no hardcoded
  fallback") is rewritten to the new policy with the corpus rationale.

- **FR3 — Worker semantics drift (comment-only):
  `service/nfe_service/worker.go::processDraftInvoice`** — legacy incomplete
  snapshot rows (no `nfe_invoice_tax_rates` row / partially-set rows) used to
  be refused by the prepare-step's rate validation; they now merge with the
  farm config / fallbacks and rebuild. Update the comment block; the
  `MergeRates` call itself is untouched.

- **FR4 — Persistence contract:`model/nfe_model/model.go
  ::insertInvoiceTaxRates`** — with all five rates resolved on every emission,
  the per-axis "user provided / not provided" distinction of
  `nfe_invoice_tax_rates` is retired: every emission persists the effective
  snapshot (including resolved defaults/zeros). Update the function's doc
  comment accordingly (the `all nil → skip` guard becomes unreachable in the
  emission path and stays as a defensive no-op).

- **FR5 — tPag default (bundled companion phase, user-approved):** the
  departure emission default payment changes from `90`/"sem pagamento" to
  `99`/"Outros" declaring `vPag` = total, mirroring the corpus G1 shape
  (84 files: `tPag 99` + `xPag "Outros"` + `vPag = total`; `tPag 90` appears
  only in the 12 docs with no payment). Touches
  `service/nfe_service/departure_inputs.go::departurePayment` (default
  `"90"` → `"99"`, set `XPag = "Outros"` when the user did not choose one)
  and the modal default in `router/nfe_router/router.go::getNFeEmitModal`
  (`"DefaultTPag": "90"` → `"99"`). Choices 01 / 90 / 99 stay available and
  validated exactly as today (`validateDeparturePayment` unchanged).

### Non-Functional

- Snapshot idempotency preserved: `MergeRates` on an already-resolved snapshot
  returns it unchanged (worker rebuilds / SVC contingency reproducibility).
- No schema migration, no template layout change, no `pkg/nfe/xml` change —
  `defaultPayment("99", total)` already renders `vPag = total` (see
  `default_payment_test.go`) and `XPag` emission is supported
  (registry key `xPag`, retired as gap).
- Tests stay offline — no SEFAZ contact (AGENTS.md rules). DB-gated tests use
  only the local e2e test database.
- Existing pt-br toasts for tariff/other validation remain;
  "BTX de taxas incompleta" (rates) disappears entirely.

## Acceptance Criteria

| # | Acceptance criterion | Verification |
|---|----------------------|--------------|
| AC1 | Unconfigured farm + empty user rates resolve to all non-nil: ICMS 0, PIS 0, COFINS 0, IBS `0.001`, CBS `0.009` | unit `TestMergeRates_UnconfiguredFarmResolvesDefaults` (`service/nfe_service/rates_test.go` rewrite) |
| AC2 | Precedence holds: user-typed zero beats farm and constants; non-zero farm config beats the constants; user value beats farm | same file (adapted existing tests) |
| AC3 | Snapshot idempotency: `MergeRates` on a resolved snapshot returns it unchanged | existing `TestMergeRates_SnapshotIsIdempotent` stays green |
| AC4 | No code path returns the incomplete-rates block; the rates-related toast is gone from both prepare paths | compile-level + `grep -r "Configuração de taxas incompleta"` returns only historical docs; `validateResolvedRates` symbol absent (`go vet ./...` clean) |
| AC5 | Emission of a farm with defaultless rates persists the effective snapshot: `nfe_invoice_tax_rates` row with ICMS/PIS/COFINS 0.0000, IBS 0.0010, CBS 0.0090 | DB-gated `TestInvoiceTaxRates_DefaultResolvedRatesPersist_DBRoundTrip` (`model/nfe_model/invoice_tax_snapshot_model_test.go` extension, `RUN_DB_TESTS=1`) |
| AC6 | tPag default: a departure emission without explicit payment produces `tPag 99`, `xPag "Outros"`, `vPag = total`; explicit 01/90/99 overrides still work (90 keeps vPag 0.00) | unit extension in `service/nfe_service/default_payment_test.go` / `departure_inputs_test.go` (`TestDeparturePayment_DefaultIs99Outros`) |
| AC7 | Documentation: AGENTS.md rate-policy paragraph rewritten (waterfall user > farm > corpus fallback; no gate); `docs/sdd/nfe_rate_snapshot/sdd.md` gains a superseded-notice on its R10 gate | doc review (mechanical grep for the old sentence) |

## Verification Results
- `go build ./...` — PASS
- `go test ./service/nfe_service/` — PASS
- `go test ./model/nfe_model/ -run 'TestInvoice_TaxSnapshot_DBRoundTrip|TestDetachedInvoice_TaxSnapshot_DBRoundTrip'` with RUN_DB_TESTS=1 + DB_* (local e2e Postgres) — PASS
- `go test ./...` — PARTIAL: pre-existing failures in router/nfe_router (TestParseDepartureFormOverrides_PaymentBillingAndTransport, TestDepartureNFeModalRendersInputsAndTaxSelects, TestDepartureNFePreviewCarriesInputsToConfirm) unrelated to this SDD; nfe_service and model packages green.
- `grep -rn "validateResolvedRates" --include="*.go"` — no matches (symbol absent)
- `grep -rn "Configuração de taxas incompleta" --include="*.go"` — no matches
