# sdd:resptec_system_defaults

## Status
Done

## Goal
Wire `<infRespTec>` (Responsável Técnico, NT 2022.002 / E28a) with the **system-scoped** data of the company behind Armazenda — the vendor of the software that generated the XML, constant for every tenant. This resolves the infRespTec item the `detached_nfe_input_wiring` SDD left deferred "globally (both flows)".

User decision driving this spec (discussion 2026-10-14): `infRespTec` is **not** farm-scoped — it identifies the software vendor, which is system-scoped by construction (the corpus itself carries the vendor's data in all 115 files). Provider values supplied by the human:

- CNPJ: `47962538000140`
- xContato: `VAGNER SIMONE MARTINS FILHO CONSULTORIA EM TECNOLOGIA DA INFORMACAO LTDA`
- fone: `66996615930`
- email: `vagner.filhopro@gmail.com` (supplied afterwards after the first homolog attempt showed `<email>` is **1-1** in `infRespTec` — SEFAZ 215 "One of 'email' is expected" when missing; the initial "email optional" premise was only true for `dest/email` E20, a different element. The gate now requires the complete set: an incomplete vendor identity omits the whole block).

## Background

- `<infRespTec>` is mandatory in production (E28a; most UFs also require it in homologação — the same vendor data is used in both environments).
- Builder capability already exists (`entity.RespTecData` +
  `buildRespTec` as the last child of `infNFe`, regression-covered by
  `TestBuilder_SmallRealworldFields/infRespTec_*` and the corpus goldens,
  whose adapter populates `RespTec` from fixtures).
- Import direction: `entity` imports `defaults` — so the system constants
  must live in `pkg/nfe/defaults` using a package-local struct; the service
  maps it into `entity.RespTecData`.
- `idCSRT`/`CSRT` (SEFAZ admission number issued to the registered vendor)
  stays out of scope; it is registered by the vendor and can be added later.

## Requirements

### Functional
1. **System constants** — `pkg/nfe/defaults/resptec.go`:
   - `SystemRespTecData{CNPJ, XContato, Email, Fone}` (plain strings).
   - Exported `RespTec SystemRespTecData` carrying the provider values above
     (email empty for now).
   - `RespTecConfigured() bool` — true when CNPJ and fone are non-empty
     (the two mandatory fields of the E28a group).
2. **Builder robustness** — `buildRespTec` omits `<xContato>`, `<email>`
   when empty (today it emits an empty `<email>` element, which violates
   the TString min-length). CNPJ/fone stay unconditional (mandatory).
3. **Service wiring (both flows)** — `prepareInvoiceBuildData` (departure)
   and `prepareDetachedBuildData` (detached) set
   `invoiceInput.RespTec = &entity.RespTecData{...}` from
   `defaults.RespTec` when `RespTecConfigured()`. System-scoped: no new
   inputs, no DB change, no UI.

### Non-Functional
- pt-br not needed (no user-facing text; values are vendor constants).
- No SEFAZ traffic; guard untouched; offline tests only.
- Real-world goldens unaffected (adapter populates `RespTec`; the diff set
  must remain byte-identical — verified).
- gofmt: no new flags.

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | `defaults.RespTec` carries the provider values; `RespTecConfigured()` gates on CNPJ+fone. | `go test -run 'TestRespTec' ./pkg/nfe/defaults/` |
| AC2 | `buildRespTec`: with the provider shape (no email) the block emits `CNPJ → xContato → fone` (no `<email>` element); a full shape keeps `CNPJ → xContato → email → fone`. | `go test -run 'TestBuilder_SmallRealworldFields' ./pkg/nfe/xml/` (new subtest) |
| AC3 | Both emission paths set `input.RespTec` from the system constants (detached + departure). Structural (code review); the wiring is two assignments around the already-tested builder emission. | recorded in Verification Results |
| AC4 | Corpus goldens keep their diff set unchanged after the builder tweak. | `go test -run '^TestRealWorld' ./pkg/nfe/xml/` (+ `RW_DEBUG=1` diff audit) |
| AC5 | No regressions repo-wide; gofmt unchanged. | `make test-go`; `gofmt -l <changed files>` |

Manual checks: none (constants + builder capability; a homolog smoke would be a separate human-driven step with `NFE_SEFAZ_GUARD=off`).

## Verification Results

| Command | Result | Notes |
|---------|--------|-------|
| `go test ./pkg/nfe/... ./service/nfe_service/` (baseline) | PASS | Green before implementation. |
| `go test -run 'TestRespTec' ./pkg/nfe/defaults/` (AC1) | PASS | Provider constants present; CNPJ 14 digits; fone digits-only; `RespTecConfigured()` true for the shipped values and false when CNPJ or fone is blanked (`TestRespTecUnconfigured`). |
| `go test -run 'TestBuilder_SmallRealworldFields' ./pkg/nfe/xml/` (AC2) | PASS | New `infRespTec_email_omitted_when_empty`: provider shape (no email) emits `CNPJ → xContato → fone` with no `<email>` element; fully-empty xContato likewise skipped (`CNPJ → fone`); the pre-existing full-shape subtest (`CNPJ → xContato → email → fone`) stays green. |
| Structural review (AC3) | PASS | Both prepare paths wired: `prepareDetachedBuildData` (detached) and `prepareInvoiceBuildData` (departure) set `input.RespTec` from `defaults.RespTec` when `RespTecConfigured()`; two assignments around the already-emission-tested builder block. |
| `go test -run '^TestRealWorld' ./pkg/nfe/xml/` + `RW_DEBUG=1` diff audit (AC4) | PASS | Goldens green; the diff set stays inside the existing allowed registry — no new paths (the adapter populates `RespTec` from fixtures, so the defaults never reach the goldens). |
| `make test-go` (AC5) | PASS | Zero `^FAIL` lines across the repo. |
| `gofmt -l <changed files>` (AC5) | PASS | New/changed files clean; only the documented pre-existing `pkg/nfe/xml/builder.go` item remains (same 14 comment-alignment lines as HEAD). |

## Follow-up correction (2026-10-14, first SEFAZ homolog attempt)

The first real homolog emission arrived at a **SEFAZ 215** —
"cvc-complex-type.2.4.a: Invalid content was found starting with element
'fone'. One of 'email' is expected" — proving `<email>` inside `infRespTec`
is **mandatory (1-1)** in the 4.00 XSD. The Field premise in this SDD ("email
optional 0-1") was only valid for `dest/email` (E20), not the RespTec block.

Resolution (quick-fix, no new SDD per the operator):
- `defaults.RespTec.Email = "vagner.filhopro@gmail.com"` (provider
  supplied the address).
- `RespTecConfigured()` now gates on **all four** mandatory fields
  (CNPJ, xContato, email, fone) — an incomplete vendor identity omits the
  entire block instead of emitting one that SEFAZ will reject.
- The `infRespTec_email_omitted_when_empty` subtest was replaced by
  `infRespTec_full_provider_shape` (locks the complete 4-element block —
  the shape every real emission now carries).
- Re-verified: `make test-go` green (0 failures).

## Block disabled (2026-10-14, second homolog attempt)

Second real attempt arrived at **SEFAZ cvc-type.3.1.3 on `xContato`**: the
provider razão-social string is 72 chars while NF-e `TString` carries
maxLength 60 (SEFAZ*registration* is irrelevant to schema validation).
Outcome — `<infRespTec>` is emitted **nowhere** for now:
`RespTecConfigured()` returns `false` (values preserved for re-enable);
both flows omit the block, matching the pre-feature "working" state. To
re-enable: set a ≤60-char `xContato` (e.g. `VAGNER SIMONE MARTINS FILHO`)
and restore the four-field gate per the TODO in `pkg/nfe/defaults/resptec.go`.
