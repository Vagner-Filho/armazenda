import { describe, expect, it } from 'bun:test'
import {
  adoptIntoManifest,
  buildReportText,
  diffIndex,
  normalizeSourceHref,
  slugifyTitle,
  sortIndexItems,
  sortManifestDocs,
  type Manifest,
  type ManifestDoc,
  type PortalIndex,
  type PortalIndexItem,
} from '../e2e/tools/sefaz-docs-sync/core.ts'

function item(over: Partial<PortalIndexItem> = {}): PortalIndexItem {
  return {
    rootKey: 'notas-tecnicas',
    href: 'https://www.nfe.fazenda.gov.br/portal/exibeArquivo.aspx?tipoConteudo=04BIflQt1aY=',
    key: '04BIflQt1aY=',
    section: 'Notas Técnicas',
    title: 'NT 2025.002-RTC v1.51',
    date: undefined,
    ...over,
  }
}

function manifestDoc(over: Partial<ManifestDoc> = {}): ManifestDoc {
  return {
    key: '04BIflQt1aY=',
    rootKey: 'notas-tecnicas',
    section: 'Notas Técnicas',
    title: 'NT 2025.002-RTC v1.51',
    sourceHref: 'https://www.nfe.fazenda.gov.br/portal/exibeArquivo.aspx?tipoConteudo=04BIflQt1aY=',
    adoptedAt: '2026-10-14T10:00:00.000Z',
    ...over,
  }
}

describe('normalizeSourceHref', () => {
  it('extracts the raw token (preserving "+" and "=")', () => {
    const out = normalizeSourceHref(
      'https://www.nfe.fazenda.gov.br/portal/exibeArquivo.aspx?tipoConteudo=ndIjl+iEFdE=',
    )
    expect(out?.key).toBe('ndIjl+iEFdE=')
  })

  it('extracts document tokens from exibirArquivo links (with spaces)', () => {
    const out = normalizeSourceHref(
      'exibirArquivo.aspx?conteudo=9hgG%20%20cH8gA=',
    )
    expect(out?.key).toBe('9hgG  cH8gA=')
    expect(out?.href).toContain('exibirArquivo.aspx?conteudo=')
  })

  it('normalizes relative links against the portal base', () => {
    const out = normalizeSourceHref('exibeArquivo.aspx?tipoConteudo=pGfPgyEaNGg=%20')
    expect(out?.href).toBe('https://www.nfe.fazenda.gov.br/portal/exibeArquivo.aspx?tipoConteudo=pGfPgyEaNGg=%20')
    expect(out?.key).toBe('pGfPgyEaNGg= ')
  })

  it('strips AspxAutoDetectCookieSupport noise', () => {
    const out = normalizeSourceHref(
      'exibeArquivo.aspx?tipoConteudo=vZzcObX90uc=&AspxAutoDetectCookieSupport=1',
    )
    expect(out?.href).not.toContain('AspxAutoDetectCookieSupport')
    expect(out?.key).toBe('vZzcObX90uc=')
  })

  it('returns null for javascript links and links without tokens', () => {
    expect(normalizeSourceHref('javascript:void(0)')).toBeNull()
    expect(normalizeSourceHref('principal.aspx')).toBeNull()
    expect(normalizeSourceHref('')).toBeNull()
  })
})

describe('diffIndex', () => {
  const baseIndex: PortalIndex = {
    schemaVersion: 1,
    fetchedAt: '2026-10-14T12:00:00.000Z',
    items: [
      item({ key: 'a1', title: 'NT 2025.002-RTC v1.51' }),
      item({ key: 'b2', title: 'Doc B' }),
    ],
  }
  const baseManifest: Manifest = {
    schemaVersion: 1,
    docs: [manifestDoc({ key: 'a1' })],
  }

  it('classifies NEW, UNCHANGED and REMOVED', () => {
    const report = diffIndex(
      { ...baseIndex, items: [...baseIndex.items, item({ key: 'c3', title: 'Doc C', rootKey: 'informes' })] },
      baseManifest,
    )
    expect(report.counts.NEW).toBe(2) // b2 + c3 (b2 is not adopted yet)
    expect(report.counts.UNCHANGED).toBe(1)
    expect(report.counts.REMOVED).toBe(0)
    const kinds = report.entries.map((e) => [e.key, e.kind])
    expect(kinds).toEqual(
      expect.arrayContaining([
        ['c3', 'NEW'],
        ['b2', 'NEW'],
        ['a1', 'UNCHANGED'],
      ]),
    )
  })

  it('detects title/date changes as CHANGED with previous values', () => {
    const idx: PortalIndex = {
      ...baseIndex,
      items: [item({ key: 'a1', title: 'NT 2025.002-RTC v1.52', date: '02/02/2026' })],
    }
    const report = diffIndex(idx, baseManifest)
    expect(report.counts.CHANGED).toBe(1)
    const changed = report.entries.find((e) => e.kind === 'CHANGED')
    expect(changed?.key).toBe('a1')
    expect(changed?.previousTitle).toBe('NT 2025.002-RTC v1.51')
  })

  it('marks manifest entries missing from the index as REMOVED', () => {
    const emptyIndex: PortalIndex = {
      schemaVersion: 1,
      fetchedAt: '2026-10-14T12:00:00.000Z',
      items: [],
    }
    const report = diffIndex(emptyIndex, baseManifest)
    expect(report.counts.REMOVED).toBe(1)
  })

  it('is deterministic regardless of item order', () => {
    const reversed: PortalIndex = {
      schemaVersion: 1,
      fetchedAt: '2026-10-14T12:00:00.000Z',
      items: [item({ key: 'b2', title: 'Doc B', section: 'Notas Técnicas' }), item({ key: 'a1' })],
    }
    const a = diffIndex(reversed, baseManifest)
    const b = diffIndex({ ...reversed, items: [...reversed.items].reverse() }, baseManifest)
    expect(a.entries.map((e) => e.key)).toEqual(b.entries.map((e) => e.key))
  })

  it('sorts NEW entries before CHANGED before REMOVED', () => {
    const manifest: Manifest = {
      schemaVersion: 1,
      docs: [manifestDoc({ key: 'gone' })],
    }
    const index: PortalIndex = {
      schemaVersion: 1,
      fetchedAt: '2026-10-14T12:00:00.000Z',
      items: [item({ key: 'a1' }), item({ key: 'fresh', title: 'Fresh doc' })],
    }
    const report = diffIndex(index, manifest)
    const kinds = report.entries.map((e) => e.kind)
    expect(kinds.indexOf('NEW')).toBeLessThan(kinds.indexOf('REMOVED'))
  })
})

