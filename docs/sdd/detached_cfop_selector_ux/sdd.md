# sdd:detached_cfop_selector_ux

## Status
WIP

## Goal
Improve the detached NF-e CFOP selector and registration dialog: search by code, description, or Origem/Destino; keep search available while scrolling the option list; and derive Origem/Destino from the entered CFOP instead of letting the user choose it.

**Parent specification:** [docs/sdd/detached_cfop_selector/sdd.md](../detached_cfop_selector/sdd.md). This iteration supersedes the parent requirements for selector layout/search and editable Origem/Destino (including the corresponding portion of its registration-validation acceptance criterion); all other parent requirements remain in force.

## Requirements

### Functional

#### CFOP selector search and list

1. The detached emission page and the rascunho editor continue to use the shared CFOP selector.
2. Search matches each option's CFOP code, description, and `origin_destination`, using case- and diacritic-insensitive substring matching. The displayed option text remains `code — description`.
3. Search affects only the current item row. The selected CFOP remains available while filtering, clearing the query restores the options, and typing a complete CFOP code continues to support direct keyboard selection.
4. The search input is inside the CFOP options panel, above the results. Scrolling the results must not scroll the search input out of view; the results area scrolls independently.
5. The selector remains usable with keyboard and assistive technology, and works for both initial rows and dynamically cloned rows.
6. The submitted CFOP field remains `items[i].cfop`. Farm ordering, optgroups, default CFOP behavior, post-registration refresh, and row-specific selection remain unchanged.

#### Derived Origem/Destino

7. The CFOP registration dialog does not provide an editable Origem/Destino control. It displays a non-editable value that updates as the CFOP is entered: `1`/`5` → `Mesmo estado`, `2`/`6` → `Outro estado`, and `3`/`7` → `Exterior`.
8. For an incomplete or invalid CFOP, the dialog does not present a valid derived value and does not bypass the existing server-side CFOP validation.
9. The registration endpoint derives the persisted `origin_destination` from the validated CFOP code. It does not trust a client-supplied classification; a request without an `originDestination` field succeeds when otherwise valid, and a supplied value cannot cause a classification inconsistent with the code.
10. Description validation, catalog and farm duplicate handling, farm scoping, and the successful modal refresh/auto-selection behavior remain unchanged.

### Non-Functional

- User-facing text remains in Brazilian Portuguese.
- The option panel and search provide visible focus and selected states, appropriate ARIA semantics, and keyboard operation.
- No database schema or migration change is expected.
- No new SEFAZ calls are introduced. Tests must not seed certificate/configuration data or exercise emission/cancellation paths that could contact SEFAZ.
- This iteration does not change `templates/crop/crop-selector.html`, the crop/Safra flow, or departure-linked NF-e behavior.

## Acceptance Criteria

1. **Search fields:** on the detached CFOP selector, searching by code, description, or Origem/Destino returns matching options; description search is accent-insensitive. Verify with focused Playwright coverage in `test/e2e/tests/detached_cfop_selector.spec.js` and table-driven search-normalization tests where practical (`make test-js`).
2. **Per-row behavior:** filtering one item row does not filter another; the selected value remains available; clearing restores options; exact-code keyboard selection still works. Verify in the focused selector E2E test.
3. **Sticky search:** the search input appears at the top of the open option panel and remains stationary while the result list scrolls. Verify in Playwright by scrolling the results container and checking the search input's position.
4. **Form contract and both surfaces:** a selected CFOP submits under `items[i].cfop` and continues through preview; new rows use the configured/default CFOP; the selector works in both detached emission and rascunho editing. Verify with existing/focused Go parser tests and Playwright coverage for both surfaces.
5. **Automatic classification:** the registration dialog displays the correct Origem/Destino as CFOP digits are entered and offers no editable classification field. Verify with a table-driven test for families 1, 2, 3, 5, 6, and 7 plus Playwright dialog assertions.
6. **Server authority:** a valid registration request without `originDestination` persists the classification derived from its CFOP; a client-supplied value cannot persist a mismatch. Invalid CFOP families remain rejected. Verify with router/service tests and a SEFAZ-free E2E database assertion.
7. **Registration regressions:** description limits, catalog/farm duplicate rejection, farm scoping, and option refresh/auto-selection continue to pass. Verify with `go test ./router/nfe_router/...` and the focused CFOP E2E suite.
8. **Scope and safety:** the crop selector and departure-linked NF-e flow are unchanged; no test triggers real emission or cancellation, and no certificate/configuration data is added to a test database. Verify by changed-file/code-path review and the applicable Go/E2E suites.
9. **Completion checks:** run and record exact results for `go build ./...`, focused Go tests, `make test-go`, `make test-js`, focused CFOP E2E, `make test-e2e`, and `gofmt -l` on changed Go files before setting Status to Done. Record unrelated baseline failures accurately.

## Verification Results

Baseline before implementation (2026-10-06):

| Check | Command | Result |
|---|---|---|
| Focused Go tests | `go test ./service/nfe_service/... ./router/nfe_router/...` | PASS (cached) |
| JavaScript unit tests | `make test-js` | PASS — 19 tests |
| Focused CFOP E2E | `cd test/e2e && bun run test --grep "CFOP"` | NOT RUN — the E2E fixture seeds `nfe_farm_config`; the SEFAZ isolation rules prohibit running the setup that seeds this row while the app is running. Fixture unchanged. |

Status remains WIP; the E2E acceptance criteria remain unverified until a safe test setup is available.

Implementation verification (2026-10-06):

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | PASS |
| Focused Go tests | `go test ./service/nfe_service/... ./router/nfe_router/...` | PASS |
| Full Go suite | `make test-go` | PASS |
| JavaScript unit tests | `make test-js` | PASS — 33 tests |
| JavaScript syntax | `node --check assets/js/cfopSelector.js && node --check assets/js/cfopDialog.js && node --check test/e2e/tests/detached_cfop_selector.spec.js && node --check test/e2e/tests/detached_nfe_rascunho.spec.js` | PASS |
| Formatting | `gofmt -l router/nfe_router/cfop_router.go router/nfe_router/cfop_router_test.go` | PASS — no output |
| Diff whitespace | `git diff --check` | PASS — no output |
| Focused CFOP E2E | `cd test/e2e && bun run test --grep "CFOP"` | NOT RUN — Playwright global setup executes `test/e2e/fixtures/test-user.sql`, which seeds `nfe_farm_config` while starting the app; prohibited by the SEFAZ isolation rules. No E2E setup or app was started. |
| Full E2E | `make test-e2e` | NOT RUN — same fixture/setup restriction as focused E2E. |

### Verification notes

- The selector E2E cases were updated for the custom option panel and server-derived classification, but were not executed. Browser rendering, sticky positioning, and the end-to-end behavior on both surfaces remain unverified; therefore Status remains WIP.
- The router test uses a fake registrar to verify that a request with no `originDestination`, or one with a forged value, persists the server-derived classification. Go unit tests and the JS pure-matcher tests are offline and passed.
- The E2E fixture was not changed, and no certificate/configuration data was added. Do not run E2E until its setup can comply with the SEFAZ isolation rules.
