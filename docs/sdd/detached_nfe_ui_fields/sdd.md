# sdd:detached_nfe_ui_fields

## Status
<!-- One of: Backlog | WIP | Done -->
Done

## Goal
Complete the detached NF-e (NF-e Avulsa) emission form (`templates/pages/nfe-emit.html`) so the UI delivers the control the feature promises. Today the backend (`router/nfe_router/detached_router.go` + `service/nfe_service/detached_service.go`) accepts per-emission tax rate overrides, CST/cClassTrib overrides, freight modality, recipient e-mail and farm-product selection — but the form sends none of them, so every emission silently falls back to farm-config defaults. This spec also covers two correctness/UX defects found in the current form (hidden `required` inputs blocking submission; no feedback/download after a successful emission).

## Background (current state)

The form posts to `POST /nfe/emitir/build` (`buildDetachedNFe`), which parses fields the form does not contain:

| Field group | Parsed at | Form state | Effect of absence |
|---|---|---|---|
| `icmsRate` `pisRate` `cofinsRate` `ibsRate` `cbsRate` | `parseUserTaxRates` (router.go:779) | missing | `TaxRates` all-nil → `MergeRates` falls back to farm config → 2026 symbolic rates; user has no per-emission control |
| `icmsCST` `pisCST` `cofinsCST` `ibsCST` `cbsCST` `cClassTrib` | `parseInvoiceOverrides` (router.go:834) | missing | CSTs always farm-config defaults |
| `modFrete` | router.go:858 | missing | always farm-config default |
| `recipientEmail` | detached_router.go:168 | missing | e-mail never captured |
| `items[i].farmProductId` | detached_router.go:202 | missing | products must be typed manually |

`docs/DETACHED_NFE_IMPLEMENTATION.md` (lines 95–98) already claims "Per-invoice overrides for all tax types" and "CST codes configurable per invoice" — the UI is behind its own docs.

### Side findings (defects, in scope here)

1. **Hidden `required` inputs block submission.** `recipientName`/`recipientDocument` are `required` inside `#new-recipient-section`, which is toggled only via the Tailwind `hidden` class (still in the DOM). If the user selects "Selecionar existente" and submits, Chrome blocks submit with "An invalid form control … is not focusable". The `required` attributes must be toggled together with visibility.
2. **No result affordance.** `buildDetachedNFe` returns the signed XML as a raw string with `hx-swap="none"`; the user sees only a toast and has no download link (the romaneio flow offers "Baixar XML Assinado" in `nfe-preview.html`). *(The result fragment is now specified in `docs/sdd/detached_nfe_preview/` — see its plan Phase 4.)*
3. **Item removal corrupts indexes.** After adding items, removing a row leaves index gaps (e.g. rows `0,1,3` with `itemCount=3`): `parseDetachedItems` loops `0..itemCount-1`, reads a nonexistent row and misses the last one — submission fails ("Quantidade inválida") or silently drops items. The clone/removal script must reindex remaining rows contiguously. This is a prerequisite for the preview flow's confirm step (`docs/sdd/detached_nfe_preview/`).

## Requirements

### Functional

1. **Tributação fieldset** in the detached form, mirroring the romaneio modal (`templates/nfe/nfe-emit-modal.html:134–258`):
   - CST inputs: `icmsCST`, `pisCST`, `cofinsCST`, `ibsCST`, `cbsCST` and `cClassTrib`, pre-filled with farm-config defaults.
   - "Usar taxa padrão" checkbox (checked by default) controlling five rate inputs `icmsRate`, `pisRate`, `cofinsRate`, `ibsRate`, `cbsRate`, pre-filled with farm-config defaults:
     - Checked → inputs `disabled` → not submitted → nil rates → farm-config/2026 fallback (matches `TaxRates` nil semantics).
     - Unchecked → inputs enabled and editable; submitted values are per-emission overrides (percent display, e.g. `17,00` → 0.17).
   - IBS/CBS fields labeled with the "Reforma Tributária" hint like the romaneio modal.
2. **Transporte fieldset**: `modFrete` select (0 CIF / 1 FOB / 2 terceiros / 3 remetente / 4 destinatário / 9 sem frete), pre-selected with the farm-config default.
3. **New-recipient section gains an `recipientEmail` field** (optional), parsed and persisted through `CreatePersonForDetachedNFe`.
4. **Farm product selector per item**: an optional select per item row populated with the farm's products (`model/farm_product_model.GetFarmProductsByFarm`, `entity_public.FarmProduct{Id, Name, NCM}`); selecting one posts `items[i].farmProductId` and auto-fills Descrição + NCM (FarmProduct carries no CEST — CEST stays manual). Clearing the selection restores free-text entry.
5. **Required-attribute correctness**: `required` on new-recipient inputs is added/removed in sync with section visibility, so "existing recipient" submissions are never blocked by hidden controls.
6. **Post-emission feedback**: on success the user gets a success toast **and** a download link for the signed XML (same affordance as the romaneio flow).
7. All new user-facing strings in **pt-br**; field naming must match exactly what the router parses (no renames).
8. **Item-row reindexing on removal** (Side finding 3): after any removal, remaining rows are renumbered contiguously (`items[0..n-1]`) and the `itemCount` hidden input matches the row count.

### Non-Functional

- Styling follows the project's glass-panel design language (`.agents/skills/ui-design`); reuse the romaneio modal's field classes/markup patterns.
- No new client-side calculation logic — totals/IBS/CBS remain computed server-side (`pkg/nfe`, `nfe_service`); the UI only collects inputs.
- Inline `<script>` must carry the `{{ .CSPNonce }}` nonce (page already does).
- Backward compatible: submitting the form with the "use default" checkbox untouched must produce XML identical in tax content to today's behavior (all-nil rates).
- XML free text continues to pass through `SanitizeSchemaString` server-side (already in `pkg/nfe/xml/builder.go`) — no change needed, do not bypass.

## Acceptance Criteria

1. With farm config present, the detached form renders the Tributação fieldset with all six CST inputs and five rate inputs pre-filled from the farm config (`getDetachedNFePage` passes the defaults).
2. Submitting with "Usar taxa padrão" checked produces the same tax values as before the change (nil `TaxRates`, farm fallback) — regression guard.
3. Submitting with unchecked defaults and e.g. ICMS 12 % / CBS 1,8 % results in those values persisted in `detached_nfe_tax_rates` and visible in the built XML (`pICMS`, `pCBS` ×100 semantics per NT 2025.002).
4. Overriding `cClassTrib`/CST fields reflects the new codes in the per-item `<IBSCBS>`/ICMS/PIS/COFINS groups.
5. `modFrete` submitted value appears in the emitted transport group; with no selection the farm default applies.
6. New-recipient emission with e-mail filled stores the e-mail on the created person record.
7. Selecting a farm product in an item row posts `items[i].farmProductId` and pre-fills Descrição/NCM; the emission succeeds end-to-end.
8. Choosing "Selecionar existente" and submitting does not trigger the browser "not focusable" validation error (required toggling verified).
9. After a successful emission the page shows a success toast and a working XML download link.
10. `go build ./...` and `gofmt -l` (no diffs) pass; existing suites `go test ./service/nfe_service/... ./model/nfe_model/...` still pass.
11. Add 3 items, remove the middle one, submit — the emission (or preview, per `detached_nfe_preview`) contains exactly the 2 remaining items with no lost or shifted products.
