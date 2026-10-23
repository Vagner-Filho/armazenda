# plan:detached_nfe_tab_validation

Implementation plan for the guided-stepper validation of the detached NF-e
emission form. Spec: `docs/sdd/detached_nfe_tab_validation/sdd.md`.

## Test strategy (defined before implementation)

- **Layer**: E2E only. The behaviour is browser-level (native constraint
  validation, focus, bubbles), and the Bun unit harness
  (`test/unit/setup.js`) has no real DOM — no JS unit test is practical here.
- **Behaviours to cover** (mapped to the spec's ACs): step-gated forward
  validation with native bubble + focus (AC1–AC3), footer button visibility and
  the preview button's hidden+disabled rule (AC4), the hidden-tab backstop on
  the final submit with a no-POST assertion (AC5), live chip status (AC6), the
  rascunho request actually reaching the server (AC7) and the happy-path preview
  render (AC8).
- **New/updated spec files**:
  - `test/e2e/tests/detached_nfe_validation.spec.js` (new) — all ACs above;
    request assertions via `page.on('request')` for `/nfe/emitir/preview` and
    `/nfe/rascunho/salvar`; native bubble assertions via `validationMessage`
    plus `toBeFocused()` (the bubble itself is UA chrome, not DOM).
  - `test/e2e/tests/detached_nfe_tabs.spec.js` — re-run, expected to keep
    passing unchanged (chips remain clickable; a hidden+disabled submit button
    still satisfies `toBeDisabled()`/`toBeEnabled()`).
- **Focused commands during development**:
  `cd test/e2e && bun run test --grep "validação"`, plus
  `bun run test --grep "Tabs|rascunho"` for the regression net.
- **Before completion**: `make test-js`, `make test-e2e`, `make test-go`,
  `go build ./...`, `gofmt -l <changed Go files>` (no Go change expected), and
  `git diff --check`.
- **Environment**: Docker-backed PostgreSQL fixture on `localhost:5433`, app on
  `http://localhost:8100`, started by `test/e2e/global-setup.js` with
  `NFE_SEFAZ_GUARD=block`. Do **not** add certificate data or new
  `nfe_farm_config` rows to the fixture; the preview endpoint is SEFAZ-free.

## Phase 0 — Baseline and SDD gate

1. Run the focused Go/JS commands above where practical and record the exact
   baseline in `sdd.md` (including any pre-existing WebKit flakes).
2. Flip `sdd.md` Status to `WIP` only after the baseline is recorded.

## Phase 1 — Tab module becomes reusable

**Files:** `assets/js/detachedNFeTabs.js`, `templates/pages/nfe-emit.html`

1. Convert `assets/js/detachedNFeTabs.js` into an ES module (the page already
   imports `cfopSelector.js` the same way) exporting:
   - `TAB_ORDER = ['dest', 'transac', 'transp', 'trib']` and
     `TAB_LABELS` (pt-br: Destinatário, Transação, Transporte, Tributação);
   - the existing `tabFieldsets` map unchanged;
   - `initDetachedNFeTabs()` — idempotent wiring of the chip clicks plus the
     initial reveal (today's `active-dest` boot);
   - `showTab(id)` — reveal the fieldset group, move the `active-*` class on
     `#tabs`, set `aria-current="step"` on the active chip, apply the footer
     button rules (Voltar/Avançar visibility, preview submit `hidden` +
     `disabled` off the last step), and move focus to the first focusable
     control of the newly shown step;
   - `activeTab()`, `tabForElement(el)` (element → tab id via the owning
     fieldset) and `stepControls(id)` (the step's controls in DOM order,
     skipping non-candidates such as disabled inputs).
2. The footer-button rules move out of `handleTabClick` into `showTab`, so the
   click path and the programmatic jump share one implementation.
3. Template: remove the classic `<script src="…/detachedNFeTabs.js" defer>` tag
   and import the module from the existing inline `<script type="module">`.

**Phase gate:** the existing `detached_nfe_tabs.spec.js` (tab reveal + preview
button rule) still passes unchanged.

## Phase 2 — Step validation module

**Files:** `assets/js/detachedNFeValidation.js` (new),
`assets/static/css/nfe-emit.css` (new), `templates/pages/nfe-emit.html`

1. `initDetachedNFeValidation({ form, tabs })`:
   - `validateStep(tabId)`: iterate `stepControls(tabId)`, `checkValidity()`
     each, red-ring every invalid control, then `reportValidity()` +
     `focus()` the first invalid one; return whether the step is valid.
   - `Avançar` click → `if (validateStep(activeTab())) showTab(next)`.
   - `invalid` capture listener (the backstop for the final submit): reveal the
     owning tab of the **first** invalid control of the pass, focus it, mark it;
     cancel the remaining `invalid` events of the same pass. Idempotency: a
     per-pass guard reset with `setTimeout(…, 0)` (htmx validates a form twice
     per request, so `invalid` fires twice per control).
   - `updateStepStatus()`: per-chip `✓` / `⚠ n` from `checkValidity()` over
     `stepControls`, recomputed on delegated `input`/`change` and after each
     step change; the status text is rendered into the chip with `aria-live`.
   - marker cleanup: a delegated `input` listener removes the ring once the
     control becomes valid.
2. Template: add the stepper status slot and `aria-current` hooks to the chips,
   and the marker classes for invalid controls (a red ring that does not fight
   the existing `peer` floating-label styles).
3. **Deviation from the plan — the marker styles are hand-written**
   (`assets/static/css/nfe-emit.css`, linked after `output.css`, like
   `cfop-selector.css`): the Tailwind utilities the marker needs (`ring-2`,
   `ring-red-500`) are not in the generated `output.css`, the `tailwindcss` CLI
   is not available in this environment, and `border-white/20` would otherwise
   win over a red border (same specificity, later in the sheet, inside
   `@layer utilities`). Unlayered rules in the hand-written file take
   precedence, so the marker works without a CSS rebuild. The badge also gets a
   fixed `min-width` so it never resizes a chip under the pointer (a shifting
   chip row is what made the WebKit clicks hang under parallel load).
4. Add the E2E cases for AC1–AC3 and AC6 to the new spec.

**Phase gate:** focused `--grep "validação"` passes for the step-gated cases in
chromium/firefox/webkit.

## Phase 3 — Wizard footer, backstop and rascunho fix

**Files:** `templates/pages/nfe-emit.html`

1. Replace `#finish-container` with the sticky footer: `Voltar`, `Avançar`,
   `Gerar Pré-visualização` (keep `data-test-id="preview-detached-nfe"`),
   `Salvar como rascunho` (keep `data-test-id="open-save-rascunho"`). The
   preview button starts `hidden` + `disabled` in the markup and is only
   revealed by `showTab('trib')`.
2. Add `novalidate` to `#save-rascunho-form` so the draft always reaches the
   server (the server's existing pt-br toasts report bad data; the empty-name
   case is answered by `Nome obrigatório`).
3. Wire the modules in the inline `<script type="module">`
   (`initDetachedNFeTabs()` then `initDetachedNFeValidation()`).
4. Add the E2E cases for AC4, AC5, AC7 and AC8 to the new spec.

**Phase gate:** focused E2E green for the new spec plus `Tabs` and `rascunho`.

## Phase 4 — Recipient contract (increment)

**Files:** `templates/pages/nfe-emit.html`,
`test/e2e/tests/detached_nfe_validation.spec.js`,
`router/nfe_router/nfe_templates_test.go`

1. The server accepts a recipient either as a registered person
   (`recipientPersonId`) or as the inline name/document pair
   (`parseDetachedRecipient`); the client only enforced the second half, so
   "Selecionar existente" with nothing picked reported the step as valid.
   Extend the page script's recipient-mode handler so the person select
   carries `required` only while "Selecionar existente" is active (and the
   inline fields only while "Novo destinatário" is). The attribute is toggled
   in the script, not baked into the markup: a hidden required control blocks
   the submission with "An invalid form control ... is not focusable", and the
   no-JS fallback must stay the server toast.
2. E2E: seed a contact for the farm (a `person` row with a unique IE — the
   `(farm, ie)` index is unique — plus a `natural_person` row), assert the
   ⚠ 1 badge, the blocked `Avançar` with the select focused, and the ✓ badge
   plus the advance once a contact is picked.
3. Go: assert the recipient select renders with its id and its placeholder
   option in the stepper contract test.

## Phase 5 — Completion verification

1. Run and record in `sdd.md`: `go build ./...`, `make test-go`, `make test-js`,
   `cd test/e2e && bun run test --grep "validação"`, `make test-e2e`, and
   `git diff --check`. Record unrelated baseline failures accurately.
2. Review the diff for: no field-name/payload change, no Go change, no SEFAZ
   path, pt-br wording, and that the existing `detached_nfe_*` specs are
   untouched unless a real behaviour change required it.
3. Flip Status to `Done` only when every AC is verified; otherwise keep `WIP`
   and document what remains unverified.

## Risks / notes

- **Native validation cannot be scoped to a form subset**, so the step check
  iterates the step's controls instead of calling `form.reportValidity()`. This
  is the whole point: only visible fields are ever asked for a message.
- **Enter key**: pressing Enter on a non-final step must stay a no-op. This is
  why the preview button is `hidden` **and** `disabled`; hiding it alone makes
  Enter issue a premature POST (verified in the three engines).
- **htmx double value collection** makes `invalid` fire twice per control —
  the per-pass guard is mandatory, not defensive.
- **Firefox** does not re-fire `invalid` for untouched fields on a second
  attempt; the live chip status and the persistent ring are the reliable
  signals there.
- **Cancelling the extra `invalid` events** deviates from the spec letter (a
  cancelled `invalid` event means "not enforced"). Verified: all three engines
  still block the submission. If an engine submitted anyway, the server's toast
  is the answer — no data-integrity risk.
- **Tailwind cannot be regenerated here**: `tailwindcss` is not installed and
  the generated `output.css` is committed, so any new utility class would be
  missing at runtime. The invalid-field marker and the status badge therefore
  live in a hand-written stylesheet (`assets/static/css/nfe-emit.css`), whose
  unlayered rules take precedence over `@layer utilities`.
- **Focus move on step change** must not fight the Transporte auto-derivation
  (volume prefill, plate-UF default), which only writes values, never focus.

## Out of scope

- Enter = `Avançar` on the non-final steps (a follow-up; would need
  `formnovalidate` handling or a keydown handler).
- Loading indicator on the preview submit (htmx `hx-indicator` + spinner).
- Per-row "Item N" labels in the items grid (the server already numbers items
  in its toasts).
- Blur-only field validation (validate on blur, re-validate on input).
- Offering "Salvar como rascunho" on every step instead of the last one.
- "Calculado automaticamente" vs "personalizado" hints on the Transporte step.
- Client-side format hints for `nfRefs` (44-digit keys) and CPF/CNPJ — the
  server already reports these; duplicating its rules risks divergence.
