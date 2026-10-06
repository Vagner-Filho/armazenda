# sdd:detached_nfe_operation_profiles

## Status
Done

## Goal
Let farm users save, manage, and reuse named **"Rascunhos de NF-e"** to start a detached NF-e (NF-e avulsa) with recurring data already filled in — a reusable transaction scenario, not a CFOP record. A rascunho includes a recipient, product lines and their unit prices, operation and freight defaults, and optional tax overrides, and every detached NF-e item carries its own CFOP. The wording and lifecycle mirror the existing "Rascunho de Entrada"/"Rascunho de Saída" so users land on familiar ground. To free the word "rascunho" from its invoice-status meaning inside the NF-e area, the NF-e `draft` status label becomes **"Não enviada"** (display-only). The change is isolated to detached NF-e; the departure-linked NF-e flow is unchanged.

## Naming

| Surface | Term |
|---|---|
| User-facing concept | **"Rascunho de NF-e"** — panel "Rascunhos", actions "Novo Rascunho", "Gerar NF-e a partir do rascunho", "Editar Rascunho", "Excluir Rascunho", "Salvar como rascunho" |
| Internal identifiers | English `profile` (`DetachedProfile`, `detached_profile_model.go`, `detached_profile_router.go`). `draft` is deliberately avoided in code: `detached_nfe_invoice.status='draft'` already drives `GetDraftDetachedInvoicesForRetry` / `processDetachedDraftInvoice` (`service/nfe_service/worker.go`), so a `DetachedDraft` type would read as "a draft-status invoice" |
| NF-e invoice status `draft` | UI label **"Não enviada"** (was "Rascunho"); the internal status string stays `'draft'` |
| Romaneio rascunhos | "Rascunho de Entrada" / "Rascunho de Saída" — wording, lifecycle, and UI stay untouched; they are the pattern to mirror |

## Requirements

### Functional

#### Rascunhos de NF-e

1. Rascunhos are scoped to a farm and available to users authorized for that farm. The name is required and unique, case-insensitively, among the farm's rascunhos. Users can create, list, edit, and **delete** rascunhos — deletion asks for confirmation, exactly like romaneio rascunhos (hard delete; there is no archive state).
2. A saved rascunho contains:
   - An optional recipient reference to a person registered for the same farm.
   - One or more ordered item defaults. Each item stores its optional farm-product reference and the product fields collected by the detached form (description, NCM, CEST, and unit), its **four-digit CFOP**, and its unit price as an exact decimal.
   - Optional operation defaults: `Natureza da operação` (`natOp`), freight modality, and complementary information.
   - Optional invoice-level tax overrides: ICMS/PIS/COFINS/IBS/CBS rates and the corresponding CST/CSOSN and `cClassTrib` fields currently exposed by the detached form.
3. A rascunho stores product fields and prices as rascunho values, not as live product-price lookups. Selecting a rascunho fills the form with those values; subsequent catalog or rascunho changes do not silently change an invoice already in progress or an issued invoice.
4. Quantity, gross weight, vehicle, invoice number/series, environment, signing data, authorization data, and SEFAZ state are not rascunho values. Quantity and gross weight remain inputs for each emission. Rascunho item unit prices are starting values and remain editable for the current emission.
5. A rascunho may reference a farm-scoped registered recipient. When saving a rascunho from the inline-recipient branch of the emission form, the user must explicitly save that recipient to the farm's contacts; the rascunho then references the resulting person ID. The save action may create that contact, but must not emit an invoice. Preview remains free of writes.
6. A rascunho can be saved from the populated detached form and can be edited without emitting an invoice. Editing form values after applying a rascunho does not update the saved rascunho; updating it requires an explicit action.
7. The management UI lives in its own page listing existing rascunhos and rendering a "Novo Rascunho" action, and per-rascunho actions "Gerar NF-e a partir do rascunho" (fills the emission form), "Editar Rascunho", and "Excluir Rascunho" (similar to row item buttons of `templates/entry/entry-content.html` and `templates/departure/departure-content.html`) with the same confirmation pattern (`data-delete-text` + `hx-trigger="confirmed"`). Applying a rascunho by clicking the button icon "Gerar NF-e a partir do rascunho" navigates the user to the Emitir page of the detached NF-e feature.
8. Field resolution is: explicitly configured rascunho values → farm NF-e configuration → existing system fallback. Unset rascunho tax values inherit farm defaults; an explicitly supplied zero rate is preserved as zero, not treated as unset. A rascunho is a starting point and is not reapplied over edits at preview or confirmation time.
9. Selecting no rascunho continues to support detached emission. Each item row is prefilled with the current detached/farm CFOP default as an editable item value; no invoice-level CFOP is submitted or persisted.

