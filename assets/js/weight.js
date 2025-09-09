/**
 * Formats a numeric value as a weight string in kilograms with Brazilian Portuguese locale.
 * Handles non-numeric or invalid inputs gracefully by returning 'N/A'.
 *
 * @param {number | string} value - The numeric value to format.
 * @returns {string} The formatted weight string (e.g., "1.234,56 kg") or "N/A" for invalid input.
 * @example
 * formatWeight(1234.56); // "1.234,56 kg"
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

document.addEventListener("DOMContentLoaded", function() {
  const weightElements = document.querySelectorAll(".weight-value");
  weightElements.forEach(el => {
    el.textContent = formatWeight(el.textContent);
  });
});
