/**
 * pt-BR decimal input mask for htmx forms.
 *
 * The backend contract receives dot-decimal numbers ("1250.90") and rejects
 * the comma separator ("1250,90") with its existing 400 toasts. The
 * placeholders of the NF-e forms, however, teach the comma ("0,000",
 * "17,00"), so this module is the single translation layer between what the
 * user sees and what htmx submits:
 *
 * - DISPLAY: pt-BR comma decimals with full pt-BR grouping ("1.250,90"),
 *   same visual language the `weight.js` inputs already use on the
 *   entry/departure forms.
 * - RAW: the canonical dot-decimal string ("1250.90"), always free of
 *   grouping separators, stored in `dataset.raw` (sets the `data-raw`
 *   attribute) exactly like `weight.js` does.
 * - SUBMISSION: `wireDecimalParamOverride()` replaces matching htmx
 *   parameters with the raw value so the dot-only contract is respected.
 *
 * If this module never runs (stale cached JS, no script), the display value
 * reaches the server as-is and the existing toasts fire — a loud failure,
 * never a silent value mutation.
 *
 * The core functions are DOM-free and unit-testable (same pattern as
 * `assets/js/detachedNFeModes.js`); the cursor math is generalized from
 * `assets/js/weight.js`.
 *
 * No new dependencies; assets are served statically (no bundling step).
 */

/** Cached Intl formatters per decimals cap. */
const formatters = new Map();

function displayFormatter(decimals) {
    let formatter = formatters.get(decimals);
    if (!formatter) {
        formatter = new Intl.NumberFormat('pt-BR', { maximumFractionDigits: decimals });
        formatters.set(decimals, formatter);
    }
    return formatter;
}

/**
 * Parses a masked/editor value into digits around the decimal separator.
 *
 * Dots and commas are interchangeable as separators, so dots are unified to
 * commas first (this also makes the grouping dots of a committed display —
 * e.g. re-parsing "1.250,90" after an edit — inert the same way the
 * `weight.js` display strings are). The LAST comma is the decimal separator;
 * earlier separators are grouping noise.
 *
 * @param {string} value raw input value (may contain grouping separators).
 * @returns {{intDigits: string, fracDigits: string, hasSep: boolean, trailing: boolean}}
 */
function parseMaskedValue(value) {
    const text = (value == null ? '' : String(value)).replace(/\./g, ',');
    const lastSep = text.lastIndexOf(',');
    let intDigits = '';
    let fracDigits = '';
    if (lastSep === -1) {
        for (const ch of text) {
            if (ch >= '0' && ch <= '9') intDigits += ch;
        }
        return { intDigits, fracDigits, hasSep: false, trailing: false };
    }
    for (let i = 0; i < text.length; i++) {
        const ch = text[i];
        if (ch >= '0' && ch <= '9') {
            if (i < lastSep) {
                intDigits += ch;
            } else {
                fracDigits += ch;
            }
        }
    }
    return { intDigits, fracDigits, hasSep: true, trailing: fracDigits === '' };
}

/** Groups thousands with dots: "12500" → "12.500". */
function groupIntDigits(intDigits) {
    return intDigits.replace(/\B(?=(\d{3})+(?!\d))/g, '.');
}

/**
 * Formats a numeric editor value for display and keeps the canonical raw.
 *
 * Live mode (default) keeps the user's committed decimal point visible
 * while typing ("1250," stays "1250,"). Final mode (`final: true`, used on
 * blur) drops the trailing separator. Display decimals are capped at
 * `decimals` (truncated); the raw keeps every typed digit.
 *
 * @param {string|number|null} value - the value to format.
 * @param {{decimals?: number, final?: boolean}} [options]
 * @returns {{formatted: string, raw: string}} display string + dot-decimal raw.
 *   Both strings are empty when the value has no digits.
 */
