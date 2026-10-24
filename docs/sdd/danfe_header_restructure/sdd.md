# sdd:danfe_header_restructure

## Status
Done

## Goal
Relocate the DANFE barcode to the right side of the second row and restructure the DANFE header so information that would be lost when the receiver signs the first row (entrada/saída, folha, DANFE title) is moved to the second row. The first row keeps only `NF-e`, number, and série. The second row becomes three columns: emitter logo (uploadable, fallback to Armazenda's), moved header information, and barcode plus its existing access-key/protocol information. Logo storage uses a dedicated `farm_logo` table.

## Requirements

### Functional
- First row, second column renders only `NF-e`, `Nº`, and `SÉRIE`.
- Second row has 3 columns: logo, moved header info, barcode + existing info.
- Barcode is rendered in the rightmost column.
- Moved info includes `DANFE`, `Documento Auxiliar da Nota Fiscal Eletrônica`, `0 - ENTRADA`/`1 - SAÍDA` with active type, `Nº`, `SÉRIE`, and `FOLHA 1/1`.
- Emitters can upload a logo via NF-e config; when absent, Armazenda's default logo renders.
- Logo is persisted in dedicated `farm_logo` table.
- Preview and download DANFE paths both render the logo when available.

### Non-Functional
- Migration is idempotent.
- Barcode remains at least 6cm wide.
- No real SEFAZ calls in tests.
- User-facing text in pt-br where applicable.

## Acceptance Criteria
- `go test ./pkg/nfe/... ./service/nfe_service/... ./router/nfe_router/...` passes.
- `make test-go` passes.
- Generated PDFs remain valid (`%PDF-`).
- No panic when `LogoData` is empty or populated.

## Verification Results
- `go test ./pkg/nfe/service/...` — PASS
- `go test ./service/nfe_service/...` — PASS
- `go test ./router/nfe_router/...` — FAIL due 3 pre-existing unrelated failures (departure input payment/billing tests) present on baseline before this change.
- `make test-go` — FAIL due the same 3 pre-existing unrelated failures; no failures introduced by this change.