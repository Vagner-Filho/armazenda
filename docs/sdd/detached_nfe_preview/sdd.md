# sdd:detached_nfe_preview

## Status
<!-- One of: Backlog | WIP | Done -->
Done

## Goal
Add a **preview step** to the detached NF-e (NF-e Avulsa) emission, matching the departure-based (romaneio) flow: the user edits the form, clicks "Gerar Pré-visualização", sees a non-fiscal DANFE PDF ("Documento de Pré-visualização — Sem Valor Fiscal") built from the exact data they entered, and only on explicit "Confirmar e Emitir" does the system allocate a number, sign, persist, and send to SEFAZ. Today `POST /nfe/emitir/build` does all of that atomically with zero visual feedback — the user never sees the document before it is signed and transmitted.

## Background (how the romaneio flow does it — the pattern to mirror)

```
GET  /nfe/modal/:departureID   → nfe-emit-modal.html (edit form)
POST /nfe/preview/:departureId → previewNFe (router.go:80)
      parses unitPrice/cfop/rates/overrides → svc.GeneratePreviewDANFE(...)
      → renders nfe-preview.html: PDF iframe + hidden fields + Voltar/Confirmar
POST /nfe/build/:departureID   → full emission with the same parsed fields
```

Key mechanics of the romaneio preview (`service/nfe_service/service.go:625` `GeneratePreviewDANFE`):
- Builds invoice data via `prepareInvoiceBuildData` but produces only a `entity.DANFEData` with `AccessKey: ""`, `Numero: 0` — **no number allocation, no signing, no persistence, no SEFAZ call**.
- The preview fragment carries **all** form inputs as hidden fields so confirm is stateless (no server-side session of the input).

### Current detached flow (what this SDD changes)

`POST /nfe/emitir/build` → `buildDetachedNFe` (detached_router.go:45) → `BuildDetachedInvoice` (detached_service.go:67) is monolithic: farm-config fetch → **person creation** (`resolveOrCreateRecipient` writes when recipient is inline, detached_service.go:296–322) → validations → items → **number allocation** (`AllocateDetachedNumber`, line 188) → sign → persist → send. There is no way to see the document before all side effects happen.

### Side effects the preview must NOT have

| Side effect | Where today | Preview behavior |
|---|---|---|
| Number allocation | `AllocateDetachedNumber` (detached_service.go:188) | forbidden — Numero stays 0, AccessKey empty |
| Person creation (inline recipient) | `CreatePersonForDetachedNFe` (detached_service.go:318) | forbidden — dry-run mapping only |
| Invoice persistence | `CreateDetachedInvoice` + XML updates (lines 236–250) | forbidden |
| XML signing | `BuildAndSign` (line 218) | forbidden — PDF built from `DANFEData`, not signed XML |

### Related finding (dependency for confirm correctness)

**Item removal corrupts indexes** (pre-existing in `templates/pages/nfe-emit.html`): after adding items, removing a middle row leaves index gaps (e.g. rows `0,1,3` with `itemCount=3`); `parseDetachedItems` loops `0..itemCount-1`, so it reads a nonexistent row and misses the last one — submission fails with "Quantidade inválida" or silently drops items. This directly breaks the confirm step of the preview flow, so the reindex fix (see `detached_nfe_ui_fields` Side finding 3) is a prerequisite.

### Related finding (dependency for preview viability — fixed here)

**Detached emitter mapping was incomplete** (`mapFarmToEmitterForDetached`, pre-existing in `service/nfe_service/detached_service.go`): it built `entity.EmitterData` with name, address fields and IBGE municipality code always empty, so `validateEmitter` rejected every detached emission with "Configuração de NF-e incompleta" before signing. The preview could therefore never reach DANFE generation, and the Background premise above ("today `/nfe/emitir/build` does all of that atomically") was not actually true. Fixed by delegating the detached mapping to the same `mapFarmToEmitter` used by romaneio invoices, fed with `farm_config_model.GetFarmConfig` (read-only; used by both preview and emission). Covered by the HTTP smoke check below.

## Requirements

### Functional

