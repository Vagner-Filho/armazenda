# Implementation Plan: detached_nfe_product_selector

## Phases

### Phase 1 — Model & service types
- `model/nfe_model/detached_model.go`: add `GlobalProductID *uint8` with JSON tag
  `global_product_id,omitempty` to `DetachedInvoiceItem`.
- `service/nfe_service/detached_service.go`: add `GlobalProductID *uint8` to
  `DetachedItemInput`; in `buildDetachedItems`, copy `input.GlobalProductID`
  into `itemForDB.GlobalProductID`.
- Add unit tests for JSON round-trip of `GlobalProductID`.

### Phase 2 — Router parsing & view rendering
- `router/nfe_router/detached_router.go`:
  - Add `parseNamespacedProductID(raw string) *uint16` (farm extraction).
  - Add `parseNamespacedGlobalProductID(raw string) *uint8`.
  - Add `namespacedProductRef(farm *uint16, global *uint8) string`.
  - Replace inline parsing in `parseDetachedItems`.
  - Update `buildDetachedPageView`, `detachedPreviewItems` to emit namespaced refs.
- `router/nfe_router/detached_profile_router.go`:
  - Replace inline parsing in `parseDetachedProfileItems`.
  - Update `detachedProfileFormRows` to emit `farm:<id>`.
- Templates:
  - `templates/pages/nfe-emit.html`: namespaced option values + selected conditions.
  - `templates/nfe/nfe-rascunho-form.html`: namespaced farm option values.

### Phase 3 — Tests
- Update/add `TestParseDetachedItems_*` cases for namespaced values.
- Run `go test ./router/nfe_router/ ./service/nfe_service/ ./model/nfe_model/`.
- Run `make test-go`.

## Risks
- Profile items have no `GlobalProductID`; view helper prefers farm then global.
- Legacy bare numeric values remain accepted as farm IDs for backward compat.
```

Now implement. Let's start Phase 1.

Edit model.<|end|><|start|> Assistant to=edit<|channel|>analysis <|constrain|>xml<|message|><rfl:parameter name="path">
/home/vfilho/projects/armazenda/model/nfe_model/detached_model.go