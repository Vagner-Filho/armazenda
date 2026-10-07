# plan:detached_cfop_selector_ux

Implementation plan for the CFOP selector search/list and derived Origem/Destino dialog refinement. This iteration inherits the catalog, farm scoping, ordering, registration, and use-counting requirements from [the parent SDD](../detached_cfop_selector/sdd.md); it does not re-implement those features.

## Test Strategy

Run the focused existing tests before implementation, record any pre-existing failures in `sdd.md`, and only then change Status from `Backlog` to `WIP`.

| Behavior | Test location/layer | Environment / command |
|---|---|---|
| Search normalization and matching across code, description, and Origem/Destino | `test/unit/cfopSelector.test.js` (new pure-helper tests if the matcher is extracted); browser behavior in `test/e2e/tests/detached_cfop_selector.spec.js` | `make test-js`; `cd test/e2e && bun run test --grep "CFOP"` |
| Per-row filtering, selected-option behavior, exact-code selection, and `items[i].cfop` preview round trip | `test/e2e/tests/detached_cfop_selector.spec.js`; existing detached parser tests | `cd test/e2e && bun run test --grep "CFOP"`; focused router Go tests |
| Search placement and persistence while results scroll | Playwright assertion in `test/e2e/tests/detached_cfop_selector.spec.js` | Focused CFOP E2E in Chromium and the configured browser projects; no SEFAZ-backed action |
| Prefix-to-Origem/Destino mapping and server-side derivation | `service/nfe_service/cfop_service_test.go`, `router/nfe_router/cfop_router_test.go`, plus registration E2E DB assertion | `go test ./service/nfe_service/... ./router/nfe_router/...`; focused CFOP E2E |
| Selector clone/reindex, default selection, modal refresh, and rascunho surface | `test/e2e/tests/detached_cfop_selector.spec.js`, `test/e2e/tests/detached_nfe_rascunho.spec.js`, and existing detached NF-e tests | Focused CFOP E2E; `make test-e2e` before completion |
| Unrelated behavior/regression coverage | Full Go, JS, and E2E suites | `make test-go`, `make test-js`, `make test-e2e` |

**Environment and safety:** Go 1.25.4, Bun, and the Docker-backed Playwright PostgreSQL fixture. Do not add or seed `nfe_farm_config` rows, certificates, or certificate passwords. Keep E2E to page rendering, selector interaction, registration, and preview; never confirm emission or cancellation. The app starts a retry worker during E2E setup, so preserve the fixture's existing safe state and stop any manually started app after a smoke run.

## Phase 1 — Baseline and specification gate

**Artifacts:**

- `docs/sdd/detached_cfop_selector_ux/sdd.md`
- `docs/sdd/detached_cfop_selector_ux/plan.md`

1. Review these iteration-specific artifacts and the parent SDD; keep parent requirements outside this iteration unchanged.
2. Run the focused Go, JS, and CFOP E2E commands from the Test Strategy where practical. Record exact baseline outcomes and known unrelated failures.
3. Change Status to `WIP` only after baseline results are recorded.

## Phase 2 — Searchable CFOP option panel

**Likely files:**

- `templates/nfe/cfop-selector.html`
- `templates/nfe/cfop-options.html`
- `templates/nfe/cfop-option.html`
- `assets/js/cfopSelector.js`
- `templates/pages/nfe-emit.html` and `templates/nfe/nfe-rascunho-form.html` only if clone/reindex integration needs adjustment
- `test/unit/cfopSelector.test.js` (if a pure matcher is extracted)
- `test/e2e/tests/detached_cfop_selector.spec.js`
- `test/e2e/tests/detached_nfe_rascunho.spec.js` (use the visible custom selector instead of interacting with the visually hidden select)

1. Make each option's code, description, and `origin_destination` available to the selector without changing the visible `code — description` label.
2. Replace the externally placed search/native-menu interaction with a custom option panel: search at the top, an independently scrollable result area, and the search kept visible while results scroll.
3. Keep a form control synchronized with the visible selection and preserve the exact `items[i].cfop` name/value contract, default selection, groups, row-local search, and exact-code behavior.
4. Preserve initialization on dynamically cloned rows and rebuild visible options after the existing farm-options refresh. Reset cloned search state and keep each row's selection independent.
5. Add unit coverage for pure text normalization/matching if practical. Add Playwright coverage for code, accent-insensitive description, Origem/Destino, row isolation, selected option, exact-code selection, and sticky search while scrolling.

**Phase gate:** `make test-js` and focused CFOP E2E pass for the selector behavior.

## Phase 3 — Derived registration classification

**Likely files:**

- `templates/nfe/nfe-cfop-form.html`
- `assets/js/cfopDialog.js`
- `router/nfe_router/cfop_router.go`
- `router/nfe_router/cfop_router_test.go`
- `service/nfe_service/cfop_service_test.go` (extend existing mapping coverage if needed)
- `test/e2e/tests/detached_cfop_selector.spec.js`

1. Replace the editable Origem/Destino `<select>` with a non-editable output that updates from the CFOP prefix; show a neutral Portuguese prompt until a recognized family is entered.
2. Keep existing four-digit/family and description validation. In the router, derive `origin_destination` from the validated code using the existing service helper and pass only that derived value to persistence; never use a posted classification as the source of truth.
3. Update tests so a valid request with no `originDestination` succeeds and the saved row has the derived value; verify a forged value cannot produce a mismatched saved row. Invalid code families remain rejected.
4. Update the E2E modal assertions and retain the existing duplicate, refresh, auto-selection, and rascunho coverage. Do not add database schema changes.

**Phase gate:** `go test ./service/nfe_service/... ./router/nfe_router/...` and focused CFOP E2E pass; the E2E exercises no emission/cancellation path.

## Phase 4 — Integration and completion verification

1. Verify both detached surfaces, new-row default selection, row reindexing, modal refresh, and preview submission continue to work with the shared component.
2. Run completion checks and record exact results in `sdd.md`:

   ```bash
   go build ./...
   go test ./service/nfe_service/... ./router/nfe_router/...
   make test-go
   make test-js
   cd test/e2e && bun run test --grep "CFOP"
   make test-e2e
   gofmt -l <changed Go files>
   ```

3. Review the diff for crop/Safra or departure-linked NF-e changes and SEFAZ network paths. Update Status to `Done` only when acceptance criteria pass; otherwise keep it `WIP` and document any unverified criteria.

## Risks and Out of Scope

- A custom option panel can be clipped by row/dialog overflow or behave differently across browser engines. Keep result scrolling separate from the search header and verify the open panel and sticky position in Playwright.
- The option data contains both "Mais utilizados" and "Todos os CFOPs" groups. Preserve existing ordering, selected-value behavior, and farm metadata when constructing the visible panel.
- Existing detached CFOP E2E results in the parent SDD include pre-existing WebKit/full-suite flakes. Compare against the focused baseline, record exact final outcomes, and do not mask or relabel unrelated failures.
- `templates/crop/crop-selector.html`, the Safra flow, departure-linked NF-e, CFOP catalog/migration, farm ranking/counting semantics, and any real SEFAZ interaction are explicitly out of scope.
