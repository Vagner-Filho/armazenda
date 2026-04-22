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
 * Parses a weight string, extracting integer and decimal parts.
 * Handles both dot (server format) and comma (user format) as decimal separators.
 * Dots are treated as thousands separators unless they are the last separator
 * with 1-2 digits after them (server decimal format).
 *
 * @param {string} value
 * @returns {{isEmpty: boolean, intNum: number, decimalStr: string, hasDecimal: boolean}}
 */
function parseWeightValue(value) {
  // Strip the " kg" suffix and any whitespace from previous formatting
  let workingValue = value.replace(/\s*kg\s*$/i, '').replace(/\s/g, '');

  // Step 1: Detect decimal separator
  const commaIndex = workingValue.indexOf(',');

  if (commaIndex !== -1) {
    // Comma is the decimal separator — remove all dots (thousands separators)
    workingValue = workingValue.replace(/\./g, '');
  } else {
    // No comma — check if the last dot is a decimal separator
    const lastDotIndex = workingValue.lastIndexOf('.');
    if (lastDotIndex !== -1) {
      const afterDot = workingValue.substring(lastDotIndex + 1);
      // If 1-2 digits after the last dot, treat it as a decimal separator
      if (/^\d{1,2}$/.test(afterDot)) {
        // Keep this dot as decimal, remove other dots (thousands separators)
        const beforeDot = workingValue.substring(0, lastDotIndex).replace(/\./g, '');
        workingValue = beforeDot + '.' + afterDot;
      } else {
        // All dots are thousands separators
        workingValue = workingValue.replace(/\./g, '');
      }
    }
  }

  // Step 2: Clean to only digits, comma, and dot
  const cleaned = workingValue.replace(/[^0-9,.]/g, '');

  if (cleaned === '') {
    return { isEmpty: true, intNum: 0, decimalStr: '', hasDecimal: false };
  }

  // Step 3: Extract integer and decimal parts
  const finalCommaIndex = cleaned.indexOf(',');
  const finalDotIndex = cleaned.indexOf('.');

  let integerStr = cleaned;
  let decimalStr = '';
  let hasDecimal = false;

  if (finalCommaIndex !== -1) {
    integerStr = cleaned.substring(0, finalCommaIndex);
    decimalStr = cleaned.substring(finalCommaIndex + 1).replace(/[,.]/g, '').substring(0, 2);
    hasDecimal = true;
  } else if (finalDotIndex !== -1) {
    integerStr = cleaned.substring(0, finalDotIndex);
    decimalStr = cleaned.substring(finalDotIndex + 1).replace(/[,.]/g, '').substring(0, 2);
    hasDecimal = true;
  }

  const intNum = parseInt(integerStr || '0', 10);
  return { isEmpty: false, intNum, decimalStr, hasDecimal };
}

/**
 * Formats a weight value for display without padding decimals.
 *
 * @param {string} value - The raw or partially formatted input value.
 * @returns {{formatted: string, raw: string}}
 */
function formatForDisplay(value) {
  const parsed = parseWeightValue(value);

  if (parsed.isEmpty) {
    return { formatted: '', raw: '' };
  }

  let formatted = parsed.intNum.toLocaleString('pt-BR');
  let raw = parsed.intNum.toString();

  if (parsed.hasDecimal) {
    formatted += ',' + parsed.decimalStr;
    if (parsed.decimalStr) {
      raw += '.' + parsed.decimalStr;
    }
  }

  formatted += ' kg';
  return { formatted, raw };
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
    if (/[0-9]/.test(newValue[i]) || newValue[i] === ',') {
      count++;
    }
    if (count >= significantCount) {
      return i + 1;
    }
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
  const parsed = parseWeightValue(initialValue);
  if (!parsed.isEmpty) {
    const raw = parsed.intNum + (parsed.decimalStr ? '.' + parsed.decimalStr : '');
    input.dataset.raw = raw;
    const result = formatForDisplay(initialValue);
    input.value = result.formatted;
  }

  input.addEventListener('input', handleInput, { signal });
  input.addEventListener('blur', handleBlur, { signal });
  input.addEventListener('focus', handleFocus, { signal });
  input.addEventListener('keydown', handleKeydown, { signal });
}
