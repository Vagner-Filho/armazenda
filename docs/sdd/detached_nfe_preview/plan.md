# plan:detached_nfe_preview

Implementation plan for the detached NF-e preview step. Spec: [sdd.md](./sdd.md).

Sibling SDD: `docs/sdd/detached_nfe_ui_fields/` (form fields + defaults). **Sequencing:** this SDD's confirm step depends on the field vocabulary introduced there (rates, CSTs, `modFrete`, `recipientEmail`, `farmProductId`, item reindex fix). Land its Phases 1–4 first; the confirm/result work of its Phase 5 is **absorbed here** (Phase 4 below) to avoid building the XML-download affordance twice.

Reference implementations to mirror:
- Router: `previewNFe` (`router/nfe_router/router.go:80–167`) — parse → preview service → fragment render with hidden-field pass-through.
- Service: `GeneratePreviewDANFE` (`service/nfe_service/service.go:625–745`) — `DANFEData` with `AccessKey ""`, `Numero 0`, non-fiscal generation.
- Template: `templates/nfe/nfe-preview.html` — banner, PDF iframe, hidden fields, Voltar/Confirmar.
- DANFE generator: `pkg/nfe/service/danfe.go` — `GeneratePreview` already loops over `data.Products` (line 670).

## Test strategy

- **Unit (service, DB-free)**: `service/nfe_service/detached_preview_test.go` (internal `nfe_service` package) covers `buildDetachedDANFEData` (multi-item products + summed totals, `Numero 0`/empty access key, CSOSN fallback, PIBS derivation), `buildDetachedItems` (fallback values, farm/user rates, overrides, IBS/CBS totals) and `detachedDocumentType`. Command: `go test ./service/nfe_service/...`.
- **Regression gate**: `go test ./service/nfe_service/... ./model/nfe_model/...` must stay green after the service refactor.
- **HTTP smoke (manual, not committed)**: test Postgres + `go run .`; `POST /nfe/emitir/preview` plus DB queries proving no number/person/invoice side effects. Recorded in `sdd.md` Verification Results.
- **Browser flow (focused Playwright, temporary)**: preview panel + hidden fields + Voltar + item-removal reindex. Recorded in `sdd.md`; not committed because the e2e fixture has no `nfe_farm_config`.
- **Completion sweep**: `go build ./...`, `gofmt -l` on changed files, `go test ./...`.

## Phase 1 — Service refactor: extract shared build preparation (behavior-preserving)

**File:** `service/nfe_service/detached_service.go`

1. Extract from `BuildDetachedInvoice` (lines 67–207) into:
   ```go
   func (s *NFeService) prepareDetachedBuildData(input DetachedInvoiceInput) (
       entity.InvoiceInput,            // serie/environment/naturezaOp/emitter/recipient/items/transport/payment/totalValue/infCpl (Numero=0, CNF=0, TpEmis normal)
       *entity_public.FarmConfig,      // needed downstream for cert/env/serie
       entity_public.Toast,
   )
   ```
   Contents: `GetFarmConfig`, `mapFarmToEmitterForDetached` + `validateEmitter`, recipient resolution, `validateRecipient`, effective CFOP, `buildDetachedItems`, naturezaOp/modFrete resolution, transport volumes from weights, infCpl, payment block.
2. **Dry-run recipient**: extract the inline-recipient → `entity.RecipientData` mapping (detached_service.go:296–384) into a helper used by both paths:
   - `mapInlineRecipientData(input DetachedRecipient, nfeModel) entity.RecipientData` — builds the same struct (type from document length, IE→`IndIEDest`, `S/N` numero default, municipality code resolution) **without** creating the person.
   - `resolveOrCreateRecipient` keeps creating, then delegates to the shared mapping for the non-ID path; existing-person path unchanged.
3. Rewire `BuildDetachedInvoice` = `prepareDetachedBuildData` + `AllocateDetachedNumber` + `generateRandomCNF` + `BuildAndSign` + persist + send. Zero behavior change.
4. Run `go test ./service/nfe_service/...` — detached suites must pass unmodified (regression gate before continuing).

Risk: number allocation order vs. `prepare` — keep allocation after preparation exactly as today so validation failures don't burn numbers.

## Phase 2 — Service: preview DANFE generator

**File:** `service/nfe_service/detached_service.go`

