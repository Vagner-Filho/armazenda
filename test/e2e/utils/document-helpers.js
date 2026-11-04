/**
 * Check-digit helpers for e2e fixtures.
 *
 * Contacts created for the transportadora selector go through
 * validateDetachedTransport, which validates CPF/CNPJ check digits. Faker's
 * random numeric strings fail that validation, so fixtures use these
 * generators instead.
 */

/** Returns the two CPF check digits for a nine-digit base. */
function cpfCheckDigits(base) {
  const digits = String(base).padStart(9, '0').slice(0, 9).split('').map(Number);

  let sum = 0;
  for (let i = 0; i < 9; i += 1) sum += digits[i] * (10 - i);
  let dv1 = 11 - (sum % 11);
  if (dv1 >= 10) dv1 = 0;

  sum = 0;
  for (let i = 0; i < 10; i += 1) sum += (i === 9 ? dv1 : digits[i]) * (11 - i);
  let dv2 = 11 - (sum % 11);
  if (dv2 >= 10) dv2 = 0;

  return `${dv1}${dv2}`;
}

/** Returns the two CNPJ check digits for a twelve-digit base. */
function cnpjCheckDigits(base) {
  const digits = String(base).padStart(12, '0').slice(0, 12).split('').map(Number);
  const dv1Weights = [5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2];
  const dv2Weights = [6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2];

  const compute = (values, weights) => {
    let sum = 0;
    for (let i = 0; i < weights.length; i += 1) sum += values[i] * weights[i];
    const dv = 11 - (sum % 11);
    return dv >= 10 ? 0 : dv;
  };

  const dv1 = compute(digits, dv1Weights);
  const dv2 = compute([...digits, dv1], dv2Weights);
  return `${dv1}${dv2}`;
}

/** Generates a random nine-digit base that is not all-identical digits. */
function randomBase(length) {
  for (;;) {
    let base = '';
    for (let i = 0; i < length; i += 1) base += Math.floor(Math.random() * 10);
    if (!/^(\d)\1+$/.test(base)) return base;
  }
}

/** Generates a valid 11-digit CPF. */
function generateCPF() {
  const base = randomBase(9);
  return `${base}${cpfCheckDigits(base)}`;
}

/** Generates a valid 14-digit CNPJ. */
function generateCNPJ() {
  const base = randomBase(12);
  return `${base}${cnpjCheckDigits(base)}`;
}

module.exports = {
  cpfCheckDigits,
  cnpjCheckDigits,
  generateCPF,
  generateCNPJ,
};
