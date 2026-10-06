# plan: detached_sync_authorized_xml

Implementation plan for storing the `<nfeProc>` wrapper in `xml_authorized` on synchronous authorization of detached NF-e invoices. Spec: `docs/sdd/detached_sync_authorized_xml/sdd.md`.

## Test Strategy

**Behaviors to cover** (all new, all in Go unit tests — no DB needed):

1. Authorized response → status updated to `authorized` **and** `<nfeProc>` wrapper persisted with correct `<infProt>` fields; wrapper round-trips through `nfe_xml.ParseDANFEData` (AC1, AC4).
2. Authorized response with unparseable signed XML → status still updated, no `xml_authorized` write, success toast (AC2).
3. Pending / rejected / unknown responses → no `xml_authorized` write, status/toast behavior unchanged (AC3).

**Test locations/layers:** `service/nfe_service`. New file: `service/nfe_service/detached_service_response_test.go` containing a hand-written stub for the new model interface — no mocking framework, per project convention (reference pattern: `service/entry_service/test/mocks.go`). **Implementation note:** the file uses the *internal* test package (`package nfe_service`), not the external `nfe_service_test` used by sibling files — `handleDetachedSefazResponse` is unexported and inaccessible from the external package. Go permits both packages in one directory, so the existing files are unaffected.

**Fixtures/environment:** no database, no certificates. `sefaz.AutorizacaoResponse` and `nfe_service.DetachedInvoiceResult` are plain structs. A minimal signed-shaped XML is only needed for the failure test (any non-`<NFe>` string triggers the `BuildAuthorizedXML` error path). For the success test the captured XML is asserted via the parser, never by substring (canonical serialization may alter formatting).

**Environment needs:** none beyond Go toolchain. `InitNFeModel` is not called; the stub replaces the model entirely.

**Commands:**

```bash
# focused, during development
go test -run '^TestHandleDetachedSefazResponse' ./service/nfe_service/

# guards while working
go test ./service/nfe_service/ ./pkg/nfe/xml/

# before completion
make test-go
gofmt -l service/nfe_service/
```

## Phase 1 — Test seam (model interface)

**Files:** `service/nfe_service/detached_service.go`

1. Define an unexported interface near `handleDetachedSefazResponse`:

```go
// detachedResponseModel is the slice of the NFe model that
// handleDetachedSefazResponse needs; narrow so tests can stub it.
type detachedResponseModel interface {
	UpdateDetachedInvoiceStatus(id int, status, protocol, sefazCode, sefazMotive string) error
	UpdateDetachedInvoiceAuthorizedXML(id int, xmlAuthorized string) error
}
```

   Signatures must match `model/nfe_model/detached_model.go:250` and `:273` exactly.
2. Change `handleDetachedSefazResponse`'s parameter from `nfeModel *nfe_model.NFeModel` to `nfeModel detachedResponseModel`. The method is unexported — no external ripple. Call sites `detached_service.go:333` and `:1015` already pass `*nfe_model.NFeModel`, which satisfies the interface implicitly; no call-site edits.

**Tests in this phase:** compile only (`go build ./...`, `go vet ./service/nfe_service/`). Behavior tests come with the fix in Phase 2 (same slice), per "tests alongside implementation".

## Phase 2 — Fix + regression tests

**Files:** `service/nfe_service/detached_service.go`, new `service/nfe_service/detached_service_response_test.go`

1. In the `IsAuthorized()` branch, after `UpdateDetachedInvoiceStatus`, add the wrapper build/store — a direct port of the departure-flow block (`service.go:384-392`), using the function's existing `accessKey` parameter:

```go
// Build and store the <nfeProc> wrapper (signed NFe + protocol) so
// consumers can read nProt/dhRecbto from the stored XML.
if authXML, buildErr := nfe_xml.BuildAuthorizedXML(signedXML, accessKey, sefazResp.Protocol, sefazResp.DhRecbto, sefazResp.StatusCode, sefazResp.StatusMotive); buildErr == nil {
	if xmlErr := nfeModel.UpdateDetachedInvoiceAuthorizedXML(invoiceID, authXML); xmlErr != nil {
		model_error.GetLoggerModel().Log(fmt.Sprintf("UpdateDetachedInvoiceAuthorizedXML error: %v", xmlErr.Error()))
	}
} else {
	model_error.GetLoggerModel().Log(fmt.Sprintf("BuildAuthorizedXML error: %v", buildErr.Error()))
}
```

   Notes: `signedXML` is guaranteed non-empty here (persisted before sending at `detached_service.go:326`); no `nil` guard needed, unlike the worker's DB-record path. Returns/log messages in English (internal-only), consistent with the rest of the file. No new user-facing strings — the success toast already exists.
2. Write `detached_service_response_test.go` (package `nfe_service_test`) with:
   - `stubDetachedResponseModel` implementing `nfe_service`'s new interface: records calls to both methods (status/protocol pairs, captured XML, call counts), returns a settable error for `UpdateDetachedInvoiceAuthorizedXML` to cover the DB-failure branch of AC2.
   - `TestHandleDetachedSefazResponse_Authorized_StoresNfeProcXML` — cStat 100 + nProt + dhRecbto; assert status call args, one authorized-XML call, parse captured XML with `nfe_xml.ParseDANFEData`, assert `data.Protocol`, `data.ProtocolDate` and success toast (AC1, AC4).
   - `TestHandleDetachedSefazResponse_Authorized_BuildFailureNonFatal` — invalid `signedXML` ("not xml") and a stub DB-error variant: status call recorded, zero authorized-XML calls (or one failed call for the DB-error variant), success toast, no panic (AC2).
   - `TestHandleDetachedSefazResponse_NonAuthorized_DoesNotStoreXML` — table over pending (105), rejected (999), unknown (""): correct status recorded, zero authorized-XML calls (AC3).

**Run:** `go test -run '^TestHandleDetachedSefazResponse' ./service/nfe_service/` — expect PASS.

## Phase 3 — Verification & bookkeeping

1. `go test ./service/nfe_service/ ./pkg/nfe/xml/` (guards: behavior + untouched builder).
2. `make test-go` (full Go suite) — record PASS/FAIL in `sdd.md` → Verification Results.
3. `gofmt -l service/nfe_service/` — expect empty output.
4. Update `sdd.md` `Status: Done` only if all ACs are verified; record exact commands and results.

## Risks

- **Interface mismatch:** stub must mirror the model's exact method signatures; a future signature change in `nfe_model` would break the stub at compile time (acceptable — compile-time safety, not drift).
- **Parameter type change** (`*NFeModel` → interface) touches both call sites' inferred types only; both compile unchanged. If `go vet`/build flags anything unexpected, it will surface immediately in Phase 1.
- **Serialization differences:** the stored XML is re-serialized by etree (canonical end tags); the plan asserts via the parser, never byte-for-byte against the signed XML.
- **Legacy rows:** existing sync-authorized detached invoices keep `xml_authorized = NULL` (see Out of scope). `downloadDetachedNFeXML`'s existing fallback covers them.

## Out of Scope

- Detached DANFE endpoint (`downloadDetachedNFeDANFE` — currently a 501 stub) and its protocol DB-column fallback; separate SDD.
- Backfill of `xml_authorized` for already-emitted detached invoices (e.g., a one-off migration or worker pass) — deliberate: root-cause fix only; if wanted, do it as its own SDD.
- Any change to `worker.go` (pending/draft processing) or to the departure flow (`handleSefazResponse`) — they are already correct and serve as the reference implementation.
- Departure-flow fallbacks (`router.go:514-535`) — unchanged by this fix.
