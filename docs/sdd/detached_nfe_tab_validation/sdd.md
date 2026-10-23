# sdd:detached_nfe_tab_validation

## Status
Done

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

11. **The recipient is exclusive, and the client says so.** The server accepts
    a recipient either as a registered person (`recipientPersonId`) or as the
    inline name/document pair — never a mix (`parseDetachedRecipient`,
    `router/nfe_router/detached_router.go`). The client mirrors it exactly:
    while "Selecionar existente" is active the person select carries
    `required`, and while "Novo destinatário" is active the inline name and
    document do. The hidden side must never keep `required` (a hidden
    required control blocks the submission with "An invalid form control ...
    is not focusable"), so the attribute is toggled by the page script rather
    than baked into the markup — which also keeps the no-JS behaviour as it
    is today (the server toast). Consequently the Destinatário step reports
    itself incomplete (⚠ 1) until a recipient is picked, instead of claiming
    to be valid and failing later on the server.

### Non-Functional

- All new user-facing text in pt-br (chip labels/status, `aria-live` text).
- Behaviour verified in Chromium, Firefox and WebKit (the e2e projects).
- No new dependencies, no migrations, no SEFAZ call: the preview endpoint is
  SEFAZ-free and the e2e app runs with `NFE_SEFAZ_GUARD=block`. No
  certificate/`nfe_farm_config` data may be added to a test database. No Go
  production code changes (only a template-contract test in
  `router/nfe_router/nfe_templates_test.go`).
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
| AC8 | Happy path unaffected: walking the steps and submitting still reaches the preview endpoint exactly once (the DANFE fragment itself cannot be rendered in this environment — the e2e fixture seeds no certificate and `nfeCertGate` answers 400; verified on a clean tree). | Same spec (request assertion) + the existing `detached_nfe_*` specs |
| AC9 | No repo-wide regressions. | `make test-go`, `make test-js`, `make test-e2e`, `go build ./...`, `git diff --check` |
| AC10 | "Selecionar existente" with nothing picked: the Destinatário chip shows ⚠ 1, `Avançar` keeps the user on step 1 with the person select focused and a native validation message, and no request is issued; picking a contact flips the chip to ✓ and lets the step advance. "Novo destinatário" keeps its own requirement (name + document, ⚠ 2). | Same spec (new case, seeds a contact for the farm) |

Manual checks: none — every criterion above is covered by automated E2E in the
three engine projects.

## Verification Results

Baseline before implementation (2026-10-21):

| Command | Result | Notes |
|---------|--------|-------|
| `go build ./...` | PASS | |
| `make test-js` | PASS | 49 tests across 3 files |
| `make test-go` | FAIL | 3 **pre-existing, unrelated** failures in `armazenda/router/nfe_router`: `TestParseDepartureFormOverrides_PaymentBillingAndTransport`, `TestDepartureNFeModalRendersInputsAndTaxSelects` (missing `name="indPag"` / `name="includeCobranca"` in `nfe-emit-modal.html`), `TestDepartureNFePreviewCarriesInputsToConfirm`. Departure NF-e flow — untouched by this SDD. Not masked, not fixed here. |
| `cd test/e2e && bun run test --grep "Tabs"` | PASS | 6 tests (2 cases × chromium/firefox/webkit) |

Implementation verification:

| Command | Result | Notes |
|---------|--------|-------|
| `go build ./...` | PASS | |
| `node --check assets/js/detachedNFeTabs.js assets/js/detachedNFeValidation.js test/e2e/tests/detached_nfe_validation.spec.js` | PASS | |
| `make test-js` | PASS | 49 tests across 3 files (unchanged — no JS unit tests added; the behaviour is browser-level) |
| `go test ./router/nfe_router/ -run TestNFeEmitPageRendersStepperContract` | PASS | new structural guard for the stepper markup contract |
| `make test-go` | FAIL | same 3 **pre-existing** failures as the baseline, no new ones |
| `cd test/e2e && bun run test detached_nfe_validation` | PASS | 21 tests (7 cases × chromium/firefox/webkit) after the recipient increment; the earlier 18 also passed repeatedly, also under 3 workers |
| `cd test/e2e && bun run test --grep "stepper\|Tabs"` | PASS | 24 tests (the new spec + the existing tabs spec), 3 workers |
| `cd test/e2e && bun run test --grep "Tabs\|Rascunho de NF-e\|CFOP\|Transporte"` | FAIL | identical failure set to the clean-tree baseline (see below) |
| `cd test/e2e && bun run test --grep "Tabs\|Rascunho de NF-e\|CFOP\|Transporte\|stepper"` (after the recipient increment) | FAIL | the failure set is **identical** to the same command's baseline before the increment — no new failure, none removed |
| `make test-e2e` | FAIL | 121 passed, 4 skipped, 28 failed — the same 28 failures as the clean-tree baseline (which scored 103 passed / 28 failed); the only difference is two flaky entries swapping sides. The new spec is **18/18 green** in this run. |
| `git diff --check` | PASS | no whitespace errors |

### Notes on the E2E results

- **Full-suite comparison (clean tree vs. this change, same command
  `bun run test`):** 28 failures on the clean tree, 28 with this change, with
  two flaky entries swapping sides (`detached_nfe_transport:131` /
  `detached_cfop_selector:305` in the baseline against `detached_nfe_tabs:56`
  and `login:83` here). The new spec passes 18/18 in the full run.
- **Pre-existing failures, confirmed by a clean-tree run of the same commands**
  (changes stashed, same Playwright projects): `detached_cfop_selector` (111,
  207, 305), `detached_nfe_rascunho` (104, 232), `detached_nfe_transport` (131,
  162, 215) and `detached_nfe_tabs` (34, 56) in WebKit, plus the
  preview-dependent cases in chromium/firefox and `departure_nfe_inputs` (the
  departure modal is missing `indPag`/`includeCobranca` — the same defect the
  three failing Go tests report). None of them are caused by this change.
- **The preview endpoint cannot succeed in this environment**: the e2e fixture
  seeds `nfe_farm_config` with an empty `certificate_data`, and
  `nfeCertGate` (`service/nfe_service/cert_gate.go`) rejects the preview with
  400 "Certificado digital não configurado". Verified on a clean tree with the
  existing `detached_inf_ad_prod.spec.js`, which fails identically. Seeding
  certificate bytes is prohibited by the SEFAZ isolation rules, so AC8 is
  verified as far as the environment allows: the last step's submit issues
  exactly one POST to `/nfe/emitir/preview`, and the request is no longer
  blocked client-side.
- **WebKit parallel-load flakes**: the same specs fail intermittently under
  3-worker runs on a clean tree (the chip and CFOP-trigger clicks hang in
  "performing click action"). The new spec uses `evaluate`-clicks for the
  icon-bearing footer buttons, the stepper chips and the CFOP trigger (the
  pattern the other detached NF-e specs already use), and the status badge has
  a fixed `min-width` so it never resizes a chip under the pointer.
- **Operational note**: `test/e2e/global-teardown.js` deletes the tracked
  `test/e2e/.test-server` binary on every E2E run. Restore it with
  `git checkout -- test/e2e/.test-server` after running the suite.
- **A manually started dev app occupies port 8100** while this was written;
  the E2E suite then silently runs against it (login fails). The suite honours
  `PORT`/`BASE_URL`, so it can be run aside with
  `PORT=8110 BASE_URL=http://localhost:8110 bun run test …`.
