# plan:detached_nfe_input_wiring

Implementation plan for wiring the builder capabilities into the detached
NF-e (NF-e Avulsa) flow. Spec:
`docs/sdd/detached_nfe_input_wiring/sdd.md`.

## Test strategy (defined before implementation)

- **Layers**: builder unit tests (`pkg/nfe/xml/builder_test.go`), defaults
  helper tests (`pkg/nfe/defaults/agriculture_test.go`), service-level unit
  tests over pure helpers/struct mapping in `service/nfe_service` (no DB, no
  HTTP). No mocking framework — pure functions + hand-written stubs where a
  seam already exists (`detachedResponseModel` precedent).
- **Test selection during development**: focused
  `go test -run <pattern> <pkg>` per phase.
- **Before completion**: `make test-go` (AC11) +
  `gofmt -l pkg/nfe/xml/ pkg/nfe/entity/ pkg/nfe/defaults/ service/nfe_service/ router/nfe_router/`
  — expect only the documented pre-existing `builder.go` item.
- **Behaviors to cover** (mapped to SDD ACs):
  - AC1/AC2: builder defaults (dhSaiEnt=dhEmi equality; xPag tPag map).
  - AC3: tPag whitelist + `Payment` assembly.
  - AC4: email propagation both branches.
  - AC5: discount math per item, totals coherence, JSONB zero-default.
  - AC6: FinNFe table over curated CFOP set + corpus anchors + homogeneity.
  - AC7: nfRefs parse/validate.
  - AC8: entrega parse/mandatory-set/city-resolution mapping.
  - AC9: CST whitelist per tax.
  - AC10: migration + persistence writes (structural).
  - AC11: goldens diff-set audit + `make test-go`.
- **Fixtures/environment**: none beyond the repo; offline; SEFAZ guard
  untouched; no `nfe_farm_config` seeding anywhere (SEFAZ isolation rule).

## Phase 0 — Workflow baseline
1. Run `go test ./pkg/nfe/... ./service/nfe_service/` — note any
   pre-existing failures (none expected; baseline was green at
   `nfe_builder_replay_gaps` wrap-up).
2. Flip `sdd.md` Status → WIP.

## Phase 1 — Builder defaults (AC1, AC2 + AC11 registry audit)
1. `pkg/nfe/xml/builder.go` `buildIDE`: capture `now` once; `dhEmi` uses it;
   `input.DhSaiEnt == nil` → `<dhSaiEnt>` = `now` (RFC3339).
2. `pkg/nfe/xml/builder.go` `buildPag`: default `xPag` = `"Outros"` only when
   `XPag == ""` && `TPag == "99"`.
3. Tests (`builder_test.go`):
   - Rework `dhSaiEnt_omitted_when_nil` → `dhSaiEnt_defaults_to_dhEmi_when_nil`
     (assert equal text to `dhEmi`).
   - New subtests: xPag default for tPag 99; explicit XPag preserved; tPag
     01/90 keep omission.
4. Registry audit (AC11 part): `RW_DEBUG=1 go test -run
   '^TestRealWorldBuilder_PerGroupGolden' -v ./pkg/nfe/xml/` — adapter
   populates `DhSaiEnt`/`XPag`, so the diff set must be unchanged. If a diff
   appears: investigate implementation first; registry edit only with a
   documented justification in `realworld_test.go`.

## Phase 2 — Detached payment composition (AC3)
1. `router/nfe_router/detached_router.go`: `parseDetachedForm` reads `tpag`
   (default `"90"` when absent/empty); whitelist `{01,90,99}` with pt-br
   warning on other values.
2. `DetachedInvoiceInput` (or a small struct the router fills) carries the
   selected tPag; `prepareDetachedBuildData` payment assembly
   (`detached_service.go:240-249`) uses it (IndPag stays 1).
3. Tests: payment whitelist + assembly (`TestDetachedPayment_*` in a new
   `service/nfe_service/detached_wiring_test.go`).

## Phase 3 — Recipient email (AC4)
1. Person branch: `mapFullPersonToRecipient` (`service.go:1273-1345`) sets
   `Email` from `person.Email` — keep the shared function generic (benefits
   the future departure wiring too; no behavior there today since
   `RecipientData.Email` stays empty for departures).
2. Inline branch: `mapInlineRecipientData` (`detached_service.go:448-509`)
   consumes `DetachedRecipient.Email`.
3. Templates: nothing (field already posts `recipientEmail`).
4. Tests: `TestDetachedRecipientEmail_*` covering both branches + empty.

## Phase 4 — Item discount (AC5)
1. `DetachedItemInput` + `nfe_model.DetachedInvoiceItem`: add
   `Discount decimal.Decimal` (`json:"discount"` / `json:"discount,omitempty"`
   respectively; zero → `omitempty` keeps old JSONB rows valid).
2. `parseDetachedItems`: read `items[N].discount` (decimal parse; negative →
   pt-br reject).
3. `buildDetachedItems`: `ProdutoData.VDesc = Discount`; item net total =
   gross (`Quantity × UnitPrice` — keep the existing precision approach) −
   discount; validation discount ≥ gross → pt-br reject; totals coherent
   (`input.TotalValue = Σ net`).
4. Persistence: `items_json` automatically carries the new key (model write
   already serializes the struct).
5. Tests: extend `TestBuildDetachedItems_DefaultsAndTaxTotals` /
   `_UserRatesAndOverrides` (`detached_preview_test.go`) with discount rows;
   new `_DiscountBounds` rejection test; zero-discount no-op test.

