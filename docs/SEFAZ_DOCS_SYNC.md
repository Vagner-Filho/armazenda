# SEFAZ NF-e docs sync — developer how-to

Manual-only workflow to keep `pkg/nfe/resources/` aligned with SEFAZ portal
publications (Notas Técnicas, Informes Técnicos, Manuais, Esquemas XML).
Spec: `docs/sdd/sefaz_docs_sync/`.

> **Isolation rule:** this tool only reads the public documentation portal
> (`www.nfe.fazenda.gov.br`) via a browser session. It is invoked explicitly
> by a human through `make` targets and is never wired into tests, CI, or the
> app runtime. It must never be used against NF-e authorization/SVC web
> services (see AGENTS.md — any SEFAZ-reaching check stays a deliberate,
> human-driven, `NFE_SEFAZ_GUARD=off` action).

## Found portal facts

- The four configured roots are **leaf list pages** (`listaConteudo.aspx?tipoConteudo=<token>`),
  fetched directly with a real browser session:
  | root key | section | URL token |
  |----------|---------|-----------|
  | `notas-tecnicas` | Notas Técnicas | `04BIflQt1aY=` |
  | `informes-tecnicos` | Informes Técnicos | `hXzemuyNHW4=` |
  | `manuais` | Manuais | `ndIjl+iEFdE=` |
  | `esquemas-xml` | Esquemas XML | `BMPFMBoln3w=` |
- Document anchors are `exibirArquivo.aspx?conteudo=<token>` (note the
  spelling and the `conteudo` parameter). The RAW token (may contain `+`,
  `=` and spaces) is the identity key; publication date is parsed from the
  anchor text ("… - Publicada em dd/mm/yyyy").
- Agents/`webfetch`/plain HTTPS GETs struggle here: the app redirects on
  missing cookies and serves the generic landing body when navigated
  outside a browser session — hence the Playwright implementation.

## Files

- `test/e2e/tools/sefaz-docs-sync/core.ts` — pure logic (identity
  normalization, diff, report, manifest merge; offline-tested by
  `test/unit/sefaz_docs_sync.test.ts`).
- `test/e2e/tools/sefaz-docs-sync/cli.ts` — Playwright-driven CLI + modes.
- `pkg/nfe/resources/sefaz_docs/manifest.json` — committed baseline of known
  documents (identity, section, title, listing date, sha256/path when
  downloaded).
- `pkg/nfe/resources/sefaz_docs/inbox/` — working artifacts (gitignored):
  `portal_index.json` (latest harvest), `reports/report-*.md`,
  `pdf/`, `meta/`, `md/`.

## Workflow

```bash
# 1. check — harvest the portal lists and diff against the manifest (no downloads)
make sefaz-docs-check        # prints + writes inbox/reports/report-*.md

# 2. sync — download NEW/CHANGED docs into inbox; PDFs become markdown skeletons
make sefaz-docs-sync         # run after reviewing the check report

# 3. adopt — promote inbox items into the committed manifest.json (atomic write)
make sefaz-docs-adopt

# re-run check afterwards: expect "NEW 0 · CHANGED 0 · REMOVED 0"
```

Bootstrap (first adoption without downloading everything):
```bash
make sefaz-docs-check        # fills inbox/portal_index.json
cd test/e2e
bun tools/sefaz-docs-sync/cli.ts --mode adopt --adopt-from-index pkg/nfe/resources/sefaz_docs/inbox/portal_index.json
```
(then optionally delete a few manifest entries and run `sync` to fetch just
those, as an end-to-end sanity check.)

Useful flags:
```bash
# offline report re-run (no browser) from a stored index
make -C . -f Makefile sefaz-docs-check  # normal path needs the portal
bun test/e2e/tools/sefaz-docs-sync/cli.ts --mode check --from-index pkg/nfe/resources/sefaz_docs/inbox/portal_index.json

# custom roots (key=URL), e.g. monitoring a NFCom list found on the portal menu
bun test/e2e/tools/sefaz-docs-sync/cli.ts --root nfcom="https://www.nfe.fazenda.gov.br/portal/listaConteudo.aspx?tipoConteudo=gHveCSDQhSM="

#/watch the browser while it walks the lists (debugging)
bun test/e2e/tools/sefaz-docs-sync/cli.ts --mode check --headed --timeout-ms 120000
```

## Curation flow

Inbox `md/*.md` skeletons are `pdftotext -layout` outputs — good enough for
grepping and diffing, not a finished structured reference. Promote curated
versions into `pkg/nfe/resources/<area>/` (e.g. `cbs_ibs/res_1.md`) by hand
or agent-assisted extraction, referencing the source sha256 from the
manifest. New NT versions (e.g. `NT 2025.002 v.1.51 → v.1.52`) should be
diffed against the current materials before updating specs/tests that cite
them.