1. **Preview endpoint** `POST /nfe/emitir/preview` (`previewDetachedNFe` in `router/nfe_router/detached_router.go`):
   - Reuses the existing parsers: `parseDetachedRecipient`, `parseDetachedItems`, `parseUserTaxRates`, `parseInvoiceOverrides` (same package) — the previewed data is exactly what confirm will re-parse.
   - Runs full validation **with the same toasts as emission**: emitter config validation, `validateRecipient`, item validation — so users fix errors before emitting.
   - Returns the `nfe-detached-preview` HTML fragment (HTTP 200).
2. **Service**: new `GenerateDetachedPreviewDANFE(input nfe_service.DetachedInvoiceInput) ([]byte, entity_public.Toast)` in `service/nfe_service/detached_service.go`, built on an extracted `prepareDetachedBuildData` (see plan Phase 1). It must:
   - Map **all items** to `entity.DANFEData.Products` (the generator already loops over the slice — `pkg/nfe/service/danfe.go:670`; the romaneio function maps only `Items[0]`, which is departure-specific).
   - Sum invoice-level totals (VBC, VICMS, VPIS, VCOFINS, VBCIBSCBS, VIBS, VCBS) across **all items** — not item[0]'s imposto block as the romaneio version does.
   - Set `AccessKey: ""`, `Numero: 0`, serie, naturezaOp, tpAmb from farm config, emission date = now.
   - Generate via `service.NewDANFEGenerator().GeneratePreview(data)` (non-fiscal variant already implemented).
3. **Preview fragment** `templates/nfe/nfe-detached-preview.html` (new block, mirroring `templates/nfe/nfe-preview.html`):
   - Banner "Documento de Pré-visualização — Sem Valor Fiscal".
   - PDF iframe (`data:application/pdf;base64,{{ .PDFBase64 }}`).
   - Hidden inputs carrying **every** field the form collected, so confirm is stateless:
     - recipient: `recipientPersonId` **or** the full inline set (`recipientName`…`recipientEmail`);
     - `itemCount` + `items[i].{farmProductId, productName, ncm, cest, quantity, grossWeight, unitPrice, unit}` — rendered by looping over the parsed items (multi-item-safe);
     - `cfop`, `naturezaOp`, `infCpl`, `modFrete`;
     - rates `icmsRate`…`cbsRate` (empty string = "use default" semantics, matching nil `TaxRates`);
     - CST overrides `icmsCST`…`cClassTrib` (empty = not provided).
   - "Voltar" button: clears the preview region and returns the user to the (still populated) form.
   - "Confirmar e Emitir" button posting to the **existing** `POST /nfe/emitir/build`.
4. **Form retarget**: the main form posts to `/nfe/emitir/preview` with `hx-target="#nfe-detached-preview"` (a dedicated region below the form) instead of swapping into nothing. Consequence: the form stays in the DOM — "Voltar" does not re-render the form, so **user input is never lost** (unlike the romaneio modal, whose Voltar can afford a server re-render because defaults come from the departure).
5. **Confirm-step result**: on success, `/nfe/emitir/build` returns a result fragment (success toast + link "Baixar XML Assinado" → `GET /nfe/avulsa/download/xml/:accessKey`) instead of the raw XML body with `hx-swap="none"`. This absorbs Phase 5 of `detached_nfe_ui_fields` (see plan cross-reference).
6. Failure paths (preview or confirm): current behavior preserved — `HX-Trigger` toast + 4xx/5xx status, no fragment swap.

### Non-Functional

- **Single-source logic**: item building, rate resolution (`MergeRates`), CST chains, and validations live in the extracted `prepareDetachedBuildData` used by **both** preview and emission — no forked copies (project rule: shared logic is never duplicated).
- **Behavior-preserving refactor**: `BuildDetachedInvoice`'s externally visible behavior (XML, toasts, DB rows) unchanged; existing `service/nfe_service` detached tests must stay green.
- All new user-facing strings in **pt-br**; inline scripts carry `{{ .CSPNonce }}`.
- Glass-panel styling per `.agents/skills/ui-design`.
- No client-side calculation; the PDF and totals come from the server.

## Acceptance Criteria

