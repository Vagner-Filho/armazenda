# sdd:detached_infadprod_wiring

## Status
Done

## Goal
Let the user type per-item additional information (infAdProd — MOC leiaute Z / det group, last child of `<det>`) on the Detached NF-e (NF-e Avulsa) form, and have it flow end-to-end: form → parse → preview round-trip → XML emission → items_json persistence → (already-supported) DANFE print. The builder and the DANFE mapping already support the field (`pkg/nfe/xml/builder.go:314-315`, `service/nfe_service/detached_service.go:917`) — the gap is purely the upper layers: UI, router parse, service input mapping and DB model.

The reference emitter corpus uses infAdProd on 2/115 documents (e.g. serial-number notes under items); after the homolog comparison (`docs/research_nfe_use_case/HOMOLOG_REPLAY_COMPARISON.md`) this is the only remaining corpus tag with zero wired path in the UI.

## Requirements

### Functional
1. **Form:** each item row gains an "Informações adicionais do item" text input (`items[i].infAdProd`, optional). Optional per row; blank → never emitted.
2. **Router:** `parseDetachedItems` reads the field and validates length (≤ 500 and ≤ det schema bounds — stays inside the v4.00 TString for infAdProd); a field too long yields the standard pt-br warning toast, naming the item index.
3. **Service:** `DetachedItemInput` gains `InfAdProd string`; `buildDetachedItems` copies it into `entity.ItemData.InfAdProd` unchanged (the builder already sanitizes via `setSchemaText`).
4. **Preview round-trip:** `detachedPreviewItem` + `buildDetachedPreviewViewData` + the preview fragment carry the parsed value back in the hidden `items[i].infAdProd` inputs so the confirm step emits exactly the previewed text (the same contract as every other field).
5. **Persistence:** `nfe_model.DetachedInvoiceItem` gains `InfAdProd string` (`json:"inf_ad_prod,omitempty"`); existing rows without the key retain empty value via JSONB omitempty (same backward-compat pattern as the discount key).
6. **XML:** zero code change — `<infAdProd>` is already emitted as the last `<det>` child when non-empty and sanitized by `setSchemaText`.
7. **DANFE:** zero code change — the DANFE item mapping already carries `InfAdProd` (`detached_service.go:917`).

