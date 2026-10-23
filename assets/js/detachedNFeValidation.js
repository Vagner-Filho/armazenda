/**
 * Step validation for the detached NF-e emission form
 * (templates/pages/nfe-emit.html, stepper in detachedNFeTabs.js).
 *
 * Two responsibilities:
 *
 * 1. `Avançar` validates **only the current step**, using the browser's own
 *    constraint validation (`checkValidity()` per control, `reportValidity()`
 *    on the first invalid one). Every control of the active step is visible,
 *    so the native bubble is always shown on a focusable element.
 *
 * 2. The final submit keeps the browser's whole-form validation, which runs
 *    *before* the submit event. A required control inside a non-active step is
 *    `display:none`, so the UA cannot focus it and aborts the submission with
 *    "An invalid form control ... is not focusable." — silently, with no
 *    request. The `invalid` listener below reveals the step that owns the
 *    first invalid control, which restores the native bubble and the focus.
 *
 * No client-side toast is raised for constraint violations: the native bubble
 * is the message, and the server keeps answering with its pt-br toasts.
 */

import {
    TAB_LABELS,
    TAB_ORDER,
    activeTab,
    showTab,
    stepControls,
    tabForElement,
} from './detachedNFeTabs.js';

/** Class applied to invalid controls (see assets/static/css/nfe-emit.css). */
const INVALID_CLASS = 'nfe-field-invalid';

/**
 * Wires the wizard navigation and the validation. Idempotent.
 */
export function initDetachedNFeValidation() {
    const form = document.querySelector('form#detached-nfe-form');
    const tabs = document.querySelector('#tabs');
    if (!form) return;

    // True while this module runs its own `checkValidity()` calls, which fire
    // `invalid` events per failing control. Without the flag those events
    // would reach the submit backstop and reveal tabs / steal focus while the
    // user is merely typing.
    let selfValidating = false;
    // One native validation pass fires `invalid` once per failing control
    // (htmx even collects a form's values twice per request). Only the first
    // one drives the step jump; the rest are cancelled to keep the console
    // free of "... is not focusable." noise.
    let handledInPass = false;

    /**
     * @param {() => T} fn
     * @returns {T}
     * @template T
     */
    function quietly(fn) {
        selfValidating = true;
        try {
            return fn();
        } finally {
            selfValidating = false;
        }
    }

    function markInvalid(control) {
        control.classList.add(INVALID_CLASS);
    }

    function clearInvalid(control) {
        control.classList.remove(INVALID_CLASS);
    }

    /**
     * Validates a single step with the browser's constraint validation.
     * @param {string} identifier step id.
     * @returns {boolean} true when the step has no invalid control.
     */
    function validateStep(identifier) {
        let firstInvalid = null;
        quietly(() => {
            for (const control of stepControls(identifier)) {
                if (control.checkValidity()) {
                    clearInvalid(control);
                } else {
                    markInvalid(control);
                    if (!firstInvalid) firstInvalid = control;
                }
            }
            if (firstInvalid) {
                firstInvalid.reportValidity();
                firstInvalid.focus();
            }
        });
        return !firstInvalid;
    }

    /**
     * Recomputes every chip's status badge: `✓` when the step has no invalid
     * control, `⚠ n` with the count otherwise. This is the signal users can
     * rely on before submitting anything (Firefox does not re-fire `invalid`
     * for untouched fields on a second attempt).
     */
    function updateStepStatus() {
        if (!tabs) return;
        quietly(() => {
            for (const tab of TAB_ORDER) {
                const slot = tabs.querySelector(`[data-step-status="${tab}"]`);
                if (!slot) continue;

                const invalid = stepControls(tab).filter((control) => !control.checkValidity());
                const state = invalid.length === 0 ? 'complete' : 'error';
                slot.dataset.state = state;
                slot.innerHTML = '';

                const badge = document.createElement('span');
                badge.setAttribute('aria-hidden', 'true');
                badge.textContent = state === 'complete' ? '✓' : `⚠ ${invalid.length}`;
                slot.append(badge);

                const text = document.createElement('span');
                text.className = 'nfe-sr-only';
                const label = TAB_LABELS[tab] || tab;
                text.textContent = state === 'complete'
                    ? `${label}: completo`
                    : `${label}: ${invalid.length} campo(s) pendente(s)`;
                slot.append(text);
            }
        });
    }

    // Wizard navigation: forward is validated, backward never is.
    const nextButton = form.querySelector('#wizard-next');
    if (nextButton) {
        nextButton.addEventListener('click', () => {
            const current = activeTab();
            if (!validateStep(current)) return;
            // The step just passed: refresh the badges so they agree with the
            // markers before the user moves on.
            updateStepStatus();
            const index = TAB_ORDER.indexOf(current);
            if (index >= 0 && index < TAB_ORDER.length - 1) showTab(TAB_ORDER[index + 1]);
        });
    }

    const backButton = form.querySelector('#wizard-back');
    if (backButton) {
        backButton.addEventListener('click', () => {
            const index = TAB_ORDER.indexOf(activeTab());
            if (index > 0) showTab(TAB_ORDER[index - 1]);
        });
    }

    // Backstop for the final submit (and any other whole-form validation the
    // browser performs): reveal the step owning the first invalid control so
    // the UA can show its bubble and focus it.
    form.addEventListener('invalid', (event) => {
        const control = event.target;
        if (!(control instanceof Element)) return;
        // Our own checks (step validation, status badges) fire `invalid` per
        // failing control too; they must not mark fields or jump steps.
        if (selfValidating) return;
        markInvalid(control);

        if (handledInPass) {
            event.preventDefault();
            return;
        }
        handledInPass = true;
        setTimeout(() => { handledInPass = false; }, 0);

        const tab = tabForElement(control);
        if (tab) showTab(tab);
        control.focus();
    }, true);

    // Live status plus marker cleanup as the user fixes a field.
    form.addEventListener('input', (event) => {
        const control = event.target;
        if (control instanceof Element && control.classList.contains(INVALID_CLASS)) {
            quietly(() => {
                if (control.checkValidity()) clearInvalid(control);
            });
        }
        updateStepStatus();
    });
    form.addEventListener('change', updateStepStatus);

    updateStepStatus();
}
