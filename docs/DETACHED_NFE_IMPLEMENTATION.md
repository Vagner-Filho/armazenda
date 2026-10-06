# NF-e Avulsa (Detached NF-e) - Implementation Summary

## Overview
Implemented a new feature allowing users to emit NF-e (Brazilian electronic invoices) independently from departures (romaneios). This gives users full control over invoice data without requiring a departure record.

Farm users can also save named **"Rascunhos de NF-e"** (internally `DetachedProfile`, a reusable operation scenario — not a CFOP record) to start a detached emission with recurring data already filled in. Each detached NF-e item carries its own CFOP, and the NF-e `'draft'` status is displayed as **"Não enviada"** so the word "Rascunho" means exactly one thing in the NF-e area.

## What Was Implemented

### 1. Database Layer
- **Migration 000021**: Created `detached_nfe_invoice` and `detached_nfe_tax_rates` tables
- **Migration 000022**: Moved CFOP from `detached_nfe_invoice.cfop` to each item in `items_json` (backfills missing item values from the legacy column, then drops only the detached header column; `nfe_invoice` is untouched)
- **Migration 000023**: Created `detached_nfe_profile` (farm-scoped rascunhos) with an optional recipient, ordered `items_json` snapshot, operation/freight defaults, optional tax overrides/rates, and a case-insensitive unique index on `(farm_id, LOWER(name))`
- Separate from existing `nfe_invoice` table to maintain clean separation of concerns
- Supports multi-item invoices via JSONB storage
- Links to person table for recipient normalization

### 2. Model Layer (`model/nfe_model/detached_model.go`, `detached_profile_model.go`)
- `DetachedInvoice` struct with all invoice fields (no invoice-level CFOP)
- `DetachedInvoiceItem` struct for line items (includes `CFOP`)
- `DetachedProfile` / `DetachedProfileItem` structs for rascunhos
- Detached invoice CRUD operations:
  - `CreateDetachedInvoice`
  - `GetDetachedInvoicesByFarm`
  - `GetDetachedInvoiceByAccessKey`
  - `UpdateDetachedInvoiceStatus`
  - `UpdateDetachedInvoiceSignedXML`
  - `UpdateDetachedInvoiceAuthorizedXML`
  - `UpdateDetachedInvoiceCancelled`
  - `SupersedeDetachedInvoice`
- Rascunho CRUD, all scoped by `farm_id`:
  - `CreateDetachedProfile`, `GetDetachedProfile`, `GetDetachedProfilesByFarm`, `UpdateDetachedProfile`, `DeleteDetachedProfile`
  - `ValidateDetachedProfileReferences` (recipient + farm products must belong to the farm)
- Retry support:
  - `GetPendingDetachedInvoicesForRetry`
  - `GetDraftDetachedInvoicesForRetry`
  - `IncrementDetachedRetryCount`

### 3. Service Layer (`service/nfe_service/detached_service.go`, `detached_profile_service.go`)
- `DetachedRecipient` struct supporting both existing persons and inline creation
- `DetachedItemInput` struct for item data (includes a required four-digit `CFOP`)
- `DetachedInvoiceInput` struct for complete invoice input (no invoice-level CFOP)
- `prepareDetachedBuildData`: single-source validation/mapping shared by preview and emission (dry-run mode performs no writes). Validates each item CFOP as exactly four ASCII digits.
- `resolveDetachedNaturezaOp`: form value → farm default → derivation from the item CFOPs only when all items share one CFOP; mixed CFOPs without an explicit/farm value are rejected
- `GenerateDetachedPreviewDANFE`: non-fiscal preview PDF with all items and summed totals
- `BuildDetachedInvoice`: Main method that builds, signs, and sends to SEFAZ
- `CancelDetachedInvoice`: Cancels authorized invoices
- Rascunho service (`DetachedProfileInput`, `SaveDetachedProfile`, `GetDetachedProfile`, `ListDetachedProfiles`, `DeleteDetachedProfile`): create/update/list/delete with farm-scoped references; an inline recipient is saved as a farm contact only when the user explicitly presses "Salvar como rascunho"
- Helper methods for recipient resolution and item building

