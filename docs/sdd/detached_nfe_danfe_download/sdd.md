# sdd:detached_nfe_danfe_download

## Status
<!-- One of: Backlog | WIP | Done -->
Done

## Goal
Implement the fiscal DANFE download for detached NF-e invoices (NF-e Avulsa). Today `GET /nfe/avulsa/download/danfe/:accessKey` is a stub returning 501 `"DANFE para NF-e avulsa ainda não implementado"` (`router/nfe_router/detached_router.go:851`), so the list page's DANFE button (rendered for `authorized`/`cancelled` invoices by `templates/nfe/nfe-list-item.html:62`) fails for Avulsa invoices. Authorized detached invoices must download a DANFE PDF built from the stored XML, and cancelled ones must download it with the red "NF-e CANCELADA" banner — the same behavior the departure-based (romaneio) flow already provides via `downloadNFeDANFE` (`router/nfe_router/router.go:497`).

## Background

Everything around the stub already exists; only the handler body is missing.

### What is already wired (no work needed)
- **Route**: `router.GET("/avulsa/download/danfe/:accessKey", downloadDetachedNFeDANFE)` (`detached_router.go:890`).
- **UI**: `templates/nfe/nfe-list-item.html:62` renders the DANFE button for `authorized` and `cancelled` invoices of both types, pointing Avulsa invoices at `/nfe/avulsa/download/danfe/...`. Detached invoices are listed on `/nfe/list` through this shared fragment.
- **Model**: `nfe_model.GetDetachedInvoiceByAccessKey` (`model/nfe_model/detached_model.go:219`) returns everything the handler needs: `FarmID`, `Status`, `XMLAuthorized`, `XMLSigned`, `Protocol`.
- **Parser**: `nfe_xml.ParseDANFEData` (`pkg/nfe/xml/danfe_parser.go:17`) is generic — it accepts signed `<NFe>` and authorized `<nfeProc>` documents and iterates **all** `<det>` elements, so multi-item detached invoices are handled. Contingency fields (`dhCont`/`xJust`) are parsed too (`TestParseDANFEData_ContingencyFields`).
- **Generator**: `nfe_pdf.NewDANFEGenerator().Generate(data)` / `.GenerateCancelled(data)` (`pkg/nfe/service/danfe.go:42,53`).

### Detached-specific gotcha (why the fallbacks are mandatory)

The **synchronous** authorization path `handleDetachedSefazResponse` (`service/nfe_service/detached_service.go:898`) only calls `UpdateDetachedInvoiceStatus` (status + `protocol` DB column) — it does **not** build the `<nfeProc>` wrapper. Only the retry worker does (`service/nfe_service/worker.go:438` → `UpdateDetachedInvoiceAuthorizedXML`). Consequence for a synchronously authorized detached invoice:

- `XMLAuthorized` may be empty → the handler must fall back to `XMLSigned` (always stored on emission);
- the signed XML has no `<protNFe>` → `ParseDANFEData` yields `Protocol == ""` → the handler must fall back to the `invoice.Protocol` DB column so DANFE Campo 2 is filled.

This is exactly the fallback pair `downloadNFeDANFE` already implements for romaneio invoices (router.go:519–535) — the detached handler mirrors it.

**Scope boundary:** fixing that persistence gap (storing the `<nfeProc>` wrapper on the synchronous authorization path) is **out of scope here** and planned as a separate change. This SDD only makes the download correct with today's persistence behavior; if the wrapper starts being stored later, this handler needs no adjustment — it already prefers `XMLAuthorized`, and the DB-protocol fallback fires only when the XML lacks `<protNFe>`.

### Why not the preview path

The detached preview (`GenerateDetachedPreviewDANFE` → `GeneratePreview`) is **non-fiscal**: no access key, number 0, "Sem Valor Fiscal" banner, built from in-memory form data. The download is the **fiscal** DANFE rendered from the persisted XML (what SEFAZ authorized), with protocol and real number — same distinction as romaneio preview vs. `downloadNFeDANFE`. The two must not share a code path.

## Requirements

### Functional

