# plan:detached_nfe_operation_profiles

Implementation plan for "Rascunhos de NF-e" (named reusable emission rascunhos) and item-scoped CFOP in detached NF-e. Specification: [sdd.md](./sdd.md).

## Design Summary

- The user-facing concept is **"Rascunho de NF-e"**, mirroring the existing "Rascunho de Entrada"/"Rascunho de Saída" in wording and lifecycle (create / edit / delete with confirmation; no archive). Internal identifiers stay English `profile` (`DetachedProfile`, `detached_profile_model.go`, `detached_profile_router.go`) — see the Naming table in `sdd.md`.
- A rascunho is farm-scoped and stores an **optional** recipient reference, an ordered `items_json` snapshot (product details, per-item CFOP, exact unit price), and optional operation/freight/tax defaults.
- Rascunhos are managed on their **own page** and are applied to the detached emission form by navigation: `GET /nfe/emitir/rascunho/:rascunhoId` renders the Emitir page with the rascunho's values prefilled. This differs from entry/departure rascunhos, which generate documents from a modal without leaving the list page — the NF-e domain is much more complex, so a full page is chosen instead. The route validates farm ownership, sets no quantity/weight, and leaves every field editable.
- CFOP moves to each detached item throughout the input, XML, persistence, and retry paths. The detached invoice header has no CFOP. The romaneio NF-e flow and its CFOP handling are untouched.
- The NF-e `draft` status badge becomes **"Não enviada"** (display-only: `nfe-list-item.html`, `nfe-existing-modal.html`, plus a template sweep) so "rascunho" means exactly one thing in the NF-e area. The internal status string `'draft'` and worker logic are unchanged.
- Inline recipient data must be explicitly saved as a farm contact when the user saves a rascunho; rascunho management and preview do not otherwise create contacts or invoices.

## Test Strategy

Run the existing focused suites before implementation and record any pre-existing failures before changing the spec status from `Backlog` to `WIP`:

```bash
go test ./service/nfe_service/... ./model/nfe_model/...
```

Test the behavior at the narrowest useful layer as each phase lands:

| Behavior | Test location/layer | Environment / command |
|---|---|---|
| Per-item CFOP mapping, four-digit validation, mixed-CFOP `NaturezaOp` resolution, rascunho/default/edit precedence, and nil-vs-zero rates | `service/nfe_service/detached_service_test.go` and new `detached_profile_service_test.go` | `go test ./service/nfe_service/...` |
| Item/rascunho JSON serialization, exact unit prices, optional farm-product reference, and absence of invoice-level detached CFOP | `model/nfe_model/detached_model_test.go` and new `detached_profile_model_test.go` | `go test ./model/nfe_model/...` |
| Form parsing (per-item CFOP fields), farm scoping, `GET /nfe/emitir/rascunho/:rascunhoId` prefilled rendering (rascunho → farm → fallback resolution, editable fields, no side effects), and preview hidden-field fidelity | New tests in `router/nfe_router/` | `go test ./router/nfe_router/...` |
| "Não enviada" status label, rascunho page lifecycle (create/apply-via-navigation/edit/delete), migration backfill/drop, farm isolation, explicit inline-contact save, and no issuance side effects | `test/e2e/tests/detached_nfe_rascunho.spec.js` (new) plus a template render guard test for the label and fragments | Test Postgres/app managed by Playwright global setup; add isolated NF-e config, farm/contact/product fixtures because the current fixture does not seed `nfe_farm_config`. Run `cd test/e2e && bun run test --grep "Rascunho de NF-e"` and `make test-e2e`. |
| Departure-flow and repository regressions | Existing Go, JavaScript, and browser suites | `make test` before completion (Docker required for E2E). |

Do not claim a SEFAZ-authorized emission test: rascunho application and per-item XML construction can be verified without a real A1 certificate or SEFAZ availability. Use service/XML tests plus a non-fiscal preview flow instead.

## Phase 1 — Baseline, migrations, and persistence models

**Files:**

- `model/armazenda_database/migrations/000022_detached_nfe_item_cfop.sql` (new)
- `model/armazenda_database/migrations/000023_detached_nfe_operation_profiles.sql` (new)
- `model/nfe_model/detached_model.go`
- `model/nfe_model/detached_profile_model.go` (new)
- `model/nfe_model/detached_model_test.go`
- `model/nfe_model/detached_profile_model_test.go` (new)