### 4. Person Model Extension (`model/person_model/model.go`)
- Added `CreatePersonForDetachedNFe` method
- Creates person records inline when users explicitly save a new recipient (emission or rascunho save)
- Handles both natural (CPF) and legal (CNPJ) persons

### 5. Router Layer (`router/nfe_router/detached_router.go`, `detached_profile_router.go`)
- `GET /nfe/emitir` - Full-page emission form
- `GET /nfe/emitir/rascunho/:rascunhoId` - Emitir page prefilled from a farm-scoped rascunho (cross-farm/missing → "Rascunho não encontrado" warning, never a 500)
- `POST /nfe/emitir/preview` - Non-fiscal DANFE preview (no number allocation, no person creation, no persistence, no SEFAZ call)
- `POST /nfe/emitir/build` - Submit and emit invoice
- `GET /nfe/rascunhos` - Rascunho management page
- `GET /nfe/rascunho/list` - Rascunho list fragment
- `GET /nfe/rascunho/form` / `GET /nfe/rascunho/form/:id` - New/edit rascunho editor
- `POST /nfe/rascunho` / `PUT /nfe/rascunho/:id` / `DELETE /nfe/rascunho/:id` - Rascunho CRUD (delete asks for confirmation like romaneio rascunhos)
- `POST /nfe/rascunho/salvar` - "Salvar como rascunho" from the populated emission form (204, no emission side effects)
- `GET /nfe/avulsa/list` - List detached invoices (pre-existing broken template reference, out of scope)
- `GET /nfe/avulsa/download/xml/:accessKey` - Download XML
- `GET /nfe/avulsa/download/danfe/:accessKey` - Download DANFE (fiscal PDF from the stored XML; "NF-e CANCELADA" banner for cancelled invoices)
- `POST /nfe/avulsa/cancel/:accessKey` - Cancel invoice

### 6. Worker Integration (`service/nfe_service/worker.go`)
- Added `processDetachedPendingInvoices` function
- Processes pending detached invoices (status polling)
- Processes draft detached invoices (SVC retry - basic implementation)
- Integrated into `StartRetryWorker` for automatic background processing

### 7. UI Layer
- **Menu**: "Emitir" and "Rascunhos" links in the NF-e submenu
- **Emission Form** (`templates/pages/nfe-emit.html`):
  - Recipient selection (existing or new, including optional e-mail)
  - Multi-item support with dynamic add/remove (rows are reindexed contiguously on removal)
  - Per-item farm product selector that auto-fills description and NCM
  - **Per-item CFOP selector** — searchable, farm-ordered select (catalog ∪ farm CFOPs, "Mais utilizados" + "Todos os CFOPs") with a "+" register modal; new rows start from the farm/default CFOP
  - All standard NF-e fields (nature operation, complementary info)
  - Transporte fieldset (modalidade do frete)
  - Tributação fieldset (CST/cClassTrib + "Usar taxa padrão" rate overrides)
  - Renders either blank or prefilled from a rascunho (recipient, ordered items, CFOPs, prices, operation/tax values); quantity and gross weight stay empty
  - "Gerar Pré-visualização" submits to the preview endpoint; the form stays
    in the DOM so its values survive a "Voltar" from the preview
  - "Salvar como rascunho" dialog (name + optional explicit update of the applied rascunho)
- **Rascunho Management** (`templates/pages/nfe-rascunhos.html`, `templates/nfe/nfe-rascunho-*.html`):
  - "Rascunhos" panel with "Novo Rascunho"
  - Per-rascunho actions: "Gerar NF-e a partir do rascunho" (navigates to the prefilled Emitir page), "Editar Rascunho", and "Excluir Rascunho" with the same `data-delete-text` + `hx-trigger="confirmed"` confirmation pattern as romaneio rascunhos
