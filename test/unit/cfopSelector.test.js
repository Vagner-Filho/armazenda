import { describe, expect, it } from 'bun:test'
import {
  cfopOptionMatches,
  deriveOriginDestination,
} from '../../assets/js/cfopSelector.js'

describe('CFOP selector matching', () => {
  const option = {
    value: '1202',
    textContent: '1202 — Devolução de venda de produção do estabelecimento',
    dataset: { originDestination: 'Outro estado' },
    selected: false,
  }

  it('matches the CFOP code and description without case or accent sensitivity', () => {
    expect(cfopOptionMatches(option, '120')).toBe(true)
    expect(cfopOptionMatches(option, 'DEVOLUCAO')).toBe(true)
    expect(cfopOptionMatches(option, 'producao')).toBe(true)
  })

  it('matches Origem/Destino metadata', () => {
    expect(cfopOptionMatches(option, 'outro estado')).toBe(true)
    expect(cfopOptionMatches(option, 'OUTRO ESTADO')).toBe(true)
    expect(cfopOptionMatches(option, 'exterior')).toBe(false)
  })

  it('derives Origem/Destino from the code when option metadata is absent', () => {
    const fallbackOption = { ...option, value: '6102', dataset: {} }
    expect(cfopOptionMatches(fallbackOption, 'outro estado')).toBe(true)
    expect(cfopOptionMatches({ ...option, value: '7101', dataset: {} }, 'exterior')).toBe(true)
  })

  it('keeps selected options available and matches an empty query', () => {
    expect(cfopOptionMatches({ ...option, selected: true }, 'não corresponde')).toBe(true)
    expect(cfopOptionMatches(option, '')).toBe(true)
  })
})

describe('Origem/Destino derived from CFOP', () => {
  it.each([
    ['1101', 'Mesmo estado'],
    ['5101', 'Mesmo estado'],
    ['2102', 'Outro estado'],
    ['6102', 'Outro estado'],
    ['3101', 'Exterior'],
    ['7101', 'Exterior'],
    ['7', ''],
    ['710', ''],
    ['4101', ''],
    ['', ''],
  ])('maps %s to %s', (code, expected) => {
    expect(deriveOriginDestination(code)).toBe(expected)
  })
})
