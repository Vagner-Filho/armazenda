# plan:decimal_comma_mask

Implementation plan for the decimal comma mask (spec: `docs/sdd/decimal_comma_mask/sdd.md`).

**Core constraint:** the server contract is dot-decimal (`xxx.xx`) and must be **respected, not changed**. No Go source files are modified anywhere in this plan; the mask + htmx parameter override is the only comma→dot translation layer. Comma input that bypasses the mask keeps failing with the existing (correct) server toasts.

## Test strategy

Layers and locations (before implementation starts):

| Behavior | Layer | Location | Fixtures / env | Command |
|----------|-------|----------|----------------|---------|
| Mask pure core: comma/dot/mixed/invalid, caps, grouping, cursor math, raw = dot-decimal | JS unit (DOM-free) | `test/unit/decimalMask.test.js` | Bun only, no DOM library | `cd test && bun test unit/decimalMask.test.js` |
| Full JS unit suite sanity (weight.js etc. untouched) | JS unit | `test/unit/*` | Bun | `cd test && bun test unit/` (`make test-js`) |
| Backend contract regression: parsers unchanged (comma still rejected) | Go (no code change) | existing `router/nfe_router` tests | None — read-only regression reference | `go test ./router/nfe_router/` (Phase 0 baseline) → `make test-go` (gate, same result set) |
| User journey: comma typing → dot submission → preview OK, no 400 toast | E2E (reusable) | `test/e2e/tests/detached_decimals.spec.js` (new); optional extension of `detached_nfe_rascunho.spec.js` | Playwright global-setup auto-starts Docker DB (localhost:5433) + app (localhost:8100) and seeds fixtures; `NFE_SEFAZ_GUARD=block` is set by `global-setup.js`. The flows exercised here (`/nfe/emitir/preview`, `/nfe/rascunho/salvar`) are SEFAZ-free — no NF-e emission reaches SEFAZ. No `nfe_farm_config` may be seeded (AGENTS rule 1). | `cd test/e2e && bunx playwright test tests/detached_decimals.spec.js` |
| Completion gates | JS + E2E (+ Go as unchanged backend sanity) | — | Docker for E2E | `make test-js`, `make test-e2e`, `make test-go` |

Behaviors to cover (mapped to ACs in sdd.md):

1. `"1250,90"` / `"1250.90"` / `"1.250,90"` → raw `"1250.90"`, grouped comma display (AC1)
2. `"0,9"` → raw `"0.9"`; `"0.900"` → raw `"0.900"` (scale untouched) (AC1)
3. Empty / only-separator values → both raw and formatted empty (AC1)
4. Non-numeric chars stripped on input; keydown allowlist (digits, `,`, `.`; control keys pass) (AC1)
5. `decimals: 3` vs `2` display caps; raw preserves typed extra digits (AC1)
6. Cursor preserved across reformat with grouping (AC1)
7. `htmx:configRequest` override: `input[data-raw]` name present in parameters → replaced by `dataset.raw` (dot-decimal); disabled controls skipped (AC2, AC4, AC7)
8. E2E payload interception: masked submission carries dot-decimal for item quantity/price, volumes and rates (AC2–AC4, AC7)
9. E2E: comma typing end-to-end + cloned row + programmatic prefill display (AC3, AC5)
10. Regression: entry/departure forms unchanged (AC8); Go parsers unchanged (AC6)

## Phase 0 — Baseline

- Run and record pre-existing results of the affected suites (update spec `Status` to **WIP** first):
  - `go test ./router/nfe_router/` — the **unchanged backend contract reference**: the recorded baseline is what AC6 must reproduce at the gate.
  - `cd test && bun test unit/`
- Note any unrelated failures in the spec's Verification Results; do not fix them in this feature.
- No DB migration needed (no schema change).

## Phase 1 — JS mask module + unit tests (AC1)

Files:

