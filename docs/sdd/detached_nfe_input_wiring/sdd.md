# sdd:detached_nfe_input_wiring

## Status
Done

## Goal
Wire the NF-e XML-builder capability inputs delivered by
`nfe_builder_replay_gaps` into the **detached NF-e (NF-e Avulsa)** flow — the
only flow targeted by this SDD. Specifically: recipient email, payment
composition (`tPag` selection, which enables the builder's new `xPag`
default), per-item discounts, `finNFe` + referenced NF-e keys (devolução),
the `<entrega>` delivery group, and ICMS/PIS/COFINS CST selection hardening.
The XML builder itself gains two small, corpus-faithful defaults that apply
to every call immediately (`dhSaiEnt = dhEmi` when unset; `xPag` default for
tPag 99).

User decisions driving this spec:
- **dhSaiEnt** defaults to the `dhEmi` instant when `DhSaiEnt == nil`
  (corpus always emits `dhSaiEnt == dhEmi`; SEFAZ rejects `dhSaiEnt < dhEmi`
  — equality avoids the race of two separate `time.Now()` calls).
- **infRespTec**: deferred (left for a later SDD).
- **XPag**: builder default per `tPag` (`99` → `"Outros"`, else omitted),
  which requires exposing `tPag` in the detached form (today `TPag` is
  hardcoded `"90"` at `detached_service.go:245`).
- **Departure-linked NF-e is NOT targeted by this SDD** — detached feature
  only (see Out of scope note below).
- All phases in one batch; acceptance verified offline (no SEFAZ traffic).

## Out of scope — departure-linked NF-e (explicit note)

The departure emission flow (`prepareInvoiceBuildData`,
`service/nfe_service/service.go`) deliberately receives **no** wiring in this
SDD. When that follow-up happens it must address, at minimum:

- `DhSaiEnt` semantics for departures (the builder default `dhSaiEnt = dhEmi`
  already covers it with zero wiring; departure-date-based values would
  violate the `dhSaiEnt ≥ dhEmi` SEFAZ rule because dhEmi is emission-now).
- `RespTec` from farm config (deferred globally).
- `ProdutoData.VDesc` + the discount form field + the per-emission
  persistence columns the worker rebuild path needs (the rebuild currently
  re-derives from `nfe_invoice` columns and passes `cfop=""` → farm default —
  a pre-existing rebuild-fidelity gap that devolução/CFOP choices would
  amplify).
- CFOP select options in `nfe-emit-modal.html` (today hardcodes 5101/5102/5901
  — return CFOPs are not reachable there).
- `NatOp`, `Entrega`, `NFRefs`, `xPag` caller inputs are configured/free in
  the detached path but unused in the departure path.

infRespTec is likewise deferred **globally** (both flows). *(Resolved: the
system-scoped provider data is now wired in both flows by
`docs/sdd/resptec_system_defaults/`.)*

## Background

- Builder capabilities are already in place and regression-tested by the
  corpus goldens (`pkg/nfe/xml/realworld_test.go`) + `builder_test.go`
  (`TestBuilder_Devolution_Extras`, `TestBuilder_SmallRealworldFields`,
  `TestBuilder_ICMS`, `TestBuilder_PIS_COFINS_Variants`,
  `TestBuilder_VItem_VNFTot_NumericFormula`).
- Detached chain: `POST /nfe/emitir/preview` → `previewDetachedNFe`
  (`router/nfe_router/detached_router.go:766-791`) → build ▸
  `POST /nfe/emitir/build` → `buildDetachedNFe` (`detached_router.go:513+`)
  → `parseDetachedForm` (`detached_router.go:380-424`) →
  `detachedInputFromForm` (499-510) → `NFeService.BuildDetachedInvoice`
  → `prepareDetachedBuildData(input, dryRun)`
  (`detached_service.go:104-264`).
- Recipient email already exists in the data layer: `contact` table via
  `entity_public.FullPerson.Email` (`model/person_model/model.go:54`), the
  detached form posts `recipientEmail` (`detached_router.go:874-876`) —
  neither is propagated to `entity.RecipientData.Email` today.
- Payment is hardcoded `Payment{IndPag: 1, Detalhes:[{TPag:"90"}]}`
  (`detached_service.go:240-249`).
