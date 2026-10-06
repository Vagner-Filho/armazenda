# sdd:detached_sync_authorized_xml

## Status
Done

## Goal
When a detached (NF-e avulsa) invoice is **synchronously authorized** by SEFAZ, store the canonical `<nfeProc>` wrapper (signed `<NFe>` + `<protNFe>`) in `detached_nfe_invoice.xml_authorized`, exactly like the departure-linked flow already does. Today only the retry worker builds that wrapper (`worker.go:438`), and the worker never revisits invoices that went straight to `authorized` — so every synchronously authorized detached invoice is left permanently with `XMLAuthorized = NULL` and a signed XML that contains no `<protNFe>`.

This is a root-cause fix only. The detached DANFE endpoint is out of scope.

## Background

- Departure flow, synchronous path: `handleSefazResponse` (`service/nfe_service/service.go:384-392`) builds and stores the wrapper on authorization, with the comment "so the DANFE parser can extract nProt/dhRecbto from the stored XML".
- Detached flow, synchronous path: `handleDetachedSefazResponse` (`service/nfe_service/detached_service.go:898-933`) only calls `UpdateDetachedInvoiceStatus(invoiceID, "authorized", protocol, ...)` and returns.
- Detached flow, async path: `processDetachedInvoice` (`service/nfe_service/worker.go:430-445`) does build and store the wrapper — but only for invoices it polls from `pending` (`GetPendingDetachedInvoicesForRetry`). A sync-authorized invoice is already `authorized` when the worker next runs, so it is never processed.
- `BuildAuthorizedXML` (`pkg/nfe/xml/authorized.go:16`) requires a signed `<NFe>` root document and appends `<protNFe><infProt>` with `chNFe`, `dhRecbto`, `nProt`, `cStat`, `xMotivo`. It already has unit tests including a parser round-trip (`pkg/nfe/xml/authorized_test.go`).
- Consumers of `xml_authorized` today: `downloadDetachedNFeXML` (`router/nfe_router/detached_router.go:835-843`) falls back to `xml_signed`, so XML download is unaffected — but it silently serves a protocol-less XML for sync-authorized invoices.

## Requirements

### Functional
1. On the `IsAuthorized()` branch of `handleDetachedSefazResponse`, after updating the status/protocol, build the `<nfeProc>` wrapper with `nfe_xml.BuildAuthorizedXML(signedXML, accessKey, sefazResp.Protocol, sefazResp.DhRecbto, sefazResp.StatusCode, sefazResp.StatusMotive)` and persist it via `nfeModel.UpdateDetachedInvoiceAuthorizedXML(invoiceID, authXML)`.
2. Persisting the wrapper must be **non-fatal**: if `BuildAuthorizedXML` or the DB update fails, the authorization result (status update, success toast) must not change; the failure is logged via `model_error.GetLoggerModel().Log(...)` — mirroring `service.go:386-392` and `worker.go:438-444`.
3. Non-authorized outcomes (`pending`, `denied`, unknown) must remain unchanged and must not write `xml_authorized`.
4. The fix must also cover the SVC synchronous path, since `attemptDetachedSVCContingency` delegates to `handleDetachedSefazResponse` (`detached_service.go:1015`) — i.e., SVC-authorized detached invoices also get the wrapper stored.
5. The returned `DetachedInvoiceResult.XML` stays the signed XML (not the wrapper), consistent with the departure flow.