export function formatDecimalInput(value, options = {}) {
    const decimals = options.decimals ?? 3;
    const final = options.final ?? false;
    const { intDigits, fracDigits, hasSep, trailing } = parseMaskedValue(value);

    if (intDigits === '' && fracDigits === '') {
        return { formatted: '', raw: '' };
    }

    // Display integer part without leading zeros (single zero kept).
    const intDisplay = intDigits.replace(/^0+(?=\d)/, '');
    const grouping = groupIntDigits(hasSep && intDisplay === '' ? '0' : intDisplay);

    let fracDisplay = fracDigits;
    if (decimals != null && fracDisplay.length > decimals) {
        fracDisplay = fracDisplay.slice(0, decimals);
    }

    let formatted;
    if (hasSep && !trailing) {
        formatted = `${grouping},${fracDisplay}`;
    } else if (hasSep && trailing && !final) {
        formatted = `${grouping},`;
    } else {
        formatted = grouping;
    }

    const raw = hasSep && !trailing ? `${intDigits || '0'}.${fracDigits}` : intDigits;
    return { formatted, raw };
}

/**
 * Computes the new cursor position after reformatting by matching the number
 * of significant characters (digits + decimal separator) before the old cursor.
 * Generalized from `assets/js/weight.js`.
 *
 * @param {string} oldValue
 * @param {number} oldCursor
 * @param {string} newValue
 * @returns {number}
 */
export function computeCursorPosition(oldValue, oldCursor, newValue) {
    if (oldCursor <= 0) return 0;

    // Count significant chars (digits and the decimal separator) before old cursor
    let significantCount = 0;
    for (let i = 0; i < oldCursor && i < oldValue.length; i++) {
        if (/[0-9]/.test(oldValue[i]) || oldValue[i] === ',' || oldValue[i] === '.') {
            significantCount++;
        }
    }

    // Find the position in newValue where we've passed the same number of significant chars
    let count = 0;
    for (let i = 0; i < newValue.length; i++) {
        if (/[0-9]/.test(newValue[i]) || newValue[i] === ',' || newValue[i] === '.') {
            count++;
        }
    }

    if (count >= significantCount) {
        return count;
    }

    return newValue.length;
}

/** Inputs already wired by this module (cloned rows must not double-bind). */
const boundInputs = new WeakSet();

/** Targets that already have the htmx parameter override wired. */
const boundTargets = new WeakSet();

/**
 * Reads the mask options from an input's dataset (`data-decimals`).
 *
 * @param {HTMLInputElement} input
 * @returns {{decimals: number}}
 */
function maskOptions(input) {
    const attr = input.dataset.decimals;
    const decimals = attr == null || attr === '' ? 3 : Number(attr);
    return { decimals: Number.isFinite(decimals) ? decimals : 3 };
}

/**
 * Recomputes the display and the `data-raw` of a single masked input from its
 * current value (no cursor bookkeeping). Exported for programmatic value
 * writes, which do not fire `input` events.
 *
 * @param {HTMLInputElement} input
 */
export function refreshMask(input) {
    if (!input || !boundInputs.has(input)) return;
    const result = formatDecimalInput(input.value, maskOptions(input));
    input.dataset.raw = result.raw;
    if (result.formatted !== input.value) {
        input.value = result.formatted;
    }
}

/** Re-run the display formatting without moving the caret (blur-style). */
function formatFinal(input) {
    const hasDigits = /[\d]/.test(input.value || '');
    if (!hasDigits) {
        input.value = '';
        input.dataset.raw = '';
        return;
    }
    const result = formatDecimalInput(input.value, { ...maskOptions(input), final: true });
    input.value = result.formatted;
    input.dataset.raw = result.raw;
}

/** Show the editable comma-decimal raw (no grouping) on focus. */
function showRaw(maskedInput) {
    const raw = maskedInput.dataset.raw;
    if (raw && raw !== '') {
        maskedInput.value = raw.replace('.', ',');
    }
}

/**
 * Wires a single mask input. Registering the handlers per element keeps the
 * pagination markers of the surrounding forms (`required`, validation) intact.
 *
 * @param {HTMLInputElement} input
 * @param {AbortSignal} [signal] cleanup signal from the host form setup.
 */
