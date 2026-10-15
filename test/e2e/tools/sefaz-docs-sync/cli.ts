// SEFAZ NF-e documentation sync (docs/sdd/sefaz_docs_sync/).
//
// Modes:
//   check  harvest the portal document lists, diff against the committed
//          manifest, print + write a change report (no downloads).
//   sync   check + download NEW/CHANGED documents into inbox/ and convert
//          PDFs to markdown skeletons via pdftotext.
//   adopt  merge inbox download metadata into the committed manifest.json.
//
// Evidence-based portal facts (docs/sdd/sefaz_docs_sync probes):
//   - the four configured root URLs are LEAF LIST pages of the portal; a
//     plain browser GET renders their document lists (the side rail h1
//     "Estatísticas da NF-e" is unrelated chrome);
//   - document anchors are `exibirArquivo.aspx?conteudo=<token>`; the token
//     (raw, may contain "+", "=" and spaces) is the identity key;
//   - publication date is part of the anchor text ("… - Publicada em
//     dd/mm/yyyy"), also captured separately.
//
// Manual-only workflow (never wired into tests/CI) — SEFAZ isolation rule:
// this tool talks ONLY to the public documentation portal via a browser
// session; it must never be used against authorization/SVC web services.
//
// Usage: bun tools/sefaz-docs-sync/cli.ts [--mode check|sync|adopt] ...

import { spawnSync } from 'node:child_process'
import { createHash } from 'node:crypto'
import {
  existsSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  renameSync,
  writeFileSync,
} from 'node:fs'
import path from 'node:path'
import { chromium, type BrowserContext } from 'playwright-core'

import {
  PORTAL_BASE_URL,
  adoptIntoManifest,
  buildReportText,
  diffIndex,
  normalizeSourceHref,
  sortIndexItems,
  slugifyTitle,
  type Manifest,
  type ManifestDoc,
  type PortalIndex,
  type PortalIndexItem,
} from './core.ts'

interface RootConfig {
  key: string
  url: string
  /** Human-facing section label; defaults to the key. */
  label: string
}

const DEFAULT_ROOTS: RootConfig[] = [
  {
    key: 'notas-tecnicas',
    label: 'Notas Técnicas',
    url: 'https://www.nfe.fazenda.gov.br/portal/listaConteudo.aspx?tipoConteudo=04BIflQt1aY=',
  },
  {
    key: 'informes-tecnicos',
    label: 'Informes Técnicos',
    url: 'https://www.nfe.fazenda.gov.br/portal/listaConteudo.aspx?tipoConteudo=hXzemuyNHW4=',
  },
  {
    key: 'manuais',
    label: 'Manuais',
    url: 'https://www.nfe.fazenda.gov.br/portal/listaConteudo.aspx?tipoConteudo=ndIjl+iEFdE=',
  },
  {
    key: 'esquemas-xml',
    label: 'Esquemas XML',
    url: 'https://www.nfe.fazenda.gov.br/portal/listaConteudo.aspx?tipoConteudo=BMPFMBoln3w=',
  },
]

const REPO_ROOT = path.resolve(import.meta.dir, '../../../..')
const DOCS_DIR = path.join(REPO_ROOT, 'pkg/nfe/resources/sefaz_docs')
const MANIFEST_PATH = path.join(DOCS_DIR, 'manifest.json')
const INBOX_DIR = path.join(DOCS_DIR, 'inbox')
const INDEX_PATH = path.join(INBOX_DIR, 'portal_index.json')
const PDF_DIR = path.join(INBOX_DIR, 'pdf')
const META_DIR = path.join(INBOX_DIR, 'meta')
const MD_DIR = path.join(INBOX_DIR, 'md')
const REPORT_DIR = path.join(INBOX_DIR, 'reports')

const USER_AGENT =
  'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0 Safari/537.36'

interface CliOptions {
  mode: 'check' | 'sync' | 'adopt'
  headed: boolean
  timeoutMs: number
  roots: RootConfig[]
  fromIndex?: string
  adoptFromIndex?: string
  help: boolean
}