1. `GenerateDetachedPreviewDANFE(input DetachedInvoiceInput) ([]byte, entity_public.Toast)`:
   - `prepareDetachedBuildData(input)` → toast on failure.
   - Build `entity.DANFEData` mirroring `GeneratePreviewDANFE` (service.go:643–736) but:
     - `Products`: loop over **all** `input.Items`-derived `entity.ItemData` (per-item CST/PIS/COFINS/IBS/CBS from each item's imposto; per-item `PIBS` via `perItemRate(VIBSUF, VBC)`).
     - Invoice-level totals: **sum across items** (VBC, VICMS, VPIS, VCOFINS, VBCIBSCBS, VIBS, VCBS).
     - `AccessKey: ""`, `Numero: 0`; serie, naturezaOp, TpAmb, emitter/recipient/transport as prepared.
   - `service.NewDANFEGenerator().GeneratePreview(data)` → pdfBytes.
2. No DB access in this function beyond what `prepare` does (reads only).

## Phase 3 — Router: preview endpoint

**Files:** `router/nfe_router/detached_router.go`, route registration (`UseDetachedNFeRoutes`)

1. New handler `previewDetachedNFe`:
   - `parseDetachedRecipient` / `parseDetachedItems` / `parseUserTaxRates` / `parseInvoiceOverrides` (identical to `buildDetachedNFe` — extract a tiny shared `parseDetachedForm(c)` returning `(recipient, items, cfop, rates, overrides, toast)` so preview and confirm can't drift).
   - Build `nfe_service.DetachedInvoiceInput` (farmID from session cookie; `vehicleId` stays unsupported per out-of-scope).
   - `svc.GenerateDetachedPreviewDANFE(input)` → on failure: `HX-Trigger` toast + status (current error pattern).
   - Render `"nfe-detached-preview"` with:
     - `PDFBase64`, `Serie`, `Environment` (for display), and **every parsed field** for hidden inputs: recipient fields (branch on `recipient.PersonID` vs inline), `itemCount`, per-item loop data (including `FarmProductID`), cfop, rates via `rateDisplayString`, all overrides via `safeFormValue`.
2. Register: `router.POST("/emitir/preview", previewDetachedNFe)`.

## Phase 4 — Template: preview fragment + form retarget + confirm result

**Files:** `templates/pages/nfe-emit.html`, `templates/nfe/nfe-detached-preview.html` (new), `router/nfe_router/detached_router.go` (`buildDetachedNFe` response)

1. Main form changes:
   - `hx-post="/nfe/emitir/preview"`, `hx-target="#nfe-detached-preview"` (add an empty `<div id="nfe-detached-preview">` after the form), `hx-swap="innerHTML"`.
   - Submit button label → "Gerar Pré-visualização" (icon `mdi:eye` or `mdi:receipt-text`).
2. `nfe-detached-preview.html` block:
   - Banner + PDF iframe (same markup as `nfe-preview.html:3–9`).
   - Compact summary grid (recipient name, total value, item count) for context.
   - Hidden inputs per sdd.md Functional 3 (items rendered with `{{ range .Items }}` and explicit `{{ .Index }}`).
   - "Voltar" (`type="button"`, JS/hx: clear `#nfe-detached-preview` innerHTML, scroll to form) and "Confirmar e Emitir" (`hx-post="/nfe/emitir/build"`, `hx-target="#nfe-detached-preview"`).
   - Spinner indicator for the confirm call.
3. `buildDetachedNFe` success response: replace `c.String(200, signedXML)` with a result fragment (success toast + "Baixar XML Assinado" anchor to `/nfe/avulsa/download/xml/:accessKey`). Access key must be returned by the service — extend `BuildDetachedInvoice`'s success return from `(string, Toast)` to `(string, string /*accessKey*/, Toast)` (call sites: router only; update tests).
   - This **supersedes** Phase 5 of `detached_nfe_ui_fields`; mark that phase as absorbed in that plan.
4. Failure responses unchanged.

## Phase 5 — Tests & docs

1. Unit:
   - `service/nfe_service/detached_service_test.go`: `prepareDetachedBuildData` validation parity (config/recipient/item failures); `GenerateDetachedPreviewDANFE` multi-item totals + Numero 0/empty access key; dry-run mapping (no person fields mutated unexpectedly).
   - Existing detached suites stay untouched-green (refactor gate, AC 11).
2. E2E (optional, `test/e2e`): preview → confirm happy path; preview-not-consuming-number; voltar-preserves-input.
3. Commands: `gofmt -w`, `go build ./...`, `go test ./service/nfe_service/... ./model/nfe_model/...`, then `go test ./...`.
4. Update `docs/DETACHED_NFE_IMPLEMENTATION.md` usage flow (steps now include preview + confirm) and mark SDD `Status: Done` when ACs pass.

## Out of scope (explicit)

- Vehicle selection (`vehicleId`) — backend TODO (detached_service.go:137–140).
- Payment conditions — service hardcodes `IndPag 1` / `TPag "90"`.
- Volume/frete-value editing — volumes auto-derived as "Granel".
- DANFE final (fiscal) rendering of the signed XML — unchanged behavior; preview is always non-fiscal.
- Editing data **inside** the preview (Vincenzina-style inline edits) — Voltar → edit form → re-preview covers it.

## Phase 6 — Discovered prerequisites (added during implementation)

1. **Emitter data for detached invoices** (`service/nfe_service/detached_service.go`, `service/nfe_service/service.go`): `mapFarmToEmitterForDetached` produced an emitter with empty name/address/IBGE code, so `validateEmitter` rejected **every** detached emission (preview included) with "Configuração de NF-e incompleta". It now delegates to `mapFarmToEmitter` (identical to romaneio invoices) using `farm_config_model.GetFarmConfig`; `mapFarmToEmitter` gained a nil-farm guard. Verified via HTTP smoke (see `sdd.md` Verification Results).
2. **Item reindex fix** (prerequisite from `detached_nfe_ui_fields` Side finding 3): implemented in `templates/pages/nfe-emit.html` — delegated removal plus `reindexItems()`, which renumbers `items[0..n-1]` and syncs `itemCount`. Verified via the focused Playwright smoke.

## Risks

- **Hidden-field drift between preview and confirm**: field names must match `parseDetachedRecipient`/`parseDetachedItems`/`parseUserTaxRates`/`parseInvoiceOverrides` exactly; a rename breaks stateless confirm. Mitigation: shared `parseDetachedForm` + acceptance criterion 6.
- **Refactor side-effect leakage**: moving code out of `BuildDetachedInvoice` must not reorder allocation/signing/persistence; compile-time types won't catch reordering — review diff carefully (AC 2/4).
- **PDF fidelity**: `GeneratePreview` renders `DANFEData`, not the signed XML — any field the builder derives (e.g. CNF, dates) differs by design; that is acceptable for a preview, but per-item tax values must come from the same `buildDetachedItems` output to guarantee AC 7 totals match the emitted XML.
- **Cross-farm recipient lookup** (pre-existing): `GetFullPersonById` is not farm-scoped; out of scope here, noted for a future hardening pass.
- **CSP nonce** on all new inline script blocks.
