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
- `BuildDetachedInvoice`: Main method that builds, signs, and sends to SEFAZ
- `CancelDetachedInvoice`: Cancels authorized invoices
- Helper methods for recipient resolution and item building

### 4. Person Model Extension (`model/person_model/model.go`)
- Added `CreatePersonForDetachedNFe` method
- Creates person records inline when users provide new recipient data
- Handles both natural (CPF) and legal (CNPJ) persons

### 5. Router Layer (`router/nfe_router/detached_router.go`)
- `GET /nfe/emitir` - Full-page emission form
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
  - Recipient selection (existing or new)
  - Multi-item support with dynamic add/remove
  - All standard NF-e fields (CFOP, nature operation, complementary info)
  - Tax rate overrides
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

### Farm Product Integration
- Form doesn't yet integrate with `farm_product` table
- All products entered manually
- Can be enhanced to allow selection from farm products

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
- `router/nfe_router/detached_router.go`
- `templates/pages/nfe-emit.html`

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
6. Clicks "Emitir NF-e"
7. System builds, signs, and sends to SEFAZ
8. On success, invoice appears in unified list with "Avulsa" badge
9. User can download XML, view status, or cancel if authorized

## Future Enhancements

1. Implement DANFE PDF generation for detached invoices
2. Add farm product selector integration
3. Store full InvoiceInput for complete SVC rebuild capability
4. Add filtering/search in unified list
5. Add export/reporting features for detached invoices
