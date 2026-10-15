# sdd:sefaz_docs_sync

## Status
Done

## Goal
A developer-triggered workflow to track NF-e documentation changes on the
SEFAZ portal (`www.nfe.fazenda.gov.br`): detect NEW/CHANGED/REMOVED reference
documents (Notas Técnicas, Legislação/Manuais, Downloads/Esquemas XML,
Informes Técnicos), download the new/updated ones, and produce deterministic
markdown skeletons for human curation into `pkg/nfe/resources/`. The tool is
never integrated into app runtime or automated tests; every run is an explicit
human action via Makefile targets.

User-driven motivation: keep `pkg/nfe/resources/` (moc layouts, NT/RT markdown
extracts like `cbs_ibs/res_*.md`) in sync with SEFAZ publications (NT
2025.002-RTC etc.), so future rule-enforcement changes (e.g., the
"Implementação futura" vItem/vNFTot family) get detected and replayed into
Armazenda's spec/tests in time.

## Background

Probes performed while scoping (all read-only):

- The `webfetch` agent tool fails on the portal (transport error) — consistent
  with the known difficulty.
- Plain `curl` GET hits the classic ASP.NET cookie-less redirect loop
  (`…&AspxAutoDetectCookieSupport=1`, 302 ×50). With a **cookie jar** the same
  URL returns HTTP 200 (42 KB HTML).
- The document lists are rendered via **ASP.NET postback menu state**: direct
  GET on the four content URLs returns generic shells (landing
  "Estatísticas da NF-e"), so URL-level fetching cannot enumerate documents —
  a browser session clicking through the menu is required.
- Section mapping of the four configured URLs (observed anchor labels):
  - `tipoConteudo=hXzemuyNHW4=` → **Informes Técnicos**
  - `tipoConteudo=ndIjl+iEFdE=` → **Legislação** (branch incl. *Manuais*)
  - `tipoConteudo=BMPFMBoln3w=` → **Downloads** (branch incl. *Esquemas XML*)
  - `tipoConteudo=04BIflQt1aY=` → **Documentos** (branch incl. *Notas Técnicas*)
  The same token observed with two different labels in different menu contexts
  ("Documentos"/"Notas Técnicas" for `04BIflQt1aY=`) — the tool must walk the
  UI and record labels as encountered, not hard-code token semantics.
- Environment already has everything: `bun` 1.3.0, `node` 24, Playwright +
  chromium browsers (`test/e2e/node_modules`, `~/.cache/ms-playwright`), and
  `pdftotext` (poppler 24.02) for PDF→text.
- Existing conventions to reuse: `Makefile` kebab-case targets
  (`test-go`, `build-css`…), JS unit tests under `test/unit/` executed by
  `cd test && bun test unit/` via `make test-js`.