1. **Handler** `downloadDetachedNFeDANFE` (`router/nfe_router/detached_router.go:851`), replacing the 501 stub, in this order:
   1. Resolve `farmID` from the `session_id` cookie via `user_service.GetFarmFromToken`.
   2. `nfeModel.GetDetachedInvoiceByAccessKey(accessKey)` → `404 "NF-e não encontrada"` when nil or DB error.
   3. Ownership: `invoice.FarmID != farmID` → `403 "Acesso negado"` (direct column check, like `downloadDetachedNFeXML` — no departure lookup).
   4. Status gate: DANFE only for `authorized` or `cancelled` → `403 "DANFE somente disponivel para NF-e autorizada pela SEFAZ"` (identical string to `downloadNFeDANFE`).
   5. XML selection: `XMLAuthorized` when non-empty, else `XMLSigned`, else `404 "XML não disponível"` (same precedence and wording as `downloadDetachedNFeXML`).
   6. `nfe_xml.ParseDANFEData(xmlContent)` → `500 "Falha ao processar os dados da NF-e"` on parse error.
   7. **Protocol fallback**: `if data.Protocol == "" && invoice.Protocol != nil { data.Protocol = *invoice.Protocol }`.
   8. Generate: `GenerateCancelled(*data)` when status is `cancelled` (red "NF-e CANCELADA" banner per MOC Anexo II), else `Generate(*data)` → `500 "Falha ao gerar a DANFE"` on error.
   9. Respond `c.Data(http.StatusOK, "application/pdf", pdfBytes)` with `Content-Disposition: attachment; filename=danfe_<accessKey>.pdf`.
2. **No changes** to the departure-flow handler `downloadNFeDANFE`, the route registration, or any template — the wiring already exists (see Background).
3. Small **local pure helpers** in `detached_router.go` for testability (not shared with `router.go` — see Non-Functional):
   - `danfeStatusAllowed(status string) bool` — the authorized/cancelled gate.
   - `detachedDANFEXML(invoice *nfe_model.DetachedInvoice) (string, bool)` — the XML precedence selection; `ok == false` when neither XML is stored.

### Non-Functional

- **Deliberate duplication**: per explicit decision, the handler duplicates `downloadNFeDANFE`'s logic instead of extracting a shared helper across the two download paths. The two flows use different models and ownership resolution; a ~25-line duplication is accepted and no refactor of the working romaneio path is in scope.
- **Language**: all user-facing strings in **pt-br** (AGENTS.md requirement). Note this diverges from the pre-existing English strings inside `downloadNFeDANFE` ("Invoice not found", "Failed to parse invoice data"); the detached router is already pt-br-consistent and stays that way.
- **Reuse tested components**: parsing and PDF generation come from `pkg/nfe/xml` and `pkg/nfe/service`, both already covered by unit tests. No new dependencies.
- **No schema/DB/template changes**; no new packages or routes.
- Error logging via `model_error.GetLoggerModel().Log` for internal failures (parse/generate), matching the service-layer pattern.

## Acceptance Criteria

| # | Criterion | Verified by |
|---|---|---|
| 1 | Authorized detached invoice with a stored `<nfeProc>` `xml_authorized` downloads a `200` response, `application/pdf`, filename `danfe_<accessKey>.pdf`, and the DANFE shows the protocol (Campo 2) | Manual HTTP smoke (see justification below); `ParseDANFEData` protNFe extraction already covered by `pkg/nfe/xml/authorized_test.go` |
| 2 | Authorized detached invoice with **only** `xml_signed` (no `<nfeProc>` — the synchronous-authorization case) still downloads the PDF, with the protocol taken from the `protocol` DB column | Manual HTTP smoke; unit test of `detachedDANFEXML` covers the signed-only selection; parser's empty-protNFe behavior covered by `TestParseDANFEData_Signed` |
| 3 | Cancelled detached invoice downloads the DANFE with the red "NF-e CANCELADA" banner | Manual HTTP smoke; `GenerateCancelled` covered by `pkg/nfe/service` suite |
| 4 | `draft`, `pending`, `denied` and `superseded` invoices get `403` and no PDF | Unit test `TestDanfeStatusAllowed` (automated); manual smoke confirms the wired handler |
| 5 | An invoice belonging to another farm gets `403 "Acesso negado"` | Manual HTTP smoke (ownership check needs DB + session; no automated handler harness exists in the repo — router tests are DB-free pure-function tests only) |
| 6 | Missing access key → `404 "NF-e não encontrada"`; invoice with no stored XML → `404 "XML não disponível"` | Unit test `TestDetachedDANFEXML` covers the no-XML branch (automated); manual smoke confirms the wired handler |
| 7 | Multi-item and contingency (SVC, `tpEmis` 6/7) detached invoices render correctly (all product rows, contingency banner) | Existing coverage: `ParseDANFEData` loops all `<det>` (`danfe_parser_test.go`) and parses contingency fields (`TestParseDANFEData_ContingencyFields`); generator contingency banner covered by `pkg/nfe/service` |
| 8 | Departure-flow download unchanged; no regressions elsewhere | `make test-go` (`go test ./...`) |
| 9 | Build and formatting clean | `go build ./...`, `gofmt -l` on changed files |