function usage(): string {
  return [
    'SEFAZ NF-e documentation sync — manual-only developer tool.',
    '',
    'Usage:',
    '  bun tools/sefaz-docs-sync/cli.ts [--mode check|sync|adopt] [flags]',
    '',
    'Modes:',
    '  check  harvest the document lists + diff/report (no downloads) [default]',
    '  sync   check + download new/updated docs + convert PDFs to markdown',
    '  adopt  merge inbox metadata into the committed manifest.json',
    '',
    'Flags:',
    '  --root <key=URL>      add/override a root section (repeatable; combine with --replace-roots to reset)',
    '  --replace-roots       replace the default roots with the provided --root set',
    '  --from-index <file>   offline: diff an existing index JSON (no browser/network needed)',
    '  --adopt-from-index <file>  adopt every index item as baseline doc (bootstrap, no sha256)',
    '  --headed              show the browser while walking the portal',
    '  --timeout-ms <n>      navigation timeout per page (default 90000)',
    '  --help                this text',
  ].join('\n')
}

class CliUsageError extends Error {}

function parseArgs(argv: string[]): CliOptions {
  const opts: CliOptions = {
    mode: 'check',
    headed: false,
    timeoutMs: 90_000,
    roots: [...DEFAULT_ROOTS],
    help: false,
  }
  const statedRoots: RootConfig[] = []
  let replaceRoots = false
  for (let i = 0; i < argv.length; i++) {
    const arg = argv[i]
    switch (arg) {
      case '--mode': {
        const value = argv[++i]
        if (value !== 'check' && value !== 'sync' && value !== 'adopt') {
          throw new CliUsageError(`invalid --mode ${value ?? '(missing)'}`)
        }
        opts.mode = value
        break
      }
      case '--headed':
        opts.headed = true
        break
      case '--timeout-ms': {
        const value = Number(argv[++i])
        if (!Number.isFinite(value) || value <= 0) throw new CliUsageError('invalid --timeout-ms')
        opts.timeoutMs = value
        break
      }
      case '--root': {
        const value = argv[++i]
        const eq = value.indexOf('=')
        if (eq <= 0) throw new CliUsageError(`--root must be key=URL, got ${value ?? '(missing)'}`)
        const key = value.slice(0, eq)
        const rest = value.slice(eq + 1)
        const slash = rest.indexOf('=')
        // key=URL or key=label=URL
        if (slash > 0) {
          statedRoots.push({ key, label: rest.slice(0, slash), url: rest.slice(slash + 1) })
        } else {
          statedRoots.push({ key, label: key, url: rest })
        }
        break
      }
      case '--replace-roots':
        replaceRoots = true
        break
      case '--from-index':
        opts.fromIndex = argv[++i]
        break
      case '--adopt-from-index':
        opts.adoptFromIndex = argv[++i]
        break
      case '--help':
        opts.help = true
        break
      default:
        throw new CliUsageError(`unknown argument ${arg}`)
    }
  }
  if (statedRoots.length > 0) {
    if (replaceRoots) {
      opts.roots = statedRoots
    } else {
      for (const r of statedRoots) {
        const idx = opts.roots.findIndex((o) => o.key === r.key)
        if (idx >= 0) opts.roots[idx] = r
        else opts.roots.push(r)
      }
    }
  }
  return opts
}

function ensureDirs(): void {
  for (const dir of [DOCS_DIR, INBOX_DIR, PDF_DIR, META_DIR, MD_DIR, REPORT_DIR]) {
    if (!existsSync(dir)) mkdirSync(dir, { recursive: true })
  }
}

function readCommittedManifest(): Manifest {
  if (!existsSync(MANIFEST_PATH)) {
    const initial: Manifest = { schemaVersion: 1, docs: [] }
    writeAtomic(MANIFEST_PATH, JSON.stringify(initial, null, 2) + '\n')
    return initial
  }
  return JSON.parse(readFileSync(MANIFEST_PATH, 'utf8')) as Manifest
}

function writeAtomic(targetPath: string, contents: string): void {
  const tmp = `${targetPath}.tmp-${process.pid}`
  writeFileSync(tmp, contents)
  renameSync(tmp, targetPath)
}

function timestampSlug(): string {
  return new Date().toISOString().replace(/[-:T]/g, '').slice(0, 12)
}

// ---------------------------------------------------------------------------
// Harvest

/** Result of parsing a listing anchor. */
function parseListingText(text: string): { title: string; date?: string } {
  const trimmed = text.replace(/\s+/g, ' ').trim()
  const dateMatch = /(?:Publicada|Publicado|Atualizado|atualizado|publicad[oa]s?) em (\d{2}\/\d{2}\/\d{4})/.exec(trimmed)
  return {
    title: trimmed,
    date: dateMatch ? dateMatch[1] : undefined,
  }
}

