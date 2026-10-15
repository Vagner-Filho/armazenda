# sdd:nfe_builder_replay_gaps

## Status
Done

## Goal
Fix the NF-e XML-builder gaps surfaced by the real-world replay
(`docs/sdd/nfe_use_case_replay/`, `docs/research_nfe_use_case/USE_CASES.md`) so
that the builder can emit schema-valid XML for every use case the farm's
reference emitter uses, at SEFAZ-acceptable fidelity. Scope: `pkg/nfe/entity`
(input fields) + `pkg/nfe/xml/builder.go` + test registry updates. No
service-layer wiring, no DB changes, no SEFAZ traffic.

User decisions driving this spec:
- `vItem` uses the **full NT 2025.002 formula** (not the flattened variant the
  reference emitter uses).
- **One combined batch** (all phases in this SDD).
- **Capability only** — callers are NOT wired to populate the new inputs in
  this effort.

## Background

The real-world replay locked a per-group gap registry in
`pkg/nfe/xml/realworld_test.go` (fixtures in `pkg/nfe/xml/testdata/realworld/`).
Current builder divergences to fix, with registry keys:

- `ICMSTag` — `buildICMS` always emits `ICMS00` (builder.go:256-287); corpus
  is 93% `ICMS51` plus `ICMS90`/`ICMS40`/`ICMS20` (CST 51/90/40/41/50/20).
  No `pRedBC` input slot exists.
- `pisTag`/`cofinsTag` — always `PISAliq`/`COFINSAliq`; corpus uses
  `PISNT`/`COFINSNT` (CST 08) everywhere.
- `vItem`/`vNFTot` — never emitted (mandatory per reform layout: res_1.md H01
  `vItem` 1-1; res_4.md:2074 W60 `vNFTot`, "soma do total dos itens (vItem)").
- `dhSaiEnt`, `infRespTec`, `destEmail`, `paisFields` (cPais/xPais), `xPag`,
  `indPag`, `veicTranspUFEmpty` — missing input fields / cosmetic deltas.
- `finNFe` (hardcoded "1"), `nfref` (no input), `prodVDesc`/`icmsTotDesc` (no
  discount input; totals hardcode vDesc=0.00), `entrega` (no delivery-site
  input) — G3/G2 capabilities.
- `pIBSUFDrift`, `ibsRateDP`, `cobr`, `gMonoTot`, `dhEmiNow`, `accessKeyDV`,
  `verProc` — accepted divergences, unchanged by this SDD.

Verified element placements (from the accepted real XMLs):
- `prod`: `vDesc` sits between `vUnTrib` and `indTot` (G3 file 512607…1064…).
- `det` child order: `prod → imposto → infAdProd? → vItem` (vItem last).
- `ide`: `NFref` (with `refNFe` children) sits inside `ide`, after `verProc`.
- `total`: `ICMSTot → IBSCBSTot → vNFTot` (vNFTot last sibling).
- `detPag`: `tPag → xPag? → vPag`. `indPag` omitted by real files when 0.
- `enderEmit`/`enderDest`: `… CEP → cPais(1058) → xPais(Brasil) → fone?`.
- `dest`: `email` element after `IE`.
- ICMS groups (all zero-valued in corpus; exact child sequences):
  - ICMS51 (simplified): orig, CST, modBC, pRedBC, vBC, pICMS, vICMS
    (vICMSOp/pDif/vICMSDif omitted).
  - ICMS20: orig, CST, modBC, pRedBC, vBC, pICMS, vICMS
    (vICMSDeson/motivoDesoneracao omitted).
  - ICMS40: orig, CST only (CST 40/41/50).
  - ICMS90: orig, CST, modBC, vBC, pRedBC, pICMS, vICMS, modBCST, vBCST,
    pICMSST, vICMSST.
- PISNT/COFINSNT: a single `<CST>` child.

## Requirements

### Functional
1. **ICMS variants** — `buildICMS` switches on `ICMS.CST` (regime normal):
   `00` → ICMS00 (unchanged); `20` → ICMS20; `40|41|50` → ICMS40; `51` →
   ICMS51; `90` → ICMS90, with the child sequences above. `ICMSData` gains
   `PRedBC`, `ModBCST`, `VBCST`, `PICMSST`, `VICMSST`.
2. **PIS/COFINS variants** — CST `04`-`09` → `PISNT`/`COFINSNT` (single CST
   child); `01`/`02` keep `PISAliq`/`COFINSAliq` (unchanged behavior for
   existing flows); other CSTs keep the current Aliq default, documented in
   code as unsupported-by-corpus.
