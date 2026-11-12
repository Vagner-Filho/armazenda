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
 * - TYPING: the comma is the only typable decimal separator — the keydown
 *   filter blocks the dot. Dots reach the field only through paste, where
 *   the parser digests them (as decimal or grouping) instead of choking.
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

export { parseMaskedValue, groupIntDigits, splitRaw };

/**
 * Canonical raw of the tracked input state: digits with at most one decimal
 * point ("47148", "47148.", "47148.55"). A trailing point means the user
 * committed a decimal separator whose fraction is still empty.
 */

function digitsOnly(text) {
    let out = '';
    for (const ch of text) {
        if (ch >= '0' && ch <= '9') out += ch;
    }
    return out;
}

/** Splits a canonical raw into its digit parts. */
function splitRaw(raw) {
    const text = raw == null ? '' : String(raw);
    const point = text.indexOf('.');
    if (point === -1) {
        return { intDigits: digitsOnly(text), fracDigits: '', hasPoint: false };
    }
    return {
        intDigits: digitsOnly(text.slice(0, point)),
        fracDigits: digitsOnly(text.slice(point + 1)),
        hasPoint: true,
    };
}

/** Joins the digit parts back into a canonical raw. */
function buildRaw(intDigits, fracDigits, hasPoint) {
    let int = intDigits;
    if (hasPoint && int === '') int = '0';
    if (int === '' && fracDigits === '') return '';
    if (!hasPoint) return int;
    return `${int}.${fracDigits}`;
}

/** Number of digits before position p of value. */
function digitCountBefore(value, p) {
    let n = 0;
    const end = Math.min(p, value.length);
    for (let i = 0; i < end; i++) {
        if (value[i] >= '0' && value[i] <= '9') n++;
    }
    return n;
}

/**
 * Full parse of a value that did not come from this mask's own rendering
 * (pasted text, a Playwright/multi-char fill, a programmatic write):
 *
 * - a committed comma makes the LAST comma the decimal separator and every
 *   dot grouping noise ("47.148,55" → 47148.55);
 * - a comma-less value matching strict pt-BR grouping (groups of exactly
 *   three) is grouping noise ("50.000" → 50000 — the mask's own display and
 *   pt-BR-formatted pastes round-trip);
 * - otherwise the LAST dot is the decimal separator ("47148.55" → 47148.55).
 *
 * The dot-only cases are inherently ambiguous; the strict-3 rule resolves
 * them towards the pt-BR display convention, which is the common one.
 *
 * @param {string|number|null} value
 * @returns {string} canonical raw ("" when there are no digits).
 */
export function parseToRaw(value) {
    const cleaned = String(value ?? '').replace(/[^0-9.,]/g, '');
    if (cleaned === '') return '';
    if (cleaned.includes(',')) {
        const { intDigits, fracDigits, hasSep } = parseMaskedValue(cleaned);
        if (!hasSep) return intDigits;
        return `${intDigits || '0'}.${fracDigits}`;
    }
    if (/^\d{1,3}(?:\.\d{3})+$/.test(cleaned)) {
        return digitsOnly(cleaned);
    }
    const { intDigits, fracDigits, hasSep } = parseMaskedValue(cleaned);
    if (!hasSep) return intDigits;
    return `${intDigits || '0'}.${fracDigits}`;
}

/**
 * Applies one input event to the tracked mask state without re-parsing the
 * mask's own display.
 *
 * The mask renders formatted values (grouping dots, unit suffixes) into the
 * field, and re-parsing that display is ambiguous — a grouping dot of
 * "50.000" would read back as the decimal 50.000. Instead, the caller
 * remembers the exact string it last rendered (`prevValue`) together with its
 * canonical raw (`prevRaw`), and this function:
 *
 * - single-character edits (typing and one-char pastes): locate the edit by
 *   diffing against `prevValue`, map its position into the raw's digit
 *   stream, and apply it there;
 * - larger edits (multi-char pastes/fills, selection replacements, clears):
 *   fall back to the full `parseToRaw` of the new value.
 *
 * A comma keystroke commits the decimal point at its digit position (a
 * trailing point marks an empty fraction); deleting the display comma
 * un-commits it; deleting a grouping dot or unit text leaves the raw alone.
 *
 * @param {string} prevRaw canonical raw of the previously rendered value.
 * @param {string} prevValue exact string previously rendered into the field.
 * @param {string} newValue the field's value after the edit.
 * @returns {{raw: string, caretDigits: number}} new canonical raw plus the
 *   number of digits that sit before the caret after the edit (feed it to
 *   `caretAfterDigit` once the new display is rendered).
 */