async function collectItems(
  page: import('playwright-core').Page,
  root: RootConfig,
): Promise<PortalIndexItem[]> {
  const raw = await page.evaluate(() => {
    const out: { text: string; href: string }[] = []
    for (const a of Array.from(document.querySelectorAll('a')) as HTMLAnchorElement[]) {
      const href = a.getAttribute('href') ?? ''
      if (!href.includes('exibirArquivo.aspx?conteudo=')) continue
      const text = (a.textContent ?? '').replace(/\s+/g, ' ').trim()
      if (!text) continue
      out.push({ text, href })
    }
    return out
  })
  const items: PortalIndexItem[] = []
  const seen = new Set<PortalIndexItem['key']>()
  for (const entry of raw) {
    const norm = normalizeSourceHref(entry.href)
    if (!norm) continue
    if (seen.has(norm.key)) continue
    seen.add(norm.key)
    const { title, date } = parseListingText(entry.text)
    items.push({
      rootKey: root.key,
      href: norm.href,
      key: norm.key,
      section: root.label,
      title,
      date,
    })
  }
  return items
}

async function harvestRoot(context: BrowserContext, root: RootConfig, opts: CliOptions): Promise<PortalIndexItem[]> {
  const page = await context.newPage()
  page.setDefaultTimeout(opts.timeoutMs)
  try {
    await page.goto(root.url, { waitUntil: 'domcontentloaded', timeout: opts.timeoutMs, referer: PORTAL_BASE_URL })
    await page.waitForLoadState('domcontentloaded').catch(() => undefined)
    const items = await collectItems(page, root)
    if (items.length === 0) {
      // tolerance: pagination / listing shape drift — warn loudly
      const h1 = (await page.locator('h1').first().textContent().catch(() => '')) ?? ''
      console.warn(
        `[harvest] root ${root.key}: 0 document anchors found (h1="${h1.replace(/\s+/g, ' ').trim()}") — if the portal changed, inspect the page and update the harvest selector.`,
      )
    }
    // pagination: follow "próxima" links while present (bounded)
    let nextPage = items.length > 0
    let guard = 0
    while (nextPage && guard < 20) {
      const next = page.locator('a:visible', { hasText: /^pr[óo]xim/i }).first()
      if ((await next.count().catch(() => 0)) === 0) break
      guard++
      try {
        await next.click()
        await page.waitForLoadState('domcontentloaded').catch(() => undefined)
        const more = await collectItems(page, root)
        nextPage = more.length > 0
        items.push(...more)
      } catch {
        break
      }
    }
    return items
  } catch (err) {
    console.warn(`[harvest] root ${root.key} failed: ${(err as Error).message}`)
    return []
  } finally {
    await page.close().catch(() => undefined)
  }
}

async function harvestAll(context: BrowserContext, opts: CliOptions): Promise<PortalIndex> {
  const items: PortalIndexItem[] = []
  for (const root of opts.roots) {
    console.log(`[harvest] root ${root.key} (${root.url})`)
    const found = await harvestRoot(context, root, opts)
    console.log(`[harvest] root ${root.key}: ${found.length} items`)
    items.push(...found)
  }
  const dedup = new Map<string, PortalIndexItem>()
  for (const it of items) dedup.set(`${it.rootKey}|${it.key}`, it)
  return {
    schemaVersion: 1,
    fetchedAt: new Date().toISOString(),
    items: sortIndexItems([...dedup.values()]),
  }
}

async function makeContext(opts: CliOptions): Promise<BrowserContext> {
  const browser = await chromium.launch({ headless: !opts.headed })
  return browser.newContext({
    userAgent: USER_AGENT,
    locale: 'pt-BR',
  })
  // NOTE: the context keeps the app running; caller closes via closeContext
}

// ---------------------------------------------------------------------------
// Download + conversion

function slugFor(section: string, title: string, key: string): string {
  const shortKey = key.replace(/[^A-Za-z0-9]/g, '').slice(0, 10)
  return `${slugifyTitle(`${section}-${title}`)}-${shortKey}`.slice(0, 120)
}

