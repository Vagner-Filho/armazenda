# sdd:detached_nfe_product_selector

## Status
Done

## Goal
Remove the value collision in the detached NF-e item product selector. Today
`buildDetachedPageView` appends global catalog products (`product.id`, `uint8`)
after farm products (`farm_product.id`, `SMALLINT`) into one
`<select name="items[N].farmProductId">`. Because both key spaces are small
integers, duplicate option values are essentially guaranteed, making the
selected option ambiguous on render, submit, and rascunho round-trips.

This SDD implements namespaced option values (`farm:<id>` / `global:<id>`) and
adds durable tracking of global catalog product usage via a new
`GlobalProductID` field on detached invoice items.

## Requirements

### Functional
- The combined product `<select>` must never render two options with the same
  `value`.
- Selecting a farm product persists its farm product ID (`FarmProductID`).
- Selecting a global catalog product persists its global product ID
  (`GlobalProductID`) and leaves `FarmProductID` nil.
- Selecting "none" persists neither ID.
- Rascunho (profile) items continue to support only farm products; global
  selections are not persisted in rascunhos.
- Preview/confirm round-trips must preserve both farm and global selections.
- Backward compatibility: a bare numeric value (legacy) is treated as a farm
  product ID.

### Non-Functional
- No database migration is required; IDs remain small ints internally.
- No change to the SEFAZ XML output.
- Existing E2E flows (which select by label) remain green.

## Acceptance Criteria
- AC1: `parseDetachedItems` and `parseDetachedProfileItems` correctly parse
  `farm:<id>`, `global:<id>`, bare numeric, and empty values.
- AC2: A global selection yields `FarmProductID == nil` and
  `GlobalProductID != nil`.
- AC3: A farm selection yields `FarmProductID != nil` and
  `GlobalProductID == nil`.
- AC4: `DetachedInvoiceItem` persists `global_product_id` in JSONB.
- AC5: Preview hidden inputs round-trip global selections.
- AC6: `go test ./router/nfe_router/ ./service/nfe_service/ ./model/nfe_model/` passes.
- AC7: `make test-go` passes.

## Verification Results
- `go test ./router/nfe_router/ -run 'Detached|detached|ParseDetached'` → PASS
- `go test ./service/nfe_service/ ./model/nfe_model/` → PASS
- `make test-go` → FAIL (3 pre-existing failures in `departure_input_router_test.go` /
  `departure_input_template_test.go`, unrelated to this change; verified by stashing
  the change and re-running the same tests on the baseline)
```

Now plan.md.<|end|><|start|> Assistant to=write<|channel|>analysis <|constrain|>xml<|message|><rfl:parameter name="path">
/home/vfilho/projects/armazenda/docs/sdd/detached_nfe_product_selector/plan.md