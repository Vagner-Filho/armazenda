# sdd:detached_nfe_tab_validation

## Status
Backlog

## Goal
Fix the detached (avulsa) NF-e emission form so a required field that lives in a
hidden tab is never silently swallowed, and turn the tab bar into a guided
stepper: `Avançar` validates the current step with the browser's own constraint
validation, `Voltar` never validates, and the single network request only
happens on the last step.

Today, clicking **Gerar Pré-visualização** while a required input sits in a
hidden tab does nothing visible: no toast, no error, no request. The user cannot
tell what is wrong, and the browser does not help.

## Background

- `templates/pages/nfe-emit.html` renders the form as four tabs
  (`#tabs` buttons `dest`, `transac`, `transp`, `trib`); each tab reveals a
  group of `<fieldset>`s and hides the rest with Tailwind's `hidden`
  (`display:none`). `assets/js/detachedNFeTabs.js` owns the reveal and also
  enables the preview submit button **only** while the `trib` tab is active.
- **Root cause of the silent failure:** the browser runs *native constraint
  validation before the `submit` event*. An invalid control inside a
  `display:none` subtree cannot be focused, so the UA aborts the submission and
  logs `An invalid form control with name='…' is not focusable.` The `submit`
  event never fires, therefore htmx never issues the POST. Verified with a
  standalone harness in Chromium 1208, Firefox 1509 and WebKit 2248 (the builds
  the e2e suite uses): `invalid` events fire per control, **no** submit event,
  **no** request.
- The `invalid` event (capture phase on the form) is the hook that can rescue
  this: if the handler reveals the fieldset that owns the invalid control, the
  control becomes focusable and the UA then shows its native validation bubble
  and focuses it, while still blocking the submission. Verified in all three
  engines — no console error, focus on the invalid input, no request.
- htmx 2.0.8 (bundled `assets/js/htmx.min.js`) collects a form's values twice
  per request, so `invalid` fires **twice per control** during one pass; any
  `invalid` handler must therefore be idempotent (a per-pass guard).
- Firefox does not re-fire `invalid` on a *second* submit attempt for fields the
  user has not touched (Chrome/WebKit do). A live, recomputed status is the
  reliable signal there.
- **Related silent failure (same family):** "Salvar como rascunho"
  (`#save-rascunho-form`, `hx-include="#detached-nfe-form"`) is silently halted
  whenever any required field of the emission form is empty — htmx validates
  the included form's elements and aborts with `htmx:validation:halted`, with no
  toast. Verified in the harness. The server already answers with proper pt-br
  toasts (`Nenhum item informado`, `Item %d: informe um CFOP…`), so the request
  should simply be allowed to reach it.
- Division of labour agreed for this SDD: **HTML constraints** (`required`,
  `type=email`) are reported by the browser's native validation bubble on the
  (now visible) field — **no client-side toast**; everything else (CPF/CNPJ
  format, interstate plate, decimal parsing, CFOP rules…) keeps being reported
  by the server through the existing `HX-Trigger` toast responses.

## Requirements

### Functional

1. **Stepper, not tabs.** `#tabs` and the button ids `dest`, `transac`,
   `transp`, `trib` are preserved (the e2e suites address them). Each chip shows
   a number, its pt-br label and a live status slot. The chips stay clickable so
   the user can navigate freely backwards *and* forwards, as today.
2. **Wizard footer.** A sticky footer inside the form replaces
   `#finish-container`:
   - `Voltar` (`type="button"`, hidden on the first step);
   - `Avançar` (`type="button"`, hidden on the last step);
   - `Gerar Pré-visualização` (`type="submit"`, the only htmx-triggered action):
     `hidden` **and** `disabled` on steps 1–3, visible and enabled on the last
     step. Keeping it `disabled` (not merely hidden) preserves today's behaviour
     where pressing Enter on a non-final step does nothing — verified in the
     three engines;
   - `Salvar como rascunho` stays available on the last step, next to the
     submit.
3. **Forward validation only.** `Avançar` validates **only the current step's**
   controls and advances only when they are all valid; `Voltar` never
   validates. Validation is delegated to the browser: `checkValidity()` per
   control of the step, `reportValidity()` + `focus()` on the first invalid one
   (native bubble, native message text).