async function downloadAndConvert(
  context: BrowserContext,
  doc: ManifestDoc,
  opts: CliOptions,
): Promise<{ doc: ManifestDoc; processed: boolean }> {
  const slug = slugFor(doc.section, doc.title, doc.key)
  console.log(`[sync] downloading: ${doc.title}`)
  const res = await context.request.get(doc.sourceHref, {
    headers: { Referer: PORTAL_BASE_URL },
    timeout: opts.timeoutMs,
  })
  if (res.status() !== 200) {
    console.warn(`[sync] HTTP ${res.status()} for ${doc.sourceHref} — skipped`)
    return { doc, processed: false }
  }
  const bytes = await res.body()
  if (bytes.length === 0) {
    console.warn(`[sync] empty body for ${doc.sourceHref} — skipped`)
    return { doc, processed: false }
  }
  const contentType = (res.headers()['content-type'] ?? '').toLowerCase()
  const disposition = res.headers()['content-disposition'] ?? ''
  const fnameMatch = /filename\*=utf-8''([^;]+)|filename="?([^";]+)"?/i.exec(disposition)
  const portalName = fnameMatch ? decodeURIComponent(fnameMatch[1] ?? fnameMatch[2] ?? '') : ''
  const ext = contentType.includes('pdf')
    ? 'pdf'
    : contentType.includes('zip')
      ? 'zip'
      : contentType.includes('xml')
        ? 'xml'
        : portalName.toLowerCase().endsWith('.xlsx')
          ? 'xlsx'
          : 'html'
  const pdfPath = path.join(PDF_DIR, `${slug}.${ext}`)
  writeFileSync(pdfPath, bytes)
  const sha256 = createHash('sha256').update(bytes).digest('hex')
  console.log(`[sync] saved ${path.relative(REPO_ROOT, pdfPath)} (${bytes.length} bytes, ${ext})`)

  let mdPath: string | undefined
  if (ext === 'pdf') {
    const mdTarget = path.join(MD_DIR, `${slug}.md`)
    const proc = spawnSync('pdftotext', ['-layout', '-enc', 'UTF-8', '-nopgbrk', pdfPath, mdTarget], {
      stdio: 'pipe',
    })
    if (proc.status !== 0) {
      console.warn(`[sync] pdftotext failed for ${slug} (exit ${proc.status}); poppler-utils needed.`)
    } else {
      const header = [
        '<!--',
        `source: ${doc.sourceHref}`,
        `section: ${doc.section}`,
        `title: ${doc.title}`,
        `sha256: ${sha256}`,
        `fetchedAt: ${new Date().toISOString()}`,
        '-->',
        '',
      ].join('\n')
      writeFileSync(mdTarget, header + readFileSync(mdTarget, 'utf8'))
      mdPath = mdTarget
      console.log(`[sync] converted → ${path.relative(REPO_ROOT, mdTarget)}`)
    }
  }

  const enriched: ManifestDoc = {
    ...doc,
    sha256,
    pdfPath: path.relative(REPO_ROOT, pdfPath),
    mdPath: mdPath ? path.relative(REPO_ROOT, mdPath) : undefined,
  }
  writeFileSync(path.join(META_DIR, `${slug}.meta.json`), JSON.stringify(enriched, null, 2) + '\n')
  return { doc: enriched, processed: true }
}

// ---------------------------------------------------------------------------
// Modes

async function runCheck(opts: CliOptions): Promise<void> {
  let index: PortalIndex
  if (opts.fromIndex) {
    index = JSON.parse(readFileSync(opts.fromIndex, 'utf8')) as PortalIndex
    console.log(`[check] (offline) index from ${opts.fromIndex}`)
  } else {
    ensureDirs()
    const context = await makeContext(opts)
    try {
      index = await harvestAll(context, opts)
    } finally {
      await context.close().catch(() => undefined)
    }
  }
  ensureDirs()
  writeAtomic(INDEX_PATH, JSON.stringify(index, null, 2) + '\n')

  const manifest = readCommittedManifest()
  const report = diffIndex(index, manifest)
  const text = buildReportText(report)
  const reportPath = path.join(REPORT_DIR, `report-${timestampSlug()}.md`)
  writeFileSync(reportPath, text)
  console.log(text)
  console.log(`[check] report written to ${path.relative(REPO_ROOT, reportPath)}`)
}

