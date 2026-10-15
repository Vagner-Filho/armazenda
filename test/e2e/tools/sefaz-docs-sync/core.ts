// Pure logic for the SEFAZ docs sync workflow (docs/sdd/sefaz_docs_sync/).
// This module MUST stay free of Playwright/DOM/process imports so it can be
// unit-tested offline from test/unit (bun test unit/).

export const PORTAL_BASE_URL = 'https://www.nfe.fazenda.gov.br/portal/'

export interface PortalIndexItem {
  /** Configured root section this item was found under (e.g. "notas-tecnicas"). */
  rootKey: string
  /** Absolute source URL as normalized by normalizeSourceHref. */
  href: string
  /** tipoConteudo token extracted raw (canonical identity key). */
  key: string
  /** Section label as encountered in the UI (e.g. "Notas Técnicas"). */
  section: string
  /** Document title as listed on the portal. */
  title: string
  /** Date shown in the listing, if any. */
  date?: string
}

export interface PortalIndex {
  schemaVersion: number
  fetchedAt: string
  items: PortalIndexItem[]
}

export interface ManifestDoc {
  key: string
  rootKey: string
  section: string
  title: string
  date?: string
  sourceHref: string
  /** sha256 of the downloaded artifact; empty until synced and adopted. */
  sha256?: string
  /** Relative repo path of the downloaded PDF, if synced. */
  pdfPath?: string
  /** Relative repo path of the converted markdown skeleton, if converted. */
  mdPath?: string
  adoptedAt: string
}

export interface Manifest {
  schemaVersion: number
  docs: ManifestDoc[]
}

export type DiffKind = 'NEW' | 'CHANGED' | 'REMOVED' | 'UNCHANGED'

export interface DiffEntry {
  kind: DiffKind
  key: string
  rootKey: string
  section: string
  title: string
  date?: string
  /** Previous listing values for CHANGED/REMOVED entries. */
  previousTitle?: string
  previousDate?: string
}

export interface DiffReport {
  generatedAt: string
  entries: DiffEntry[]
  counts: { NEW: number; CHANGED: number; REMOVED: number; UNCHANGED: number }
}

const tokenRegex = /tipoConteudo=([A-Za-z0-9+/=% ]+)/
const documentTokenRegex = /exibirArquivo\.aspx\?conteudo=([A-Za-z0-9+/=% ]+)/

/**
 * Normalizes a link found on the portal and extracts its canonical identity.
 *
 * Two link families exist on the portal:
 * - navigation/content pages: `…aspx?tipoConteudo=<token>`
 * - DOCUMENT links (on list pages): `exibirArquivo.aspx?conteudo=<token>` —
 *   this is the family that identifies downloadable documents (NT/RT/…).
 *
 * Tokens are taken RAW from the query string — portal tokens contain "+",
 * "=" and even spaces, which a URLSearchParams decode would mangle. The raw
 * string is stable across listings and is the identity key.
 */
export function normalizeSourceHref(
  href: string,
  baseUrl: string = PORTAL_BASE_URL,
): { href: string; key: string } | null {
  if (!href) return null
  let raw = href.trim()
  // strip postback/javascript wrappers; keep only http(s) targets
  if (raw.startsWith('javascript:')) return null
  // drop ASP.NET cookie-support noise before matching
  raw = raw.replace(/&AspxAutoDetectCookieSupport=1/, '').replace('?AspxAutoDetectCookieSupport=1', '?')
  // relative vs absolute
  const abs = raw.startsWith('http://') || raw.startsWith('https://')
    ? raw
    : new URL(raw, baseUrl).toString()
  const docMatch = documentTokenRegex.exec(abs)
  const m = docMatch ?? tokenRegex.exec(abs)
  if (!m) return null
  // canonical: decode percent-encoding only, keep "+/=/space…" payload as-is
  let token = m[1]
  try {
    token = decodeURIComponent(token)
  } catch {
    // keep raw token if malformed URI encoding
  }
  // absolute-ish link to show users (same anchor without noise)
  const cleanHref = abs.replace(/&AspxAutoDetectCookieSupport=1/, '')
  return { href: cleanHref, key: token }
}

export function sortIndexItems(items: PortalIndexItem[]): PortalIndexItem[] {
  return [...items].sort((a, b) =>
    a.rootKey.localeCompare(b.rootKey) ||
    a.section.localeCompare(b.section) ||
    a.title.localeCompare(b.title) ||
    a.key.localeCompare(b.key),
  )
}

export function sortManifestDocs(docs: ManifestDoc[]): ManifestDoc[] {
  return [...docs].sort((a, b) =>
    a.rootKey.localeCompare(b.rootKey, 'pt-BR') ||
    a.section.localeCompare(b.section, 'pt-BR') ||
    a.title.localeCompare(b.title, 'pt-BR') ||
    a.key.localeCompare(b.key),
  )
}

/**
 * Diffs a freshly harvested index against the committed manifest.
 * - token present in index and manifest: UNCHANGED or CHANGED (title/date)
 * - token only in index: NEW
 * - token only in manifest: REMOVED
 */