SEFAZ isolation compliance: this workflow only reads the public documentation
portal; it never touches NF-e authorization/cancellation service endpoints
(no `pkg/nfe/sefaz.Client` usage — transport is the browser session), and is
invoked manually by a human — aligning with AGENTS.md ("never cause network
calls to SEFAZ/SVC endpoints… unless the human explicitly asks").

## Requirements

### Functional
1. **CLI** — `test/e2e/tools/sefaz-docs-sync/cli.ts` executable via
   `cd test/e2e && bun tools/sefaz-docs-sync/cli.ts --mode check|sync|adopt`.
   Modes:
   - `check` (default, read-only): harvest the portal index into
     `pkg/nfe/resources/sefaz_docs/inbox/portal_index.json`; diff it against
     the committed `manifest.json`; print a NEW/CHANGED/REMOVED report and
     write `inbox/report-<YYYYMMDD-HHMMSS>.md`.
   - `sync` = `check` + download NEW/CHANGED docs into
     `inbox/pdf/<slug>.pdf` with `<slug>.meta.json` (source URL, section,
     title, fetch date, HTTP byte size, sha256) and convert each to
     `inbox/md/<slug>.md` via `pdftotext -layout`.
   - `adopt` = review-time promotion: merge inbox entries into the committed
     `manifest.json` (atomic tmp+rename so a crash cannot corrupt it).
2. **Harvest** — headless Playwright chromium driving the public portal:
   load each configured section URL, expand/click the menu items needed to
   reach each document list, collect every item whose link targets
   `exibeArquivo.aspx?tipoConteudo=<token>`, capturing
   `{ rootKey, section, title (as listed), date (if shown), source href }`;
   follow pagination ("próxima página" pattern) when present. Miss-shaped
   pages (layout drift) must downgrade to a warning list, never crash the
   run silently.
3. **Identity & diff** — canonical item key: the `tipoConteudo` token (post
   query-string normalization: reject/rename `AspxAutoDetectCookieSupport`
   noise); sensitive to title/date changes for the same token; ordering
   deterministic (sorted by rootKey, section, title). `changed` detection
   powered by previously recorded metadata (title/date); a same-title silent
   content edit is NOT detectable from the listing (documented limitation;
   see risks).
4. **Conversion** — `pdftotext -layout -enc UTF-8 -nopgbrk` into the `.md`
   skeleton, prefixed with a metadata block (source URL, sha256, fetch date,
   section, title) for curator traceability. Missing `pdftotext` → explicit
   non-zero exit with install hint (poppler-utils).
5. **Make targets** — `sefaz-docs-check` and `sefaz-docs-sync` (manual-only
   wrappers of the CLI; not hooked into `test`/`build` chains).
6. **Docs** — `docs/SEFAZ_DOCS_SYNC.md` developer how-to: preconditions, the
   four section URLs, example run commands, curation flow into
   `pkg/nfe/resources/<area>/`, and the isolation rule (manual-only).

### Non-Functional
- No new dependencies: Bun + Playwright (already in `test/e2e/node_modules`)
  and poppler (`pdftotext`) from the system.
- Offline-first: pure logic (normalize/diff/report) lives in
  `test/e2e/tools/sefaz-docs-sync/core.ts` with no DOM/Playwright imports,
  so it can be unit-tested offline in `test/unit/`
  (`test/unit/sefaz_docs_sync.test.ts` via `bun test unit/`, i.e. part of
  `make test-js` once landed… note: any failing it would sit inside the
  normal suite — it must be hermetic).
- Network politeness: single-threaded, sequential navigation, bounded page
  count per run, explicit 90s default timeouts (configurable `--timeout-ms`),
  no retries storm.
- Runs must be idempotent: same portal state → same index, report says
  "NO CHANGES"; `adopt` guarded to only promote inbox items with metadata
  + sha256 recorded.
- Browsers: headless chromium already installed; a `--headed` flag exists for
  debugging only.

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | Core diff/normalize/report functions are pure and offline-tested: same-token normalization (with/without `AspxAutoDetectCookieSupport` noise), NEW/CHANGED/REMOVED classification, deterministic ordering, report text matches expected template. | `cd test && bun test unit/` (new `unit/sefaz_docs_sync.test.ts`) — part of `make test-js`; network not required. |
| AC2 | CLI mode parsing + guardrails: invalid mode args exit ≠ 0 with usage text; `--headed`, `--timeout-ms` accepted; default mode `check`. | `bun tools/sefaz-docs-sync/cli.ts --help` and an invalid-arg run returning exit 1 (offline). |
| AC3 | `make sefaz-docs-check` runs against the live portal end-to-end and reports all four sections with harvested document listings (Notas Técnicas list present, incl. NT 2025.002-RTC or successors); writes only `inbox/portal_index.json` + `inbox/report-*.md`; leaves the repo tree otherwise untouched. | Manual run, output + `git status --short` diff review. Automated justification: the portal is an interactive postback menu, untestable offline; pure-logic correctness is covered by AC1. |
| AC4 | `make sefaz-docs-sync` downloads ≥1 document (e.g. an NT PDF) recording sha256/meta and producing the `.md` skeleton via `pdftotext -layout`. | Manual run; inspect `inbox/pdf`, `inbox/meta`, `inbox/md` contents (offline check of produced files; PDF fetch is the network-dependent step of AC3's same family). |
| AC5 | `--adopt` promotes inbox items into `manifest.json` atomically; a second `check` afterwards reports NO NEW (idempotence) offline re-parsing of the manifest. | Run with the same inbox twice; assert manifest única version + zero NEW on second report; file-write crash safety via tmp+rename code review. |
| AC6 | Isolation: no `pkg/nfe/sefaz` imports inside the tool; no writes outside `pkg/nfe/resources/sefaz_docs/` + `test/e2e/tools/` + `docs/`; Makefile targets are not part of `test`/`build` chains. | `grep -rn 'pkg/nfe/sefaz' test/e2e/tools/sefaz-docs-sync` empty; `grep -n 'sefaz-docs' Makefile` shows targets only near the manual docs section; full `make test-unit` unaffected. |
| AC7 | Developer documentation exists (`docs/SEFAZ_DOCS_SYNC.md`) describing preconditions, URLs, commands, curation flow and the manual-run isolation rule. | Read; example commands match the implemented CLI flags alphabetically. |
| AC8 | Repo hygiene: browser artifacts stay uncommitted (inbox paths added to `.gitignore`) while `manifest.json` remains committed. | `git status --short` after a run shows only expected tree changes. |

## Notes / limitations

- Same-title silent content edits are undetectable from listing diffs; the
  pragmatic complement is a periodic manual visual skim or a later
  complementary `Last-Modified`/`ETag` probe on `exibeArquivo.aspx` (future
  work, not this SDD).
- Portal layout changes may require selector updates — the tool must emit
  actionable warnings (including the diff of expected-vs-found anchors) so
  the failure is diagnosable by a human, not silent.
- Converting PDFs is a skeleton step: tables in the reference docs render
  roughly via `pdftotext -layout`; faithful structured conversion (like the
  existing `cbs_ibs/res_*.md`) stays a human/agent-assisted curation task
  (out of scope here).

## Verification Results

| Command | Result | Notes |
|---------|--------|-------|
| `cd test && bun test unit/` (baseline, pre-implementation) | PASS | 33/33 green (offline hermetic). |
| `cd test && bun test unit/` (post-implementation) | PASS | 49/49 green — 16 new `sefaz_docs_sync` cases (token normalization incl. `exibirArquivo?conteudo=` family with spaces, NEW/CHANGED/REMOVED classification, deterministic ordering, report template, idempotent adopt merge, manifest sort, slugs). |
| `bun tools/sefaz-docs-sync/cli.ts --help` | PASS | Usage text; default mode `check`. |
| `bun tools/sefaz-docs-sync/cli.ts --mode bogus` | PASS | Exit 1 + usage error message (guarded CLI). |
| `make sefaz-docs-check` | PASS | AC3: live harvest — 513 items (notas-tecnicas 320, informes-tecnicos 49, manuais 12, esquemas-xml 132); report `NEW 513` (первый run, empty manifest); only `inbox/portal_index.json` + `inbox/reports/report-*.md` written. |
| `bun … --mode adopt --adopt-from-index inbox/portal_index.json` | PASS | AC5 bootstrap: 513 docs adopted as baseline (atomic write). |
| `make sefaz-docs-sync` (after removing 3 newest NT entries) | PASS | AC4: exactly the 3 removed docs downloaded (775KB/461KB/712KB PDFs), each converted to a `.md` skeleton with metadata header via `pdftotext -layout`; report `NEW 3 · UNCHANGED 510`. |
| `make sefaz-docs-adopt` | PASS | 3 docs merged with sha256 + pdfPath + mdPath recorded. |
| `make sefaz-docs-check` (second live run) | PASS | AC5 idempotence: `NEW 0 · CHANGED 0 · REMOVED 0 · UNCHANGED 513`. |
| Bonus pull: NT 2025.002 v.1.52 (published 01/10/2026) | PASS | 2.1MB PDF downloaded + converted; **live portal shows v.1.52 is newer than the v.1.51-based extracts in `pkg/nfe/resources/cbs_ibs/`** — flag for curation in the ongoing gap-fix effort. |
| `make -n sefaz-docs-check` + `git status --short` | PASS | AC8: Makefile targets wired; inbox gitignored (`git status` shows only `manifest.json` tree under `sefaz_docs/`, no inbox artifacts). |
| `grep -rn 'pkg/nfe/sefaz' test/e2e/tools/sefaz-docs-sync/` | PASS (no matches) | AC6 isolation: tool never imports the NF-e web-service client; browser-session transport only; not chained into test/build (targets live in a manual-only Makefile section). |
| `make test-go` | PASS (other session's scope) | Baseline suite remains green; this SDD changed no Go files. |