- **Preview Fragment** (`templates/nfe/nfe-detached-preview.html`):
  - Non-fiscal DANFE ("Documento de Pré-visualização — Sem Valor Fiscal")
  - Carries every parsed field (including `items[i].cfop`) as a hidden input, so "Confirmar e Emitir" re-parses the exact data the preview was generated from
  - Rascunho id travels informational only; the rascunho is never reloaded/reapplied at confirmation
  - Success feedback with signed XML download link (`nfe-detached-result`)
- **List View** (`templates/nfe/nfe-list.html`):
  - Updated to show both departure-based and detached invoices
  - Added "Tipo" column showing "Romaneio" or "Avulsa"
  - Color-coded badges for visual distinction
  - `'draft'` status renders as **"Não enviada"** (display-only; internal status string and worker behavior unchanged)

### 8. CFOP Catalog & Selector (Migration 000025)
- **Catalog** (`cfop`): system-wide, read-only reference table seeded by the migration with the complete Contabilizei CFOP list (167 codes) plus every CFOP assumed by `pkg/nfe/defaults`; each row carries `description` and the derived `origin_destination` (`Mesmo estado` / `Outro estado` / `Exterior`, from the first digit)
- **Farm CFOPs** (`farm_cfop`): farm-scoped extra codes registered through the selector modal (`GET /nfe/cfop/form`, `POST /nfe/cfop`); catalog duplicates and same-farm duplicates are rejected with pt-br warning toasts; no per-farm description overrides, so catalog ∪ farm is disjoint
- **Use ranking** (`farm_cfop_use`): per-farm counters bumped once per distinct item CFOP at the emission persist step (`BuildDetachedInvoice`), never by preview/rascunho saves or worker retries; failures only log
- **Selector component** (`templates/nfe/cfop-selector.html`): per-row search input (code + accent-insensitive description, never hides the selected option) + `<select name="items[i].cfop">` + register "+"; `GET /nfe/cfop/options` refreshes the farm-ordered option list after a registration
- **Model/view**: `model/cfop_model/` (merge + grouping + counters), `view/cfop/` (adds the resolved default and prepends it when missing)

### 9. Tests
- **Model Tests** (`model/nfe_model/detached_model_test.go`, `detached_profile_model_test.go`):
  - DetachedInvoiceItem serialization (per-item CFOP and exact prices)
  - DetachedProfile item JSON round trip with distinct CFOPs and prices
  - Nil-vs-explicit-zero tax rate distinction
  - Farm-product reference dedupe
- **Service Tests** (`service/nfe_service/detached_service_test.go`, `detached_profile_service_test.go`, `detached_preview_test.go`):
  - DetachedRecipient with PersonID / inline fields
  - Per-item CFOP mapping and four-digit validation
  - `Natureza da operação` resolution (form → farm → common CFOP; mixed CFOPs rejected)
  - Rascunho item validation (CFOP, price, explicit zero)
  - Tax rates/defaults and exact decimals
- **Router Tests** (`router/nfe_router/detached_router_test.go`, `cfop_router_test.go`):
  - Form parsing of per-item CFOP fields, rascunho items/recipient parsing
  - Preview hidden-field round trip (CFOP + exact price strings)
  - Rascunho editor view data (explicit zero vs inherit)
  - CFOP register-modal validation (malformed codes, description, Origem/Destino mismatch) with pt-br toasts
  - CFOP fragment rendering (`cfop-options` groups/ordering, `cfop-option`, `cfop-selector` form contract, modal)
- **CFOP Model/View/Service Tests** (`model/cfop_model/`, `view/cfop/`, `service/nfe_service/cfop_service_test.go`):
  - Seed integrity over the migration file (167 codes, first-digit derivation, system-assumed extras)
  - Merge grouping/ordering and default-CFOP prepend
  - Selector validation/derivation helpers
  - Deduped use counting with a hand-written incrementer mock (once per emission action, errors only log)