3. **vItem + vNFTot (full NT formula, user decision)** —
   `vItem = vProd − vDesc + VIBSUF + VIBSMun + VCBS` (per item, 2dp; the
   remaining NT formula terms — vICMSST, vFCPST, vFrete, vSeg, vOutro, vII,
   vIPI, vIPIDevol, vServ, PISST/COFINSST, vIS — are not representable in the
   current entity and default to 0; documented in code). Emitted as the LAST
   child of `det`, after `infAdProd` when present. `vNFTot = Σ vItem` emitted
   after `IBSCBSTot` — **deliberately ≠ vNF** (NT formula); `vNF`/`vPag`
   remain sourced from `input.TotalValue`. `ProdutoData` gains `VDesc`;
   `prod/vDesc` emitted when non-zero (between `vUnTrib` and `indTot`);
   `ICMSTot/vDesc` = Σ item vDesc (no longer hardcoded 0.00 when items carry
   discounts).
4. **Small fields** — `DhSaiEnt *time.Time` (emit RFC3339 when set);
   `RespTec` input struct → `infRespTec` block last inside `infNFe`
   (CNPJ/xContato/email/fone; omit block when input nil); `cPais`="1058",
   `xPais`="Brasil" after CEP in both enders (before fone); `Recipient.Email`
   after `IE`; `XPag` after tPag when set; emit `indPag` only when ≠ 0;
   skip `veicTransp/UF` when empty.
5. **Devolução / delivery capabilities** — `FinNFe` input (default "1",
   emitted as-is); `NFRefs []string` → `NFref` block inside `ide` after
   `verProc` (one `refNFe` per 44-digit key); `Entrega *EntregaData` (flat
   group: CNPJ/CPF, xNome, xLgr, nro, xBairro, cMun, xMun, UF, CEP, IE)
   emitted as a sibling of `dest`.
6. **Registry flip in the same phases** — after each fix, the gap registry in
   `pkg/nfe/xml/realworld_test.go` retires the corresponding required-gap
   assertion and (where built-XML values intentionally differ from the
   fixtures) introduces the documented allowed divergence key `vItemFormula`
   for `det/vItem` and `total/vNFTot` value diffs. USE_CASES.md gap table is
   updated to match the final registry at wrap-up.

### Non-Functional
- No DB migrations, no service/router changes, no template changes; SEFAZ
  guard untouched; all tests offline.
- Existing builder tests must keep passing unchanged (CST "00" + PIS/COFINS
  "01" default paths are current behavior).
- Formatting must keep `formatDecimal` precision consistent with the schema
  places used today (2dp values, 4dp rates); `vItem`/`vNFTot` arithmetic uses
  the emitted (rounded) per-item values to avoid ±0.01 drift.
- New free-text fields flow through `SanitizeSchemaString` like existing ones.
- Changed files gofmt-clean (`builder.go` has a documented pre-existing gofmt
  flag at HEAD — leave untouched; record as before).

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | ICMS groups: for CST 51/90/40/20/00 the builder emits ICMS51/ICMS90/ICMS40/ICMS20/ICMS00 with the exact child sequences verified against the corpus. | New unit tests in `pkg/nfe/xml/builder_test.go` (table over CSTs asserting tag + children); `go test -run ^TestBuilder_ICMS ./pkg/nfe/xml/` |
| AC2 | PIS/COFINS: CST 08 (and 04-09) emit PISNT/COFINSNT single-child; CST 01 keeps `PISAliq` shape (existing tests unchanged and green). | `go test -run ^TestBuilder_PIS ./pkg/nfe/xml/` |
| AC3 | `vItem` present as last `det` child = `vProd − vDesc + VIBSUF + VIBSMun + VCBS`; `vNFTot` present in `total` = Σ vItem; `vProd`-minus-discount and discount totals coherent for G3-style inputs. | New test `TestBuilder_VItem_VNFTot_NumericFormula`; also the existing goldens. |
| AC4 | Small fields: `dhSaiEnt`, `infRespTec`, `cPais/xPais`, `dest/email`, `xPag` emitted in verified positions; `indPag` omitted when 0; empty `veicTransp/UF` omitted. | New test `TestBuilder_SmallRealworldFields` (golden segment assertions). |
| AC5 | Devolução/extras: `finNFe` honored; `NFref/refNFe` inside ide after verProc; `<entrega>` flat group sibling of dest; `prod/vDesc` + summed `ICMSTot/vDesc`. | New test `TestBuilder_Devolution_Extras`. |
| AC6 | Golden registry: after all phases, `TestRealWorldBuilder_PerGroupGolden` passes with retired required gaps (long list → only intentional divergences `vItemFormula`, `pIBSUFDrift`, `ibsRateDP`, `pIBSMun`-style cosmetics, `cobr`, `gMonoTot`, `dhEmiNow`, `accessKeyDV`, `verProc`, `finNFe` for G3 stays retired once AC5 lands… final set recorded in the registry file). No unclassified diff paths. | `go test -run ^TestRealWorld ./pkg/nfe/xml/` |
| AC7 | Parser round-trip over all 115 fixtures still green; corpus inventory unchanged. | `go test -run ^TestRealWorld_ParseDANFEData ./pkg/nfe/xml/` |
| AC8 | No regressions anywhere. | `make test-go` |
| AC9 | Code is gofmt-clean (changed files only; pre-existing `builder.go` flag documented). | `gofmt -l pkg/nfe/xml/` output limited to the pre-existing list |

