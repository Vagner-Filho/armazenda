# plan:detached_cfop_selector

Implementation plan for the system-wide CFOP catalog, farm-scoped CFOP selector with search, farm CFOP registration modal (with Origem/Destino), and per-farm use ranking for the detached NF-e surfaces. Specification: [sdd.md](./sdd.md).

## Test Strategy

Run the existing focused suites before implementation and record any pre-existing failures before changing Status from `Backlog` to `WIP` (commands in sdd.md Verification Results).

| Behavior | Test location/layer | Environment / command |
|---|---|---|
| Catalog seed list integrity (all codes present, `origin_destination` matches first digit, extras `5101/5102/5901/6202`) | `model/cfop_model/cfop_seed_test.go` (pure Go over the embedded migration file's INSERT list; no DB) | `go test ./model/cfop_model/...` |
| Merge/order/group of `ListCfopsForFarm` output handling, default-CFOP prepend, going `use_count` bucketing | `model/cfop_model/` unit test over the pure grouping/sorting helpers | `go test ./model/cfop_model/...` |
| 4-digit + first-digit-family validation and origin/destination derivation/consistency | `service/nfe_service/` validation helper tests | `go test ./service/nfe_service/...` |
| Register-modal parse/validation, duplicate rejection, pt-br toast payloads, options fragment rendering | `router/nfe_router/cfop_router_test.go` (new; handler parse paths + template render) | `go test ./router/nfe_router/...` |
| Dedup increment semantics with the emission flow | `service/nfe_service/` test with a hand-written `cfopModel` mock (pattern of `service/entry_service/test/mocks.go`): assert `IncrementFarmCfopUse` called once per action with the deduped codes, and that errors only log | `go test ./service/nfe_service/...` |
| Migration applies twice cleanly; tables/keys/CHECKs as specified; seed idempotence | scratch-DB bash script (baseline schema → migrations → assert → re-run idempotence), as prior SDDs did | Docker DB (`make test-e2e` infra), documented in Verification Results |
| Selector search per-row scoping, modal round trip (create → appears + auto-select), default CFOP selectable, `items[i].cfop` round trip through preview | `test/e2e/tests/detached_cfop_selector.spec.js` (new, Playwright, Chromium-focused run) | `cd test/e2e && bun run test --grep "CFOP"`; DB fixtures must **not** contain `nfe_farm_config` (SEFAZ isolation) |
| Departure-flow regression | Existing suites | `make test-go`, `make test-js`, `make test-e2e` before completion |

SEFAZ note: no new code may call SEFAZ. Use counters are written right after `CreateDetachedInvoice` (local DB). E2E stays SEFAZ-free; never seed `nfe_farm_config` into test DBs.

## Phase 1 — Migration + catalog data

**Files:**

- `model/armazenda_database/migrations/000025_cfop_catalog.sql` (new)

1. Run baseline focused tests; record in sdd.md; flip Status to `WIP`.
2. Create tables idempotently:
   ```sql
   CREATE TABLE IF NOT EXISTS cfop (
       code               TEXT PRIMARY KEY,
       description        TEXT NOT NULL,
       origin_destination TEXT NOT NULL,
       CONSTRAINT cfop_origin_destination_check
           CHECK (origin_destination IN ('Mesmo estado','Outro estado','Exterior'))
   );
   CREATE TABLE IF NOT EXISTS farm_cfop (
       farm_id            INTEGER NOT NULL REFERENCES farm(id),
       code               TEXT    NOT NULL,
       description        TEXT NOT NULL,
       origin_destination TEXT NOT NULL,
       PRIMARY KEY (farm_id, code),
       CONSTRAINT farm_cfop_origin_destination_check
           CHECK (origin_destination IN ('Mesmo estado','Outro estado','Exterior'))
   );
   CREATE TABLE IF NOT EXISTS farm_cfop_use (
       farm_id    INTEGER NOT NULL REFERENCES farm(id),
       code       TEXT    NOT NULL,
       use_count  INTEGER NOT NULL DEFAULT 0,
       updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
       PRIMARY KEY (farm_id, code)
   );
   ```
3. Seed the catalog with one `INSERT ... VALUES (...),(...) ON CONFLICT (code) DO NOTHING` block containing **every code from the referenced Contabilizei page** (code, description, origem/destino for each table section of that page — families 1xxx entries, 5xxx/6xxx exits, including 7xxx exterior variants where the page lists them), plus `5101`, `5102`, `5901`, `6202` if not in the page list (they are; still assert). Derive `origin_destination` per first digit while building the seed (1/5→Mesmo estado, 2/6→Outro estado, 3/7→Exterior) and double-check against the page's third column.
4. Assertions for the scratch-DB script: expected row count (~page count + extras), all seeded codes can be classified by first digit, second run leaves row count unchanged, `farm_cfop`/`farm_cfop_use` empty initially.

**Phase gate:** scratch-DB migration script PASS; `go test ./model/cfop_model/...` (seed test) PASS.

## Phase 2 — Entity + model (merge/order/counts)

**Files:**

- `entity/public/cfop.go` (new)
- `model/cfop_model/model.go` (new)
- `main.go` (init wiring, one import + one line)
- `model/cfop_model/model_test.go` (new)

1. `entity/public/cfop.go`: `CfopOption{Code string; Description string; OriginDestination string; UseCount int; IsFarm bool}`.
2. `model/cfop_model` singleton (`InitCfopModel(pool)` + `GetCfopModel()`):
   - `ListCfopsForFarm(farmID uint32) ([]CfopOption, *model_error.ModelError)` — one query `UNION`-style:
     ```sql
     -- catalog rows
     SELECT code, description, origin_destination, 0 AS use_count, false AS is_farm FROM cfop
     UNION ALL
     -- farm rows (disjoint by construction, reject-at-register policy)
     SELECT code, description, origin_destination, COALESCE(u.use_count,0), true
     FROM farm_cfop f LEFT JOIN farm_cfop_use u ON u.farm_id=f.farm_id AND u.code=f.code
     WHERE f.farm_id=@farm
     ```
     then the caller (view/controller) splits "Mais utilizados" (`UseCount>0`, count DESC, code ASC) vs "Todos os CFOPs" (code ASC) via a pure function `GroupCfopOptions([]CfopOption) (mostUsed, allCfops []CfopOption)` exported from the model package; the farm default CFOP prepend (from `nfe_farm_config.default_cfop` else `5101`) happens in the view layer.
   - `AddFarmCfop(farmID, code, description, originDestination)` — INSERT with unique-violation → `ModelError{IsServerErr:false, Message:"CFOP já cadastrado para esta fazenda"}`; server faults → `IsServerErr:true` ("Falhamos ao cadastrar o CFOP").
   - `IncrementFarmCfopUse(farmID uint32, codes []string)` — batch upsert:
     ```sql
     INSERT INTO farm_cfop_use (farm_id, code, use_count, updated_at)
     SELECT @farm, c.code, 1, CURRENT_TIMESTAMP FROM unnest(@codes::text[]) AS c(code)
     ON CONFLICT (farm_id, code) DO UPDATE
       SET use_count = farm_cfop_use.use_count + 1, updated_at = CURRENT_TIMESTAMP
     ```
     dedup happens upstream (single SQL pass already guarantees one bump per emission given deduped input).
3. Wire init in `main.go`.
4. Pure-unit tests: grouping function ordering (most-used incl. farm rows, ties by code), empty most-used omits group, farm default prepend helper, validation of `origin_destination`/first-digit derivation.

**Phase gate:** `go test ./model/cfop_model/... ./model/nfe_model/...` PASS.

## Phase 3 — Service + routes + view

**Files:**

- `service/nfe_service/cfop_service.go` (new; or fold into `detached_service.go` — decide when code structure is visible)
- `pkg/nfe/defaults/agriculture.go` untouched; a shared validation helper `ValidSelectorCFOP(code) bool` (4 digits + first in 1–3/5–7) lives beside `IsFourDigitCFOP` in `service/nfe_service/`
- `view/cfop/view.go` (new): `CfopOptions{MostUsed, All []CfopOption}` + `GetCfopOptions(farmID)` importing `model/cfop_model` and farm config default prepend
- `router/nfe_router/cfop_router.go` (new)
- `service/nfe_service/detached_service.go` (increment hook, ~5 lines after `CreateDetachedInvoice` success, also after SVC supersede within the same action — pass code set through, call once)
- `service/nfe_service/test/mocks.go` or a local mock (extend interfaces if the mock pattern requires)
- Tests: `router/nfe_router/cfop_router_test.go`, `service/nfe_service/cfop_service_test.go`

1. Routes registered by a call inside `UseNFeRoutes` (`router/nfe_router/router.go` — inside the `/nfe` group, therefore tier `fiscal`):
   - `GET /nfe/cfop/options` → renders `cfop-options` block (both optgroups) with farm ordering; used post-creation to refresh rows.
   - `GET /nfe/cfop/form` → renders `nfe-cfop-form` modal (farm-aware nothing; needs no data beyond nonce).
   - `POST /nfe/cfop` → validate (`ValidSelectorCFOP`, farm duplicate, catalog duplicate via `cfop` table lookup, origin/destination consistency), create `farm_cfop`, 201 `cfop-option` fragment; warnings/errors via `HX-Trigger` toasts in pt-br.
2. Farm ID always from `user_service.GetFarmFromToken(c.Cookie("session_id"))`; never trust form farm identifiers.
3. Increment hook: dedup item CFOPs, call `cfopModel.IncrementFarmCfopUse(farmID, dedupedCodes)` after detached persistence success; failure → `model_error.GetLoggerModel().Log(...)` only; must not run on preview or rascunho-save paths. Do not touch `BuildInvoiceFromDeparture`.
4. Router tests: modal fragment renders; parse/validation failures produce the right toast types + no persistence; catalog-duplicate and same-farm-duplicate paths; options fragment shows groups.

**Phase gate:** `go test ./router/nfe_router/... ./service/nfe_service/...` PASS.

## Phase 4 — Templates + JS wiring

**Files:**

- `templates/nfe/cfop-selector.html` (new; block `cfop-selector`)
- `templates/nfe/cfop-option.html` (new; single `<option>` fragment)
- `templates/nfe/cfop-options.html` (new; the optgrouped `<option>` list fragment)
- `templates/nfe/nfe-cfop-form.html` (new; modal, mirrors `templates/crop/crop-form.html`)
- `assets/js/cfopSelector.js` (new; module)
- `templates/pages/nfe-emit.html` (item-row CFOP cell → selector include; keep names/`data-default-cfop` contract)
- `templates/nfe/nfe-rascunho-form.html` (same swap)
- `test/e2e/tests/detached_cfop_selector.spec.js` (new)

1. Component structure (per row):
   ```
   <div class="flex flex-col">
     <label class="...">CFOP</label>
     <div class="flex flex-wrap w-full gap-2">
       <input class="cfop-search grow ..." placeholder="Buscar código ou descrição…">
       <select name="items[i].cfop" class="cfop-select grow ..." required>
         <optgroup label="Mais utilizados">…</optgroup>
         <optgroup label="Todos os CFOPs">…</optgroup>
       </select>
       <button class="icon-btn ..." type="button"
               hx-get="/nfe/cfop/form" hx-target="body" hx-swap="beforeend"
               data-originating-row="{{ $index }}"></button>
     </div>
   </div>
   ```
   Option text `code — description`; farm rows flagged via `data-cfop-farm="true"` (for tests/future UX).
2. `cfopSelector.js`: search filters `closest('.item-row, .rascunho-item-row')`'s own select (normalize diacritics via `String.prototype.normalize('NFD')` + strip combining marks; match code prefix or description substring; keep selected option visible); the "+" click records the originating row (weak-set/dataset) so the modal success handler can auto-select; after successful creation, fetch/htmx-refresh that row's option list (`GET /nfe/cfop/options`) preserving the selection; add-item clone keeps `defaultItemCFOP` behavior (existing JS, ensure cloned search value resets and names reindex — already handled by `reindexItems`; only the search input value needs clearing on clone).
3. Modal template mirrors `crop-form.html` (dialog, `hx-post="/nfe/cfop"` on the form, cancel button, CSPNonce'd inline script for dialog open/close lifecycle). On `htmx:afterRequest` success → close dialog; the JS module handles option insertion/auto-select via the options refresh rather than duplicating insertion logic.
4. Submission strategy (HTMX-native, agreed): the modal form posts with `hx-post="/nfe/cfop"`; the 201 response carries the new `<option>` fragment, and per-row farm-ordered option refreshes (`GET /nfe/cfop/options`) take care of ordering and selection — the JS module coordinates which row triggered the creation. No dynamic per-row `hx-target` hacks.
5. E2E (Chromium run): search per row; register modal flow (create → option appears → auto-selected in originating row → persists across reload with farm ordering); duplicate-toast path; default CFOP present; `items[0].cfop` survives preview round trip (SEFAZ-free area). Fixtures unchanged (no `nfe_farm_config`).

**Phase gate:** `make test-js` PASS (no new JS unit tests strictly required; e2e covers behavior), focused E2E PASS:
```bash
cd test/e2e && bun run test --grep "CFOP"
```

## Phase 5 — Documentation + completion verification

1. Update `docs/DETACHED_NFE_IMPLEMENTATION.md`: CFOP selector, catalog, farm CFOPs, use ranking, Origem/Destino column, and the no-management-page convention.
2. Focused + full checks:

```bash
go test ./model/cfop_model/... ./router/nfe_router/... ./service/nfe_service/...
go build ./...
gofmt -l <changed go files>
make test-go
make test-js
make test-e2e
```

3. Record exact commands/results in sdd.md Verification Results; flip `Status: Done` only when all ACs verified.

## Risks and Mitigations

- **Seed list completeness/accuracy:** the page's table is multi-section; build the VALUES list programmatically from the fetched text, cross-check every row's origem/destino column, and add a scratch-DB assertion comparing seeded count vs the extracted list length. Accented descriptions (ex: "não especificada") are fine for the DB and templates, but note `pkg/nfe/xml/sanitize.go` folding rules apply **only** to XML fields — catalog text is template-only data, never concatenated into XML.
- **HTMX option-append vs farm ordering:** the creation response appends the new `<option>` to all rows' selects; per-row options refresh re-renders the farm-ordered list so ordering self-heals without full page reloads.
- **Row identity across reindex:** originating-row tracking must survive row add/remove/reindex; use a dataset marker resolved at click time (row element reference), not index-based lookup.
- **Use-count inflation paths:** counting is bound to the emission-persist success site only (including its SVC supersede within one action); tests must assert the worker retry path does not call the increment.
- **252-length descriptions:** maxlength validated server-side responsive to the longest catalog entries; farm modal input maxlength mirrors it (200) to keep UX tight.
- **Cross farm leakage:** `ListCfopsForFarm` filters `farm_cfop`/`farm_cfop_use` by session farm; catalog is shared by design (system-wide, read-only).
- **SEFAZ isolation:** no new SEFAZ-calling code; increment is local-DB only; e2e fixtures keep no `nfe_farm_config`; if any verification seems to require SEFAZ, stop and ask the human (per repo rules).
- **Departure regression:** new files/routes only; existing parser test files untouched (their passing is an AC gate).
