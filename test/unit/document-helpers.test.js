import { describe, expect, it } from 'bun:test'
import {
  cpfCheckDigits,
  cnpjCheckDigits,
  generateCPF,
  generateCNPJ,
} from '../e2e/utils/document-helpers.js'

/** Independent CPF validator (mirrors the service's isValidCPF). */
function isValidCPF(digits) {
  if (!/^\d{11}$/.test(digits) || /^(\d)\1+$/.test(digits)) return false
  const d = digits.split('').map(Number)
  let sum = 0
  for (let i = 0; i < 9; i += 1) sum += d[i] * (10 - i)
  let dv1 = 11 - (sum % 11)
  if (dv1 >= 10) dv1 = 0
  sum = 0
  for (let i = 0; i < 10; i += 1) sum += d[i] * (11 - i)
  let dv2 = 11 - (sum % 11)
  if (dv2 >= 10) dv2 = 0
  return dv1 === d[9] && dv2 === d[10]
}

/** Independent CNPJ validator (mirrors the service's isValidCNPJ). */
function isValidCNPJ(digits) {
  if (!/^\d{14}$/.test(digits) || /^(\d)\1+$/.test(digits)) return false
  const d = digits.split('').map(Number)
  const compute = (values, weights) => {
    let sum = 0
    for (let i = 0; i < weights.length; i += 1) sum += values[i] * weights[i]
    const dv = 11 - (sum % 11)
    return dv >= 10 ? 0 : dv
  }
  const dv1 = compute(d, [5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2])
  const dv2 = compute([...d.slice(0, 12), dv1], [6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2])
  return dv1 === d[12] && dv2 === d[13]
}

describe('document check digits', () => {
  it('computes the canonical CPF check digits', () => {
    expect(cpfCheckDigits('529982247')).toBe('25')
    expect(`529982247${cpfCheckDigits('529982247')}`).toBe('52998224725')
  })

  it('computes the canonical CNPJ check digits', () => {
    expect(cnpjCheckDigits('112223330001')).toBe('81')
    expect(`112223330001${cnpjCheckDigits('112223330001')}`).toBe('11222333000181')
  })

  it('generates valid CPFs', () => {
    for (let i = 0; i < 25; i += 1) {
      const cpf = generateCPF()
      expect(cpf).toMatch(/^\d{11}$/)
      expect(isValidCPF(cpf)).toBe(true)
    }
  })

  it('generates valid CNPJs', () => {
    for (let i = 0; i < 25; i += 1) {
      const cnpj = generateCNPJ()
      expect(cnpj).toMatch(/^\d{14}$/)
      expect(isValidCNPJ(cnpj)).toBe(true)
    }
  })
})
