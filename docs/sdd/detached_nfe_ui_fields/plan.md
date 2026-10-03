# plan:detached_nfe_ui_fields

Implementation plan for completing the detached NF-e emission form UI. Spec: [sdd.md](./sdd.md).

Follows the Layered architecture (Entity → Model → Service → Router). Backend input contracts already exist — most of this is Router/template work; the service layer is touched only where noted.

Reference implementations to copy from:
- `templates/nfe/nfe-emit-modal.html` (lines 114–258): Transporte select, Tributação fieldset (CST grid + "Usar taxa padrão" + disabled rate inputs), default values via `{{ .Default* }}`.
- `router/nfe_router/router.go` (lines 590–674): farm-config defaults resolution + `percentDisplay`/`safePtrString` for template values.
- `templates/nfe/nfe-preview.html` (lines 17–38): exact field-name vocabulary expected by the backend.

## Phase 1 — Router: pass farm defaults + farm products to the page

**File:** `router/nfe_router/detached_router.go` (`getDetachedNFePage`)

1. Load `farmNFeConfig` (already fetched) and mirror the romaneio modal handler's resolution (router.go:600–630):
   - Rates: `ICMSRate`, `PISRate`, `COFINSRate`, `IBSRate`, `CBSRate` → template via `percentDisplay` (helper already in package `nfe_router`).
   - CSTs/cClassTrib: `DefaultICMSCST`, `DefaultPISCST`, `DefaultCOFINSCST`, `DefaultIBSCST`, `DefaultCBSCST`, `DefaultCClassTrib` → `safePtrString`.
   - `DefaultModFrete`, `DefaultNaturezaOp` (derive with `defaults.NaturezaOpForCFOP` when unset), `DefaultCEST`.
2. Fetch farm products: `farm_product_model.GetFarmProductModel().GetFarmProductsByFarm(farmID)` → pass as `FarmProducts` (Id, Name, NCM).
3. No service/model changes: `GetFarmConfig` and `GetFarmProductsByFarm` already exist.

Risk: none (read-only). If farm config is nil the form renders with empty defaults (same non-fatal behavior as the romaneio modal, router.go:595–597).

## Phase 2 — Template: Tributação fieldset

**File:** `templates/pages/nfe-emit.html` ("Dados da NF-e" fieldset area)

1. Add a "Tributação" fieldset cloned from `nfe-emit-modal.html:134–258`:
   - Grid with `icmsCST` (placeholder 00), `pisCST` (01), `cofinsCST` (01), `ibsCST` (000, maxlength 3), `cbsCST` (000, maxlength 3), `cClassTrib` (000001, maxlength 10) — pre-filled with `{{ .Default* }}`.
   - "Usar taxa padrão" checkbox `useDefaultTaxRates` (checked) + rate inputs `icmsRate`, `pisRate`, `cofinsRate`, `ibsRate`, `cbsRate` with `{{ .Default* }}` values and `disabled` when the checkbox is checked.
   - IBS/CBS labels get the "— Reforma Tributária" hint span.
2. Inline script (nonce-carrying) toggles `disabled` on the five rate inputs when `useDefaultTaxRates` changes — same behavior as `nfe-emit-modal.html:296+` (mirror that logic).
   - Key semantic: disabled inputs are **not submitted** → `parseUserTaxRates` yields nil → `MergeRates` falls back to farm config. Unchecked + filled → authoritative override (including explicit zeros).
3. Note: "Dados da NF-e" fieldset already has `cfop`, `naturezaOp`, `infCpl` — keep those; `productDesc`/`ncm`/`cest`/`unit` invoice-level overrides stay **omitted** (per-item fields cover them).

## Phase 3 — Template: Transporte + recipient fixes

**File:** `templates/pages/nfe-emit.html`

1. Add "Transporte" fieldset with `modFrete` select (options 0/1/2/3/4/9, pre-selected via `{{ .DefaultModFrete }}`), copied from `nfe-emit-modal.html:114–133`.
2. Add `recipientEmail` input to `#new-recipient-section` (optional, `type="email"`).
3. Fix hidden-`required` bug: the recipient-type radio handler (inline script) must toggle the `required` attribute on `recipientName`/`recipientDocument` in sync with the section visibility:
   - "existing" selected → `required` removed from new-recipient inputs.
   - "new" selected → `required` restored.
   - Keep the Tailwind `hidden` class toggle as is.