4. **Live step status.** Every chip shows `✓` when its step has no invalid
   control and `⚠ n` with the number of invalid controls otherwise, recomputed
   on `input`/`change` and whenever the step changes. The status text is exposed
   to assistive technology (`aria-live`).
5. **Persistent field marker.** Invalid controls get a visible marker (red
   ring) that stays until the control becomes valid, so a problem can also be
   discovered in a tab the user is not looking at.
6. **Backstop on the final submit.** The final submit keeps the browser's
   native whole-form validation. An `invalid` listener (capture) reveals the
   fieldset of the **first** invalid control of the pass, focuses it and marks
   it, so the bubble is shown on a visible field instead of being swallowed.
   The remaining `invalid` events of the same pass are cancelled so the console
   stays free of `… is not focusable.` noise. The submission is still blocked by
   the UA in all three engines (verified); in the unlikely case an engine
   followed the spec literally and submitted, the server answers with a toast.
7. **No client-side toast for constraint violations.** The native bubble is the
   message; the server keeps raising toasts for everything it validates.
8. **Rascunho fix.** `#save-rascunho-form` gets `novalidate`, so htmx stops
   validating the `hx-include`d emission form and the draft always reaches the
   server, which reports bad data with its existing pt-br toasts.
9. **Focus and semantics.** Switching steps moves focus to the first focusable
   control of the newly shown step; the active chip carries `aria-current="step"`.
10. **No contract change.** Field names, the POST payload, the preview/confirm
    flow and the rascunho payload are unchanged. The existing `required`
    toggling driven by the `recipientType` radios keeps working (single form).

### Non-Functional

- All new user-facing text in pt-br (chip labels/status, `aria-live` text).
- Behaviour verified in Chromium, Firefox and WebKit (the e2e projects).
- No new dependencies, no Go changes, no migrations, no SEFAZ call: the preview
  endpoint is SEFAZ-free and the e2e app runs with `NFE_SEFAZ_GUARD=block`. No
  certificate/`nfe_farm_config` data may be added to a test database.
- Keyboard and assistive-technology support must not regress: the chips remain
  real buttons, the step status is announced, focus is managed on step change.
- The DOM contract the existing e2e suites rely on (`#tabs button#<id>`,
  `#detached-nfe-form`, `[data-test-id="preview-detached-nfe"]`,
  `[data-test-id="open-save-rascunho"]`) is preserved.

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | Step 1 with "Novo destinatário" and an empty name: `Avançar` keeps the user on step 1, the name input is focused and reports a native validation message, the Destinatário chip shows the error state, and no request is issued. | New E2E `test/e2e/tests/detached_nfe_validation.spec.js` (chromium/firefox/webkit) |
| AC2 | With the required fields filled, `Avançar` reveals the next step and its footer. | Same spec |
| AC3 | Step 2 with an empty item description/quantity/unit price: `Avançar` keeps the user on step 2, the offending input is focused and marked. | Same spec |
| AC4 | On the last step the preview submit is visible and enabled; on steps 1–3 it is hidden and disabled; `Voltar` is hidden on step 1 and `Avançar` hidden on the last step. | Same spec + existing `test/e2e/tests/detached_nfe_tabs.spec.js` keeps passing (`toBeDisabled()` still holds for a hidden+disabled button) |
| AC5 | Slip-through: a required field invalidated *after* its step was passed, then a forward jump via the chip and a submit → the owning tab is revealed, the field is focused and marked, and **no** POST to `/nfe/emitir/preview` is sent. | Same spec (request listener assertion) |
| AC6 | Live status: a step with missing required fields shows `⚠ n`; after they are filled it shows `✓`. | Same spec |
| AC7 | Rascunho with an incomplete emission form: the POST to `/nfe/rascunho/salvar` is actually sent (no silent `htmx:validation:halted`) and the server's toast is displayed. | Same spec + existing `test/e2e/tests/detached_nfe_rascunho.spec.js` stays green |
| AC8 | Happy path unaffected: walking the steps and submitting still returns the DANFE preview fragment into `#nfe-detached-preview`. | Same spec + existing `detached_nfe_*` specs |
| AC9 | No repo-wide regressions. | `make test-go`, `make test-js`, `make test-e2e`, `go build ./...`, `git diff --check` |

Manual checks: none — every criterion above is covered by automated E2E in the
three engine projects.

## Verification Results

| Command | Result | Notes |
|---------|--------|-------|
| *(baseline to be recorded when Status flips to WIP)* | | |
