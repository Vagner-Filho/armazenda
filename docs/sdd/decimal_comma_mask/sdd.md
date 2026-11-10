# sdd:decimal_comma_mask

## Status
Done

## Goal
The NF-e numeric inputs teach (via placeholders like `0,000`, `17,00`, `0,90`) the pt-BR comma decimal separator, but the visual teaching and the submission format disagree for many users: the backend contract is **dot-decimal only** (`xxx.xx`). A comma typed directly (`1250,90`) reaches the server unchanged and is rejected with a 400 toast ("Quantidade inválida", "Preço unitário inválido", "Falha ao identificar as alíquotas", "Informe um valor válido para …"). The page is also internally inconsistent: the client-side `parseDecimalInput` helper in `templates/pages/nfe-emit.html` handles commas, while the submission path does not.

This feature keeps the established server contract (dot-decimal in, comma rejected) and fixes the input experience purely on the client:

1. **Client mask** — a shared JS decimal mask that displays pt-BR formatted values (comma decimals, full pt-BR grouping) live in the input, keeps the canonical dot-decimal in the element's `data-raw` (the same proven pattern as `assets/js/weight.js` for the entry/departure form weights), and submits the raw dot-decimal value to the backend.
2. **Server unchanged** — no Go source changes; the backend continues to receive and parse dot-decimal strings exactly as it does today. Comma input that bypasses the mask (e.g. JS unavailable) keeps failing with the existing, loud pt-br toasts.

Placeholders stay unchanged; with the mask active they finally tell the truth. `weight.js`-formatted fields on the entry/departure forms get the identical behavior "for free" through the same attribute convention, without touching `weight.js`.

## Requirements

### Functional

**F1 — Shared decimal mask module (`assets/js/decimalMask.js`)**

- Pure, DOM-free core exported for unit tests: `formatDecimalInput(value, {decimals})` → `{formatted, raw}`:
  - `value` may contain digits, `.` and `,`; both `.` and `,` are accepted as decimal separators on input; other characters are stripped.
  - `raw` is the canonical dot-decimal string (e.g. `"1250.90"`), empty string for empty input, typed extra digits preserved beyond the display caps (matching `weight.js`).
  - `formatted` is the full pt-BR grouping display (comma decimals, `.` thousands grouping, e.g. `"1.250,90"`, `"17,00"`), empty string for empty input.
  - `raw` never contains grouping separators — it always matches the dot-decimal form the server contract expects.
- Generalizes the existing `weight.js` logic (`formatForDisplay` + `computeCursorPosition`, cursor restored after reformat); no `" kg"` unit suffix.
- DOM wiring: `setupDecimalMasks(root, signal)` attaches handlers to every `[data-mask="decimal"]` input inside `root`:
  - `keydown`: digits, `,` and `.` allowed (comma and dot interchangeable); blocks other characters while allowing control/meta combos and navigation/editing keys (same rules as `weight.js` `handleKeydown`).
  - `input`: live reformat, `dataset.raw` kept in sync, cursor position preserved.
  - `blur`: formatted display; empty/digits-only values cleared consistently.
  - `focus`/`focusin`: shows the editable raw value with comma decimals (no grouping), like `weight.js` `handleFocus`.
  - `dataset.raw` always holds the dot-decimal (this sets the `data-raw` attribute) or empty string.
- Submission: `wireDecimalParamOverride(signal)` — one body-level `htmx:configRequest` listener that, for every element matching `input[data-raw]` whose `name` exists in `evt.detail.parameters`, replaces the parameter value with `dataset.raw`. Semantics:
  - Disabled controls are untouched (htmx does not serialize them).
  - Works for dynamic item rows and for secondary forms that include mask inputs (`hx-include="#detached-nfe-form"` in the draft-save form, preview→confirm flow).
  - Same-value overrides on `grossWeight`/`tare` (entry/departure forms already carry `data-raw` via `weight.js`) are harmless.
- `refreshMask(input)` exported for programmatic value writes (they do not fire `input` events and must still produce comma display + synced `data-raw`).
- No new dependencies; no JS bundling step required (assets are served statically).