- Detached items are persisted as JSONB (`items_json`,
  migration `000021_detached_nfe.sql`), item struct
  `nfe_model.DetachedInvoiceItem` (`model/nfe_model/detached_model.go:16`),
  service input `DetachedItemInput` / `DetachedInvoiceInput`
  (`detached_service.go:60`), item build in `buildDetachedItems`
  (`detached_service.go:580-740`).
- `detached_draft` rebuild is a `TODO` in the worker
  (`detached_service.go worker path: processDetachedDraftInvoice`) — new
  persistence columns for finNFe/NFRefs/entrega are written for record
  coherence today and become the inputs of that future rebuild.
- Next migration number: `000029` (latest is
  `000028_farmer_name.sql`).
- SEFAZ rule of engagement: all tests offline; the SEFAZ network guard
  (`pkg/nfe/sefaz/guard.go`) is untouched; no homolog smoke in this SDD.

## Requirements

### Functional

#### F1 — Builder defaults (global, small)
1. **dhSaiEnt default**: in `pkg/nfe/xml/builder.go` `buildIDE`, capture
   `now := time.Now()` once; `dhEmi` uses `now`; when
   `input.DhSaiEnt == nil` emit `<dhSaiEnt>` = `now` (RFC3339); when set,
   emit the given value as today. `dhSaiEnt == dhEmi` must hold textually
   whenever the default is used.
2. **xPag default**: in `buildPag`, when `det.XPag == ""` and
   `det.TPag == "99"` emit `<xPag>Outros</xPag>` (corpus G1 shape); explicit
   `XPag` is always used verbatim; all other tPag keep xPag omitted.

#### F2 — Detached payment composition
3. Detached form gains `tpag` `<select>` with options `01` (Dinheiro),
   `90` (Sem pagamento — default, preserves current behavior), `99`
   (Outros). `parseDetachedForm` reads it; values outside the set are
   rejected with a pt-br warning toast. `Payment` assembly uses the selected
   tPag; `IndPag` stays `1` and `vPag` semantics stay as today (builder
   forces `vPag=0.00` for tPag 90).

#### F3 — Recipient email (detached)
4. Person branch (`resolveOrCreateRecipient` →
   `mapFullPersonToRecipient`): set `RecipientData.Email` from
   `person.Email` when present. Inline branch
   (`mapInlineRecipientData`): consume the already-parsed
   `DetachedRecipient.Email` (form `recipientEmail`). Empty → element
   omitted (builder rule unchanged).

#### F4 — Item discount (detached)
5. Form: `items[N].discount` (pt-br "Desconto (R$)"). Parsed per item into
   `DetachedItemInput`; carried into `DetachedInvoiceItem.Discount` (JSONB
   key `discount`; absent/old records default to zero).
6. Math: per-item `VProd = Quantity × UnitPrice` stays gross; `VDesc` =
   discount; item **net** total = gross − discount; `input.TotalValue` =
   Σ net totals (monetary coherence; IBS/CBS amounts stay informational and
   excluded, as today). The existing rates/IBS plumbing is untouched.
7. Validation (pt-br): discount ≥ 0; discount < gross per item (net total
   must stay positive); non-numeric values rejected by the form parser.
   Discount `0`omits the element (builder rule) and leaves totals unchanged.

#### F5 — Devolução: finNFe + NFRefs (detached)
8. `pkg/nfe/defaults` gains `IsDevolucaoCFOP(cfop) bool` and
   `FinNFeForCFOP(cfop) string` over the curated return-CFOP set
   `{1202, 2202, 5202, 6202, 5555, 5556, 5557, 5558, 5559, 6555, 6556, 6557,
   6558, 6559}`; everything else → `"1"`. Corpus anchors: `5556 → 4`,
   `5913/5949/5551/5118/5101 → 1`.
9. A detached emission must have a homogeneous finNFe across items: if a
   mixed set is submitted, reject with a pt-br warning naming the offending
   CFOPs (extends the existing mixed-CFOP homogeneity treatment in
   `buildDetachedItems`/`resolveDetachedNaturezaOp` area).
10. Form: `nfRefs` textarea (optional). Parsing splits on whitespace and/or
    commas/semicolons; every entry must be a 44-digit numeric key or the
    whole input is rejected with a pt-br message identifying the bad entry.
    The (already validated) list flows to `input.NFRefs` → builder
    `NFref/refNFe`.
