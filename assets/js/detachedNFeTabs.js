/**
 * Detached NF-e emission form stepper (templates/pages/nfe-emit.html).
 *
 * The emission form is a guided stepper: exactly one step's fieldset group is
 * visible at a time and the footer carries Voltar / Avançar. The preview
 * submit button only exists on the last step — it is hidden *and* disabled
 * elsewhere, so pressing Enter on a non-final step stays a no-op (hiding it
 * alone would make Enter issue a premature POST).
 *
 * Exposed for the validation module (detachedNFeValidation.js), which needs to
 * reveal a step from an `invalid` event and to know which controls belong to
 * the active step.
 */

/**
 * Step order, also the tab button ids in #tabs.
 * @type {readonly string[]}
 */
export const TAB_ORDER = ['dest', 'transac', 'transp', 'trib'];

/**
 * Step labels, used for status text (pt-br).
 * @type {Readonly<Record<string, string>>}
 */
export const TAB_LABELS = {
    dest: 'Destinatário',
    transac: 'Transação',
    transp: 'Transporte',
    trib: 'Tributação',
};

/**
 * Fieldset group per step. The pagamento (detPag/tPag) and entrega
 * (<entrega>) fieldsets were introduced by the NF-e emission wiring
 * (docs/sdd/detached_nfe_input_wiring): payment is transaction data, so it
 * shares the Transação step, and the delivery site belongs with the parties,
 * sharing the Destinatário step.
 * @type {Readonly<Record<string, readonly string[]>>}
 */
const TAB_FIELDSETS = {
    dest: ['dest', 'entrega-fieldset'],
    transac: ['transac', 'pagamento-fieldset'],
    transp: ['transporte-fieldset'],
    trib: ['trib'],
};

/** Id of the preview submit button (the only htmx-triggered action). */
const PREVIEW_BUTTON_SELECTOR = 'button[type="submit"]#preview-nfe';

let tabs = null;
let form = null;

/**
 * @returns {string} the active step id, or '' when no step is active.
 */
export function activeTab() {
    if (!tabs) return '';
    const active = Array.from(tabs.classList).find((cls) => cls.startsWith('active-'));
    return active ? active.replace('active-', '') : '';
}

/**
 * Maps a form control to the step that owns it, via its closest fieldset.
 * @param {Element|null} element any node inside the emission form.
 * @returns {string|null} the step id, or null when the element is outside the
 * stepper (or inside no fieldset).
 */
export function tabForElement(element) {
    const fieldset = element && element.closest ? element.closest('fieldset') : null;
    if (!fieldset || !fieldset.id) return null;
    for (const [tab, fieldsetIds] of Object.entries(TAB_FIELDSETS)) {
        if (fieldsetIds.includes(fieldset.id)) return tab;
    }
    return null;
}

/**
 * Controls of a step, in DOM order, skipping controls the browser never
 * validates (disabled, unnamed, submit buttons).
 * @param {string} identifier step id.
 * @returns {HTMLElement[]} the step's candidate controls.
 */
export function stepControls(identifier) {
    const scope = form || document;
    const controls = [];
    for (const fieldsetId of TAB_FIELDSETS[identifier] || []) {
        const fieldset = scope.querySelector(`fieldset#${fieldsetId}`);
        if (!fieldset) continue;
        for (const control of fieldset.querySelectorAll('input, select, textarea')) {
            if (control.willValidate) controls.push(control);
        }
    }
    return controls;
}

/**
 * Reveals the fieldset group of a step, hides every other fieldset, marks the
 * step chip as current and applies the footer button rules.
 * @param {string} identifier step id.
 * @param {{focus?: boolean}} [options] when focus is false the first control
 * of the step is not focused (used for the initial reveal).
 */
export function showTab(identifier, options = {}) {
    if (!tabs) return;
    const focus = options.focus !== false;

    for (const cls of Array.from(tabs.classList)) {
        if (cls.startsWith('active-')) tabs.classList.remove(cls);
    }
    tabs.classList.add(`active-${identifier}`);

    const fieldsets = form
        ? form.querySelectorAll('fieldset')
        : document.querySelectorAll('form#detached-nfe-form fieldset');
    for (const fieldset of fieldsets) {
        fieldset.classList.add('hidden');
    }
    for (const fieldsetId of TAB_FIELDSETS[identifier] || []) {
        const fieldset = form
            ? form.querySelector(`fieldset#${fieldsetId}`)
            : document.querySelector(`form#detached-nfe-form fieldset#${fieldsetId}`);
        if (fieldset) fieldset.classList.remove('hidden');
    }

    // Chip semantics: the active step is announced as the current one.
    for (const button of tabs.querySelectorAll('button')) {
        if (button.id === identifier) {
            button.setAttribute('aria-current', 'step');
        } else {
            button.removeAttribute('aria-current');
        }
    }

    applyFooterState(identifier);

    if (focus) {
        const first = stepControls(identifier).find((control) => !control.disabled);
        if (first && typeof first.focus === 'function') first.focus();
    }
}

/**
 * Footer rules: Voltar is useless on the first step, Avançar on the last one,
 * and the preview submit only exists on the last step. Keeping it disabled
 * (not merely hidden) preserves the "Enter does nothing" behaviour of the
 * non-final steps.
 * @param {string} identifier active step id.
 */
function applyFooterState(identifier) {
    if (!form) return;
    const isFirst = TAB_ORDER.indexOf(identifier) === 0;
    const isLast = TAB_ORDER.indexOf(identifier) === TAB_ORDER.length - 1;

    const back = form.querySelector('#wizard-back');
    if (back) back.classList.toggle('hidden', isFirst);

    const next = form.querySelector('#wizard-next');
    if (next) next.classList.toggle('hidden', isLast);

    const preview = form.querySelector(PREVIEW_BUTTON_SELECTOR);
    if (preview) {
        preview.classList.toggle('hidden', !isLast);
        if (isLast) {
            preview.removeAttribute('disabled');
        } else {
            preview.setAttribute('disabled', 'true');
        }
    }
}

/**
 * Wires the stepper. Idempotent: a second call is a no-op.
 */
export function initDetachedNFeTabs() {
    tabs = document.querySelector('#tabs');
    form = document.querySelector('form#detached-nfe-form');
    if (!tabs || !form) return;

    for (const button of tabs.querySelectorAll('button')) {
        button.addEventListener('click', () => {
            showTab(button.id);
        });
    }

    // The template ships with `active-dest` on the tabs bar; apply the same
    // reveal on load so co-tabbed fieldsets (entrega, pagamento) are reachable
    // without a first click — their own markup starts hidden.
    showTab(activeTab() || TAB_ORDER[0], { focus: false });
}