export function diffIndex(index: PortalIndex, manifest: Manifest): DiffReport {
  const counts = { NEW: 0, CHANGED: 0, REMOVED: 0, UNCHANGED: 0 }
  const byKeyManifest = new Map(manifest.docs.map((d) => [d.key, d]))
  const indexKeys = new Set<string>()
  const entries: DiffEntry[] = []

  for (const item of sortIndexItems(index.items)) {
    indexKeys.add(item.key)
    const prev = byKeyManifest.get(item.key)
    if (!prev) {
      counts.NEW++
      entries.push(entryFrom(item, 'NEW'))
      continue
    }
    const titleChanged = prev.title !== item.title
    const dateChanged = (prev.date ?? '') !== (item.date ?? '')
    if (titleChanged || dateChanged) {
      counts.CHANGED++
      entries.push({
        ...entryFrom(item, 'CHANGED'),
        previousTitle: prev.title,
        previousDate: prev.date,
      })
      continue
    }
    counts.UNCHANGED++
    entries.push(entryFrom(item, 'UNCHANGED'))
  }

  for (const doc of manifest.docs) {
    if (indexKeys.has(doc.key)) continue
    counts.REMOVED++
    entries.push({
      kind: 'REMOVED',
      key: doc.key,
      rootKey: doc.rootKey,
      section: doc.section,
      title: doc.title,
      date: doc.date,
      previousTitle: doc.title,
      previousDate: doc.date,
    })
  }

  entries.sort((a, b) => kindOrder(a.kind) - kindOrder(b.kind) ||
    a.rootKey.localeCompare(b.rootKey, 'pt-BR') ||
    a.section.localeCompare(b.section, 'pt-BR') ||
    a.title.localeCompare(b.title, 'pt-BR'))

  return { generatedAt: new Date().toISOString(), counts, entries }
}

function kindOrder(kind: DiffKind): number {
  switch (kind) {
    case 'NEW':
      return 0
    case 'CHANGED':
      return 1
    case 'REMOVED':
      return 2
    default:
      return 3
  }
}

function entryFrom(item: PortalIndexItem, kind: DiffKind): DiffEntry {
  return {
    kind,
    key: item.key,
    rootKey: item.rootKey,
    section: item.section,
    title: item.title,
    date: item.date,
  }
}

/** Markdown report rendered from a diff (kept deterministic apart from the timestamp). */
export function buildReportText(report: DiffReport): string {
  const lines: string[] = []
  lines.push('# SEFAZ NF-e documentation report')
  lines.push('')
  lines.push(`Generated: ${report.generatedAt}`)
  lines.push('')
  lines.push(
    `NEW ${report.counts.NEW} · CHANGED ${report.counts.CHANGED} · REMOVED ${report.counts.REMOVED} · UNCHANGED ${report.counts.UNCHANGED}`,
  )
  lines.push('')
  for (const heading of ['NEW', 'CHANGED', 'REMOVED'] as const) {
    const group = report.entries.filter((e) => e.kind === heading)
    lines.push(`## ${heading} (${group.length})`)
    lines.push('')
    if (group.length === 0) {
      lines.push('- _n/a_')
      lines.push('')
      continue
    }
    lines.push('| Root | Section | Title | Date (listing) | Key |')
    lines.push('|------|---------|-------|----------------|-----|')
    for (const e of group) {
      const prevSuffix = e.kind === 'CHANGED'
        ? ` (was: ${e.previousTitle ?? ''})`
        : ''
      lines.push(
        `| ${e.rootKey} | ${e.section} | ${markdownCell(e.title + prevSuffix)} | ${e.date ?? ''} | ${e.key} |`,
      )
    }
    lines.push('')
  }
  return lines.join('\n') + '\n'
}

function markdownCell(text: string): string {
  return text.replace(/\|/g, '\\|').replace(/\n/g, ' ')
}

/**
 * Merges inbox items (harvested + optionally enriched with download metadata)
 * into the manifest. Pure merge: existing entries win for fields the inbox
 * entry does not carry (sha256/pdfPath/mdPath survive); `adoptedAt` records
 * the merge time. Idempotent for identical input.
 */
export function adoptIntoManifest(
  manifest: Manifest,
  adoptedAt: string,
  incoming: ManifestDoc[],
): Manifest {
  const byKey = new Map(manifest.docs.map((d) => [d.key, { ...d }]))
  for (const inc of incoming) {
    const prev = byKey.get(inc.key)
    if (!prev) {
      byKey.set(inc.key, { ...inc, adoptedAt })
      continue
    }
    byKey.set(inc.key, {
      ...prev,
      rootKey: inc.rootKey || prev.rootKey,
      section: inc.section || prev.section,
      title: inc.title || prev.title,
      date: inc.date ?? prev.date,
      sourceHref: inc.sourceHref || prev.sourceHref,
      sha256: inc.sha256 || prev.sha256,
      pdfPath: inc.pdfPath || prev.pdfPath,
      mdPath: inc.mdPath || prev.mdPath,
      adoptedAt,
    })
  }
  return { schemaVersion: manifest.schemaVersion, docs: sortManifestDocs([...byKey.values()]) }
}

/** Slug used for inbox file names (derived from the listed title). */
export function slugifyTitle(title: string): string {
  const ascii = title
    .normalize('NFKD')
    .replace(/[\u0300-\u036f]/g, '')
    .replace(/[^A-Za-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .toLowerCase()
  return ascii || 'documento'
}
