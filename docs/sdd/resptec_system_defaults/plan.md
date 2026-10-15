# plan:resptec_system_defaults

Implementation plan for the system-scoped `<infRespTec>` values.
Spec: `docs/sdd/resptec_system_defaults/sdd.md`.

## Test strategy (defined before implementation)

- **Layer**: pure Go unit tests — `pkg/nfe/defaults` (constants/gate) and
  `pkg/nfe/xml` (builder emission shape). No DB/HTTP/SEFAZ.
- **Focused commands during development**: `go test -run <pattern> <pkg>`.
- **Before completion**: `make test-go` (AC5) + `gofmt -l` (changed files
  must add no flags).
- **Behaviors to cover**: provider constants present + threshold-gated
  `RespTecConfigured()` (AC1); builder omits empty `xContato`/`email`, keeps
  mandatory CNPJ/fone unconditional, full shape unchanged (AC2); goldens
  diff-audit (AC4).

## Phase 1 — constants (AC1)
1. New file `pkg/nfe/defaults/resptec.go`: `SystemRespTecData` struct,
   exported `RespTec` with the provider values, `RespTecConfigured()`.
2. Test `TestRespTec` in a new `pkg/nfe/defaults/resptec_test.go`:
   values present, CNPJ 14 digits, fone digits-only, gate true; gate false
   when CNPJ removed (copied struct in the test).

## Phase 2 — builder + service wiring (AC2, AC3)
1. `pkg/nfe/xml/builder.go` `buildRespTec`: skip `<xContato>`/`<email>`
   when empty.
2. `service/nfe_service/service.go` `prepareInvoiceBuildData` and
   `service/nfe_service/detached_service.go` `prepareDetachedBuildData`:
   set `invoiceInput.RespTec` from `defaults.RespTec` when configured
   (respecting the import direction — map defaults → entity here).
3. Builder test subtest `infRespTec_email_omitted_when_empty` in
   `builder_test.go`.

## Phase 3 — wrap-up (AC4, AC5)
1. `go test -run '^TestRealWorld' ./pkg/nfe/xml/` + `RW_DEBUG=1` audit.
2. `make test-go`; `gofmt -l` on changed files.
3. Fill SDD Verification Results; Status → Done.

## Risks / notes
- **Import cycle**: `entity` → `defaults` — constants live in defaults with
  a local struct; the service maps into `entity.RespTecData`.
- **Departure flow included**: infRespTec is system-scoped and has no
  departure-specific entanglement (unlike discounts/CFOP), so it is wired in
  both flows here — closing the "deferred globally" item of
  `detached_nfe_input_wiring` (that SDD's out-of-scope note gets a pointer).
- **Empty email**: provider gave none; constants keep `Email` empty —
  add the value to the constant when available.
- **idCSRT/CSRT**: out of scope (vendor registration follow-up).
