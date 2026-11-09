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
 * Formats a weight value for display without padding decimals.
 *
 * @param {string} value - The completely raw value.
 * @param {boolean} hasDotDecimal - Avoid stripping decimal if it is a dot.
 * @returns {{formatted: string, raw: string}}
 */
function formatForDisplay(value, hasDotDecimal) {
  // makes sure value is raw and valid decimal.
  const withDecimal = hasDotDecimal ? value : value.replace(/[^0-9,]/g, '').replace(',', '.');
  if (withDecimal[withDecimal.length - 1] === '.') {
    return { formatted: value, raw: withDecimal.substring(0, withDecimal.length - 1) };
  }

  const formatted = kgFormatter.format(withDecimal);

  if (withDecimal === '') {
    return { formatted: '', raw: '' };
  }

  return { formatted, raw: withDecimal };
}

/**
 * Sets up a weight input with live formatting.
 * On init, formats any existing numeric value to XX.XXX kg.
 * On input, formats in real-time with thousands separators and " kg" suffix.
 * On blur, shows the final formatted value.
 * On focus, shows raw editable value.
 *
 * @param {HTMLInputElement} input - The input element.
 * @param {AbortSignal} signal - AbortSignal for cleanup.
 */
/**
 * Computes the new cursor position after reformatting by matching the number
 * of significant characters (digits + decimal separator) before the old cursor.
 *
 * @param {string} oldValue
 * @param {number} oldCursor
 * @param {string} newValue
 * @returns {number}
 */
function computeCursorPosition(oldValue, oldCursor, newValue) {
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

export function setupWeightInput(input, signal) {
  if (!input) return;

  function handleInput() {
    const oldValue = input.value;
    const oldCursor = input.selectionStart || 0;

    const result = formatForDisplay(oldValue);
    if (result.formatted === '') {
      input.value = '';
      input.dataset.raw = '';
      return;
    }

    // Skip if nothing changed (prevents cursor jumps and loops)
    if (result.formatted === oldValue) {
      input.dataset.raw = result.raw;
      return;
    }

    input.value = result.formatted;
    input.dataset.raw = result.raw;

    const newCursor = computeCursorPosition(oldValue, oldCursor, result.formatted);
    input.setSelectionRange(newCursor, newCursor);
  }

  function handleBlur() {
    if (!input.value || input.value.replace(/[^0-9]/g, '') === '') {
      input.value = '';
      input.dataset.raw = '';
      return;
    }
    const result = formatForDisplay(input.value);
    input.value = result.formatted;
    input.dataset.raw = result.raw;
  }

  function handleFocus() {
    const raw = input.dataset.raw;
    if (raw && raw !== '') {
      input.value = raw.replace('.', ',');
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
    // Allow comma or dot for decimal separator (treat dot same as comma)
    if (e.key === ',' || e.key === '.') {
      if (input.value.includes(',')) {
        e.preventDefault();
      }
      return;
    }
    // Allow digits
    if (e.key >= '0' && e.key <= '9') {
      return;
    }
    // Block everything else
    e.preventDefault();
  }

  // Initialize: handle both raw numeric and pre-formatted initial values
  const initialValue = input.value;
  const result = formatForDisplay(initialValue, true);
  if (result.raw !== '') {
    input.dataset.raw = result.raw;
    input.value = result.formatted;
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
