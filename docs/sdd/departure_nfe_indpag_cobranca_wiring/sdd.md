# sdd:departure_nfe_indpag_cobranca_wiring

## Status
Done

## Goal
Wire the already-supported `indPag` and `includeCobranca` payment/billing inputs through the departure NF-e modal, preview round-trip, and form parser, so users can mark a payment as "a prazo" and suppress `<cobr>/<fat>`.

## Requirements

### Functional
- Modal renders `name="indPag"` select and `name="includeCobranca"` checkbox (plus `includeCobrancaSpecified` marker).
- Parser populates `DepartureInvoiceOverrides.IndPag` and `IncludeCobranca`.
- Preview carries `indPag` and `includeCobranca` as hidden inputs.
- Router echoes `IndPag`/`IncludeCobranca` into the preview template.

### Non-Functional
- pt-br validation toasts; no DB/SEFAZ changes.

## Acceptance Criteria
- `go test ./router/nfe_router/` passes the 3 B1 tests.
- `departure_nfe_inputs.spec.js` (chromium) passes.

## Verification Results
- `go test ./router/nfe_router/` — PASS
- `make test-go` — PASS (no failures)
- `cd test && bun test unit/` — not run (no JS changes)