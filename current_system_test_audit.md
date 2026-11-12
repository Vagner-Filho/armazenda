# Current System Test Audit

**Date:** 2026-11-10
**Scope:** Full test battery run (`make test` = `make test-go` + `make test-js` + `make test-e2e`).
**Policy followed:** No tests were modified or fixed; batteries were run once against the current working tree (HEAD `07d6611 fix: nfe-emit page bug fixes`).

## Summary

| Battery | Command | Result |
|---|---|---|
| Go (backend) | `make test-go` (`go test ./...`) | **FAIL** — 1 package, 3 tests |
| JavaScript unit | `make test-js` (`cd test && bun test unit/`) | **PASS** — 76 pass, 0 fail |
| E2E (Playwright) | `make test-e2e` | **FAIL** — 115 failed, 82 passed, 4 skipped (201 total) |

Overall: **2 of 3 batteries failed.**

## 1. Go backend tests

Command: `go test ./...`
Result: 14 packages `ok`, 58 packages with no test files (`?`), 1 package `FAIL`.

Failing package: `armazenda/router/nfe_router` (0.166s).

Failing tests:
1. `TestParseDepartureFormOverrides_PaymentBillingAndTransport` — `departure_input_router_test.go:41`: payment options = tPag "01" / indPag 0; `:44`: `IncludeCobranca = <nil>, want explicit false`.
2. `TestDepartureNFeModalRendersInputsAndTaxSelects` — `departure_input_template_test.go:78`: departure modal must render `name="indPag"` and `name="includeCobranca"`.
3. `TestDepartureNFePreviewCarriesInputsToConfirm` — `departure_input_template_test.go:125`: departure preview must preserve `name="indPag" value="1"` and `name="includeCobranca" value="0"`.

These failures relate to the departure NF-e payment/billing/transport input wiring (migration `000030`). The template no longer renders the `indPag`/`includeCobranca` inputs and the parser no longer surfaces an explicit `false` for `IncludeCobranca`.

## 2. JavaScript unit tests

Command: `cd test && bun test unit/`
Result: **PASS**
- 76 pass, 0 fail
- 265 `expect()` calls
- 6 test files
- Runtime: ~130 ms

## 3. E2E tests (Playwright)

Command: `cd test && bun run test:e2e`
Result: **FAIL** — 115 failed, 82 passed, 4 skipped (201 total across chromium/firefox/webkit). Duration ~11.8 min.

### Root cause of the bulk of E2E failures

`GET /nfe/emitir` returns HTTP 500 ("Falha ao carregar a página de NF-e"). This is caused by a nil-interface trap in `buildDetachedPageView` (`router/nfe_router/detached_router.go`):

```go
farmNFeConfig, err := nfeModel.GetFarmConfig(farmID)   // err is type `error`
...
people, err := pModel.GetPeopleByFarm(farmID)           // GetPeopleByFarm returns *model_error.ModelError
if err != nil {                                         // typed nil pointer boxed into `error` interface => always true
    ...
}
```

`GetPeopleByFarm` returns `*model_error.ModelError`; on success it returns an untyped `nil` that becomes a typed nil pointer. Because `err` was already declared as the `error` interface (from `GetFarmConfig`), the typed nil pointer is boxed into a non-nil interface, so `err != nil` is true even on success. The logged message was `buildDetachedPageView GetPeopleByFarm error: <nil>`. This breaks every test that navigates to `/nfe/emitir`.

This regression was introduced by commit `07d6611` (which converted previously-ignored errors into checked errors while reusing a generic `error` variable).

### Other observed E2E failure signals

- `tests/departure.spec.js` — "numeric field overflow (SQLSTATE 22003)" on `add_get_departure` (test data magnitude issue) and a vehicle plate assertion failure (`Expected "ABC123", Received ""`).
- `tests/detached_decimals.spec.js` — comma-typed decimals / rate parse toast assertions (related to the H2 comma-vs-dot decimal handling noted in the prior audit).
- Many failures are downstream of the `/nfe/emitir` 500 and the departure overflow above (cascading timeouts).

### Note on SEFAZ

The E2E run used `NFE_SEFAZ_GUARD=block` (set by `global-setup.js`); no real SEFAZ/SVC traffic was attempted. No `nfe_farm_config` rows are seeded by the fixture.

## Conclusion

The JavaScript unit battery is green. The Go backend has 3 failing tests in `nfe_router` (departure NF-e input wiring). The E2E suite is largely red, primarily due to a regression in `buildDetachedPageView` (`/nfe/emitir` 500) introduced by the most recent commit, plus a departure-creation numeric overflow in test data. No tests were modified during this audit.