- **Template Guards** (`router/nfe_router/nfe_templates_test.go`):
  - `'draft'` renders "Não enviada" in the list row and existing modal; no NF-e surface labels it "Rascunho"
  - Romaneio entry/departure rascunho wording stays unchanged
- **E2E** (`test/e2e/tests/detached_nfe_rascunho.spec.js`, `detached_cfop_selector.spec.js`):
  - Rascunho lifecycle, application via navigation, farm isolation, inline-contact save, preview write-freedom, invoice-number/invoice side-effect checks, and the "Não enviada" list label
  - CFOP selector: per-row search (accent-insensitive, selected option survives, no cross-row influence), exact-code selection, preview round trip
  - Farm CFOP registration from the emission page and the rascunho editor (auto-select, farm-ordered re-render, catalog/farm duplicate toasts)
  - "Mais utilizados" ordering from the farm use counters

## Key Features

### Multi-Item Support
- Users can add multiple line items to a single invoice
- Each item can have its own product, quantity, price, and weight
- Dynamic form with add/remove buttons (no HTMX, simple JavaScript)

### Recipient Flexibility
- **Existing Recipients**: Select from dropdown of farm's persons
- **New Recipients**: Fill inline form to create new person record
- Automatic person creation preserves database normalization

### Per-item CFOP vs Natureza da Operação
- **CFOP is an item value**: every detached item carries its own required four-digit CFOP, mapped to that item's `<det><prod><CFOP>`; different items in one NF-e may use different codes
- **CFOP selection** uses a searchable, farm-ordered selector (catalog ∪ farm CFOPs) rendered per item row; the free-text input was replaced without changing the `items[i].cfop` form contract
- **Natureza da operação is invoice-level** and is resolved from the current form value, then a selected rascunho value, then the farm default; only when all items share one CFOP is it derived from that code. Mixed item CFOPs without an explicit/farm nature are rejected at preview with a pt-br warning
- The two fields are separate: no CFOP data is relabeled as Natureza da operação
- `detached_nfe_invoice.cfop` no longer exists; `nfe_invoice` and the departure-linked flow are unchanged

### CFOP Selector & Farm Catalog
- The system-wide catalog (`cfop`, migration 000025) is read-only and seeded with the complete Contabilizei table (167 codes, description + Origem/Destino); every code is classified by its first digit (1/5 same state, 2/6 other state, 3/7 exterior)
- A farm can register extra CFOPs through the selector's "+" modal (`GET /nfe/cfop/form`, `POST /nfe/cfop`); catalog duplicates and same-farm duplicates are rejected with pt-br toasts, and farm CFOPs are visible only to that farm
- Options are ordered per farm: "Mais utilizados" (use count > 0, count DESC then code ASC) and "Todos os CFOPs" (code ASC); the farm default (`nfe_farm_config.default_cfop`, else `5101`) is always present and preselected
- Each row has its own search input filtering only that row's options by code and accent-insensitive description; the selected option is never hidden, and typing a full code selects it
- Registration success returns the `cfop-option` fragment and triggers a farm-ordered refresh of every selector, auto-selecting the new code in the row that opened the modal
- Use counters (`farm_cfop_use`) are bumped once per distinct item CFOP when an emission persists; preview, rascunho saves, and worker retries never count, and counter failures only log (no SEFAZ contact, no emission block)
- There is no standalone CFOP management page — the selector modal is the only UI, mirroring the vehicle/crop/field add-on pattern

### Rascunhos de NF-e
- Named, farm-scoped reusable starting points (internally `DetachedProfile`; the word "draft" is reserved for `detached_nfe_invoice.status='draft'`)
- Store an optional registered recipient, ordered item defaults (product fields, per-item CFOP, exact unit price), operation/freight defaults, and optional invoice-level tax overrides/rates
- Quantity, gross weight, vehicle, invoice number/series, environment, signing and SEFAZ data are never stored
- Applying a rascunho fills the Emitir form; every prefilled value stays editable and is not written back unless the user explicitly saves/updates the rascunho. Preview/confirmation never reload the rascunho
- Unset rascunho tax values inherit the farm configuration; an explicit zero rate is preserved as zero
- Saving a rascunho from the inline-recipient branch explicitly creates a farm contact and links it; a plain preview creates no person
- Management operations never allocate an NF-e number, sign XML, persist an invoice, or call SEFAZ

