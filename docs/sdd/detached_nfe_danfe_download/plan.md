# plan:detached_nfe_danfe_download

Implementation plan for the detached NF-e DANFE download. Spec: [sdd.md](./sdd.md).

**Reference implementation to mirror:** `downloadNFeDANFE` (`router/nfe_router/router.go:497–549`) — lookup → farm check → status gate → XML selection → parse → protocol fallback → `Generate`/`GenerateCancelled` → PDF response.

**Explicit decision (user):** no shared helper between the two download handlers. The detached handler duplicates the ~25-line flow with its own model/ownership resolution; `downloadNFeDANFE` stays untouched.

## Test strategy

Behaviors to cover and where:

| Behavior | Layer / location | Kind |
|---|---|---|
| Status gate: only `authorized`/`cancelled` render a DANFE | `router/nfe_router/detached_router_test.go` (pure function `danfeStatusAllowed`) | Unit, DB-free |
| XML precedence: `XMLAuthorized` → `XMLSigned` → none | `router/nfe_router/detached_router_test.go` (pure function `detachedDANFEXML`) | Unit, DB-free |
| XML → `DANFEData` parsing (signed + nfeProc, multi-item, contingency) | existing `pkg/nfe/xml` suite (`danfe_parser_test.go`, `authorized_test.go`) | Unit, pre-existing |
| Fiscal + cancelled-banner PDF generation | existing `pkg/nfe/service` suite (`danfe_test.go`) | Unit, pre-existing |
| Full request path: 200 PDF, 403/404 gates, Campo 2 fallback | test Postgres + `go run .`, seeded `detached_nfe_invoice` rows, authenticated session | Manual HTTP smoke (not committed), recorded in sdd.md Verification Results |

Fixtures / environment needs:
- Unit tests need no fixtures beyond `nfe_model.DetachedInvoice` structs with pointer fields set (DB-free).
- The smoke needs the e2e test database (Docker, `test/e2e` `db:start`/`db:seed`), a seeded `nfe_farm_config` for the test farm, and detached invoice rows inserted with chosen `status` / `xml_signed` / `xml_authorized` / `protocol` combinations — same environment pattern used by `docs/sdd/detached_nfe_preview/` verification.

Commands:
- During development: `go test ./router/nfe_router/... ./pkg/nfe/...`
- Before completion: `gofmt -w` changed files, `go build ./...`, `make test-go`, manual smoke matrix.

## Phase 1 — Handler implementation + unit tests

**Files:** `router/nfe_router/detached_router.go`, `router/nfe_router/detached_router_test.go`

