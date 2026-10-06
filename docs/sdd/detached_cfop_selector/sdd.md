# sdd:detached_cfop_selector

## Status
Backlog

## Goal
Give the detached NF-e (NF-e avulsa) flow a system-wide CFOP catalog and a farm-scoped CFOP selector with search, replacing the free-text per-item CFOP input on the detached emission page and on the "Rascunho de NF-e" editor. The system cannot know in advance which nature of operation a user will build a detached NF-e around, so the selector must offer an extensive range of CFOPs, ordered per farm by the codes the farm actually uses most. Users can register farm-tied CFOPs (with the Contabilizei table's third column, **Origem/Destino**) through a small modal, exactly like the existing vehicle/crop/field add-on-selector pattern. Nothing in the departure-tied NF-e flow changes.

Reference for the catalog and the Origem/Destino column: https://www.contabilizei.com.br/contabilidade-online/tabela-cfop-completa/#tabela-cfop-completa-consulte-os-codigos

## Naming

| Surface | Term |
|---|---|
| User-facing concept | **CFOP** — selector per item row, modal "Cadastrar CFOP", groups "Mais utilizados" / "Todos os CFOPs" |
| Internal identifiers | English `Cfop` (`CFOPOption`, `model/cfop_model/`, `router/nfe_router/cfop_router.go`, `templates/nfe/cfop-*.html`) |
| Third column of the reference table | `origin_destination` — values `Mesmo estado`, `Outro estado`, `Exterior` |

## Requirements

### Functional

#### System-wide CFOP catalog

1. A new idempotent migration creates a system-wide read reference table `cfop` (`code TEXT PRIMARY KEY`, `description TEXT NOT NULL`, `origin_destination TEXT NOT NULL` with a CHECK restricting it to `Mesmo estado` / `Outro estado` / `Exterior`). It is seeded `ON CONFLICT DO NOTHING` with every code listed on the referenced Contabilizei page (each with its description and Origem/Destino value) plus the CFOPs already assumed by the system (`pkg/nfe/defaults` `NaturezaOpForCFOP` keys and agriculture defaults `5101`, `5102`, `5901`, `6202`).
2. The catalog is read-only through the app: no route creates, edits, or deletes a catalog row. The seed is a migration concern only.
3. Every code in the catalog starts with a digit in `1–3` (entries: 1 mesmo estado, 2 outro estado, 3 exterior) or `5–7` (exits: 5 mesmo estado, 6 outro estado, 7 exterior). `origin_destination` is derived from the first digit (1/5 → `Mesmo estado`, 2/6 → `Outro estado`, 3/7 → `Exterior`).

#### Farm-tied CFOPs

4. A farm can register additional CFOPs via `farm_cfop` (`farm_id + code` composite PK, `description`, `origin_destination`, same CHECKs). A farm CFOP is visible **only** to that farm. Registering does not consult SEFAZ or the catalog beyond duplicate detection.
5. Registering a farm CFOP whose code already exists in the system catalog is rejected with a warning toast ("CFOP já cadastrado") — the farm keeps using the catalog entry; there are no per-farm description overrides, so catalog ∪ farm is a disjoint merge. A duplicate within the same farm is also rejected (unique violation → warning toast).
6. A metadata transition is validated server-side: the submitted `origin_destination` must equal the first-digit derivation of the submitted `code` (mismatch → warning toast). The register modal presents it as a `<select>` with the three permitted values, auto-derived (client-side convenience) from the typed code and still user-visible/editable before submit.
7. There is no standalone page for managing CFOPs — the only user interface is the selector on the detached NF-e surfaces (same convention as vehicle/crop/field). Deleting or editing farm CFOPs is out of scope.

#### Selector (detached emission page + rascunho editor)

8. Both `templates/pages/nfe-emit.html` (detached emission items) and `templates/nfe/nfe-rascunho-form.html` (rascunho editor items) replace the per-item CFOP text input with the new `cfop-selector` component (`templates/nfe/cfop-selector.html`), structured like `templates/crop/crop-selector.html` (label + `<select>` + a "+" button opening the register modal via `hx-get` → `hx-target="body"` `hx-swap="beforeend"`) **plus a search input in the same flex row** (label on top; search input `grow`; select `grow`; "+" button).
9. The `<select>` keeps `name="items[i].cfop"` (required, exactly the current form contract): server parsers (`parseDetachedItems`, `parseDetachedProfileItems`), `reindexItems` clone/reindex JS, and the preview hidden-field round trip continue to work unchanged. Option text is `code — description`.
10. Options are ordered per **farm** (not per user) and rendered in two optgroups: **"Mais utilizados"** (farm use-count > 0, most-used first, ties by code ASC) then **"Todos os CFOPs"** (code ASC). An empty "Mais utilizados" group is not rendered. Farm-registered CFOPs participate in both groups (they live in "Todos os CFOPs" until used).
11. The farm's default CFOP (`nfe_farm_config.default_cfop`, else `5101`) is always selectable: if absent from the merged list, the server prepends it to "Todos os CFOPs" (description from the catalog or derived from `NaturezaOpForCFOP`). New item rows and the add-item clone start from this default (existing `data-default-cfop` JS behavior preserved).
12. The search input filters **only its own item row's** select options, matching code and description substrings, case/diacritics-insensitively. The currently selected option is never filtered away. Typing the full/unique code highlights the matching option for keyboard selection. Rows do not influence each other.
13. Preloading ~200 options server-side per page render is intentional (catalog lives on one farm-sized list; no client filtering of the *source list*, only of the rendered options). No pagination or lazy loading.

#### Register modal

14. The modal (`templates/nfe/nfe-cfop-form.html`, opened from any item row's "+") collects: CFOP (4-digit pattern input), Descrição (free text, maxlength matching catalog practice ~200), Origem/Destino (select of the three permitted values). HTMX-native submission (`hx-post="/nfe/cfop"`, pattern of `crop-form.html`); success returns 201 with the `cfop-option` fragment and triggers a farm-ordered options refresh for the originating row's select; failure returns a pt-br toast/diagnostic and keeps the modal open.
15. On success the new option is auto-selected in the item row whose "+" opened the modal; other rows get the option appended/refreshed but keep their current selection.

#### Use tracking

16. Farm use counters live in `farm_cfop_use` (`farm_id + code` PK, `use_count`, `updated_at`) — per farm, never per user. Counting happens **once per detached emission action**: when `BuildDetachedInvoice` persists (draft, authorized, or any persisted outcome), each **distinct** item CFOP of that emission gets `use_count + 1`. The SVC supersede path inside a single emission action must not double-count; the background worker's draft auto-send / pending retry must not count. Counting is a local DB write only — **no SEFAZ contact** — and its failure is logged, never blocking or toasting the emission.
17. Selecting an option in the selector, saving a rascunho, or previewing does **not** increment counters.

#### Scope isolation

18. Nothing about the departure-tied NF-e changes: `nfe_invoice`, `nfe-emit-modal.html`, `BuildInvoiceFromDeparture`, its CFOP source, `nfe_farm_config`, and endpoints tied to it are untouched. The new routes are registered inside the existing `/nfe` group (tier `fiscal` middleware applies) and derive farm ID from the authenticated session.

### Non-Functional

- All user-facing text in pt-br (labels, toasts, placeholders, optgroup titles).
- New inline scripts carry `{{ .CSPNonce }}`; selector/dialog JS lives in `assets/js/` following `cropDialog.js` module patterns.
- Migration is transactional and idempotent (`CREATE TABLE IF NOT EXISTS`, `INSERT ... ON CONFLICT DO NOTHING`); it adds no columns to existing NF-e tables.
- Catalog/service exposes a `is-four-digit + digits 1–3/5–7` validation helper reused by the modal route (the existing `IsFourDigitCFOP` guards the item parsers and remains untouched).
- `gofmt` clean; follows Layered architecture Entity → Model → Service → Router with singleton model init wired in `main.go`.
- No test/smoke database may ever receive `nfe_farm_config` rows (SEFAZ isolation rules remain in force; the selector feature adds no SEFAZ-calling code).

## Acceptance Criteria

1. **Catalog migration + seed:** migration `000025` creates `cfop`, `farm_cfop`, `farm_cfop_use` with the stated keys/CHECKs and seeds the catalog with all codes of the referenced page plus `5101`, `5102`, `5901`, `6202` (description + correct `origin_destination`); re-running the migration is a no-op. Verify with a scratch-DB migration script (baseline schema → migrations → assertions → idempotence re-run) as in previous SDD verify steps.
2. **Farm CFOP registration round trip:** from the detached emission page **and** from the rascunho editor, the "+" opens the modal; submitting a valid code/description/origem creates a `farm_cfop` row scoped to the farm, the option appears immediately and is auto-selected in the originating row, and a fresh page render shows it inside the farm-ordered options (in "Todos os CFOPs"). Verify with focused Playwright tests + DB assertions.
3. **Registration validation:** codes not 4 digits, not starting with 1–3/5–7, malformed origin/destination, catalog duplicates, and same-farm duplicates are rejected with specific pt-br warning toasts; nothing is persisted. Verify with router handler tests (parse/validation paths) and the E2E duplicate case.
4. **Ordering groups:** `ListCfopsForFarm` merges catalog ∪ farm (disjoint), joins counters, returns most-used first (count DESC, tie code ASC), then code ASC; an empty use bucket is omitted; the resolved default CFOP is always present (prepended when missing). Verify with view/model unit tests (offline, pure) capturing the ordering/grouping functions.
5. **Use counting semantics:** one emission action with item CFOPs `["5101","5101","6102"]` increments `5101` and `6102` by 1 each (dedup); a superseding SVC rebuild inside the same action adds nothing extra on later calls of the increment helper; worker retry paths don't count; increment failure only logs. Verify with a service-layer test using a hand-written `CfopModelInterface` mock (pattern of `service/entry_service/test/mocks.go`) and an E2E DB assertion.
6. **Selector search:** per-row typed input filters that row's options by code and description (accent-insensitive), the selected option survives filtering, other rows keep their full option sets, and the selected value is submitted correctly for preview and emission. Verify with the focused E2E suite.
7. **Form contract unchanged:** `items[i].cfop` submission from both surfaces still parses with existing validation; add-item clones start from the default CFOP; reindex keeps contiguous names; preview hidden fields round-trip per-item CFOPs. Verify with existing/focused router tests (`TestParseDetachedItems_PerItemCFOP` etc. — must keep passing) and E2E.
8. **Default CFOP availability:** with fixture defaults (no `nfe_farm_config`), option lists contain `5101`; with a farm config default set, that value is present and selectable. Verify with view unit tests and the E2E page render assertions.
9. **Scope isolation / no SEFAZ:** departure NF-e tests, template guards, and `gofmt` remain untouched and green; no new SEFAZ-touching code; the e2e fixture still contains no `nfe_farm_config`; count increment writes only `farm_cfop_use`. Verify with `go test ./...` passing and a code-path review recorded below.
10. **Completion checks:** `go build ./...`, focused Go tests (`model/cfop_model/`, `router/nfe_router/`, `service/nfe_service/`), focused E2E (`--grep "CFOP"`), `make test-go`, `make test-js`, `make test-e2e`, `gofmt -l` on changed files. Record exact commands/results below before marking Done.

## Verification Results

(To be recorded during implementation. Baseline commands to run before flipping Status to WIP:)

```bash
go test ./router/nfe_router/... ./service/nfe_service/... ./model/nfe_model/...
```