### NF-e draft status label
- The internal status string `'draft'` is unchanged, but every user-facing NF-e surface (list rows for romaneio and avulsa invoices, the existing-NF-e modal) displays **"Não enviada"**
- The word "Rascunho" in the NF-e area now refers exclusively to the reusable rascunho feature; the romaneio "Rascunho de Entrada"/"Rascunho de Saída" wording is untouched

### Tax Configuration
- Inherits farm's default tax rates
- Per-invoice overrides for all tax types (ICMS, PIS, COFINS, IBS, CBS)
- CST codes configurable per invoice

### SEFAZ Integration
- Full emission flow: build → sign → send → handle response
- SVC contingency support (automatic fallback)
- Status polling via worker
- Cancellation support

### DANFE Download (`GET /nfe/avulsa/download/danfe/:accessKey`)
- Mirrors the departure-based `downloadNFeDANFE` (deliberate duplication, no shared helper)
- Lookup by access key → farm ownership check (`detached_nfe_invoice.farm_id`) → status gate (`authorized`/`cancelled` only) → XML selection (`xml_authorized` → `xml_signed`) → `ParseDANFEData` → `Generate`/`GenerateCancelled` ("NF-e CANCELADA" banner)
- Protocol fallback: when the stored XML has no `<protNFe>` (invoices authorized synchronously, where the `<nfeProc>` wrapper is only built by the retry worker), the protocol comes from the `protocol` DB column so DANFE Campo 2 is filled
- Unit tests: `TestDanfeStatusAllowed`, `TestDetachedDANFEXML` (`router/nfe_router/detached_router_test.go`)

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
ok  	armazenda/router/nfe_router
```

Focused E2E (`cd test/e2e && bun run test --grep "Rascunho de NF-e"`):
13 passed, 2 skipped (the beforeunload guard test is Chromium-only), 0 failed.

Full `make test`: 92 passed, 2 skipped, 5 failed. The 5 failures are pre-existing
WebKit entry/departure dialog timeouts that reproduce on the unmodified `HEAD`;
no Rascunho de NF-e test failed.

Tests cover:
- Struct initialization and field access
- Tax rate handling (nil vs explicit zero)
- Multi-item scenarios with per-item CFOPs and exact prices
- Recipient variations and inline-contact save
- Rascunho lifecycle, farm isolation, no-issuance side effects
- `'draft'` status rendering as "Não enviada"
- Migration data preservation (manual scratch-DB verification)

## Files Created/Modified

### New Files
- `model/armazenda_database/migrations/000021_detached_nfe.sql`
- `model/armazenda_database/migrations/000022_detached_nfe_item_cfop.sql`
- `model/armazenda_database/migrations/000023_detached_nfe_operation_profiles.sql`
- `model/armazenda_database/migrations/000024_person_ie_unique_partial.sql`
- `model/armazenda_database/migrations/000025_cfop_catalog.sql`
- `entity/public/cfop.go`
- `model/cfop_model/model.go`
- `model/cfop_model/model_test.go`
- `model/cfop_model/cfop_seed_test.go`
- `view/cfop/view.go`
- `view/cfop/view_test.go`
- `model/nfe_model/detached_model.go`
- `model/nfe_model/detached_model_test.go`
- `model/nfe_model/detached_profile_model.go`
- `model/nfe_model/detached_profile_model_test.go`
- `service/nfe_service/detached_service.go`
- `service/nfe_service/detached_service_test.go`
- `service/nfe_service/detached_preview_test.go`
- `service/nfe_service/detached_profile_service.go`
- `service/nfe_service/detached_profile_service_test.go`
- `router/nfe_router/detached_router.go`
- `router/nfe_router/detached_profile_router.go`
- `router/nfe_router/detached_router_test.go`
- `router/nfe_router/nfe_templates_test.go`
- `router/nfe_router/cfop_router.go`
- `router/nfe_router/cfop_router_test.go`
- `service/nfe_service/cfop_service.go`
- `service/nfe_service/cfop_service_test.go`
- `assets/js/cfopSelector.js`
- `assets/js/cfopDialog.js`
- `templates/pages/nfe-emit.html`
- `templates/pages/nfe-rascunhos.html`
- `templates/nfe/cfop-selector.html`
- `templates/nfe/cfop-options.html`
- `templates/nfe/cfop-option.html`
- `templates/nfe/nfe-cfop-form.html`
- `templates/nfe/nfe-detached-preview.html`
- `templates/nfe/nfe-detached-result.html`
- `templates/nfe/nfe-rascunho-table.html`
- `templates/nfe/nfe-rascunho-list-item.html`
- `templates/nfe/nfe-rascunho-form.html`
- `test/e2e/tests/detached_nfe_rascunho.spec.js`
- `test/e2e/tests/detached_cfop_selector.spec.js`

### Modified Files
- `model/person_model/model.go` - Added CreatePersonForDetachedNFe
- `service/nfe_service/worker.go` - Added detached invoice processing
- `router/nfe_router/router.go` - Integrated detached routes, updated list handler
- `templates/layout/menu.html` - Enabled "Emitir" and added "Rascunhos" links
- `templates/nfe/nfe-list.html` - Added type column, unified list
- `templates/nfe/nfe-list-item.html` - Updated to accept Invoice/Type dict; draft label "Não enviada"
- `templates/nfe/nfe-existing-modal.html` - draft label "Não enviada"
- `test/e2e/fixtures/test-user.sql` - Added NF-e farm config and farm address
- `main.go` - Registered the `ptrString` template helper; initialized `cfop_model`

## Usage Flow

1. User navigates to NF-e → Emitir (or NF-e → Rascunhos and clicks "Gerar NF-e a partir do rascunho")
2. Selects existing recipient or fills new recipient form
3. Adds one or more items with product details; each item has its own required CFOP, chosen in a searchable selector (farm-ordered; "+" registers a farm-tied CFOP when needed)
4. Configures nature operation (invoice-level) and optional complementary info
5. Optionally overrides tax rates
6. Clicks "Gerar Pré-visualização"
7. System validates the same way emission would and renders a non-fiscal DANFE
   preview — no number allocated, no person created, nothing persisted, no
   SEFAZ call. Every form field (including each item's CFOP) travels with the
   preview as a hidden input.
8. User confirms with "Confirmar e Emitir" (or clicks "Voltar" to adjust the
   form, whose values are preserved)
9. System allocates the number, signs, persists, and sends to SEFAZ
10. On success, the result panel offers the signed XML download and the invoice
    appears in the unified list with the "Avulsa" badge; `'draft'` rows show
    "Não enviada"
11. User can download XML, view status, or cancel if authorized

### Rascunho flow

1. Fill the Emitir form (or open an existing rascunho through the panel)
2. Click "Salvar como rascunho", name it, and confirm; if the recipient was
   inline, it is explicitly saved as a farm contact
3. On NF-e → Rascunhos: "Gerar NF-e a partir do rascunho" opens the Emitir
   page prefilled (quantity/weight stay empty, everything editable); "Editar Rascunho"
   updates it through an explicit action; "Excluir Rascunho" asks for confirmation
4. Editing the form after applying a rascunho never writes back to it — save/update
   is always an explicit action

## Future Enhancements

1. Store full InvoiceInput for complete SVC rebuild capability
2. Add filtering/search in unified list
3. Add export/reporting features for detached invoices