**F2 — Templates marked and wired (all four NF-e surfaces)**

- `templates/pages/nfe-emit.html` (`#detached-nfe-form`, `hx-post="/nfe/emitir/preview"`): item decimal inputs `items[i].quantity`, `items[i].unitPrice`, `items[i].discount`, `items[i].grossWeight`; transport `volPesoL`, `volPesoB`; rates `icmsRate`, `pisRate`, `cofinsRate`, `ibsRate`, `cbsRate` get `data-mask="decimal"`.
- `templates/nfe/nfe-emit-modal.html` (form `hx-post="/nfe/preview/{{ .DepartureID }}"`): `volPesoL`, `volPesoB` and the five rates get `data-mask="decimal"`; wiring added in `assets/js/form/nfe-form.js` (`setupNFeForm`).
- `templates/nfe/nfe-config-form.html`: the five rate inputs get `data-mask="decimal"` and their init script wires setup + override.
- `templates/nfe/nfe-rascunho-form.html`: the five rate inputs get `data-mask="decimal"` and their init script wires setup + override.
- Item-row clones: the `add-item` handler runs mask setup for the new row (or the mask is fully delegated).
- Programmatic writes pass through `refreshMask`: `updateVolumePrefill` (`formatPrefillNumber` assignments) and the farm-product filled row values in `templates/pages/nfe-emit.html`; `useDefaultTaxRates`-governed disabled rate inputs keep receiving server defaults (display normalized to comma at setup, dot in `data-raw`, submission values unchanged).
- Initial server-rendered values (e.g. `{{ .DefaultICMSRate }}` = `"17.00"`) are normalized to comma display at setup; `dataset.raw` keeps the dot form so submission values are unchanged (`rateDisplayString`-fed values keep working).
- Placeholders are unchanged (`0,000`, `0,00`, `17,00`, `1,65`, `7,60`, `0,10`, `0,90`).

**F3 — Server contract respected (no backend change)**

- No Go source files are modified. All existing form-decimal parsers keep their current behavior: `parsePercentRateOrNil` (`router/nfe_router/router.go`), `formDecimalRef` and `parseDetachedItems` (`router/nfe_router/detached_router.go`), the profile (`detached_profile_router.go`) and departure-override (`departure_input_router.go`, `buildNFe`/`previewNFe`) parsers — every one continues to accept **only** dot-decimal strings.
- The mask + `htmx:configRequest` override is the single translation layer: comma digits committed by the user are submitted as dot-decimal; the server never sees a comma from masked inputs.
- If the mask or override is missing (stale cached JS, unbuilt page, no script): comma input reaches the server as-is and keeps being rejected with the existing pt-br toasts — a loud failure, identical to today's behavior. No silent mutation of values.

### Non-Functional