### Non-Functional
- All tests offline (no SEFAZ).
- All user-facing texts in pt-br.
- Backward compatible: pre-existing `items_json` rows (no key) parse fine; SVC/rebuild paths reproduce the same XML.
- gofmt clean on touched files (builder.go's pre-existing documented flag untouched; builder itself NOT touched this SDD).
- Whitespace-only values (blank after TrimSpace) are treated as absent.

## Acceptance Criteria
(every behavioral criterion automated; the user-facing flow verified in the browser)

1. **AC1** — Item row renders the new input; typing `SERIE ABC123` and previewing shows it round-tripping (fragment hidden input `items[0].infAdProd` carries it) — Playwright e2e in `test/e2e` for the admitted UI step: `make test-e2e` (focused run when practical).
2. **AC2** — Router length guard: a >limit `items[0].infAdProd` yields warning toast naming item 1 — `router/nfe_router` test.
3. **AC3** — Service mapping: `buildDetachedItems` output `ItemData[0].InfAdProd == "SERIE ABC123"` and blank stays `""` — `service/nfe_service` test (pattern of `detached_items_cst_zero_test.go`).
4. **AC4** — Builder emission: input with non-empty `InfAdProd` emits `<infAdProd>` as the **last child of det**; empty omits it → extend an existing builder det test — `pkg/nfe/xml` test.
5. **AC5** — Persistence round-trip: `CreateDetachedInvoice` row → `GetDetachedInvoice`/rebuild path reproduces `InfAdProd` from items_json — `pkg/nfe/model/nfe_model` test or service-level fixture.
6. **AC6** — Backward compat: an items_json row serialized WITHOUT the new key parses with empty InfAdProd (no error) — `pkg/nfe/model/nfe_model` test.
7. **AC7** — Full Go suite green: `make test-go`; realworld golden replay stays green (`go test -run 'TestRealWorldBuilder_PerGroupGolden' ./pkg/nfe/xml/`).
8. **AC8 (manual)** — Browser walk-through (emission preview + DαNFE print showing the per-item note) — human check; justified: visual layout confirmation.

## Verification Results

Implemented and verified 2026-10-15. Commands run from the repository root unless
noted.

### Automated checks

| # | Behavior | Command | Result |
|---|----------|---------|--------|
| AC1 | e2e: item note renders, is typed and round-trips into the preview hidden inputs (a cloned second row stays blank) | `cd test/e2e && PORT=8200 BASE_URL=http://localhost:8200 bun run test --project=chromium tests/detached_inf_ad_prod.spec.js` | **PASS** (1 passed) |
| AC2 | router: trim, whitespace-only → absent, >500 chars → warning naming item 1, 500 chars accepted | `go test -run 'TestParseDetachedItems_InfAdProd' -v ./router/nfe_router/` | **PASS** |
| AC3 | service: `buildDetachedItems` maps `InfAdProd` into `ItemData`, the `items_json` row and the DANFE product (blank stays blank) | `go test -run 'TestBuildDetachedItems_InfAdProd' -v ./service/nfe_service/` | **PASS** |
| AC4 | builder: `<infAdProd>` is the last child before `<vItem>`, omitted when empty, text sanitized (accents folded, newlines → "; ") | `go test -run 'TestBuilder_VItem_VNFTot_NumericFormula' -v ./pkg/nfe/xml/` | **PASS** |
| AC5 | persistence: `CreateDetachedInvoice` → `GetDetachedInvoiceByAccessKey` reproduces `InfAdProd` from `items_json` against a live Postgres (schema created with `InitDb` + `RunMigrations`, e2e fixture seeded) | `RUN_DB_TESTS=1 DB_HOST=localhost DB_PORT=5433 DB_USER=test DB_PASS=test DB_NAME=armazenda_test go test -run 'TestDetachedInvoiceItems_InfAdProdDBRoundTrip' -v ./model/nfe_model/` | **PASS** (live DB; SKIPs by design without the DB env) |
| AC6 | backward compat: items_json without `inf_ad_prod` parses empty; empty value omits the key on write | `go test -run 'TestDetachedInvoiceItem_InfAdProdJSON' -v ./model/nfe_model/` + the legacy sub-case of the AC5 test | **PASS** |
| AC7 | full Go suite green; realworld goldens green | `make test-go`; `go test -run 'TestRealWorldBuilder_PerGroupGolden' -v ./pkg/nfe/xml/` | **PASS** |
| — | regression sanity: adjacent detached e2e specs (tabs, transport, CFOP selector, rascunho) | `PORT=8200 BASE_URL=http://localhost:8200 bun run test --project=chromium tests/detached_nfe_tabs.spec.js tests/detached_nfe_transport.spec.js tests/detached_cfop_selector.spec.js` then `... tests/detached_nfe_rascunho.spec.js` | **PASS** (11 + 5 passed) |
| — | JS unit suite | `make test-js` | **PASS** (49 pass, 0 fail) |
| — | gofmt on touched files | `gofmt -l <touched files>` | **PASS** (clean; the pre-existing unformatted `pkg/nfe/xml/builder.go` / `pkg/nfe/defaults/*` flagged before this work stay untouched) |

Note: the e2e/DB commands ran on a spare port (`PORT=8200`, `BASE_URL`) because a
long-running `air` dev server owns `:8100`; the AC5 test used
`model/nfe_model` (the spec's `pkg/nfe/model/nfe_model` path is a typo).

### Manual / not run

- **AC8 (manual)** — browser walk-through of the emission preview and of a
  printed DANFE showing the per-item note, **NOT RUN**: requires a usable
  A1 certificate and SEFAZ access (the e2e fixture has an empty
  `certificate_data`, and the SEFAZ guard is active), plus a human visual
  check. Automated proxies: the e2e covers the form → preview round trip
  (including the PDF preview panel rendering with a 200), the service test
  asserts the note reaches `entity.DANFEProduct.InfAdProd`, and
  `TestDANFEGenerator_GenerateWithInfAdProd` /
  `TestDANFEGenerator_GenerateLongText` (`pkg/nfe/service`) exercise the
  print path with per-item notes (including wrapping) without error.
- Full `make test` (whole e2e suite) **NOT RUN**: the unit suites and every
  detached e2e spec were run; the full browser suite would need the
  `:8100` conflict above resolved.

### Follow-up (documented, out of scope)

- Rascunhos do not persist per-item info (`DetachedProfileItem` has no
  `infAdProd`), so applying a rascunho renders the note blank and
  save-from-form drops it. Adding the field to the profile's `items_json`
  is a small future slice.
- Departure-linked (romaneio) emission remains unwired for `infAdProd`; the
  builder support is shared, so it is a small future slice.