#### CFOP belongs to each detached item

10. The detached form collects a CFOP for every item. The server validates that each is exactly four ASCII digits and maps it to that item's `<det><prod><CFOP>` in the NF-e XML. Different items in one detached NF-e may have different CFOPs.
11. Remove CFOP from detached invoice-level input, model fields, persistence, and queries. Keep CFOP on `DetachedItemInput` and `DetachedInvoiceItem`, including rascunho item data and the existing `items_json` snapshot. This requirement applies **only** to detached NF-e; do not change `nfe_invoice`, departure-based NF-e behavior, or the farm default configuration used by that flow.
12. Add an idempotent migration that preserves each existing detached item's CFOP. Where a legacy item lacks a CFOP, backfill it from the legacy detached invoice-level CFOP before dropping only `detached_nfe_invoice.cfop`. Do not rewrite signed/authorized XML.
13. `Natureza da operação` remains invoice-level and is resolved from a current form value, then a selected rascunho value, then the farm default. If none exists and all items share one CFOP, retain the current derivation from that common CFOP. If an invoice has different item CFOPs and no explicit/farm `Natureza da operação`, preview asks the user to provide it instead of deriving it from an arbitrary first item.
14. Preview, confirmation, persistence, and detached pending-invoice retry all retain each item's CFOP. The preview-to-confirm hidden-field round trip includes `items[i].cfop` for every item.

#### Draft status label

15. Every user-facing rendering of the NF-e invoice status `draft` displays **"Não enviada"** — at minimum `templates/nfe/nfe-list-item.html` (used for both romaneio and avulsa rows on `/nfe/list`) and `templates/nfe/nfe-existing-modal.html`. The implementer must sweep `templates/` for any other status rendering and update it too. This is display-only: the internal status value `'draft'`, status transitions, worker retry behavior, and DANFE rules are unchanged.
16. After this change, the word "Rascunho" in the NF-e area refers exclusively to the reusable rascunho feature. The romaneio entry/departure rascunho UI keeps its existing wording and behavior.

#### Rascunho and invoice integrity

17. Every rascunho read or mutation, recipient reference, and optional farm-product reference is validated against the current farm. A rascunho or associated record from another farm cannot be loaded or used.
18. Saving, editing, or deleting a rascunho does not allocate an NF-e number, sign XML, persist an invoice, or call SEFAZ. The only allowed related write is the explicitly requested creation of an inline recipient contact. Emission happens only through the existing explicit confirmation flow.
19. Issued invoices retain the resolved recipient, product fields, unit prices, per-item CFOPs, and fiscal values used at issuance. Later rascunho edits or deletion do not alter prior invoice records or XML.
20. New user-facing text is Brazilian Portuguese. No CFOP data is relabeled as `Natureza da operação`; the two remain separate fields.

### Non-Functional

- Validate rascunho ownership and all referenced farm records server-side; never trust a client-supplied profile ID or farm ID.
- Preserve exact decimal values for rates and prices. Preserve the distinction between an unset rate (inherit) and an explicitly supplied zero rate.
- Preview and emission use the same parsed/resolved detached input so the values previewed are the values confirmed.
- Database migrations are transactional and idempotent. Existing detached XML, authorization state, invoice numbers, and item snapshots remain intact.
- The status label change is display-only (templates); no schema, status-string, or worker changes.
- Do not add client-side tax/CFOP legal inference. Prefilled fiscal values remain visible, editable, and subject to the existing server-side validation.
- Follow the existing romaneio rascunho UI patterns and `.agents/skills/ui-design` guidance.

### Side finding (pre-existing, out of scope)

