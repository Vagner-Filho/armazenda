# Implementation Plan

## Phase 1 — Parser
Edit `router/nfe_router/departure_input_router.go` `parseDepartureInvoiceOverrides` to parse `indPag` and `includeCobrancaSpecified`/`includeCobranca`.

## Phase 2 — Templates
- `templates/nfe/nfe-emit-modal.html`: add `indPag` select and `includeCobranca` checkbox + marker.
- `templates/nfe/nfe-preview.html`: add hidden `indPag`, `includeCobrancaSpecified`, `includeCobranca`.

## Phase 3 — Router data
- `getNFeEmitModal`: `DefaultIndPag`, `DefaultIncludeCobranca`.
- `previewNFe`: `IndPag`, `IncludeCobranca`.

## Test strategy
- Go: `go test ./router/nfe_router/`
- E2E: focused chromium spec.