async function runSync(opts: CliOptions): Promise<void> {
  ensureDirs()
  const context = await makeContext(opts)
  try {
    const index = await harvestAll(context, opts)
    writeAtomic(INDEX_PATH, JSON.stringify(index, null, 2) + '\n')
    const manifest = readCommittedManifest()
    const report = diffIndex(index, manifest)

    const actionable = report.entries.filter((e) => e.kind === 'NEW' || e.kind === 'CHANGED')
    let processed = 0
    for (const entry of actionable) {
      const item = index.items.find((i) => i.key === entry.key)
      if (!item) {
        console.warn(`[sync] no index item for key ${entry.key}; skipped`)
        continue
      }
      const doc: ManifestDoc = {
        key: item.key,
        rootKey: item.rootKey,
        section: item.section,
        title: item.title,
        date: item.date,
        sourceHref: item.href,
        adoptedAt: new Date().toISOString(),
      }
      const result = await downloadAndConvert(context, doc, opts)
      if (result.processed) processed++
    }

    const text = buildReportText(report)
    const reportPath = path.join(REPORT_DIR, `report-${timestampSlug()}.md`)
    writeFileSync(reportPath, text)
    console.log(text)
    console.log(`[sync] ${processed} document(s) downloaded into inbox`)
    console.log(`[sync] promote with: bun tools/sefaz-docs-sync/cli.ts --mode adopt`)
  } finally {
    await context.close().catch(() => undefined)
  }
}

function runAdopt(opts: CliOptions | { adoptFromIndex?: string }): void {
  ensureDirs()
  const incoming: ManifestDoc[] = []
  if (opts.adoptFromIndex) {
    // bootstrap: adopt every item of an index file as a known doc — sha256
    // stays empty until each document is actually downloaded by a later sync
    const index = JSON.parse(readFileSync(opts.adoptFromIndex, 'utf8')) as PortalIndex
    for (const item of index.items) {
      incoming.push({
        key: item.key,
        rootKey: item.rootKey,
        section: item.section,
        title: item.title,
        date: item.date,
        sourceHref: item.href,
        adoptedAt: new Date().toISOString(),
      })
    }
    console.log(`[adopt] adopting ${incoming.length} index items from ${opts.adoptFromIndex}`)
  } else if (existsSync(META_DIR)) {
    const metaFiles: string[] = []
    const stack = [META_DIR]
    while (stack.length > 0) {
      const dir = stack.pop()!
      for (const name of readdirSync(dir, { withFileTypes: true })) {
        const full = path.join(dir, name.name)
        if (name.isDirectory()) stack.push(full)
        else if (name.isFile() && name.name.endsWith('.meta.json')) metaFiles.push(full)
      }
    }
    for (const file of metaFiles.sort()) {
      try {
        incoming.push(JSON.parse(readFileSync(file, 'utf8')) as ManifestDoc)
      } catch (err) {
        console.warn(`[adopt] skip unreadable meta ${path.relative(REPO_ROOT, file)}: ${(err as Error).message}`)
      }
    }
  }
  if (incoming.length === 0) {
    console.log('[adopt] no inbox metadata — nothing to adopt')
    return
  }
  const manifest = readCommittedManifest()
  const next = adoptIntoManifest(manifest, new Date().toISOString(), incoming)
  writeAtomic(MANIFEST_PATH, JSON.stringify(next, null, 2) + '\n')
  console.log(`[adopt] ${incoming.length} doc(s) merged into ${path.relative(REPO_ROOT, MANIFEST_PATH)}`)
}

async function main(): Promise<number> {
  let opts: CliOptions
  try {
    opts = parseArgs(process.argv.slice(2))
  } catch (err) {
    if (err instanceof CliUsageError) {
      console.error(`Error: ${(err as Error).message}\n`)
      console.error(usage())
      return 1
    }
    throw err
  }
  if (opts.help) {
    console.log(usage())
    return 0
  }
  switch (opts.mode) {
    case 'check':
      await runCheck(opts)
      break
    case 'sync':
      await runSync(opts)
      break
    case 'adopt':
      runAdopt(opts)
      break
  }
  return 0
}

main()
  .then((code) => process.exit(code))
  .catch((err) => {
    console.error(`[sefaz-docs-sync] fatal: ${(err as Error).message}`)
    process.exit(1)
  })