`GET /nfe/avulsa/list` (`getDetachedNFeList`, `router/nfe_router/detached_router.go:565`) renders the template name `nfe-detached-list`, which is not defined anywhere in `templates/` — hitting that route fails template lookup. Detached invoices are actually listed on `/nfe/list` via `nfe-list-item`. Recorded so the status-label sweep accounts for it and the breakage is not mistaken for new work; fixing it is out of scope.

### Out of Scope

- Any change to the departure-linked NF-e/romaneio flow or its invoice-level configuration.
- Renaming the internal status string `'draft'`, the worker retry logic, or the romaneio rascunho UI/wording.
- Saving quantity, gross weight, vehicle, payment, emission identifiers, or SEFAZ status in a rascunho.
- Per-item tax rates/CST overrides; current detached tax overrides continue to apply uniformly across the items in one invoice.
- Automatic CFOP selection based on recipient state, product, or tax regime; the user selects and reviews the per-item code.
- Importing rascunhos/invoices, XML/CSV import, sharing across farms, and version history.
- Fixing the broken `nfe-detached-list` route (side finding above).

## Acceptance Criteria

1. **Rascunho lifecycle and farm scope:** an authorized farm user can create, list, edit, and delete a rascunho (confirmation on delete); duplicate names in one farm are rejected; an ID belonging to another farm cannot be read, applied, edited, or deleted. Verify with model/service tests and the focused rascunho E2E flow (`go test ./service/nfe_service/... ./model/nfe_model/...`, `go test ./router/nfe_router/...`, and `cd test/e2e && bun run test --grep "Rascunho de NF-e"`).
2. **Rascunho data round trip:** save and reload a rascunho containing one registered recipient, at least two ordered products, distinct item CFOPs, and different exact unit prices. Product description/NCM/CEST/unit, product reference where present, CFOP, price, operation/freight values, and optional tax overrides match the saved values. Verify rascunho model/service tests and the focused E2E test.
3. **Inline recipient handling:** saving a rascunho from inline recipient data explicitly creates/saves the recipient as a farm contact and links the rascunho to it; generating a preview alone creates no person. Verify with the rascunho-save and preview DB assertions in `test/e2e/tests/detached_nfe_rascunho.spec.js`.
4. **Applying a rascunho:** "Gerar NF-e a partir do rascunho" fills the recipient, products, per-item CFOPs, prices, and configured operation/tax fields; quantity and gross weight are left empty. Applying over a dirty form requires confirmation. Manual edits survive preview and confirmation and do not change the saved rascunho unless the user explicitly updates it. Verify with focused Playwright coverage.
5. **Item-scoped CFOP emission:** two items with different valid four-digit CFOPs produce their respective values in the XML and persisted `items_json`; a missing or malformed item CFOP is rejected at preview/validation. Verify with `service/nfe_service` unit tests and router/preview tests.
6. **Natureza da operação with mixed CFOPs:** a common item CFOP can retain the existing derivation; mixed item CFOPs use the explicit/rascunho/farm value and are rejected for preview when no such value exists. Verify with focused detached service tests.
7. **Data migration:** migration from the existing schema preserves each item's CFOP, backfills only missing item values from the old detached invoice column, removes `detached_nfe_invoice.cfop`, and leaves `nfe_invoice` and its CFOP unchanged. Verify by applying migrations to a seeded pre-change test database and asserting both the row data and schema.
8. **No rascunho path and regression:** detached emission without a rascunho remains usable, with the existing farm/default CFOP copied into each editable item row and current farm/tax defaults preserved. Departure-linked NF-e tests and `go test ./...` remain green.
9. **No issuance side effects from rascunho management:** create/edit/delete rascunho operations allocate no invoice numbers, create no detached invoice rows, sign no XML, and make no SEFAZ calls. The only allowed related write is the explicitly requested creation of an inline recipient contact. Verify with DB assertions in the E2E flow.
10. **Historical stability:** editing or deleting a rascunho does not change a previously emitted detached invoice's item CFOPs, unit prices, stored values, or XML. Verify with a model/DB test using a seeded invoice and a rascunho update.
11. **Draft status label:** every user-facing NF-e status rendering shows "Não enviada" for `draft` (list rows for romaneio and avulsa invoices, the existing-NF-e modal, and any other rendering found by the template sweep); no NF-e surface labels a status "Rascunho"; the internal status string `'draft'` and worker behavior are unchanged; the romaneio entry/departure rascunho wording is unchanged. Verify with a template render test plus Playwright assertions on the list and modal.
12. **Completion checks:** `go build ./...`, focused Go tests, `make test-go`, the focused E2E test, and `make test` pass; `gofmt -l` reports no formatting changes in changed Go files. Record exact commands/results below before marking Done.