## Phase 5 — Devolução: finNFe + NFRefs + migration (AC6, AC7, AC10)
1. `pkg/nfe/defaults/agriculture.go`: curated devolução set + `IsDevolucaoCFOP` /
   `FinNFeForCFOP` (+ table tests in `agriculture_test.go` over the set,
   corpus anchors, and non-members).
2. `DetachedInvoiceInput.NFRefs []string`; `parseDetachedForm` reads
   `nfRefs` textarea → split on whitespace/`,`/`;` → each entry must match
   44 digits (pt-br reject naming the bad value).
3. `buildDetachedItems` / prepare layer: per-item `FinNFeForCFOP`;
   heterogeneous result → pt-br warning (reuse/extend the mixed-CFOP
   homogeneity path); resolved finNFe set on `input.FinNFe`.
4. Migration `model/armazenda_database/migrations/000029_detached_nfe_wiring.sql`:
   `finnfe TEXT NOT NULL DEFAULT '1'`, `nf_refs JSONB`, `entrega JSONB`
   (idempotent; doc-comment header). NOTE also that the migration embed
   name in `model/armazenda_database` — check how migrations are
   discovered (`//go:embed` glob) before numbering.
5. Persistence write: `detached_model` create/update path populates the new
   columns (pgx.NamedArgs only). No reader (documented).
6. Tests: NFRefs parse/validate (`TestDetachedNFRefs_*`), homogeneity
   (`TestDetachedFinNFe_*`).

## Phase 6 — Entrega (AC8, AC10)
1. `DetachedInvoiceInput.Entrega *entity.EntregaData`;
   `parseDetachedEntrega(input)` in the router: all-empty → nil; any
   non-empty field → mandatory set {nome, documento, logradouro, número,
   bairro, cidade, UF} with pt-br missing-field warning; document type
   inferred (14→CNPJ/11→CPF; invalid lengths → pt-br reject); `entregaIE`
   / `entregaCEP` optional.
2. City → IBGE resolution: reuse the inline recipient municipality pathway
   (`mapInlineRecipientData` source); unresolvable supplied city → pt-br
   reject.
3. Persistence: writes `entrega` JSONB at emission (Phase 5 column).
4. Templates: optional delivery-address section on
   `templates/pages/nfe-emit.html` (pt-br labels: "Endereço de entrega";
   fields nome/CNPJ-CPF/IE/endereço/número/bairro/cidade/UF/CEP). Preview
   passthrough: add the entrega fields to the confirm fragment template
   chain (`nfe-detached-preview.html` hidden inputs) so build re-submits
   them.
5. Tests: `TestDetachedEntrega_*` (empty/partial/complete/type inference/
   city failure).

## Phase 7 — CST selection hardening (AC9)
1. Templates: `icmsCST`/`pisCST`/`cofinsCST` → `<select>` in both
   `templates/pages/nfe-emit.html` and `templates/nfe/nfe-emit-modal.html`
   (departure modal keeps parity of options; value still pre-filled from
   `DefaultICMSCST` etc.). Options: ICMS `{00,20,40,41,50,51,90}`;
   PIS/COFINS `{01,02,04,05,06,07,08,09}`.
2. Service: whitelist validation applied in the detached resolution path
   (after `MergeEffectiveOverrides`) — out-of-whitelist ⇒ pt-br warning
   naming the CST. Shared helper (`validateresolvedCodes`-style) so the
   future departure wiring can reuse it.
3. Tests: `TestDetachedCSTValidation_*` (in-whitelist pass; out-of-set
   pt-br; Simples regime keeps CSOSN path).

## Phase 8 — Wrap-up
1. `go test -run '^TestRealWorld' ./pkg/nfe/xml/` (goldens + round-trip).
2. `make test-go`; gofmt audit (expected output: only `pkg/nfe/xml/builder.go`).
3. Fill SDD Verification Results (exact commands + PASS/FAIL/NOT RUN).
4. Status → Done when ACs met; otherwise keep WIP with reasons.

## Risks / notes
- **Money math (discounts)**: keep the numeric approach used by the item
  totals today (decimal, no float). Truncate at the same boundaries the
  builder formats (2dp). The corpus invariant to protect: totals coherent
  (vNF = Σ net; vNFTot divergence via `vItemFormula` remains an accepted
  builder-side divergence).
- **Registry interplay**: builder defaults must not change golden diff sets
  (adapter populates the fields). Phase 1 includes the audit gate before the
  service work proceeds.
- **JSONB backward compatibility**: `DetachedInvoiceItem.Discount` with
  `omitempty` + zero-default map keeps old `items_json` values loadable.
- **Mixed-CFOP homogeneity**: finNFe rule rides the same uniformity path as
  the naturezaOp rule — one rejection place, pt-br message mentioning
  "devolução".
- **CST `03` exclusion**: document in the template helper text (pt-br) that
  CST 03 é não suportado pelo emissor — avoids operator confusion when the
  legacy value was numeric-free-text and allowed.
- **IndPag UI**: left untouched (fixed 1) — no scope creep.
- **CFOP-selector/catalog interplay**: devolução CFOPs exist in the
  catalog (`000025_cfop_catalog.sql`); the cfop-selector used by detached
  items exposes them — no new catalog entries expected.

## Out of scope
- Departure-linked NF-e (see the SDD's explicit note; includes the
  rebuild-fidelity gap where the worker passes `cfop=""`).
- infRespTec (farm config, SaaS fallback) — later SDD.
- Partial payment details (multiple detPag, IndPag selection), ICMS CST
  10/70 groups (builder default falls back to ICMS00 shape — whitelisted out
  here), CST 03 support in the PIS/COFINS builder branch.
- Detached-draft automatic rebuild consuming the new columns (worker TODO).