11. finNFe input precedence: service derives the default via
    `FinNFeForCFOP` per item; when the resolved value differs across items
    → reject (i.e. no manual finNFe override field in this phase).

#### F6 — Entrega (detached)
12. Optional `<entrega>` section on the detached form with
    `entregaNome`, `entregaDocumento` (CNPJ/CPF — type inferred by digit
    length 14/11), `entregaIE`, `entregaLogradouro`, `entregaNumero`,
    `entregaBairro`, `entregaCidade`, `entregaUF`, `entregaCEP`. All-empty
    section → no `<entrega>`; any non-empty field makes the mandatory set
    (**nome, documento, logradouro, número, bairro, cidade, UF**) required,
    with a pt-br warning listing the missing ones; IE/CEP remain optional
    (builder omits when empty).
13. `entregaCidade` (name) is resolved to the IBGE `cMun` via the same
    municipality-resolution used by the inline recipient path; failure to
    resolve an explicitly supplied city name is a pt-br warning rejected at
    parse time.
14. Mapped into `entity.EntregaData` (Type 1=CNPJ/2=CPF) and set on
    `input.Entrega`.

#### F7 — CST selection hardening (detached)
15. `icmsCST`, `pisCST`, `cofinsCST` become `<select>` fields. Allowed
    option sets (also enforced server-side at resolution, pt-br rejection):
    ICMS `{00, 20, 40, 41, 50, 51, 90}`; PIS/COFINS
    `{01, 02, 04, 05, 06, 07, 08, 09}` (CST `03` is not handled by the
    builder's PISAliq branch — excluded from options and rejected). Farm
    config defaults keep pre-filling the fields; Simples-Nacional farms keep
    the CSOSN behavior (ICMS CST input stays inert there — existing rule,
    no new UI).

#### F8 — Persistence (one migration)
16. Migration `000029_detached_nfe_wiring.sql` (idempotent, named-args
    model updates):
    - `detached_nfe_invoice.finnfe TEXT NOT NULL DEFAULT '1'`
    - `detached_nfe_invoice.nf_refs JSONB`
    - `detached_nfe_invoice.entrega JSONB`
    - `DetachedInvoiceItem.Discount` rides the existing `items_json`
      (backward-compatible: absent key → zero).
    Emission persistence writes these values so stored records stay
    coherent with the emitted XML. **No reader is added in this SDD** —
    they are the future inputs of the detached-draft rebuild (worker TODO).

### Non-Functional
- All new UI labels and server messages in **pt-br** (AGENTS.md rule).
- All free-text fields flow through the builder's `SanitizeSchemaString`
  (existing mechanism; no new handling needed).
- No SEFAZ traffic; network guard untouched; all tests offline.
- Any SQL change uses `pgx.NamedArgs` with named placeholders (no `$1`).
- gofmt: no new flags (`gofmt -l` must show only the documented pre-existing
  `pkg/nfe/xml/builder.go` item).
- Real-world goldens must keep passing without registry changes: the
  `rwFixtureToInput` adapter already populates `DhSaiEnt`/`XPag` from
  fixtures, so the new builder defaults must not alter the observed diff
  set (verified explicitly at Phase 1).

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | `dhSaiEnt` defaults to the `dhEmi` instant when `DhSaiEnt` is nil (equal text); honored when set. | `TestBuilder_SmallRealworldFields` subtests updated/added (`go test -run 'TestBuilder_SmallRealworldFields' ./pkg/nfe/xml/`) |
| AC2 | `xPag = "Outros"` emitted only for tPag 99 with empty XPag; explicit XPag preserved; tPag 01/90 omit xPag. | New subtest(s) in `TestBuilder_SmallRealworldFields` (`go test -run 'TestBuilder_SmallRealworldFields' ./pkg/nfe/xml/`) |
| AC3 | Detached form `tpag` drives `detPag/tPag`; `01`/`99` produce concrete vPag; `90` keeps `vPag=0.00`; out-of-set values rejected (pt-br). | Service-level test for payment assembly + validation helper (`go test -run 'TestDetachedPayment' ./service/nfe_service/`) |
| AC4 | Recipient email reaches `RecipientData.Email` in both branchs (person + inline); empty → omitted element. | Service tests for the recipient mapping helpers (`go test -run 'TestDetachedRecipientEmail' ./service/nfe_service/`) |
| AC5 | Per-item discount: item net = gross − discount; `TotalValue = Σ net`; discount ≥ gross or negative rejected (pt-br); zero discount → element omitted + unchanged totals; old JSONB items (no `discount` key) behave as zero. | `TestBuildDetachedItems_*` extensions + new discount tests (`go test -run 'TestBuildDetachedItems' ./service/nfe_service/`) |
| AC6 | `FinNFeForCFOP`: curated return set → "4"; corpus anchors `5556→4`, `5913/5949/5551/5118/5101→1`; mixed-CFOP finNFe → rejected (pt-br). | Table test in `pkg/nfe/defaults/agriculture_test.go` + service homogeneity test (`go test ./pkg/nfe/defaults/ && go test -run 'TestDetachedFinNFe' ./service/nfe_service/`) |
| AC7 | `nfRefs` parses 44-digit keys (whitespace/comma/semicolon separated); malformed entries rejected (pt-br, naming the bad value); valid list flows to `input.NFRefs`. | Service parse/validate tests (`go test -run 'TestDetachedNFRefs' ./service/nfe_service/`) |
| AC8 | Entrega section: all-empty → absent; partial → pt-br missing-field rejection; complete → `EntregaData` mapped (Type inferred); failed city resolution rejected (pt-br). | Service tests for entrega parse/map (`go test -run 'TestDetachedEntrega' ./service/nfe_service/`) |
| AC9 | CST selects render the allowed sets; resolution rejects out-of-whitelist values (pt-br). ICMS set `{00,20,40,41,50,51,90}`; PIS/COFINS `{01,02,04..09}`. | Service whitelist tests (`go test -run 'TestDetachedCSTValidation' ./service/nfe_service/`) + template presence check (manual/structural — records fields in the form reviews) |
| AC10 | Migration `000029` applies idempotently and emission persistence writes `finnfe`, `nf_refs`, `entrega` (+ `discount` inside `items_json`). | `go build ./...` + migration file review; applied on next app startup (recorded as structural check; no offline unit cover for DB I/O) |
| AC11 | Corpus goldens keep their diff set unchanged after the builder defaults; no regressions repo-wide. | `go test -run '^TestRealWorld' ./pkg/nfe/xml/` (diff audit with `RW_DEBUG=1` confirms same allowed set) + `make test-go` |

Manual checks: template rendering of the new form fields is verified in the
E2E/browser layer only if the environment is available — the SDD treats
template presence as a structural review point; all behavioral ACs are
offline-automated. No SEFAZ-driven check.

## Verification Results

| Command | Result | Notes |
|---------|--------|-------|
| `go test ./pkg/nfe/...` (baseline, pre-implementation) | PASS | Already green at `nfe_builder_replay_gaps` wrap-up (all packages ok); re-run at WIP flip. |
| `go test -run 'TestBuilder_SmallRealworldFields|TestBuilder_Devolution_Extras' ./pkg/nfe/xml/` (AC1–AC2 + builder defaults) | PASS | `dhSaiEnt` nil → **defaults to the dhEmi instant** (equal text, same clock); explicit value honored. `xPag` default `"Outros"` only for tPag 99 with empty XPag; explicit XPag preserved (sanitized); tPag 01/90 omit. One stale hard-coded `ide` child list in the NFref subtest was updated to include the now-always-present `dhSaiEnt`. |
| `RW_DEBUG=1 go test -run '^TestRealWorldBuilder_PerGroupGolden' -v ./pkg/nfe/xml/` (Phase-1 registry gate + AC11) | PASS (diff audit) | Golden diff set **identical before and after the builder defaults**: `verProc`, `pIBSMun`, `gMono`, `vTotTrib`, `II`, `dhEmi`, `cDV`, `cobr`, `vItem/vNFTot` (+ G3 `pIBSUFDrift`, G3 `indPag` fixture-only). No new diff paths → no registry change needed (adapter populates `DhSaiEnt`/`XPag` from fixtures). |
| `go test -run 'TestFinNFeForCFOP' ./pkg/nfe/defaults/` (AC6) | PASS | Table over the curated return set + corpus anchors (`5556→4`; `5913/5949/5551/5118/5101/5901→1`) and the `IsDevolucaoCFOP` consistency check. |
| `go test -run 'TestResolveDetachedTPag|TestValidDetachedTPag|TestValidateItemDiscounts|TestResolveDetachedFinNFe|TestParseNFRefs|TestValidateNFRefs|TestValidateDetachedCSTs|TestMapDetachedEntregaInput|TestApplyRecipientEmail|TestMapInlineRecipientEmail' ./service/nfe_service/` (AC3–AC9 helpers) | PASS | All wiring helpers: tPag default/whitelist, discount bounds (negative / ≥ gross / per-item indexing, net-positive rule), finNFe homogeneity (mixed sets rejected naming the item), 44-digit key parsing/validation, CST whitelists (ICMS `{00,20,40,41,50,51,90}`; PIS/COFINS `{01,02,04–09}`; 03/10 rejected), entrega mapping (all-empty→nil, partial→pt-br list, CNPJ/CPF inference, city resolver error/not-found), recipient email trim/blank rules + inline mapping (offline: City/State nil keeps the municipality lookup untouched). |
| `go test -run 'TestBuildDetachedItems' ./service/nfe_service/` (AC5 item math) | PASS | New `TestBuildDetachedItems_Discount` (G3-analog: gross vProd/IBS vBC untouched, `VDesc` set, DB item net total + discount, `TotalValue = Σ net`) and `TestBuildDetachedItems_DiscountZeroMatchesLegacy` (zero discount keeps totals byte-equal to the legacy behavior). |
| `go vet ./model/nfe_model/` + model tests | PASS | Pre-existing DB-backed model tests updated for the new `CreateDetachedInvoice` params (`finNFe`, `nfRefs`, `entrega`; `"1", nil, nil` for rows without wiring). |
| `go test ./router/nfe_router/` (templates/parsers) | PASS | Template compile check surfaced the missing `detachedPageItem.Discount` prefill field — added (rascunho-saved items render an empty discount). |
| `make test-go` (AC11) | PASS | Zero `^FAIL` lines across the whole repo. |
| `gofmt -l <changed pkgs>` | PASS | Only the documented pre-existing `pkg/nfe/xml/builder.go` item (same 14 comment-alignment lines as HEAD — no new flags introduced). |
| E2E/DB round-trip of migration `000029` | NOT RUN | No Docker env in this session; the migration is a plain idempotent `ADD COLUMN IF NOT EXISTS` set reviewed structurally (AC10). Columns are write-only today by design. |

## Implementation Notes (refinements recorded at wrap-up)

- **Recipient email seam**: `mapFullPersonToRecipient` (also used by the
  departure flow) was left untouched; the email lands through the shared
  `applyRecipientEmail` helper in the two *detached-only* branches
  (`resolveOrCreateRecipient` person branch + `mapInlineRecipientData`). This
  keeps the departure flow's behavior byte-identical while making both
  detached branches offline-testable (tests pass City/State nil so the
  municipality lookup stays untouched).
- **Entrega mandatory-set double gate**: the router performs the
  person-safe early rejection (before any write) and
  `mapDetachedEntregaInput` re-validates the same set — defense in depth for
  future direct callers of the service.
- **SQL house rule**: adding the wiring columns turned
  `CreateDetachedInvoice` into `pgx.NamedArgs` (the pre-existing positional
  `$1..$21` insert was converted).
- **SVC supersede**: the contingency path persists `finNFe` / `NFRefs` /
  `Entrega` straight from the rebuilt `InvoiceInput` (`input.FinNFe`, …),
  so a superseded row keeps the same wiring values as the original.
- **Rascunho profiles**: profiles intentionally do NOT store `tpag`,
  `nfRefs`, entrega or discounts (they ride the emission form and the
  hidden-field round-trip of each emission). Profile-saved items render an
  empty discount prefill (`detachedPageItem.Discount`).
- **IBGE resolver seam**: `mapDetachedEntregaInput` takes a
  `resolveCity` function (production: the nfe model's `GetMunicipio`;
  tests: stubs) so the name→IBGE resolution is testable offline.