function setupDecimalMaskInput(input, signal) {
    if (!input || boundInputs.has(input)) return;
    boundInputs.add(input);

    function handleInput() {
        const oldValue = input.value;
        const oldCursor = input.selectionStart || 0;

        const result = formatDecimalInput(oldValue, maskOptions(input));
        input.dataset.raw = result.raw;

        // Skip if nothing changed (prevents cursor jumps and loops)
        if (result.formatted === oldValue) return;

        input.value = result.formatted;
        const newCursor = computeCursorPosition(oldValue, oldCursor, result.formatted);
        input.setSelectionRange(newCursor, newCursor);
    }

    function handleBlur() {
        formatFinal(input);
    }

    function handleKeydown(e) {
        // Allow control/meta combinations (copy, paste, select all, etc.)
        if (e.ctrlKey || e.metaKey) return;
        // Allow navigation and editing keys
        if (['Backspace', 'Delete', 'Tab', 'Escape', 'Enter', 'ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End'].includes(e.key)) {
            return;
        }
        // Allow comma or dot for the decimal separator (treat dot same as comma);
        // block once the editor already committed one.
        if (e.key === ',' || e.key === '.') {
            if ((input.value || '').includes(',')) {
                e.preventDefault();
            }
            return;
        }
        // Allow digits
        if (e.key >= '0' && e.key <= '9') return;
        // Block everything else
        e.preventDefault();
    }

    input.addEventListener('input', handleInput, signal ? { signal } : undefined);
    input.addEventListener('blur', handleBlur, signal ? { signal } : undefined);
    // showRaw is a module-scoped helper taking the element; wrap it so the
    // focus handler binds the input (an addEventListener'd function receives
    // the event object, not the element).
    input.addEventListener('focus', () => showRaw(input), signal ? { signal } : undefined);
    input.addEventListener('keydown', handleKeydown, signal ? { signal } : undefined);

    // Normalize any server-prefilled value ("17.00" → display "17,00",
    // data-raw keeps the dot form). Always recomputed so cleansed clear the
    // stale clone data (cloned rows carry data-raw attributes).
    const initial = formatDecimalInput(input.value, maskOptions(input));
    input.dataset.raw = initial.raw;
    input.value = initial.formatted;
}

/**
 * Wires every `input[data-mask="decimal"]` inside `root`. Idempotent per
 * element (cloned rows bind fresh; pre-bound inputs are skipped).
 *
 * @param {Element|Document} [root=document]
 * @param {AbortSignal} [signal] cleanup signal.
 */
export function setupDecimalMasks(root, signal) {
    const scope = root ?? document;
    scope.querySelectorAll('input[data-mask="decimal"]').forEach(input => {
        setupDecimalMaskInput(input, signal);
    });
}

/**
 * Replaces matching htmx request parameters with the masked inputs' canonical
 * raw (dot-decimal) values. One delegated listener per document covers every
 * form that carries masked inputs, including secondary forms that pull them
 * in with `hx-include` (e.g. the draft-save form) and dynamically cloned rows.
 *
 * Matching is strict: a parameter is overridden only by the exact input whose
 * visible value produced it, so disabled inputs (never serialized) and
 * unrelated forms are not touched.
 *
 * @param {Window|Document|Element} [target=document] — event target to bind.
 */
export function wireDecimalParamOverride(target = document) {
    if (boundTargets.has(target)) return;
    boundTargets.add(target);

    target.addEventListener('htmx:configRequest', (evt) => {
        const params = evt.detail?.parameters;
        if (!params) return;

        target.querySelectorAll('input[data-raw]').forEach(input => {
            const name = input.name;
            if (!name) return;
            const current = params[name];
            if (typeof current !== 'string') return;
            if (current !== input.value) return;
            if (typeof input.dataset.raw !== 'string') return;
            params[name] = input.dataset.raw;
        });
    });
}