1. Run the focused baseline command above; record results in `sdd.md` and then update Status to `WIP` immediately before implementation begins.
2. In migration `000022`, preserve existing per-item `cfop` values in `detached_nfe_invoice.items_json`; fill a missing item value from that row's legacy `detached_nfe_invoice.cfop`; then drop only the detached invoice-level `cfop` column. Do not modify the departure-linked `nfe_invoice` schema or farm NF-e configuration. Test both a legacy row with existing per-item CFOP and one requiring backfill.
3. In migration `000023`, add a farm-scoped rascunho (`profile`) table with a required name, an **optional nullable `recipient_id`** reference, operation/freight fields, optional tax override/rate fields using precision compatible with `detached_nfe_tax_rates`, `items_json JSONB`, and timestamps. No archive column — deletion is hard, mirroring romaneio rascunhos. Enforce case-insensitive name uniqueness per farm with a unique index on `(farm_id, LOWER(name))`; hard-deleted rows free the name. Rascunho item JSON contains optional `farm_product_id`, product fields, `cfop`, and `unit_price`; it does not contain quantity or gross weight.
4. Add model types and CRUD/query methods (Create/Get/GetAll/Update/Delete). Scope all queries by `farm_id`; validate the optional recipient and optional farm-product ownership before saving. Persist rascunho item order and decimal values without float conversion. Use a transaction when replacing rascunho data so failed edits do not leave a partially updated rascunho.
5. Remove `DetachedInvoice.CFOP` and remove the top-level CFOP argument/column from detached insert/select/scan paths, including pending-retry queries. Keep `DetachedInvoiceItem.CFOP` in `items_json`.
6. Add model serialization tests for two rascunho items with distinct CFOPs and exact prices. Add the migration/schema/data test using the test database fixture in Phase 4 (or the repository's available migration test harness if one exists).

**Phase gate:** `go test ./model/nfe_model/...` passes, and migration testing proves old item values are retained before the legacy column is dropped.

## Phase 2 — Detached service semantics and rascunho resolution

**Files:**

- `service/nfe_service/detached_service.go`
- `service/nfe_service/detached_profile_service.go` (new, if keeping CRUD separate)
- `service/nfe_service/interfaces.go` / rascunho test mocks (new or extended, following the package's existing interface/mock pattern)
- `service/nfe_service/detached_service_test.go`
- `service/nfe_service/detached_profile_service_test.go` (new)
- `pkg/nfe/entity/invoice.go` only if the detached input needs a dedicated rascunho/resolution DTO; do not change shared departure invoice CFOP semantics.

1. Add `CFOP string` to `DetachedItemInput`; remove `CFOP` from `DetachedInvoiceInput`. Make `buildDetachedItems` use each input item's CFOP and validate four ASCII digits before XML construction. Do not apply the first item's CFOP to other items.
2. Preserve a useful no-rascunho flow: seed each new item row from the existing detached/farm default in the router/UI; the resulting value is still an item value. The service does not use a detached invoice-level CFOP fallback.
3. Resolve `NaturezaOp` from the current form, then rascunho, then farm configuration. If absent, derive from the shared CFOP only when all items have the same CFOP; reject mixed-CFOP preview with a user-facing pt-br warning when no explicit/farm nature is available.
4. Implement rascunho resolution and CRUD service behavior, with the recipient reference being optional throughout. Unset tax rascunho fields inherit farm values; explicit zero rates remain explicit. Form edits after rascunho application are authoritative, and preview/confirm do not silently reload/reapply a changed rascunho.
5. Update detached invoice creation/retry mapping to store and rebuild item CFOPs solely from the item list. Do not change `BuildInvoiceFromDeparture` or `nfe_invoice` paths.
6. Add tests for per-item XML CFOP with two different values, invalid/missing CFOP, the nature-operation resolution rules, exact price/rate defaults, nil versus zero rates, and rascunho ownership failures.

**Phase gate:** `go test ./service/nfe_service/... ./model/nfe_model/...` passes, including existing detached preview and departure-flow regressions.

## Phase 3 — Rascunho routes and farm-scoped application API

**Files:**

- `router/nfe_router/detached_router.go`
- `router/nfe_router/detached_profile_router.go` (new)
- `router/nfe_router/detached_router_test.go` / `detached_profile_router_test.go` (new)
- `service/nfe_service/detached_profile_service.go`
- `model/person_model/model.go` only if an explicit helper is needed to save an inline recipient as a farm contact.

1. Register rascunho routes under the existing detached NF-e route group, using the internal `profile` vocabulary for paths/handlers; the public path uses `rascunho`. Derive farm ID from the authenticated session; never accept it as authoritative form input:
   - `GET /nfe/emitir/rascunho/:rascunhoId` — fetches the rascunho server-side after validating farm ownership and renders the **Emitir page** (`templates/pages/nfe-emit.html`) with its values prefilled (recipient when set, ordered item rows with product fields/CFOP/price, operation/freight/tax values). Quantity and gross weight are left empty. Cross-farm or missing rascunho: "Rascunho não encontrado" warning + no render of prefilled data. This replaces the previous in-form "apply without navigation" behavior of the sdd's earlier draft — the sdd now specifies navigation (sdd.md req. 7).
   - `GET /nfe/rascunhos` — rascunho management page (list + "Novo Rascunho").
   - `GET/POST /nfe/rascunho/...` — create/update/delete handlers for the rascunho editor + delete confirmation.
2. Implement save-from-form ("Salvar como rascunho") from the emission form. Require a rascunho name and at least one item with unit prices and item CFOPs; a recipient is optional — if the recipient is inline, make contact creation an explicit consequence of the user pressing the save action; link the created person to the rascunho. Do not allocate a number, create a detached invoice, sign XML, or call SEFAZ.
3. Return rascunho data in a form suitable for filling the detached emission form. Validate rascunho ownership and all person/product references on every apply/update request.
4. Use Brazilian Portuguese errors/toasts, mirroring the romaneio rascunho wording: "Rascunho não encontrado", "Deseja excluir o rascunho … ?", duplicate-name warning, and cross-farm ID rejection.
5. Test route parsing, CRUD validation, cross-farm rejection, the `GET /nfe/emitir/rascunho/:rascunhoId` rendering path (correct prefill, quantity/weight empty, fields editable, and no person/invoice/number side effects), inline-recipient rascunho save, and no invoice/number side effects. Keep preview's existing dry-run behavior unchanged.

**Phase gate:** `go test ./router/nfe_router/... ./service/nfe_service/... ./model/nfe_model/...` passes.

## Phase 4 — Emission form, rascunho panel, status label, and preview round trip

**Files:**

- `templates/pages/nfe-emit.html` — the shared emission form template, now rendered either blank or rascunho-prefilled via route data
- `templates/nfe/nfe-rascunho-editor.html`, `templates/nfe/nfe-rascunho-list.html` (new; list mirrors `templates/entry/entry-draft-list-item.html` / `templates/departure/departure-draft-list-item.html`)
- A new rascunho management page template joining list + editor, referenced by the routes in Phase 3
- `templates/nfe/nfe-detached-preview.html`
- `templates/nfe/nfe-list-item.html`, `templates/nfe/nfe-existing-modal.html`
- `router/nfe_router/detached_router.go`, `router/nfe_router/detached_profile_router.go`
- `test/e2e/tests/detached_nfe_rascunho.spec.js` (new)

1. Replace the detached form's invoice-level CFOP select with a required CFOP control in each item row. Update add/remove/reindex behavior to keep `items[i].cfop` aligned with each product and ensure every row's CFOP reaches `parseDetachedItems`.
2. Build the rascunho management page: "Rascunhos" heading, "Novo Rascunho" action, per-rascunho actions "Gerar NF-e a partir do rascunho" (navigates to `GET /nfe/emitir/rascunho/:rascunhoId`), "Editar Rascunho", and "Excluir Rascunho" — mirroring the romaneio list items, including the delete-confirmation pattern (`data-delete-text` + `hx-trigger="confirmed"`). Unlike entry/departure, which generate documents from a modal without leaving the list, "Gerar NF-e" here **navigates** to the dedicated Emitir page (domain complexity warrants a full page). Load only the current farm's rascunhos.
3. Wire the rascunho-driven rendering of the emission page (Phase 3 route): server-rendered prefill of recipient (when the rascunho sets one), ordered item rows including product fields, per-item CFOP, and unit price, plus the rascunho's operation/tax defaults; farm defaults fill whatever the rascunho leaves unset (explicit zero stays zero, blank inherits). Quantity and gross weight stay empty; every prefilled field is editable; the form never writes back to the rascunho except through explicit "Salvar como rascunho"/update actions, which must not emit an invoice.
4. Add the "Salvar como rascunho" action on the emission form (name + current form values; inline recipient saved as a contact only when the user presses it).
5. Rename the draft status label to "Não enviada" in `templates/nfe/nfe-list-item.html` and `templates/nfe/nfe-existing-modal.html`, then sweep `templates/` for any other NF-e status rendering and update it. Do not touch the entry/departure rascunho wording or any internal status string. Note: the pre-existing `nfe-detached-list` template reference (see sdd.md side finding) renders nothing today — record it, do not fix it here.
6. Update preview view data and hidden inputs to round-trip recipient/rascunho context and every item's CFOP and price. Confirm must emit the exact item rows previewed. Keep the rascunho ID informational only during confirmation; do not reload it and overwrite form edits.
7. Add focused Playwright checks for the rascunho page lifecycle (create/apply-via-navigation/edit/delete), multi-item distinct CFOP and prices, recipient selection, quantity/weight blanks, preview/Voltar behavior, the "Não enviada" status label (list and modal; no "Rascunho" status label anywhere in the NF-e area), and no emission side effects. Add dedicated NF-e config/contact/product fixtures without weakening the shared E2E fixture.
8. Follow `.agents/skills/ui-design/SKILL.md` for the page/form UI, put all visible strings in pt-br, and carry `{{ .CSPNonce }}` on any new inline script.

**Phase gate:** focused browser command passes:

```bash
cd test/e2e && bun run test --grep "Rascunho de NF-e"
```

## Phase 5 — Documentation and completion verification

1. Update `docs/DETACHED_NFE_IMPLEMENTATION.md` with the rascunho feature, per-item CFOP semantics, the distinction between `CFOP` and `Natureza da operação`, and the "Não enviada" status label.
2. Run focused suites and formatting/build checks:

```bash
go test ./service/nfe_service/... ./model/nfe_model/... ./router/nfe_router/...
go build ./...
```

3. Run the applicable completion suites for this cross-layer change:

```bash
make test-go
make test-e2e
make test
```

4. Run `gofmt -l` on changed Go files. Record exact PASS/FAIL/NOT RUN results and any environment limitations in `sdd.md`. Mark Status `Done` only when all acceptance criteria are verified; otherwise leave it `WIP` and record the gap.

## Risks and Mitigations

- **Legacy CFOP data loss:** migration backfills missing item codes before dropping the detached header column; test both already-populated and missing item values. Do not rewrite signed XML.
- **Mixed-CFOP operation nature ambiguity:** use explicit/rascunho/farm `NaturezaOp`; only derive from CFOP when all item values agree; reject ambiguous mixed cases.
- **"Rascunho" name collision with invoice status:** the status badge rename to "Não enviada" (plus a template sweep) keeps the word unambiguous, and internal `profile` identifiers avoid clashing with `GetDraftDetachedInvoicesForRetry`/`processDetachedDraftInvoice`. The romaneio rascunho wording is untouched.
- **Navigation-based apply breaks the "dirty form" guard:** romaneio rascunhos live in a modal, so a dirty form can block in place; NF-e applies via `GET /nfe/emitir/rascunho/:rascunhoId`, meaning the user can navigate away from unsaved edits. Keep the existing beforeunload-style confirmation (or an explicit dirty-state check at the list page before navigation covers the deliberate-use case), and do not treat lost form data silently.
- **Route rendering error surface:** `GET /nfe/emitir/rascunho/:rascunhoId` must never 500 on user error — a missing/cross-farm rascunho returns a user-facing "Rascunho não encontrado" warning (pt-br); only infrastructure failures should produce server errors.
- **Optional recipient emitters:** several downstream paths (validateRecipient, address mapping, persisted `recipient_id NOT NULL`) assume a resolved person for emission. Since a rascunho may lack a recipient, the prefilled form must not emit without one — the emission-time validation already covers it, but the route/apply path must never attempt to resolve a nil recipient at render time.
- **Stale fiscal or price defaults:** rascunho values are visible, editable starting values; prices are stored as exact rascunho item values, and rascunho changes never mutate in-progress or issued invoice snapshots.
- **Hard delete safety:** issued invoices snapshot all values, so deleting a rascunho cannot alter a prior invoice (AC 10); deleting frees the farm's name for reuse, kept consistent by the DB unique index.
- **Recipient/contact side effects:** creating a rascunho from inline recipient data explicitly saves that contact; preview and ordinary rascunho reads remain side-effect free. Test the number/person/invoice side effects separately.
- **Cross-farm data access:** scope rascunho CRUD, contacts, and products by authenticated farm in both service/model access and router tests.
- **Preview/confirm drift:** carry per-item CFOP and rascunho-populated values through the existing shared parse path and hidden-field round trip; confirm uses the submitted preview values rather than reapplying the rascunho.
- **Departure flow regression:** keep migration and code paths specific to `detached_nfe_invoice`/detached input. Run the full Go suite and all repository tests before completion.