1. Submitting the detached form shows the preview: non-fiscal DANFE PDF with "Sem Valor Fiscal" banner, numero 0, empty access key.
2. A preview consumes **no** NF-e number: after N previews, the next emission's number reflects only N emissions (verifiable via the sequence/counter or the allocated `nNF`).
3. A preview with an **inline new recipient** creates **no** person row (`person` table unchanged after preview).
4. Preview performs no persistence: `detached_nfe_invoice` gains no row and no XML columns are written.
5. Validation parity: incomplete farm config / recipient / items produce the same warning toasts at preview time that emission would produce.
6. All form fields survive the round trip: confirm after preview emits with the same recipient, items, rates, CSTs, modFrete and infCpl shown in the preview (hidden-field fidelity — including the multi-item loop and empty-string rates falling back to farm config).
7. Multi-item rendering: 3 items appear as 3 product rows in the previewed DANFE with correct per-row and summed totals.
8. "Voltar" clears the preview and the form still holds all previously entered values.
9. Confirm success shows a success toast and a working XML download link; confirm failure keeps today's toast-only behavior.
10. Item removal before preview: add 3 items, remove the middle one, submit — the preview (and subsequent emission) contains exactly the 2 remaining items (reindex fix in place).
11. Refactor regression: `go test ./service/nfe_service/... ./model/nfe_model/...` pass unchanged after `prepareDetachedBuildData` extraction.
12. `go build ./...` and `gofmt -l` (no diffs) pass.

## Verification Results

Executed on 2026-10-02 against this change set.

| Check | Command / method | Result |
|---|---|---|
| Build | `go build ./...` | PASS |
| Formatting | `gofmt -l` on changed Go files (`service/nfe_service/service.go`, `service/nfe_service/detached_service.go`, `service/nfe_service/detached_preview_test.go`, `router/nfe_router/detached_router.go`) | PASS (no diffs). The repo has pre-existing `gofmt -l` diffs in untouched files (e.g. `entity/public/farm.go`, `pkg/nfe/entity/danfe.go`), confirmed present at `HEAD` |
| Focused tests (AC 11) | `go test ./service/nfe_service/... ./model/nfe_model/...` | PASS |
| Full Go suite | `go test ./...` | PASS |
| Template parse/render | Temporary guard test parsing all templates and rendering `nfe-detached-preview` (inline + existing-person branches) and `nfe-detached-result` | PASS; hidden-input vocabulary and download link verified |
| AC 1 / 7 — preview PDF | HTTP smoke against test Postgres + `go run .`: `POST /nfe/emitir/preview` (inline recipient, 2 items, rates untouched) | 200; response contains the "Documento de Pré-visualização — Sem Valor Fiscal" banner, `data:application/pdf;base64,` iframe, summed total 140000, and `items[1]` hidden fields |
| AC 2 — no number consumption | Two previews, then `SELECT * FROM nfe_numbering` | `last_number` unchanged (1, allocated only by the earlier confirm attempt) |
| AC 3 — no person creation | `SELECT count(*) FROM person` after inline preview | 0 after preview; 1 after the confirm request |
| AC 4 — no persistence | `SELECT count(*) FROM detached_nfe_invoice` after previews | 0 |
| AC 5 — validation parity | Preview with inline recipient missing address fields | 400 + `HX-Trigger` toast listing the same `validateRecipient` errors emission produces |
| AC 6 / 9 (failure) — confirm re-parses hidden fields | `POST /nfe/emitir/build` replaying all 43 hidden inputs extracted from the preview response | Parsed and validated all fields, allocated a number, reached certificate signing; dummy test certificate yields the expected 500 + "Falha ao construir e assinar NF-e" toast, and nothing was persisted |
| AC 8 / 10 — browser flow | Focused Playwright smoke (temporary config/spec, removed after run): add 3 items, remove middle, assert `itemCount=2` and survivor reindexed; submit preview (200, panel + PDF iframe + hidden `items[1]`); click Voltar (preview cleared, form values preserved) | PASS |
| AC 9 (success leg) | Confirm with an authorized SEFAZ response | NOT RUN — requires a real A1 certificate and SEFAZ availability; the success fragment (`nfe-detached-result` + access-key download link) was verified by template render, and its handler wiring is unchanged from the sibling `detached_nfe_ui_fields` work |

Environment note: the HTTP/browser smoke required a seeded `nfe_farm_config` and farm address because the e2e fixture (`test/e2e/fixtures/test-user.sql`) contains no NF-e configuration. The temporary Playwright config/spec and the smoke app instance were removed afterwards; the project dev server was not touched.