- **New** `assets/js/decimalMask.js` — no new dependencies, ES module, no bundler step:
  - Pure core (exported, DOM-free): `formatDecimalInput(value, {decimals})` → `{formatted, raw}`. Reuse the `weight.js` logic as the base: `formatForDisplay`'s strip/replace rules and `computeCursorPosition`'s algorithm (copy tonight; refactoring `weight.js` itself is out of scope).
  - `raw` is always the server-contract dot-decimal form (no grouping separators, `.` decimal); display formatter: `Intl.NumberFormat('pt-BR')` with configurable `maximumFractionDigits` (3 default for quantity/weights/volumes, 2 for prices/rates); full pt-BR grouping; comma decimals.
  - DOM wiring: `setupDecimalMasks(root, signal)` for `[data-mask="decimal"]` inputs — `keydown` allowlist, live `input` reformat + `dataset.raw` sync + cursor restore, `blur` format, `focusin` show editable comma raw; `AbortSignal` cleanup on every listener; `refreshMask(input)` for programmatic writes.
  - `wireDecimalParamOverride(signal)`: body-level `htmx:configRequest` listener — for each `document`-scoped `input[data-raw]` whose `name` key exists in `evt.detail.parameters`, set `parameters[name] = dataset.raw` (skip when `dataset.raw` is `undefined`; disabled controls are not serialized so they are naturally skipped). This listener **is the contract translator** — its correctness is proven by payload interception in E2E (Phase 2).
- **New** `test/unit/decimalMask.test.js` — table-driven; mirrors the repo's "pure state helper is unit-testable" style (`detachedNFeModes.js`). Include the cursor-restoration math as a pure function test (expose the cursor helper on the module's exported surface or test via `formatDecimalInput` + a tiny exported helper).

Tests live in this phase (not deferred). Focus commands:

- `cd test && bun test unit/decimalMask.test.js`
- `cd test && bun test unit/`

## Phase 2 — Apply the mask to the four templates (AC2–AC5, AC7)

Files and touchpoints:

1. **`templates/pages/nfe-emit.html`**
   - Add `data-mask="decimal"` to: `items[i].quantity`, `items[i].unitPrice`, `items[i].discount`, `items[i].grossWeight`, `volPesoL`, `volPesoB`, `icmsRate`, `pisRate`, `cofinsRate`, `ibsRate`, `cbsRate`.
   - In the page's inline module (near the `initCfopSelectors`/`initDetachedNFeTabs` imports): import the mask module; call `setupDecimalMasks(#detached-nfe-form)` and `wireDecimalParamOverride(signal)` once.
   - Item-row clone handling: after `itemsContainer.appendChild(template)` in the `add-item-btn` handler, run mask setup for the cloned row (`setupDecimalMasks(template)` or per-row setup).
   - After programmatic writes in `updateVolumePrefill` (`volFields.volPesoL/B.value = …`) and the farm-product `change` handler, call `refreshMask(input)` (programmatic `.value` assignment does not fire `input`).
2. **`templates/nfe/nfe-emit-modal.html` + `assets/js/form/nfe-form.js`**
   - Add `data-mask="decimal"` to `volPesoL`, `volPesoB` and the five rate inputs.
   - `setupNFeForm` calls `setupDecimalMasks(...)` + `wireDecimalParamOverride(signal)`.
   - `useDefaultTaxRates` toggle keeps its existing enabled/disabled flow; mask on disabled inputs is inert, disabled controls don't serialize.
3. **`templates/nfe/nfe-config-form.html`** — five rate inputs get `data-mask="decimal"`; its inline module imports and wires setup + override.
4. **`templates/nfe/nfe-rascunho-form.html`** — five rate inputs get `data-mask="decimal"`; inline module (`removeEmptyFields` import block) wires setup + override. The draft-save form (`#save-rascunho-form`, `hx-include="#detached-nfe-form"`) is covered by the shared override listener.

E2E test (same phase):