### Non-Functional
- **Testability:** `handleDetachedSefazResponse` currently takes the concrete `*nfe_model.NFeModel`, which cannot be mocked without a database. Refactor the parameter to a narrow, unexported interface exposing only the two model methods used (`UpdateDetachedInvoiceStatus`, `UpdateDetachedInvoiceAuthorizedXML`), following the project's service-interface + hand-written-mock pattern. `*NFeModel` satisfies it implicitly, so call sites need no change. The method is unexported, so no ripple outside the package.
- **Consistency:** behavior, error handling, and log messages mirror the existing departure-flow and worker implementations (single pattern, no new abstractions beyond the test seam).
- **No schema changes:** `xml_authorized` column already exists (migration `000021_detached_nfe.sql`).

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | An authorized response (`cStat=100` with `nProt`/`dhRecbto`) causes `UpdateDetachedInvoiceAuthorizedXML` to be called with a `<nfeProc>` document whose `<protNFe><infProt>` contains the protocol, receipt date, cStat, xMotivo, and the access key. | New Go unit test `TestHandleDetachedSefazResponse_Authorized_StoresNfeProcXML` (`service/nfe_service`, hand-written mock): assert the mock captured the XML, parse it with `nfe_xml.ParseDANFEData` and assert `data.Protocol == nProt` (round-trip, same style as `pkg/nfe/xml/authorized_test.go`). Command: `go test -run ^TestHandleDetachedSefazResponse_Authorized_StoresNfeProcXML$ ./service/nfe_service/` |
| AC2 | A synchronous authorization still reports success (toast "NF-e autorizada pela SEFAZ") and updates status to `authorized` even when wrapper building/persisting fails. | New test `TestHandleDetachedSefazResponse_Authorized_BuildFailureNonFatal` with invalid `signedXML` (parse error): mock records the status update, records no `xml_authorized` write, toast is a success toast. Command: `go test -run ^TestHandleDetachedSefazResponse_Authorized_BuildFailureNonFatal$ ./service/nfe_service/` |
| AC3 | Processing/rejected/unknown responses never write `xml_authorized`. | New test `TestHandleDetachedSefazResponse_NonAuthorized_DoesNotStoreXML` (table over cStat 105 pending, 999 rejection, empty): mock records no `UpdateDetachedInvoiceAuthorizedXML` call. Command: `go test -run ^TestHandleDetachedSefazResponse_NonAuthorized_DoesNotStoreXML$ ./service/nfe_service/` |
| AC4 | The stored wrapper is faithful to the signed XML (round-trip: `ParseDANFEData` succeeds on it and the invoice keeps the same key fields), so future consumers (e.g. the detached DANFE, out of scope here) can rely on `xml_authorized`. | Covered inside AC1's test via the parser round-trip assertions on the captured XML. |
| AC5 | No regressions in existing behavior of the detached emission flow and the XML builder. | `go test ./service/nfe_service/ ./pkg/nfe/xml/` passes; full suite `make test-go` passes. |
| AC6 | Code is gofmt-clean. | `gofmt -l service/nfe_service/` returns no files. |

Manual checks are not required for any criterion: the whole change is server-side Go logic exercised by unit tests through the mock; no template/browser behavior changes (E2E not applicable — no user-visible flow change).

## Verification Results

| Command | Result | Notes |
|---------|--------|-------|
| `go test ./service/nfe_service/ ./pkg/nfe/xml/` (baseline, pre-implementation) | PASS | Green before any change. |
| `go test -run '^TestHandleDetachedSefazResponse' -v ./service/nfe_service/` | PASS | 3 tests / 6 subtests: AC1+AC4 round-trip, AC2 (build-fail and DB-fail), AC3 (pending/rejected/unknown). |
| `go build ./...` && `go vet ./service/nfe_service/` | PASS | Phase 1 seam: both call sites compile unchanged against the interface. |
| `go test ./service/nfe_service/ ./pkg/nfe/xml/` (post-implementation) | PASS | Guards intact. |
| `make test-go` (`go test ./...`) | PASS | Exit 0, 12 packages ok. AC5 met. |
| `gofmt -l service/nfe_service/` | PASS (with pre-existing exception) | Flags `worker.go` only — pre-existing at HEAD, verified via `git stash` (file unmodified by this change; left untouched to keep the diff scoped). The files changed here (`detached_service.go`, `detached_service_response_test.go`) are gofmt-clean. AC6 met. |

No E2E or manual checks: no template/browser/HTTP behavior changed (see Acceptance Criteria note).
