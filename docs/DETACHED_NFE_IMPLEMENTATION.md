# NF-e Avulsa (Detached NF-e) - Implementation Summary

## Overview
Implemented a new feature allowing users to emit NF-e (Brazilian electronic invoices) independently from departures (romaneios). This gives users full control over invoice data without requiring a departure record.

## What Was Implemented

### 1. Database Layer
- **Migration 000021**: Created `detached_nfe_invoice` and `detached_nfe_tax_rates` tables
- Separate from existing `nfe_invoice` table to maintain clean separation of concerns
- Supports multi-item invoices via JSONB storage
- Links to person table for recipient normalization

### 2. Model Layer (`model/nfe_model/detached_model.go`)
- `DetachedInvoice` struct with all invoice fields
- `DetachedInvoiceItem` struct for line items
- CRUD operations:
  - `CreateDetachedInvoice`
  - `GetDetachedInvoicesByFarm`
  - `GetDetachedInvoiceByAccessKey`
  - `UpdateDetachedInvoiceStatus`
  - `UpdateDetachedInvoiceSignedXML`
  - `UpdateDetachedInvoiceAuthorizedXML`
  - `UpdateDetachedInvoiceCancelled`
  - `SupersedeDetachedInvoice`
- Retry support:
  - `GetPendingDetachedInvoicesForRetry`
  - `GetDraftDetachedInvoicesForRetry`
  - `IncrementDetachedRetryCount`

### 3. Service Layer (`service/nfe_service/detached_service.go`)
- `DetachedRecipient` struct supporting both existing persons and inline creation
- `DetachedItemInput` struct for item data
- `DetachedInvoiceInput` struct for complete invoice input
- `prepareDetachedBuildData`: single-source validation/mapping shared by preview and emission (dry-run mode performs no writes)
- `GenerateDetachedPreviewDANFE`: non-fiscal preview PDF with all items and summed totals
- `BuildDetachedInvoice`: Main method that builds, signs, and sends to SEFAZ
- `CancelDetachedInvoice`: Cancels authorized invoices
- Helper methods for recipient resolution and item building

### 4. Person Model Extension (`model/person_model/model.go`)
- Added `CreatePersonForDetachedNFe` method
- Creates person records inline when users provide new recipient data
- Handles both natural (CPF) and legal (CNPJ) persons

### 5. Router Layer (`router/nfe_router/detached_router.go`)
- `GET /nfe/emitir` - Full-page emission form
- `POST /nfe/emitir/preview` - Non-fiscal DANFE preview (no number allocation, no person creation, no persistence, no SEFAZ call)
- `POST /nfe/emitir/build` - Submit and emit invoice
- `GET /nfe/avulsa/list` - List detached invoices
- `GET /nfe/avulsa/download/xml/:accessKey` - Download XML
- `GET /nfe/avulsa/download/danfe/:accessKey` - Download DANFE (placeholder)
- `POST /nfe/avulsa/cancel/:accessKey` - Cancel invoice

### 6. Worker Integration (`service/nfe_service/worker.go`)
- Added `processDetachedPendingInvoices` function
- Processes pending detached invoices (status polling)
- Processes draft detached invoices (SVC retry - basic implementation)
- Integrated into `StartRetryWorker` for automatic background processing

### 7. UI Layer
- **Menu**: Enabled "Emitir" link in NF-e submenu
- **Emission Form** (`templates/pages/nfe-emit.html`):
  - Recipient selection (existing or new, including optional e-mail)
  - Multi-item support with dynamic add/remove (rows are reindexed contiguously on removal)
  - Per-item farm product selector that auto-fills description and NCM
  - All standard NF-e fields (CFOP, nature operation, complementary info)
  - Transporte fieldset (modalidade do frete)
  - Tributação fieldset (CST/cClassTrib + "Usar taxa padrão" rate overrides)
  - "Gerar Pré-visualização" submits to the preview endpoint; the form stays
    in the DOM so its values survive a "Voltar" from the preview
- **Preview Fragment** (`templates/nfe/nfe-detached-preview.html`):
  - Non-fiscal DANFE ("Documento de Pré-visualização — Sem Valor Fiscal")
  - Carries every parsed field as a hidden input, so "Confirmar e Emitir"
    re-parses the exact data the preview was generated from
  - Success feedback with signed XML download link (`nfe-detached-result`)
- **List View** (`templates/nfe/nfe-list.html`):
  - Updated to show both departure-based and detached invoices
  - Added "Tipo" column showing "Romaneio" or "Avulsa"
  - Color-coded badges for visual distinction

### 8. Tests
- **Model Tests** (`model/nfe_model/detached_model_test.go`):
  - DetachedInvoiceItem serialization
  - DetachedInvoice struct fields
  - Tax rates handling
  - Multi-item support