- **New** `test/e2e/tests/detached_decimals.spec.js`:
  - Type `1250,90` into item quantity / unit price, `17,00` into a rate (enable editing by unchecking "Usar taxa padrão" or use the draft/config surface per existing helpers), submit via the `Avançar` preview path.
  - **Payload interception (contract proof):** attach `htmx:configRequest`/request interception to assert the outgoing parameters carry dot-decimal (`items[0].quantity=1250.90`, rates `17.00`, …) — this is the acceptance evidence that the server-contract translation happened client-side.
  - Assert: no failure toast (no `HX-Trigger` warning, 200 response), preview totals reflect the comma values; volume prefill shows grouped comma display; cloned second item also accepts comma input.
  - Reuse `test/e2e/utils/helpers.js` patterns; keep fixtures SEFAZ-free (preview/draft paths only — see SEFAZ Isolation note in the strategy table).
- **Extend** `detached_nfe_rascunho.spec.js` with a comma-value save if the fixtures make it cheap; otherwise cover the draft path inside `detached_decimals.spec.js` via the save form.

Commands:

- `cd test/e2e && bunx playwright test tests/detached_decimals.spec.js` (focused; Docker must be available — global-setup handles DB/app/seed)

## Phase 3 — Completion gates & bookkeeping

- `make test-js`, then `make test-e2e` (Docker); `make test-go` as the **unchanged-backend sanity gate** — must reproduce the Phase 0 baseline exactly (AC6). Record exact commands and PASS/FAIL/NOT RUN in the spec's Verification Results; keep the spec accurate if scope moved.
- `git diff --stat` review to confirm the diff touches only `assets/js/**`, `templates/**`, `test/e2e/**`, `docs/sdd/decimal_comma_mask/**` — no Go files (AC6).
- Update spec `Status: Done` only when ACs are met and gates pass; otherwise document what remains open.

## Risks & mitigations

- **Single translation layer:** the mask + override is the only guard between the comma-friended UI and the dot-only contract. A stale cached JS bundle, a missed wiring point, or a template variant without the mask re-exposes the 400 toasts (loud, not silent — failing exactly like today). Mitigations: attribute-gated wiring (`[data-mask="decimal"]`), the wiring points enumerated per template in Phase 2, and E2E payload interception proving translation on each surface.
- **htmx listener interplay** (existing `entryForm.js`/`departureForm.js` overrides): body-level listeners coexist; the new override writes the same `dataset.raw` the weight inputs already carry for `grossWeight`/`tare` — guard strictly on `parameters[name] !== undefined` and skip disabled inputs. Watch the entry E2E specs.
- **Programmatic writes bypass `input` events**: `updateVolumePrefill`, the farm-product row fill and any future `.value =` writes must go through `refreshMask`; a missed call shows dot decimals and — without correction — would submit dot, which the server still accepts (contract-safe failure mode).
- **Ambiguity of `"1.250"` (dot, no comma) in the mask**: the mask treats typed `.` and `,` as decimal separators on input, so `"1.250"` commits as `1.250` (dot-decimal), not thousands. Accepted client-side behavior; the pt-BR grouping appears only as reformat of already-committed digits.
- **Cloned rows**: mask re-initialization must run after `cloneNode`/`htmx.process`; delegated `focusin` fallback if per-row setup proves brittle.
- **Preview → confirm flow**: the confirm step re-posts the same fields; the override listener is one-time per page and survives swaps — verify by intercepting the confirm request in E2E.
- **E2E environment**: Docker required; `global-setup.js` spawns the app with `NFE_SEFAZ_GUARD=block` — do not seed `nfe_farm_config` (AGENTS rule 1); all covered endpoints are SEFAZ-free.
- **No JS bundling step**: assets are served statically; nothing to build. If a bundler is ever introduced later, the module has no special needs.

## Out of scope

- **Any server-side change** — no comma normalization, no parser changes, no new endpoints: the dot-decimal backend contract was established by the server and is respected as-is (per explicit user decision; comma without JS keeps failing with existing toasts).
- Refactoring `weight.js` to wrap the new module (follow-up; behavior already identical).
- `type=number` humidity/damage/impurity inputs on the entry form.
- Client-side mask on the detached **profile** editor (UI follow-up if requested; its parsers are unchanged server-side).
- Display formatting of server-rendered previews (`nfe-preview`, `nfe-detached-result`) — dot-decimal rendering there is a separate concern.
