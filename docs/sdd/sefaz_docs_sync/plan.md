# plan:sefaz_docs_sync

Implementation plan for the SEFAZ documentation sync workflow.
Spec: `docs/sdd/sefaz_docs_sync/sdd.md`.

## Test strategy (defined before implementation)

- **Layers:** pure TypeScript module (`core.ts`) under offline Bun unit tests
  in `test/unit/`; interactive Playwright flow verified manually (AC3/AC4) —
  the portal is an interactive postback application, so offline automation of
  the walk is impractical; pure-logic correctness is hermetic instead.
- **Test locations:**
  - `test/unit/sefaz_docs_sync.test.ts` — normalize/diff/report cases
    (behavior list below), executed by `cd test && bun test unit/`
    (`make test-js`, hermetic).
  - `test/e2e/tools/sefaz-docs-sync/core.ts` — the pure module (no
    Playwright/DOM imports).
  - `test/e2e/tools/sefaz-docs-sync/cli.ts` — Playwright browser flow +
    download + conversion wiring.
- **Behaviors to cover offline:** URL/token normalization (query noise),
  kind classification NEW/CHANGED/REMOVED/UNCHANGED, title/date change
  detection, deterministic sort, report template text, manifest adopt merge
  (idempotent, ordered), atomic-write strategy (tmp+rename) modeled through
  bytes written.
- **Commands:** `cd test && bun test unit/` during work; `make test-js`
  before completion; manual `make sefaz-docs-check` / `make sefaz-docs-sync`
  for the network steps; `make test-go` untouched (no Go changes).

## Phase 0 — setup & baseline

1. Record baseline: `cd test && bun test unit/` (all green), `bun --version`,
   `ls test/e2e/node_modules/playwright*`, `pdftotext -v` availability.
2. Set SDD Status: WIP.

## Phase 1 — pure core (offline unit tests in same phase)

1. Create `test/e2e/tools/sefaz-docs-sync/core.ts`:
   - types: `PortalIndexItem`, `PortalIndex`, `ManifestDoc`, `Manifest`,
     `DiffEntry`, `DiffReport`;
   - `normalizeSourceHref(href)`: absolute-ize against portal base, strip
     `AspxAutoDetectCookieSupport` noise, extract `tipoConteudo` token as key;
   - `diffIndex(currentIndex, manifest)`: NEW/CHANGED/REMOVED per token
     (CHANGED on title/date divergence), deterministic order;
   - `buildReport(diff)`: template text (rooted at the four sections);
   - `adoptIntoManifest(manifest, inboxItems)`: merge by key preserving
     recorded sha256/meta, sorted output;
   - `renderIndex/parseIndex` for stable JSON shapes (schema version field).
2. Create `test/unit/sefaz_docs_sync.test.ts` with the cases above
   (fixtures inline; no network).
3. Run `cd test && bun test unit/` (this phase lands green immediately).

## Phase 2 — CLI skeleton (offline-checkable behaviors first)

1. `test/e2e/tools/sefaz-docs-sync/cli.ts` — Bun runner:
   - modes `check | sync | adopt`; flags `--headed`, `--timeout-ms`,
     `--root <key=value>` overrides, `--help`;
   - usage text; invalid mode exits 1;
   - paths resolved from repo root
     (`pkg/nfe/resources/sefaz_docs/{manifest.json,inbox/}`);
   - `check` without network: reads portal_index.json? No — check always
     harvests; offline dry reading is only for unit tests of core. CLI
     includes `--from-index <file>` debug flag to run diff/report offline
     against a stored index (supports AC2 verification without network and
     future re-analysis after portal changes).
2. Sanity runs offline: `--help`, bad mode, `--from-index` diff on a fixture.

## Phase 3 — Playwright harvest (manual verification phase)