- **No backend changes**: no Go code, no migrations, no new endpoints; the established dot-decimal submission contract is authoritative.
- **No regressions**: `weight.js` and the entry/departure forms are not modified; their behavior must remain identical (the generic `data-raw` override produces the same submission values).
- **JS required for the comma UX**: the mask is progressive UI, not a hard dependency — constraint validation, `required`, and `inputmode` semantics are unchanged, and all form fields remain plain text inputs that submit their literal value; without JS the page still works with dot input (comma keeps failing server-side, as today — documented, accepted limitation).
- **CSP**: no inline event handlers; ES module imports only (consistent with existing form modules).
- **Cleanup**: all listeners take an `AbortSignal` (consistent with `weight.js`/`entryForm.js`).
- **Performance**: the `htmx:configRequest` override is a single delegated listener; input handling is per-input with no framework cost, identical to the existing weight inputs.
- **Language**: no new user-facing strings are required in this feature (format is self-evident from the live display); any new strings must be pt-br per project rules.
- **Testability**: the mask core is pure and DOM-free so Bun unit tests need no DOM library (project pattern from `detachedNFeModes.js`).

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | `formatDecimalInput` handles comma/dot/mixed input: `"1250,90"` and `"1250.90"` and `"1.250,90"` all produce raw `"1250.90"`; display is comma-decimal with pt-BR grouping (`"1.250,90"`); empty input yields both empty; invalid chars are stripped; caps (`decimals` option: 3 for quantity/weights/volumes, 2 for prices/rates) trim display only, raw keeps typed digits. | JS unit: `test/unit/decimalMask.test.js` — `cd test && bun test unit/decimalMask.test.js` |
| AC2 | Typing `1250,90` (comma) into an item quantity / unit price shows the grouped comma display on blur, focus shows the editable comma raw, and the htmx submit carries the **dot-decimal** `items[N].quantity`/`items[N].unitPrice` parameter (payload interception) — the server contract receives dot. | E2E: `test/e2e/tests/detached_decimals.spec.js` (quantity + price rows, payload + preview totals assertions) |
| AC3 | The detached preview flow with comma-typed quantity, unit price, discount and gross weight reaches the server-side certificate gate with **neither** of the decimal parse toasts ("Quantidade inválida" / "Preço unitário inválido" / "Desconto inválido" / "Falha ao identificar as alíquotas") firing — i.e., the comma-derived dot payload parsed clean. A fully rendered 200 preview is not achievable in e2e: the fixture intentionally seeds empty certificate bytes (SEFAZ isolation), so the preview application stops at the certificate gate, which answers only after the parsers run. The full dot-decimal round-trip is proven through the draft-save path instead (AC7). | E2E: `test/e2e/tests/detached_decimals.spec.js` — preview response headers (`HX-Trigger` contains "Certificado digital", does not contain "Quantidade"/"alíquotas") + capacity assertions on the draft-save DB row (dot-decimal `'17.25'` stored) |
| AC4 | Rate fields (`icmsRate`, `pisRate`, `cofinsRate`, `ibsRate`, `cbsRate`): comma input (`"17,00"`, `"0,90"`, …) displays in comma form, `data-raw` holds the dot form, and the submission carries dot-decimal the unchanged `parsePercentRateOrNil` accepts; the rate parse toast ("Falha ao identificar as alíquotas") does not fire. The emit modal's server-rendered disabled rate defaults display comma decimals while `data-raw` keeps the dot (submission unchanged — they stay disabled/not submitted). | E2E rate rows in `detached_decimals.spec.js` + `departure_nfe_decimals.spec.js` (payload assertions + disabled-defaults display); back-end side is the unchanged contract (AC6) |
| AC5 | Item-row clones (Adicionar Item) are masked identically, and programmatic prefills (`updateVolumePrefill`, farm-product row fill, `formatPrefillNumber`) display comma format with `data-raw` synced (dot). | E2E: `detached_decimals.spec.js` (add second item, type comma, submit both; assert volume prefill display) |
| AC6 | Backend untouched: every existing form-decimal parser test keeps passing **without modification** (comma values still rejected server-side — dot-only contract intact); no Go source change in the diff. | `make test-go` (same result set as Phase 0 baseline; parsers' negative cases still fail as before, e.g. `TestParseDetachedTransport_RejectsMalformedNumber` diary) + `git diff --stat` review (no Go files) |
| AC7 | Draft save flow (`#save-rascunho-form` with `hx-include="#detached-nfe-form"`) submits comma-typed fields correctly (dot raw). | E2E: extend `test/e2e/tests/detached_nfe_rascunho.spec.js` or cover in `detached_decimals.spec.js` |
| AC8 | No behavior change for the entry/departure forms (`weight.js`, `grossWeight`/`tare`/net weights, discounts) — all existing entry/discount/departure suites stay green. | Suites: `make test-go`, `make test-js`, and existing E2E specs (`entry.spec.js`, `entry_discount.spec.js`, `departure.spec.js`) within the applicable `make test` gate |
| AC9 | Without JS (mask dead/missing), comma input is still rejected by the unchanged backend with the existing toasts — failure remains loud, dot input keeps working; no silent value coercion anywhere. | Manual check (disable JS in devtools, submit comma value): regression-only because the Go negative tests (AC6) already prove the unchanged parser behavior; automation impractical at the E2E layer without duplicating AC6 |

## Verification Results
<!-- Commands run + PASS/FAIL/NOT RUN. Feature-side gates are green; the full E2E suite carries pre-existing failures unrelated to this feature (see notes). -->

| Command | Result | Notes |
|---------|--------|-------|
| `go test ./router/nfe_router/` (Phase 0 baseline + at gate) | PASS (same result set as baseline) | 3 failures unchanged before/after: payment/billing WIP (`indPag`/`includeCobranca` — pre-existing). No Go files changed: `git diff` touches only `assets/js`, `templates`, `test/e2e` (AC6) |
| `make test-go` | PASS (same result set as baseline) | Full suite: only the same 3 pre-existing failures; parsers' negative cases still reject commas (dot-only contract intact, AC6/AC7) |
| `cd test && bun test unit/` (`make test-js`) | PASS | 76 pass / 0 fail (incl. 18 new decimalMask unit tests, no DOM library) |
| `bunx playwright test tests/detached_decimals.spec.js` (all 3 browsers) | PASS | 15/15 (chromium + firefox + webkit): group display, data-raw, payload as dot-decimals, certificate-gate toast instead of parse toasts, cloned rows, volume prefill, draft-save DB round-trip `'17.25'` |
| `bunx playwright test tests/departure_nfe_decimals.spec.js` (all 3 browsers) | PASS | 3/3: modal unitPrice/volumes comma input → dot-decimal payload → certificate-gate toast (not "Valor inválido"/"alíquotas"); server-rendered disabled rate defaults display comma decimals with dot raw |
| `bunx playwright test tests/detached_nfe_transport.spec.js tests/departure_nfe_inputs.spec.js --project=chromium` | PASS (8/10) | The 2 failures are pre-existing: transport round-trip 400 (certificate gate — fixture seeds empty certificate_data, by SEFAZ isolation) and departure_inputs timeout at `#indPag` (tests ahead of the template — matches the 3 pre-existing Go failures) |
| `make test-e2e` (full suite, all browsers) — with this feature | 145 pass / 11 skip / 45 fail | See baseline comparison below |
| `make test-e2e` — **baseline** (this feature stashed) | 131 pass / 11 skip / 41 fail | Failure set diff: ±6 items, all WebKit/Firefox click-hang & parallel-worker fixture collisions (flake class already present at baseline: tabs/cfop/validation/transport). The 6 "healed" and 6 "new" tests are untouched by this diff; my two new specs fail only inside full-parallel runs and pass single-browser. Entry/departure `netWeight` data-raw anomalies pre-exist both runs (the `kilogram mask` WIP from 4092e28) |
| `git diff --stat` review | PASS | Only `assets/js/**`, `templates/**`, `test/e2e/**`, `docs/sdd/**` — no Go files, no migrations (AC6) |

SW clearance notes recorded during development (implementation learnings, do not re-litigate):

- The certificate gate (`service/nfe_service/cert_gate.go`) requires non-empty `certificate_data`, and the e2e fixture seeds `''::bytea` on purpose (SEFAZ isolation rule). A 200 preview is therefore impossible in this environment; the decimal spec uses the certificate-gate toast (which fires *after* the form parsers) as the contract evidence, plus the draft-save DB round-trip for an end-to-end dot-decimal proof.
- Playwright 1.65's `fill` inserts at the caret and does not replace when the mask rewrites the value on focus; the deterministic test pattern is clear-then-fill (`fill('')` → `fill(value)`), which also matches human caret editing semantics.
- WebKit hangs on protocol-level `.click()` inside the emission dialog (documented in-repo in the rascunho/cfop specs); the decimal specs use `el.evaluate(el => el.click())` for tabs/buttons and `el.evaluate(el => el.focus())` for focus assertions (synthetic clicks do not focus on Firefox/WebKit).
- The main emission page's item/volume wiring lives in a classic `<script>` block; the mask helpers are bridged through `window.*` from the module block (same pattern as `window.initCfopSelectors`), with optional-call guards so load-time invocations (before the module bridge) stay no-ops.
- `parseDepartureDecimal` already accepted comma form values before this feature (its own established contract). The mask submits dot-decimals there as well, which is accepted input; its round-trip e2e expectations were updated to the echoed dot form (server code untouched).
