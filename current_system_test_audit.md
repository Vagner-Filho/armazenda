# Current System Test Audit

**Date:** 2026-11-12
**Scope:** Full test battery (`make test` = `make test-go` + `make test-js` + `make test-e2e`), run against the working tree on top of HEAD `07d6611 fix: nfe-emit page bug fixes` with four fix batches applied (uncommitted):

1. **Typed-nil fix (2026-11-10)** — `router/nfe_router/detached_router.go`: `buildDetachedPageView` now holds `GetPeopleByFarm`'s error in a concretely-typed variable (no interface boxing), plus a DB-gated regression test (`detached_router_db_test.go`).
2. **Weight-mask amendment (2026-11-11)** — documented as Amendment 1 in `docs/sdd/decimal_comma_mask/`: `weight.js`/`decimalMask.js` rebuilt on a tracked-state machine (comma-only typing, paste-digesting dots, the mask never re-parses its own display), `departure_model.AddDeparture` no longer swallows DB errors, `addDeparture`/`putDeparture` refuse the swap on failure toasts, and per-input-method E2E coverage (`fill` paste-like vs `pressSequentially` keyboard).
3. **B1 — departure NF-e `indPag`/`includeCobranca` wiring** — `router/nfe_router/departure_input_router.go` + `templates/nfe/nfe-emit-modal.html` + `templates/nfe/nfe-preview.html` + `router/nfe_router/router.go`: wires the payment/billing inputs through modal, preview round-trip and parser (SDD `docs/sdd/departure_nfe_indpag_cobranca_wiring/`).
4. **B2 — certificate gate relaxed for non-fiscal preview** — `service/nfe_service/service.go` + `service/nfe_service/detached_service.go`: preview no longer requires `certificate_data`; emission still gated. E2E specs updated to use valid recipient documents and to expect the now-rendering preview.

The previous audit (2026-11-10) ran before any of these fixes. This audit reflects the state after them; test changes made as part of the fixes are listed above and in the SDD amendment.

## Summary

| Battery | Command | Result |
|---|---|---|
| Go (backend) | `make test-go` (`go test ./...`) | **PASS** — no failures |
| JavaScript unit | `make test-js` (`cd test && bun test unit/`) | **PASS** — 95 pass, 0 fail |
| E2E (Playwright) | `make test-e2e` | **FAIL** — 35 failed, 168 passed, 4 skipped (207 test×browser instances, ~11.1 min) |

Overall: **1 of 3 batteries failed** — the Go backend and JavaScript unit batteries are green. The E2E battery is down to 35 failed (168 passed, 4 skipped), all of them the WebKit environment cluster (Group C); B1 and B2 are resolved on chromium/firefox. E2E totals count each test once per browser.

## 1. Go backend tests

Command: `go test ./...`
Result: **PASS** — all packages with tests `ok`; no failures.

The B1 departure NF-e input-wiring failures (`TestParseDepartureFormOverrides_PaymentBillingAndTransport`, `TestDepartureNFeModalRendersInputsAndTaxSelects`, `TestDepartureNFePreviewCarriesInputsToConfirm`) are fixed and now pass. The typed-nil trap guard `TestBuildDetachedPageView_DB` (DB-gated) also passes.

## 2. JavaScript unit tests

Command: `cd test && bun test unit/`
Result: **PASS**
- 95 pass, 0 fail
- 303 `expect()` calls
- 7 test files (new: `weightMask.test.js` — the weight-mask paste/typing/blur-fold contract)
- Runtime: ~0.2 s

## 3. E2E tests (Playwright)

Command: `cd test/e2e && bun run test:e2e`
Result: **FAIL** — 35 failed, 168 passed, 4 skipped (207 test×browser instances across chromium/firefox/webkit). Duration ~11.1 min.

### Fixed since the previous audit

| Previous failure | Root cause (corrected) | Fix |
|---|---|---|
| `GET /nfe/emitir` → HTTP 500 "Falha ao carregar a página de NF-e" (~74 instances) | Typed-nil interface trap in `buildDetachedPageView` (commit `07d6611`) | Concretely-typed error variable + DB-gated regression test |
| `departure.spec.js` — SQLSTATE 22003 on `add_get_departure` + `Expected "ABC123", Received ""` | **Not a test-data magnitude issue** (as the previous audit guessed): the kg mask (`weight.js`) stripped dots from paste-like `fill` input events, turning `47148.55` into `4714855` (×100), which overflowed `departure.grossWeight NUMERIC(12,6)`; `AddDeparture` then swallowed the DB error and returned a zero-value row (the empty plate) | Mask rebuilt on the tracked-state machine; `AddDeparture` returns the error; the router refuses the swap on failure toasts |
| `detached_decimals.spec.js` comma-decimal signals | Downstream of the `/nfe/emitir` 500 (the specs died at `gotoEmitPage` before their assertions) | Fixed with the 500 |

The two new per-input-method tests (keyboard `pressSequentially` in `departure.spec.js` and `detached_decimals.spec.js`) pass on chromium and firefox and join the WebKit cluster below (WebKit's environment class — every form spec fails there).

### Fixed since the previous audit (this pass)

| Previous failure | Root cause | Fix |
|---|---|---|
| B1 — departure NF-e input wiring (3 Go tests + 1 E2E spec × 3 browsers) | `indPag`/`includeCobranca` unwired by commit `07d6611` | Wired through modal, preview round-trip, parser and router (SDD `departure_nfe_indpag_cobranca_wiring`) |
| B2 — certificate gate vs preview-200 (12 instances) | `nfeCertGate` refused preview when `certificate_data` empty | Relaxed the gate for non-fiscal preview only; emission still gated; E2E specs updated to valid recipient documents |

### Remaining failure group (35 instances)

**Group C — WebKit environment cluster (35 instances, webkit only)** — all 35 remaining failures are webkit-only (0 chromium/firefox failures remain). Same signature as before: `locator.click`/`page.fill` timeouts, `frame._expect: Target page, context or browser has been closed`, URL checks receiving `""`. Environment-level, not per-spec logic; needs an isolated WebKit repro (likely the service-worker/WASM/offline layer).

### Note on SEFAZ

The E2E runs used `NFE_SEFAZ_GUARD=block` (hardcoded by `global-setup.js`); no real SEFAZ/SVC traffic was attempted. Correction to the previous audit: the fixture (`test/e2e/fixtures/test-user.sql`) **does** seed an `nfe_farm_config` row for the test farm — with **empty `certificate_data`** (`''::bytea`), deliberately (SEFAZ isolation: no certificate bytes to sign with, and the guard blocks traffic regardless). That empty-certificate row is what triggers the Group B2 gate toasts.

## Conclusion

The JavaScript unit battery is green (95/95). The Go backend is green (no failures). The E2E suite is down to 35 failed (all WebKit environment — Group C); B1 (input wiring) and B2 (certificate gate vs preview) are resolved on chromium/firefox. The two product regressions found by the previous audit (`/nfe/emitir` 500 and the departure 22003/empty-plate) remain fixed and regression-guarded.