Manual checks: none (builder capability, offline tests cover it; homolog smoke is a separate human-driven step).

## Verification Results

| Command | Result | Notes |
|---------|--------|-------|
| `go test ./pkg/nfe/...` (baseline, pre-implementation) | PASS | All packages ok — no pre-existing failures. |
| `go test -run '^TestBuilder_(ICMS|PIS|VItem|SmallRealworld|Devolution)' ./pkg/nfe/xml/` | PASS | New AC1–AC5 unit tests: ICMS switch table (CST 00/20/40/41/50/51/90 + omission rules), PIS/COFINS NT variants + compat default, vItem/vNFTot numeric formula (G1/G3-style, det child order, truncation), small fields (dhSaiEnt/infRespTec/cPais,xPais/email/xPag/indPag/veicUF), devolution extras (finNFe/NFref validation/entrega). |
| `go test ./pkg/nfe/...` (mid-implementation, before registry flip) | FAIL (expected) | Only `TestRealWorldBuilder_PerGroupGolden` failed — the designed "required gap did not appear" loud failure for the just-fixed behaviors (ICMSTag/pisTag/cofinsTag/…); parser round-trip + corpus inventory stayed green. |
| `go test -run '^TestRealWorld' ./pkg/nfe/xml/` | PASS | AC6+AC7: corpus inventory (115 files, 7 groups), parser round-trip over all fixtures, per-group goldens with the flipped registry (required gaps = `vTotTrib`/`iiGroup` only). |
| `RW_DEBUG=1 go test -run '^TestRealWorldBuilder_PerGroupGolden' -v ./pkg/nfe/xml/` | PASS (diff audit) | Observed diffs per group = exactly the final allowed set: `verProc`, `ibsRateDP`, `gMonoTot`, `vTotTrib`, `iiGroup`, `dhEmiNow`, `accessKeyDV`, `cobr`, `vItemFormula` (+ G3-only `pIBSUFDrift`, G3 `indPag` fixture-only diff for the 3 corpus files with `indPag=0`). Nothing unclassified; all retired keys now replay byte-exact. |
| `make test-go` (AC8) | PASS | 72 package results, zero FAIL. |
| `gofmt -l pkg/nfe/xml/ pkg/nfe/entity/` (AC9) | PASS | Output limited to the pre-existing `builder.go` flag (14 comment-alignment lines in `buildTotal`'s totals block, present at HEAD — left untouched per spec). `pkg/nfe/entity/invoice.go` is gofmt-clean (the two pre-existing alignment hunks in `InvoiceOverrides` were fixed in passing with this file's changes; no deviation added). |

## Follow-up correction 2 (2026-10-14, ICMS90 `modBCST` positional)

A live SEFAZ 215 (`cvc-complex-type.2.4.a`) on a real homolog attempt with
`icmsCST=90`: `<vBCST>` cannot open the ICMS-ST block inside ICMS90 —
`<modBCST>` is **positional** and must lead it ("One of 'vBCFCP, pFCPDif,
modBCST, vICMSDeson, vICMSSTDeson' is expected"). The Phase-1 premise
"modBCST 0-1, emitted only when set" was wrong for ICMS90: the corpus always
ships the zero-valued ST block WITH its `modBCST` leader.
Resolution: the ICMS90 branch always emits `<modBCST>` (default `"0"` when
the caller left it unset — the corpus reference value); the regression test
that had locked the invalid skip (`CST90_without_modBCST_skips_only_that_child`)
was replaced by `CST90_modBCST_positional_when_unset`. Re-verified: focused
ICMS tests + `make test-go` green (0 failures); gofmt unchanged (only the
documented pre-existing `builder.go` item).