export function applyInputToRaw(prevRaw, prevValue, newValue) {
    prevValue = prevValue == null ? '' : String(prevValue);
    newValue = newValue == null ? '' : String(newValue);

    // Common prefix/suffix diff: the edit lives between them.
    let start = 0;
    const maxStart = Math.min(prevValue.length, newValue.length);
    while (start < maxStart && prevValue[start] === newValue[start]) start++;
    let endPrev = prevValue.length;
    let endNew = newValue.length;
    while (endPrev > start && endNew > start && prevValue[endPrev - 1] === newValue[endNew - 1]) {
        endPrev--;
        endNew--;
    }
    const removed = prevValue.slice(start, endPrev);
    const inserted = newValue.slice(start, endNew);

    const { intDigits, fracDigits, hasPoint } = splitRaw(prevRaw);

    if (inserted.length === 1 && removed === '') {
        const k = digitCountBefore(prevValue, start);
        const ch = inserted;
        if (ch >= '0' && ch <= '9') {
            if (!hasPoint || k < intDigits.length) {
                const newInt = intDigits.slice(0, k) + ch + intDigits.slice(k);
                return { raw: buildRaw(newInt, fracDigits, hasPoint), caretDigits: k + 1 };
            }
            const fi = k - intDigits.length;
            const newFrac = fracDigits.slice(0, fi) + ch + fracDigits.slice(fi);
            return { raw: buildRaw(intDigits, newFrac, true), caretDigits: k + 1 };
        }
        if (ch === ',' || ch === '.') {
            const allDigits = intDigits + fracDigits;
            return { raw: buildRaw(allDigits.slice(0, k), allDigits.slice(k), true), caretDigits: k };
        }
        return { raw: prevRaw, caretDigits: k };
    }

    if (inserted === '' && removed.length === 1) {
        const removedChar = removed;
        const k = digitCountBefore(prevValue, start);
        if (removedChar >= '0' && removedChar <= '9') {
            if (!hasPoint || k < intDigits.length) {
                const newInt = intDigits.slice(0, k) + intDigits.slice(k + 1);
                return { raw: buildRaw(newInt, fracDigits, hasPoint), caretDigits: k };
            }
            const fi = k - intDigits.length;
            const newFrac = fracDigits.slice(0, fi) + fracDigits.slice(fi + 1);
            return { raw: buildRaw(intDigits, newFrac, true), caretDigits: k };
        }
        if (removedChar === ',') {
            // Un-commit the decimal: the digits stay in order.
            const allDigits = intDigits + fracDigits;
            return { raw: buildRaw(allDigits, '', false), caretDigits: k };
        }
        // Grouping dots and unit-suffix text are not part of the raw.
        return { raw: prevRaw, caretDigits: k };
    }

    const fullRaw = parseToRaw(newValue);
    return { raw: fullRaw, caretDigits: splitRaw(fullRaw).intDigits.length + splitRaw(fullRaw).fracDigits.length };
}

/** Position just after the n-th digit of a display value (end when fewer). */
function positionAfterDigit(display, n) {
    let seen = 0;
    for (let i = 0; i < display.length; i++) {
        if (display[i] >= '0' && display[i] <= '9') {
            seen++;
            if (seen === n) return i + 1;
        }
    }
    return display.length;
}

/**
 * Caret position after reformatting: keep the caret just after the same
 * digit index the edit touched (or at the end for appends). Counting digits
 * instead of digits-plus-separators keeps the caret from drifting one slot
 * earlier per grouping dot; a separator immediately following the caret
 * (e.g. the just-typed decimal comma) is stepped over.
 *
 * @param {string} newDisplay the freshly rendered display value.
 * @param {number} digitIndex number of digits that sit before the caret.
 * @returns {number}
 */
