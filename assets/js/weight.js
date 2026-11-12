import { applyInputToRaw, caretAfterDigit, groupIntDigits, parseToRaw } from "./decimalMask.js";

const kgFormatter = Intl.NumberFormat('pt-BR', {
  style: 'unit',
  unit: 'kilogram',
  unitDisplay: 'short' // or 'long', 'narrow'
});

/**
 * Formats a numeric value as a weight string in kilograms with Brazilian Portuguese locale.
 * Handles non-numeric or invalid inputs gracefully by returning 'N/A'.
 *
 * @param {number | string} value - The numeric value to format.
 * @returns {string} The formatted weight string (e.g., "1.234 kg") or "N/A" for invalid input.
 * @example
 * formatWeight(1234);    // "1.234 kg"
 * formatWeight("789");   // "789 kg"
 * formatWeight("abc");   // "N/A"
 * formatWeight(null);    // "N/A"
 */
export function formatWeight(value) {
  const numericValue = parseFloat(value);
  if (isNaN(numericValue)) {
    return 'N/A';
  }

  return `${numericValue.toLocaleString('pt-BR')} kg`;
}

/**
 * Formats a canonical raw ("47148", "47148.", "47148.55") for display and
 * keeps the submitted raw.
 *
 * The display is the pt-BR kg formatting ("47.148,55 kg"); a trailing point
 * (a committed-but-empty fraction) renders as a live trailing comma without
 * the unit suffix, and finalizing (blur) drops it. The returned raw is
 * submission-safe: a trailing point never reaches the server.
 *
 * DOM-free core, exported for unit tests (the decimalMask pattern). Parsing
 * of pasted/typed display text lives in decimalMask.js (`parseToRaw` /
 * `applyInputToRaw`) — this function only renders the canonical raw.
 *
 * @param {string} raw - the canonical raw value.
 * @param {{final?: boolean}} [options] - final (blur) drops a trailing separator.
 * @returns {{formatted: string, raw: string}} display + dot-decimal raw.
 *   Both strings are empty when the value has no digits.
 */
export function formatForDisplay(raw, options = {}) {
  const final = options.final ?? false;
  const text = raw == null ? '' : String(raw);
  const point = text.indexOf('.');
  const intDigits = point === -1 ? digitsOnly(text) : digitsOnly(text.slice(0, point));
  const fracDigits = point === -1 ? '' : digitsOnly(text.slice(point + 1));
  if (intDigits === '' && fracDigits === '') {
    return { formatted: '', raw: '' };
  }

  const trailing = point !== -1 && fracDigits === '';
  const rawOut = point !== -1 && fracDigits !== '' ? `${intDigits || '0'}.${fracDigits}` : intDigits;

  let formatted;
  if (trailing && !final) {
    formatted = `${groupIntDigits(intDigits)},`;
  } else {
    formatted = kgFormatter.format(point !== -1 ? `${intDigits || '0'}.${fracDigits}` : intDigits);
  }

  return { formatted, raw: rawOut };
}

/**
 * Digits of a string, in order.
 *
 * @param {string} text
 * @returns {string}
 */
function digitsOnly(text) {
  let out = '';
  for (const ch of text) {
    if (ch >= '0' && ch <= '9') out += ch;
  }
  return out;
}

/**
 * Sets up a weight input with live formatting.
 * On init, normalizes any existing server value to the display form.
 * On input, formats in real-time with thousands separators and " kg" suffix,
 * tracked against the mask's own rendering (never re-parsed).
 * On blur, shows the final formatted value.
 * On focus, shows the raw editable value.
 *
 * @param {HTMLInputElement} input - The input element.
 * @param {AbortSignal} signal - AbortSignal for cleanup.
 */