- **Service Tests** (`service/nfe_service/detached_service_test.go`):
  - DetachedRecipient with PersonID
  - DetachedRecipient with inline fields
  - DetachedItemInput variations
  - DetachedInvoiceInput with single/multiple items
  - Tax rates configuration

## Key Features

### Multi-Item Support
- Users can add multiple line items to a single invoice
- Each item can have its own product, quantity, price, and weight
- Dynamic form with add/remove buttons (no HTMX, simple JavaScript)

### Recipient Flexibility
- **Existing Recipients**: Select from dropdown of farm's persons
- **New Recipients**: Fill inline form to create new person record
- Automatic person creation preserves database normalization

### Tax Configuration
- Inherits farm's default tax rates
- Per-invoice overrides for all tax types (ICMS, PIS, COFINS, IBS, CBS)
- CST codes configurable per invoice

### SEFAZ Integration
- Full emission flow: build → sign → send → handle response
- SVC contingency support (automatic fallback)
- Status polling via worker
- Cancellation support

### List Integration
- Unified view showing both departure-based and detached invoices
- Type column with color-coded badges
- All actions available (download XML, DANFE, cancel)

## Technical Decisions

### Separate Table
- Chose to create `detached_nfe_invoice` instead of modifying `nfe_invoice`
- Keeps concerns separated and avoids complex nullable columns
- Easier to maintain and understand

### JSONB for Items
- Used JSONB column for storing items array
- Simpler than creating a separate items table
- Items are immutable once invoice is signed

### Person Normalization
- Inline recipients create actual person records
- Maintains database normalization
- Allows reuse in future invoices

### Worker Integration
- Detached invoices processed alongside departure-based invoices
- Same retry logic and backoff strategy
- Automatic status polling for pending invoices

## What's NOT Implemented

### DANFE Generation
- Download endpoint returns 501 Not Implemented
- Would require implementing PDF generation for detached invoices
- Can be added later following same pattern as departure-based DANFE

### Full SVC Rebuild for Drafts
- Draft retry only increments counter
- Full rebuild would require storing complete InvoiceInput
- Users can manually retry by re-emitting

### Vehicle Data
- The router parses `vehicleId` but the form does not expose it yet
- The service has a `TODO: fetch vehicle plate` and skips vehicle data
- Add the field only after the backend completes the vehicle lookup

### Payment Conditions
- Service hardcodes `IndPag: 1` / `TPag: "90"` (other)
- Needs a `DetachedInvoiceInput` extension before the UI can offer payment forms

### Volume / Freight Value Details
- Volumes are auto-derived ("Granel", single entry from total weights)
- Needs transport-data support in the service before the UI can collect them

## Testing

All unit tests pass:
```
ok  	armazenda/model/nfe_model
ok  	armazenda/service/nfe_service
```

Tests cover:
- Struct initialization and field access
- Tax rate handling
- Multi-item scenarios
- Recipient variations

## Files Created/Modified

### New Files
- `model/armazenda_database/migrations/000021_detached_nfe.sql`
- `model/nfe_model/detached_model.go`
- `model/nfe_model/detached_model_test.go`
- `service/nfe_service/detached_service.go`
- `service/nfe_service/detached_service_test.go`
- `service/nfe_service/detached_preview_test.go`
- `router/nfe_router/detached_router.go`
- `templates/pages/nfe-emit.html`
- `templates/nfe/nfe-detached-preview.html`
- `templates/nfe/nfe-detached-result.html`

### Modified Files
- `model/person_model/model.go` - Added CreatePersonForDetachedNFe
- `service/nfe_service/worker.go` - Added detached invoice processing
- `router/nfe_router/router.go` - Integrated detached routes, updated list handler
- `templates/layout/menu.html` - Enabled "Emitir" link
- `templates/nfe/nfe-list.html` - Added type column, unified list
- `templates/nfe/nfe-list-item.html` - Updated to accept Invoice/Type dict

## Usage Flow

1. User navigates to NF-e → Emitir
2. Selects existing recipient or fills new recipient form
3. Adds one or more items with product details
4. Configures CFOP, nature operation, and optional complementary info
5. Optionally overrides tax rates
6. Clicks "Gerar Pré-visualização"
7. System validates the same way emission would and renders a non-fiscal DANFE
   preview — no number allocated, no person created, nothing persisted, no
   SEFAZ call. Every form field travels with the preview as a hidden input.
8. User confirms with "Confirmar e Emitir" (or clicks "Voltar" to adjust the
   form, whose values are preserved)
9. System allocates the number, signs, persists, and sends to SEFAZ
10. On success, the result panel offers the signed XML download and the invoice
    appears in the unified list with the "Avulsa" badge
11. User can download XML, view status, or cancel if authorized

## Future Enhancements

1. Implement DANFE PDF generation for detached invoices
2. Store full InvoiceInput for complete SVC rebuild capability
3. Add filtering/search in unified list
4. Add export/reporting features for detached invoices
