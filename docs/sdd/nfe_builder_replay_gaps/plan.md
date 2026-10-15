# plan:nfe_builder_replay_gaps

Implementation plan for closing the NF-e builder gaps surfaced by the
real-world replay. Spec: `docs/sdd/nfe_builder_replay_gaps/sdd.md`.

## Test strategy (defined before implementation)

- **Layer:** pure Go unit tests in `pkg/nfe/xml` (no DB/HTTP/SEFAZ).
- **Test selection during development:** focused `go test -run <pattern> ./pkg/nfe/xml/` per phase.
- **Before completion:** `make test-go` (AC8) + `gofmt -l pkg/nfe/xml/` (AC9).
- **Behaviors to cover** (mapped to SDD ACs AC1–AC7):
  - AC1: table test over CST 00/20/40/41/50/51/90 → emitted tag + child sequence (+ tag absent cases for CST 41 vs 50 sharing ICMS40).
  - AC2: PIS/COFINS switch table (08 → NT; 01 → Aliq unchanged).
  - AC3: numeric vItem/vNFTot for G1-style and G3-style (discount) inputs; det child order (`vItem` after `infAdProd` when both set).
  - AC4: small-field presence/position/omission rules.
  - AC5: finNFe, NFref (multi-key), entrega flat group, item+totals discounts.
  - AC6/AC7: golden registry flips + full 115-file round-trip (regression net for the registry edits).
- **Fixtures:** existing `pkg/nfe/xml/testdata/realworld/` (no regeneration needed; generator untouched).
- **Environment:** none; offline; guard untouched.

## Phase 1 — ICMS variants (`ICMSTag` gap)

1. `entity/invoice.go`: extend `ICMSData` with `PRedBC`, `ModBCST`, `VBCST`, `PICMSST`, `VICMSST`.
2. `pkg/nfe/xml/builder.go` `buildICMS`: switch on CST → emit ICMS00 (current path, byte-compatible)/ICMS20/ICMS40/ICMS51/ICMS90 per the verified sequences. ModBC default "3" handled by callers today; ICMS40/ICMS51-without-modBC? corpus always includes modBC for 51/90/20 — emit `modBC` when `ModBC != ""`.
3. Tests: new `TestBuilder_ICMS_<CST>` table; run; registry: retire `ICMSTag` required-gap (G1–G7) — keep `pIBSUFDrift`-style unrelated keys.

## Phase 2 — PIS/COFINS variants (`pisTag`/`cofinsTag`)

1. `buildPIS`: CST 04-09 → `PISNT` (single `<CST>`); else current `PISAliq`. Same for `buildCOFINS` → `COFINSNT`.
2. Test `TestBuilder_PIS_COFINS_Variants`; registry: retire `pisTag`/`cofinsTag`.

## Phase 3 — `vItem` + `vNFTot` (full NT formula) + discounts (`prodVDesc`/`icmsTotDesc`)

1. `entity/invoice.go`: `ProdutoData.VDesc decimal.Decimal`.
2. `buildDet`: prod `vDesc` when non-zero (between `vUnTrib`, `indTot`); after `imposto`/`infAdProd` emit `vItem` = `VProd − VDesc + IBSCBS.VIBSUF + VIBSMun + VCBS` (2dp via `formatDecimal`; computed from the struct values already shaped as emitted 2dp amounts). Compute per formula: NT terms beyond the five are 0 by design — code comment documents this and names them.
3. `buildTotal`: `ICMSTot/vDesc` = Σ item VDesc (replaces hardcoded "0.00"); `vNFTot` = Σ vItem after `IBSCBSTot`.
4. Test `TestBuilder_VItem_VNFTot_NumericFormula` (G1-style input: vItem == 61999.09 analog; G3-style discount input; det child order with infAdProd). Registry: retire `vItem`/`vNFTot` required-gaps; add allowed divergence `vItemFormula` (built `det/vItem` + `total/vNFTot` values intentionally differ from the flat-formula fixtures); add allowed `prodVDesc`/`icmsTotDesc` retirement notes; `total/ICMSTot/vDesc` value-diff classification moves from `icmsTotDesc` (G3 value-diff disappears once discounts are summed — retire key).

