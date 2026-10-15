# plan — detached_infadprod_wiring

## Context

Per-item additional info (`infAdProd`) is the last corpus tag with no wired path end-to-end (the homolog residual-tags analysis, `docs/research_nfe_use_case/HOMOLOG_REPLAY_COMPARISON.md` + the 2026-10-15 resúme conversation). The lower half already exists and works:

- Builder: `<infAdProd>` conditionally emitted as the last `<det>` child, sanitized via `setSchemaText` (`pkg/nfe/xml/builder.go:314-315`).
- Entity: `entity.ItemData.InfAdProd` (`pkg/nfe/entity/invoice.go:114`).
- DANFE: the item mapping carries it (`service/nfe_service/detached_service.go:917`).
- Realworld goldens: the adapter reads fixture infAdProd into the input (`pkg/nfe/xml/realworld_test.go:660`) — the 2/115 corpus users of the tag are already exercised.

This SDD wires the top half: **form → parse → preview round-trip → service mapping → items_json**.

## Test strategy

| # | Behavior | Layer / location | Fixtures |
|---|----------|------------------|----------|
| T1 | Router parses `items[i].infAdProd`, trims; >500 chars → pt-br warning toast naming the item | `router/nfe_router/detached_router_test` (parseDetachedItems unit test if present, else via form-level test) | gin test context with POST form |
| T2 | Preview round-trip: parsed value lands in `detachedPreviewItem.InfAdProd` and renders into the fragment hidden input | router test (`buildDetachedPreviewViewData`) + template render check | existing preview-view fixtures |
| T3 | `buildDetachedItems`→`entity.ItemData.InfAdProd` (non-empty + blank→empty) | `service/nfe_service/detached_items_*_test.go` pattern | existing farm-config fixtures |
| T4 | Builder det emission: last child when non-empty; omitted when empty | `pkg/nfe/xml/builder_test.go` (det children order, idiom of `assertChildOrder`) | `minimalInvoiceInput()` |
| T5 | items_json backward-compat + round-trip through the model | `pkg/nfe/model/nfe_model` test (json round-trip of DetachedInvoiceItem with/without key) | none (pure struct JSON test) |
| T6 | Full suites: `make test-go` + realworld goldens green | `pkg/nfe/xml` | — |
| T7 | e2e: fill the item info field, preview shows it (data-test-id on the input), confirm → DαNFE shows the note under the item | `test/e2e` | existing detached-nfe e2e harness |

Baseline before starting: `go test ./pkg/nfe/xml/ ./router/nfe_router/ ./service/nfe_service/` + `make test-go` (note any pre-existing failures); then flip the SDD Status to WIP.

## Phases

### Phase 1 — Router + form input
1. `router/nfe_router/detached_router.go`:
   - `detachedPreviewItem` gains `InfAdProd string`.
   - `parseDetachedItems` reads `items[i].infAdProd` (TrimSpace; length guard ≤500 with pt-br toast "Informações adicionais do item X excede 500 caracteres").
   - `detachedPreviewView`/`buildDetachedPreviewViewData` carries it; `nfe-detached-preview.html` adds the hidden `items[i].infAdProd` input (id where items are rendered).
2. Tests T1 + T2 (with the change, not deferred).

### Phase 2 — Template (UI)
1. `templates/pages/nfe-emit.html`: add the per-item input (`name="items[{{ $index }}].infAdProd"`, `data-test-id="item-inf-ad-prod-{{ $index }}"`, placeholder "Info. adicionais do item (ex. Série, OBS)")“ in the item-row grid; add-item clone already clears inputs generically — verify `reindexItems()` renames it (name-based regex covers any input ✓).
2. e2e T7 focused run.

### Phase 3 — Service + model
1. `service/nfe_service/detached_service.go`: `DetachedItemInput` gains `InfAdProd string`; `buildDetachedItems` maps it into `entity.ItemData.InfAdProd` (blank → "").
2. `model/nfe_model/detached_model.go`: `DetachedInvoiceItem` gains `InfAdProd string json:"inf_ad_prod,omitempty"`; CreateDetachedInvoice accept func (NamedArgs pattern; no migration — JSONB) and any DB→rebuild mapping feeding `buildDetachedBuildData` for SVC retries carries it back into the input.
3. Tests T3 + T5 (with the change).

### Phase 4 — Builder assertion (no builder code change)
1. T4: extend the det-related builder test to document the emission contract (last `<det>` child, sanitize path).
2. Run T6 (full `make test-go` + goldens); gofmt; fill SDD Verification Results; flip Status Done.

## Risks / out of scope

- **Departure flow (romaneio)**: NOT wired here (matches the detached-only scope of the input-wiring decisions); the builder support is shared, so a future departure wiring is a small slice.
- **Rascunho (detached profile)**: if the profile stores item fields via the same items_json-shaped storage, the field follows automatically; verify during Phase 3 — otherwise raise a follow-up (rascunho doesn't keep item info today either, see the dialog text: "Quantidade e peso não são salvos" — infAdProd joins that bucket unless the profile storage includes it).
- **DANFE layout**: infAdProd already flows to the print mapping; no danfe-service change.
- **Length cap**: 500 chars is a pragmatic app-level guard (schema infAdProd TString allows larger; SEFAZ accepts up to ~50000 — we pick conservative + tóast; adjustable if needed).
- Cosmetic: `items_json` grows one optional key; no migration.