## Verification Results

Baseline (before implementation, 2026-10-04):

- `go test ./service/nfe_service/... ./model/nfe_model/...` — PASS (no pre-existing failures)

Implementation verification (2026-10-04):

| Check | Command | Result |
|---|---|---|
| Build | `go build ./...` | PASS |
| Go suite | `go test ./...` | PASS |
| Focused Go tests | `go test ./service/nfe_service/... ./model/nfe_model/... ./router/nfe_router/...` | PASS |
| JS unit tests | `bun test unit/` (from `test/`) | PASS — 19 tests |
| Focused E2E | `cd test/e2e && bun run test --grep "Rascunho de NF-e"` | PASS — 13 passed, 2 skipped (beforeunload guard is Chromium-only), 0 failed |
| Data migration (AC7) | scratch DB + `bash /tmp/opencode/migration_check.sh` (baseline schema + migrations 1..21, seeded legacy rows, then 000022..000024) | PASS — own item CFOP kept, missing item CFOP backfilled from the legacy column, item order preserved, `detached_nfe_invoice.cfop` dropped, `nfe_invoice` row CFOP `5999` and column unchanged, migration re-runnable, profile table + partial IE index created |
| Historical stability (AC10) | same scratch-DB script, seeded emitted invoice + rascunho update/delete | PASS — emitted item CFOPs `5101/5102`, unit price `100`, total `150.00` unchanged after the rascunho was updated and hard-deleted |
| Formatting | `gofmt -l` on changed Go files | PASS (no output) |
| Full suite | `make test` | 92 passed, 2 skipped, 5 failed — all 5 failures are pre-existing WebKit entry/departure flakes (see note); every Rascunho de NF-e test passed in all three browsers |

The focused E2E covers: rascunho lifecycle (create/list/apply/edit/delete with confirmation), duplicate-name rejection (case-insensitive), cross-farm read/apply/edit/delete rejection, two ordered items with distinct CFOPs and exact prices plus a farm-product reference, quantity/gross-weight blanks on apply, manual edits not written back, no number/invoice side effects, inline-contact save on "Salvar como rascunho", preview write-freedom, preview hidden-field CFOP round trip, dirty-form confirmation (Chromium), and the "Não enviada" list label.

Template guards: `router/nfe_router/nfe_templates_test.go` renders `nfe-list-item` and `nfe-existing-modal` with a `draft` invoice asserting "Não enviada" and no `>Rascunho<` status label, sweeps `templates/nfe/` for draft/Rascunho pairings, and asserts the romaneio entry/departure rascunho wording is unchanged.

Extra in-scope fix found by the E2E flow: migration `000024_person_ie_unique_partial.sql` replaces the `UNIQUE (farm, ie)` person constraint with a partial unique index, because two inline detached recipients without an IE (stored as `''`) previously collided with a 500 on the second save.

Full-suite note: `make test` was run three times. The new Rascunho de NF-e tests pass in every browser (13 passed, 2 browser-limited skips) and no Rascunho de NF-e test failed in the final run (`92 passed, 2 skipped, 5 failed`). The 5 remaining failures are the pre-existing WebKit entry/departure dialog-actionability timeouts (`departure.spec.js`, `entry.spec.js`, `entry_discount.spec.js`); the same failures reproduce on the unmodified `HEAD` worktree (same tests, same errors), and their count varies between runs (5–7) with parallel load, confirming they are unrelated pre-existing flakiness. One `login.spec.js` Firefox flake observed under load passed when re-run in isolation (18/18).

Status: Done — all acceptance criteria are verified; the only failing check is the documented pre-existing WebKit suite flakiness.