export function caretAfterDigit(newDisplay, digitIndex) {
    if (digitIndex <= 0) return 0;
    let pos = positionAfterDigit(newDisplay, digitIndex);
    while (pos < newDisplay.length && (newDisplay[pos] === ',' || newDisplay[pos] === '.')) pos++;
    return pos;
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
 * Per-input tracked state: exactly what the mask last rendered into the
 * field (`value`) and its canonical raw (`raw`, which may carry a trailing
 * point for a committed-but-empty fraction). The WeakMap keeps cloned rows
 * and dynamic forms independent.
 */
const inputState = new WeakMap();

/** The submitted `data-raw` never carries a trailing point ("47148."). */
function submissionRaw(raw) {
    const point = String(raw).indexOf('.');
    if (point !== -1 && String(raw).slice(point + 1) === '') return String(raw).slice(0, point);
    return String(raw);
}

/**
 * Recomputes the display and the `data-raw` of a single masked input.
 * Exported for programmatic value writes, which do not fire `input` events:
 * when the field's value differs from what the mask last rendered, the
 * written value is parsed fresh (it did not come from the mask's own
 * rendering); otherwise the tracked raw is re-rendered.
 *
 * @param {HTMLInputElement} input
 * @param {{final?: boolean}} [options] finalize (blur) drops a trailing separator.
 */
export function refreshMask(input, options = {}) {
    if (!input || !boundInputs.has(input)) return;
    const state = inputState.get(input) || { raw: '', value: '' };
    let raw = state.raw;
    if (input.value !== state.value) {
        raw = parseToRaw(input.value);
    }
    if (raw === '') {
        input.value = '';
        input.dataset.raw = '';
        inputState.set(input, { raw: '', value: '' });
        return;
    }
    const result = formatDecimalInput(raw, { ...maskOptions(input), ...options });
    input.dataset.raw = result.raw;
    if (result.formatted !== input.value) {
        input.value = result.formatted;
    }
    inputState.set(input, { raw, value: result.formatted });
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
    inputState.set(input, { raw: '', value: '' });

    function handleInput() {
        const newValue = input.value;
        const state = inputState.get(input);

        // Diff against exactly what the mask last rendered — the mask's own
        // display is never re-parsed (a grouping dot of "50.000" would read
        // back as the decimal 50.000).
        const { raw: newRaw, caretDigits } = applyInputToRaw(state.raw, state.value, newValue);
        state.raw = newRaw;

        const result = formatDecimalInput(newRaw, maskOptions(input));
        input.dataset.raw = result.raw;

        if (result.formatted === '') {
            input.value = '';
            state.value = '';
            return;
        }

        if (result.formatted !== newValue) {
            input.value = result.formatted;
        }
        const caret = caretAfterDigit(result.formatted, caretDigits);
        if (typeof input.setSelectionRange === 'function') {
            input.setSelectionRange(caret, caret);
        }
        state.value = result.formatted;
    }

    function handleBlur() {
        const state = inputState.get(input);
        if (state.raw === '') {
            input.value = '';
            input.dataset.raw = '';
            state.value = '';
            return;
        }
        // Finalize from the tracked raw — never re-parse the rendered display.
        const result = formatDecimalInput(state.raw, { ...maskOptions(input), final: true });
        input.value = result.formatted;
        input.dataset.raw = result.raw;
        state.value = result.formatted;
    }

    function handleKeydown(e) {
        // Allow control/meta combinations (copy, paste, select all, etc.)
        if (e.ctrlKey || e.metaKey) return;
        // Allow navigation and editing keys
        if (['Backspace', 'Delete', 'Tab', 'Escape', 'Enter', 'ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End'].includes(e.key)) {
            return;
        }
        // The comma is the only typable decimal separator; block once the
        // editor already committed one. The dot is a paste-only separator
        // (digested by the parser), so typing it is always blocked.
        if (e.key === ',') {
            if ((input.value || '').includes(',')) {
                e.preventDefault();
            }
            return;
        }
        if (e.key === '.') {
            e.preventDefault();
            return;
        }
        // Allow digits
        if (e.key >= '0' && e.key <= '9') return;
        // Block everything else
        e.preventDefault();
    }

    input.addEventListener('input', handleInput, signal ? { signal } : undefined);
    input.addEventListener('blur', handleBlur, signal ? { signal } : undefined);
    // Focus shows the editable comma raw and records it as the rendered
    // value, so the next input event diffs against what is actually in the
    // field. When the value needs no rewrite (the display already equals the
    // editable raw) the browser selection is left untouched — a select-all
    // (Playwright's fill selects before focusing) survives it. When the value
    // IS rewritten, the browser collapses the selection to the caret, so the
    // captured selection is re-applied clamped to the rewritten value.
    input.addEventListener('focus', () => {
        const state = inputState.get(input);
        if (state.raw !== '') {
            const editable = state.raw.replace('.', ',');
            if (input.value === editable) {
                state.value = editable;
                return;
            }
            const selStart = typeof input.selectionStart === 'number' ? input.selectionStart : null;
            const selEnd = typeof input.selectionEnd === 'number' ? input.selectionEnd : null;
            input.value = editable;
            state.value = editable;
            if (selStart !== null && selEnd !== null) {
                input.setSelectionRange(
                    Math.min(selStart, editable.length),
                    Math.min(selEnd, editable.length),
                );
            }
        }
    }, signal ? { signal } : undefined);
    input.addEventListener('keydown', handleKeydown, signal ? { signal } : undefined);

    // Normalize any server-prefilled value ("17.00" → display "17,00",
    // data-raw keeps the dot form). Always recomputed so cleansed clear the
    // stale clone data (cloned rows carry data-raw attributes).
    const initialRaw = parseToRaw(input.value);
    const initial = formatDecimalInput(initialRaw, maskOptions(input));
    input.dataset.raw = initial.raw;
    input.value = initial.formatted;
    inputState.set(input, { raw: initialRaw, value: initial.formatted });
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