describe('buildReportText', () => {
  it('renders the counts line and section headings', () => {
    const idx: PortalIndex = {
      schemaVersion: 1,
      fetchedAt: '2026-10-14T12:00:00.000Z',
      items: [item({ key: 'new1', title: 'Fresh', section: 'Downloads', rootKey: 'downloads' })],
    }
    const manifest: Manifest = { schemaVersion: 1, docs: [manifestDoc({ key: 'gone' })] }
    const report = diffIndex(idx, manifest)
    const text = buildReportText(report)
    expect(text).toContain('# SEFAZ NF-e documentation report')
    expect(text).toContain('NEW 1 · CHANGED 0 · REMOVED 1 · UNCHANGED 0')
    expect(text).toContain('## NEW (1)')
    expect(text).toContain('## REMOVED (1)')
    expect(text).toContain('| downloads')
  })
})

describe('adoptIntoManifest', () => {
  it('adds new docs with adoptedAt and preserves existing metadata', () => {
    const manifest: Manifest = {
      schemaVersion: 1,
      docs: [manifestDoc({ key: 'a1', sha256: 'deadbeef', pdfPath: 'inbox/pdf/a1.pdf' })],
    }
    const incoming: ManifestDoc[] = [
      manifestDoc({ key: 'a1', sha256: undefined, pdfPath: undefined, section: '' }),
      manifestDoc({ key: 'b2', sha256: 'cafe' }),
    ]
    const out = adoptIntoManifest(manifest, '2026-10-14T13:00:00.000Z', incoming)
    const a1 = out.docs.find((d) => d.key === 'a1')
    expect(a1?.sha256).toBe('deadbeef')
    expect(a1?.pdfPath).toBe('inbox/pdf/a1.pdf')
    expect(a1?.section).toBe('Notas Técnicas') // preserved when incoming empty
    expect(a1?.adoptedAt).toBe('2026-10-14T13:00:00.000Z')
    expect(out.docs.find((d) => d.key === 'b2')?.sha256).toBe('cafe')
  })

  it('is idempotent and keeps a sorted, deterministic order', () => {
    const docs: ManifestDoc[] = [
      manifestDoc({ key: 'z9', title: 'Z-doc' }),
      manifestDoc({ key: 'a1' }),
    ]
    const one = adoptIntoManifest({ schemaVersion: 1, docs: [] }, '2026-10-14T13:00:00.000Z', docs)
    const twice = adoptIntoManifest(one, '2026-10-14T13:00:00.000Z', docs)
    expect(twice).toEqual(one)
    expect(twice.docs.map((d) => d.key)).toEqual(['a1', 'z9'])
  })
})

describe('helpers', () => {
  it('sortIndexItems orders by root, section, title, key', () => {
    const sorted = sortIndexItems([
      item({ rootKey: 'b' }),
      item({ rootKey: 'a', key: 'z' }),
      item({ rootKey: 'a', key: 'a' }),
    ])
    expect(sorted.map((i) => i.rootKey + i.key)).toEqual(['aa', 'az', 'b04BIflQt1aY='])
  })

  it('sortManifestDocs orders Brazilian titles', () => {
    const sorted = sortManifestDocs([
      manifestDoc({ key: 'x', title: 'Nota Técnica' }),
      manifestDoc({ key: 'y', title: 'Áurea' }),
    ])
    expect(sorted.map((d) => d.title)).toEqual(['Áurea', 'Nota Técnica'])
  })

  it('slugifyTitle produces stable kebab-case slugs', () => {
    expect(slugifyTitle('NT 2025.002-RTC v1.51 – IBS/CBS')).toBe('nt-2025-002-rtc-v1-51-ibs-cbs')
    expect(slugifyTitle('')).toBe('documento')
  })
})