1. Implement harvest in `cli.ts` (browser session):
   - launch chromium (headless default), set UA/referer like probes;
   - for each configured root URL:
     a. `goto` (domcontentloaded) — wait for the menu tree;
     b. walk/click menu items to reach each documented list (labels
        discovered dynamically: Manuais / Esquemas XML / Notas Técnicas /
        Informes Técnicos / Downloads — click-driven since lists render via
        ASP.NET postback);
     c. collect anchors matching `a[href*="exibeArquivo.aspx?tipoConteudo="]`
        with candidate title/date from sibling table cells — includes
        fallback: if zero items harvested after clicking, dump a debug HTML
        snippet (`inbox/debug/<rootKey>-<step>.html`) and log warnings;
     d. pagination: loop "Próxima página" while present;
     e. per-doc record: rootKey, section label, title, date?, source href.
2. Write `inbox/portal_index.json` (sorted, schema version).
3. Manual verification: `make sefaz-docs-check` — verify all four sections
   appear, NT listing present, no repo writes beyond inbox.

## Phase 4 — download + convert (manual verification phase)

1. Downloads: after harvest, for NEW/CHANGED tokens GET the source URL via
   the same browser context (`page.request.get` keeps session cookies);
   save to `inbox/pdf/<slug>.pdf`, write `<slug>.meta.json`.
2. Conversion: spawn `pdftotext -layout -enc UTF-8 -nopgbrk <pdf> <md>`
   prefixed with metadata block; sha256 via `Bun.CryptoHasher`.
3. Manual verification: `make sefaz-docs-sync` with at least one expected
   new/updated doc; inspect inbox artifacts (AC4).

## Phase 5 — adopt + Make targets + docs

1. `adopt` mode wiring: read inbox meta/l distinct entries →
   `adoptIntoManifest` → atomic write of `manifest.json` (tmp+rename).
2. Makefile targets `sefaz-docs-check`, `sefaz-docs-sync` (manual-section
   comments; not chained into test/build).
3. `docs/SEFAZ_DOCS_SYNC.md` how-to with the commands and curation flow.
4. `.gitignore` rule for `pkg/nfe/resources/sefaz_docs/inbox/` if the repo
   keeps an ignore file (check root `.gitignore` presence first).

## Phase 6 — wrap-up

1. `cd test && bun test unit/`; `make test-js`; `make test-unit` full
   (Go untouched — but run to prove no accidental coupling).
2. Manual `make sefaz-docs-check` (+ `sync`, `adopt` cycle) — record
   PASS/FAIL in the SDD with exact commands.
3. Update USE_CASES/doc links only where they reference new docs (if any
   NT update is actually found — otherwise no doc changes).
4. Fill SDD Verification Results; Status → Done when ACs met.

## Risks

- **Portal drift:** class used for item rows may change; mitigation is the
  debug-dump + actionable warning path and `--from-index` offline replay.
- **Token↔section ambiguity:** same token seen with two labels ("Documentos"
  / "Notas Técnicas"); labels are recorded per encounter, never hardcoded.
- **Pagination shape unknown** until first live walk; tolerant page-end
  detection.
- **NFE portais are stateful:** cookie jar and referer handling required
  (proven by probe); Playwright handles both internally.
- **pdftotext tables** are rough; curation remains human/agent-assisted.
- **Detector blind spot:** silent same-title updates aren't seen from
  listings — documented in SDD; complementary `Last-Modified` POST probing
  deferred (future SDD).
- **Runtime politeness:** single browser session, sequential pages, no
  retries storm; runs are explicit human actions only.

## Out of scope

- Automatic curation into `pkg/nfe/resources/cbs_ibs/`, moc layout
  regeneration, LLM-based PDF diff summarization (can be a future
  agent-assisted workflow on top of the inbox artifacts).
- Scheduled/CI execution; `Last-Modified`/ETag content-change probing.
- Any SEFAZ authorization/SVC web-service interaction (stays forbidden
  unless the AGENTS.md homolog-session rules are followed explicitly).