## Phase 4 — Template: farm product selector per item

**Files:** `templates/pages/nfe-emit.html` (item row template + clone script)

1. Add a `select` `items[0].farmProductId` ("Produto cadastrado (opcional)") per item row, options from `{{ range .FarmProducts }}` with `value="{{ .Id }}"` and label `Name`.
2. On selection change (JS, one delegated listener on `#items-container` so cloned rows work): pre-fill the row's `productName` + `ncm` from the selected product's data embedded in option attributes (`data-name`, `data-ncm`); clearing (`value=""`) empties them back to manual entry.
3. The clone script already rewrites `[N]` indexes for inputs **and selects** (detached script queries `'input, select'`) — verify the new select's name gets reindexed.
4. Router already accepts `items[i].farmProductId` (detached_router.go:201–207); NCM fallback chain in `buildDetachedItems` (detached_service.go:449) still applies when left empty.

## Phase 5 — Post-emission feedback (XML download)

> **Superseded by `docs/sdd/detached_nfe_preview/`** — with the preview flow, the "Confirmar e Emitir" step (`/nfe/emitir/build`) response (success toast + XML download link) is specified in that SDD's plan Phase 4, including the `BuildDetachedInvoice` access-key return change. Implement it there, not here.

**Files:** `templates/pages/nfe-emit.html`, `router/nfe_router/detached_router.go` (`buildDetachedNFe`)

1. Router: on success, instead of returning the raw XML body with `hx-swap="none"`, return a small HTML fragment (or `HX-Trigger` event + fragment) containing the success message and a download link (`Content-Disposition`-style anchor pointing at `GET /nfe/avulsa/download/xml/:accessKey`, which already serves the signed/authorized XML).
   - Constraint: keep the toast via `HX-Trigger` as today; the fragment target must be inside the form panel (`hx-target="this"` already set, current swap is `none` → change to a named target for the result fragment only).
   - Access key must come from `handleDetachedSefazResponse`/persisted record — if the service returns only the XML today, extend `BuildDetachedInvoice`'s success return to also surface the access key (adjust router accordingly). This is the only service-layer change in the plan.
2. Failure paths keep current behavior (toast + status code).

## Phase 6 — Validation & docs

1. `gofmt -w` on touched Go files; `go build ./...`.
2. Regression runs:
   - `go test ./service/nfe_service/... ./model/nfe_model/...` (existing detached suites must stay green).
   - `go test ./...` for the full sweep.
3. Optional e2e (`test/e2e`): happy path with "existing recipient" + default rates; second scenario with overridden rates asserting the persisted row in `detached_nfe_tax_rates`.
4. Update `docs/DETACHED_NFE_IMPLEMENTATION.md` § What's NOT Implemented: remove "Farm Product Integration" (now implemented); note the remaining gaps (vehicle, payment forms, volumes).
5. SDD Status → `Done` when all `sdd.md` acceptance criteria pass.

## Out of scope (explicit)

- **`vehicleId` UI** — router parses it but the service has `TODO: fetch vehicle plate` (detached_service.go:137–140); add the field only after the backend completes vehicle lookup.
- **Payment conditions** — service hardcodes `IndPag: 1`/`TPag: "90"`; needs a `DetachedInvoiceInput` extension first.
- **Volume/frete value details** — volumes are auto-derived ("Granel", single entry); needs transport-data support in the service.
- **DANFE generation** — endpoint still 501; separate feature.
- **SVC full rebuild for drafts** — separate feature (needs stored `InvoiceInput`).

## Risks

- **Checkbox semantics**: if the checkbox were replaced by always-enabled inputs, untouched defaults would be submitted as explicit overrides (locking the farm config out). Preserve the disabled-unchecked-check pattern.
- **Field-name drift**: backend parses exact names (`cbsRate`, `cClassTrib`, `recipientEmail`, `items[i].farmProductId`); any rename silently breaks the flow. Validate names against `parseUserTaxRates`/`parseInvoiceOverrides`/`parseDetachedRecipient`/`parseDetachedItems` before merge.
- **CSP nonce**: new inline script blocks must include `{{ .CSPNonce }}` or they will be blocked by the CSP middleware.
- **Clone script coverage**: dynamically added fields (product select, email) must survive `cloneNode` reindexing; the product select's options rely on server-rendered `<option>` elements inside the cloned row (works as long as options are inside the row markup, not injected per row).