1. Add imports to `detached_router.go` (internal group, matching `router.go`'s ordering):
   - `nfe_xml "armazenda/pkg/nfe/xml"`
   - `nfe_pdf "armazenda/pkg/nfe/service"`
2. Add the two local pure helpers (spec Functional 3):
   ```go
   // danfeStatusAllowed reports whether a DANFE may be rendered for the given
   // invoice status. Per MOC Anexo II only authorized invoices get a DANFE,
   // plus cancelled ones rendered with the "NF-e CANCELADA" banner.
   func danfeStatusAllowed(status string) bool {
       return status == "authorized" || status == "cancelled"
   }

   // detachedDANFEXML returns the XML to render the DANFE from, preferring the
   // authorized document. ok is false when neither XML is stored.
   func detachedDANFEXML(invoice *nfe_model.DetachedInvoice) (string, bool)
   ```
3. Replace the `downloadDetachedNFeDANFE` stub body, in order (spec Functional 1):
   - `sid` cookie → `user_service.GetFarmFromToken`
   - `nfeModel.GetDetachedInvoiceByAccessKey(accessKey)`; `err != nil || invoice == nil` → `c.String(http.StatusNotFound, "NF-e não encontrada")`
   - `invoice.FarmID != farmID` → `403 "Acesso negado"`
   - `!danfeStatusAllowed(invoice.Status)` → `403 "DANFE somente disponivel para NF-e autorizada pela SEFAZ"`
   - `xmlContent, ok := detachedDANFEXML(invoice)`; `!ok` → `404 "XML não disponível"`
   - `nfe_xml.ParseDANFEData(xmlContent)`; error → log + `500 "Falha ao processar os dados da NF-e"`
   - Protocol fallback: `if data.Protocol == "" && invoice.Protocol != nil { data.Protocol = *invoice.Protocol }`
   - `generator := nfe_pdf.NewDANFEGenerator()`; `GenerateCancelled(*data)` when `invoice.Status == "cancelled"`, else `Generate(*data)`; error → log + `500 "Falha ao gerar a DANFE"`
   - `c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=danfe_%s.pdf", accessKey))` and `c.Data(http.StatusOK, "application/pdf", pdfBytes)`
4. Unit tests in `detached_router_test.go` (same phase, DB-free):
   - `TestDanfeStatusAllowed`: authorized/cancelled → true; draft, pending, denied, superseded, "" → false.
   - `TestDetachedDANFEXML`: both XMLs set → authorized wins; only signed → signed; only authorized → authorized; neither → `ok == false`; empty-string pointers treated as absent.
5. Run `go test ./router/nfe_router/...` — new tests pass, existing suites unmodified.

## Phase 2 — Verification & docs

1. Focused suites: `go test ./router/nfe_router/... ./pkg/nfe/...` (PASS expected; parser/generator untouched).
2. Completion sweep: `gofmt -w` on changed files, `go build ./...`, `make test-go`.
3. Manual HTTP smoke (test Postgres + `go run .`, authenticated as the seeded test user; record PASS/FAIL per AC in sdd.md):
   - Insert detached invoices for the test farm covering: (a) authorized + `xml_authorized` `<nfeProc>` with protocol; (b) authorized + `xml_signed` only + `protocol` column set (the synchronous-auth case); (c) cancelled (authorized XML + status `cancelled`); (d) `pending`; (e) one invoice owned by another farm.
   - `GET /nfe/avulsa/download/danfe/<accessKey>` for each: expect 200 `application/pdf` for (a)–(c) with Campo 2 filled from XML in (a) and from the DB column in (b), "NF-e CANCELADA" banner in (c); 403 for (d) and (e).
   - Unknown key → `404 "NF-e não encontrada"`; known invoice with both XML columns NULL → `404 "XML não disponível"`.
   - Eyeball one PDF: multi-item rows, contingency banner when `tpEmis` 6/7 XML is used.
4. Update `docs/DETACHED_NFE_IMPLEMENTATION.md`:
   - Route list: drop "(placeholder)" from `GET /nfe/avulsa/download/danfe/:accessKey` (line 68).
   - "DANFE Generation" section (line 198): describe the implemented flow (stored XML → `ParseDANFEData` → `Generate`/`GenerateCancelled`, protocol fallback) replacing "Can be added later…".
   - TODO list (line 319): remove the DANFE item.
5. Flip sdd.md `Status: Done` only when all ACs pass or have a documented justification.

## Risks

- **Synchronously-authorized invoices without `<nfeProc>`** — the main functional risk; covered by the mandatory XML + DB-protocol fallbacks and verified explicitly by smoke case (b). If the fallback were skipped, Campo 2 (protocolo) would render empty for those invoices.
- **Cancelled detached invoices** keep `xml_authorized`/`xml_signed` (`CancelDetachedInvoice` only flips status, detached_service.go:1021), so `GenerateCancelled` has an XML to render from — smoke case (c) confirms.
- **SVC/contingency detached invoices** (`tp_emis` 6/7, `dh_cont`/`x_just` columns): parser and generator already handle contingency fields; eyeball check only.
- **Smoke environment**: requires farm config + session fixtures (the e2e fixture has no `nfe_farm_config` — same constraint recorded by `detached_nfe_preview`); insert the rows directly via SQL.

## Out of scope (explicit)

- **Sync-authorization `<nfeProc>` wrapper fix** — `handleDetachedSefazResponse` (`service/nfe_service/detached_service.go:898`) not storing the authorized XML wrapper is handled in a **separate scope**. This SDD only guarantees a correct DANFE **with today's persistence behavior** (signed XML + `protocol` DB column, via the two fallbacks — verified by smoke case (b), which tests compatibility, not the fix). If that scope later stores `<nfeProc>` on the synchronous path, this handler needs no change: it already prefers `XMLAuthorized`, and the DB-protocol fallback fires only when the XML lacks `<protNFe>` (that scope must keep populating the `protocol` column, which is the Campo 2 safety net).
- Shared helper extraction across `downloadNFeDANFE` / `downloadDetachedNFeDANFE` (deliberate duplication, user decision).
- Fixing the broken `GET /nfe/avulsa/list` template (`nfe-detached-list` not defined — pre-existing side finding recorded in `docs/sdd/detached_nfe_operation_profiles/sdd.md:70`).
- DANFE for `denied` invoices (no DANFE per the status lifecycle), EPEC/FS-DA, any change to the romaneio flow, templates, routes, or schema.