**Automation justification**: the handler's decision inputs (session cookie → farm, DB-backed invoice lookup) sit behind singletons with no test seams — the repo has no DB-backed handler test harness (`router/nfe_router/detached_router_test.go` tests pure functions exclusively). The automatable logic is therefore extracted into the two pure helpers (AC 4, part of AC 6, and the XML-precedence rule of AC 2), while the full request path is verified by a documented manual HTTP smoke, following the same approach as `docs/sdd/detached_nfe_preview/` (Verification Results).

## Verification Results

Executed on 2026-10-05 against this change set.

| Check | Command / method | Result |
|---|---|---|
| Baseline (pre-implementation) | `go test ./router/nfe_router/... ./pkg/nfe/...` | PASS |
| AC 4 — status gate | `go test -run 'TestDanfeStatusAllowed' -v ./router/nfe_router/` | PASS (authorized/cancelled allowed; draft, pending, denied, superseded, "" denied) |
| AC 2 / 6 — XML precedence | `go test -run 'TestDetachedDANFEXML' -v ./router/nfe_router/` | PASS (authorized wins; signed fallback; neither → ok=false; empty strings absent) |
| Focused suites | `go test ./router/nfe_router/... ./pkg/nfe/...` | PASS |
| AC 8 — full Go suite | `make test-go` (`go test ./...`) | PASS |
| AC 9 — build & format | `go build ./...`; `gofmt -l` on changed files (`router/nfe_router/detached_router.go`, `detached_router_test.go`) | PASS (no diffs; pre-existing `gofmt -l` diff in untouched `router/nfe_router/router.go` confirmed at HEAD, left as-is) |
| AC 1 — authorized with `<nfeProc>` | Manual HTTP smoke (Docker test DB on :5433 + app on :8100, fixture admin login, seeded `detached_nfe_invoice` rows): `GET /nfe/avulsa/download/danfe/<key-a>` | 200, `application/pdf`, `%PDF` magic; `pdftotext` shows protocol `351250123456789` (from `<protNFe>`), emitter, recipient and product rows |
| AC 2 — signed-only XML + DB protocol | Smoke: `GET .../<key-b>` (xml_signed only, no `<protNFe>`, protocol only in DB column) | 200 PDF; `pdftotext` shows protocol `431250999888777` — **Campo 2 fallback proven** |
| AC 3 — cancelled banner | Smoke: `GET .../<key-c>` (status `cancelled`) | 200 PDF; `pdftotext` shows the red-banner text "NF-e CANCELADA" |
| AC 4 (wired) / AC 5 — gates | Smoke: `pending` → 403; cross-farm invoice → 403; authorized with no XML → 404 "XML não disponível"; unknown key → 404 | All as expected (server access log: 200×3, 403×2, 404×2) |
| AC 7 — multi-item / contingency | Existing unit coverage: `ParseDANFEData` loops all `<det>` (`danfe_parser_test.go`), contingency fields (`TestParseDANFEData_ContingencyFields`), generator suite | PASS (pre-existing) |
| Environment cleanup | `docker compose -f test/e2e/docker-compose.test.yml down -v`; server process stopped; temporary debug instrumentation in `model/user_model/model.go` reverted via `git checkout` | Done — working tree clean of smoke artifacts |

Smoke notes:
- Seeded rows covered: authorized+`<nfeProc>`, authorized+signed-only, cancelled, pending, cross-farm, authorized-without-XML, unknown key.
- A transient login failure ("Credenciais inválidas") occurred against the first server instance before instrumentation; it did not reproduce against a freshly started binary (no `DEBUG` output fired — `AuthUser` succeeded), so it was an environment artifact of the smoke setup, not a code issue. The debug patch was reverted cleanly afterwards.