## Phase 4 — small fields (`dhSaiEnt`, `infRespTec`, `paisFields`, `destEmail`, `xPag`, `indPag`, `veicTranspUFEmpty`)

1. Entity: `InvoiceInput.DhSaiEnt *time.Time`, `RespTec *RespTecData{CNPJ, XContato, Email, Fone}`, `RecipientData.Email`, `PagamentoDetalhe.XPag`.
2. Builder: `dhSaiEnt` (RFC3339, when set; after `dhEmi`); `infRespTec` last inside `infNFe` (each field `SanitizeSchemaString`d; omit block when nil); `cPais`="1058"/`xPais`="Brasil" after CEP in enderEmit and enderDest; dest `email` after `IE`; `detPag/xPag` between tPag/vPag when set; `indPag` only when ≠ 0; skip empty `veicTransp/UF`.
3. Test `TestBuilder_SmallRealworldFields`. Registry: retire the seven corresponding allowed keys.

## Phase 5 — devolução / delivery (`finNFe`, `nfref`, `entrega`)

1. Entity: `InvoiceInput.FinNFe string` (default "1"), `NFRefs []string`, `Entrega *EntregaData{Type, CNPJ, CPF, XNome, Logradouro, Numero, Bairro, CodigoMun, Municipio, UF, CEP, IE}`.
2. Builder: `ide/finNFe` = input (default "1"); `NFref` block after `verProc` (free-text safe — keys are 44 digits, validate len(44) && digits, skip+log otherwise); `entrega` flat group as sibling of `dest` (after dest; sanitize names).
3. Test `TestBuilder_Devolution_Extras`. Registry: retire `finNFe` (G3 required) and `nfref` (G3 required) and `entrega` (G2 allowed) — G3 representative no longer needs the finNFe input since fixtures carry `finNFe=4` and the adapter passes it through.
   - NOTE: `realworld_test.go` adapter (`rwFixtureToInput`) updated to map the new entity fields (dhSaiEnt/RespTec/email/xPag/finNFe/NFRefs/entrega/vDesc) — otherwise the goldens can't exercise the capabilities.

## Phase 6 — wrap-up

1. `go test -run ^TestRealWorld ./pkg/nfe/xml/`, then full `go test ./pkg/nfe/...`.
2. `make test-go`; `gofmt -l pkg/nfe/xml/` (expect only the pre-existing `builder.go` flag).
3. Update `docs/research_nfe_use_case/USE_CASES.md` gap table to the final registry (also correct its `vNFTot` evidence citation: res_4.md:2074, not res_1.md).
4. Fill SDD Verification Results; Status → Done when ACs met.

## Risks / notes

- **vItem formula vs corpus:** fixtures use the flat formula (`vItem = vProd − vDesc`); the built XML will diverge on exactly those two paths — isolated by the `vItemFormula` allowed key with a documentation pointer to the user decision (res_1.md:1617 formula, res_4.md:5929 vNFTot = Σ vItem).
- **Registry flip must be atomic per phase** — goldens stay green between phases because each phase retires exactly the keys its fix addresses.
- **`indPag` rule:** fixtures mostly omit it; G3 has `indPag=0` while the builder (post-change) omits 0 → a new fixture-only `indPag` diff at G3; classification stays allowed.
- **SEFAZ-relevant validations** (dhSaiEnt ≥ dhEmi, finNFe/refs coherence) remain the caller's responsibility; the builder validates only NFref shape.
- Existing tests asserting `PISAliq`/ICMS00 for default inputs must stay untouched (compat guard).

## Out of scope

- Service-layer wiring (departure/detached flows populating DhSaiEnt/RespTec/NFRefs/Entrega/Discounts) — follow-up SDD.
- `<gMono>`, IPI/II groups, PIS/COFINS Outr/Qtde variants, ICMS vICMSDeson/pDif forms, freight/insurance/other per-item terms in the vItem formula.
- SEFAZ homolog smoke (human-driven; requires explicit `NFE_SEFAZ_GUARD=off`).