export function setupWeightInput(input, signal) {
  if (!input) return;

  // Tracked state: exactly what the mask last rendered into the field and
  // its canonical raw (which may carry a trailing point for a committed
  // but-empty fraction). The mask's own display is never re-parsed.
  let stateRaw = '';
  let renderedValue = '';

  function handleInput() {
    const newValue = input.value;

    const { raw: newRaw, caretDigits } = applyInputToRaw(stateRaw, renderedValue, newValue);
    stateRaw = newRaw;

    const result = formatForDisplay(newRaw);
    input.dataset.raw = result.raw;

    if (result.formatted === '') {
      input.value = '';
      renderedValue = '';
      return;
    }

    if (result.formatted !== newValue) {
      input.value = result.formatted;
    }
    const caret = caretAfterDigit(result.formatted, caretDigits);
    input.setSelectionRange(caret, caret);
    renderedValue = result.formatted;
  }

  function handleBlur() {
    if (stateRaw === '') {
      input.value = '';
      input.dataset.raw = '';
      renderedValue = '';
      return;
    }
    // Finalize from the tracked raw — never re-parse the rendered display.
    const result = formatForDisplay(stateRaw, { final: true });
    input.value = result.formatted;
    input.dataset.raw = result.raw;
    renderedValue = result.formatted;
  }

  // Focus shows the raw editable value and records it as the rendered
  // value. When the value needs no rewrite (the display already equals the
  // editable raw) the browser selection is left untouched — a select-all
  // (Playwright's fill selects before focusing) survives it. When the value
  // IS rewritten, the browser collapses the selection to the caret, so the
  // captured selection is re-applied clamped to the rewritten value.
  function handleFocus() {
    if (stateRaw !== '') {
      const editable = stateRaw.replace('.', ',');
      if (input.value === editable) {
        renderedValue = editable;
        return;
      }
      const selStart = typeof input.selectionStart === 'number' ? input.selectionStart : null;
      const selEnd = typeof input.selectionEnd === 'number' ? input.selectionEnd : null;
      input.value = editable;
      renderedValue = editable;
      if (selStart !== null && selEnd !== null) {
        input.setSelectionRange(
          Math.min(selStart, editable.length),
          Math.min(selEnd, editable.length),
        );
      }
    }
  }

  function handleKeydown(e) {
    // Allow control/meta combinations (copy, paste, select all, etc.)
    if (e.ctrlKey || e.metaKey) {
      return;
    }
    // Allow navigation and editing keys
    if (['Backspace', 'Delete', 'Tab', 'Escape', 'Enter', 'ArrowLeft', 'ArrowRight', 'ArrowUp', 'ArrowDown', 'Home', 'End'].includes(e.key)) {
      return;
    }
    // The comma is the only typable decimal separator; block once the
    // editor already committed one. The dot is a paste-only separator
    // (digested by the parser), so typing it is always blocked.
    if (e.key === ',') {
      if (input.value.includes(',')) {
        e.preventDefault();
      }
      return;
    }
    if (e.key === '.') {
      e.preventDefault();
      return;
    }
    // Allow digits
    if (e.key >= '0' && e.key <= '9') {
      return;
    }
    // Block everything else
    e.preventDefault();
  }

  // Initialize: normalize any existing server or pasted value to the display
  // form ("47148.55" → "47.148,55 kg", raw keeps the dot-decimal).
  const initialRaw = parseToRaw(input.value);
  const result = formatForDisplay(initialRaw);
  stateRaw = initialRaw;
  if (result.raw !== '') {
    input.dataset.raw = result.raw;
    input.value = result.formatted;
    renderedValue = result.formatted;
  }

  input.addEventListener('input', handleInput, { signal });
  input.addEventListener('blur', handleBlur, { signal });
  input.addEventListener('focus', handleFocus, { signal });
  input.addEventListener('keydown', handleKeydown, { signal });
}

/**
 * Formats all elements with class 'weight-value' inside a container.
 *
 * @param {HTMLElement} [container=document] - The container to search within.
 */
export function formatWeightElements(container = document) {
  const elements = container.querySelectorAll('.weight-value');
  elements.forEach(el => {
    const rawValue = el.textContent.trim();
    if (rawValue) {
      el.textContent = formatWeight(rawValue);
    }
  });
}
