# NF-e Real-World Use-Case Taxonomy

Source: `docs/research_nfe_use_case/` — 115 authorized `<nfeProc>` 4.00 XMLs exported from the third-party emitter used by the farm (CPF-bearing farm producer, Colíder-MT, serie 921), covering months 04/06/07/08/09 of 2026. Companion SDD: `docs/sdd/nfe_use_case_replay/`.

Every file: `cStat=100` (authorized), `tpAmb=1` (production), `tpEmis=1` (normal), **exactly 1 item**, `finNFe` per group, `indFinal=0`, `indPres=9`, `indIntermed=0`, `tpImp=1`, `procEmi=0`, `idDest=1`, full IBS/CBS block (`CST=000`, `cClassTrib=000001`, `pIBSUF=0.1`, `pCBS=0.9`), `<II>` zero group, `<PISNT>`/`<COFINSNT>` with CST 08, `<gMono>` zero in totals, `<vNFTot>`, `indIEDest=1`, destination UF=MT, `uCom == uTrib` (no unit conversion).

**Key observation: every ICMS/II/PIS/COFINS amount is zero in all 115 files.** The farm's tax treatment (deferral per art. 573-586 RICMS/MT, exemption, suspension) is expressed exclusively through the ICMS **tag + CST choice** and free-text `infCpl` notes — never through computed tax values.

## Groups (G1–G7)

| # | Files | natOp | CFOP | finNFe | ICMS tag | ICMS CST | tPag | Representative access key |
|---|-------|-------|------|--------|----------|----------|------|---------------------------|
| G1 | 84 | 5101 VENDA DE PRODUCAO | 5101 | 1 | ICMS51 | 51 | 99 | `51260400083503463100559210000009461891752347` |
| G2 | 23 | 5118 VENDA DE PRODUCAO EM VENDAS A ORDEM | 5118 | 1 | ICMS51 | 51 | 01 | `51260900083503463100559210000011041001838853` |
| G3 | 3 | 5556 DEVOLUCAO DE COMPRA MAT USO CONSUMO | 5556 | 4 | ICMS90 | 90 | 90 | `51260700083503463100559210000010641315035171` |
| G4 | 2 | 5913 RETORNO DE MERCADORIA OU BEM RECEBIDO PARA DEMONSTRACAO | 5913 | 1 | ICMS40 | 40 / 50 | 01 | `51260700083503463100559210000010671320836294` |
| G5 | 1 | 5915 REMESSA PARA CONSERTO | 5915 | 1 | ICMS40 | 41 | 90 | `51260400083503463100559210000009721024193426` |
| G6 | 1 | 5551 VENDA IMOBILIZADO | 5551 | 1 | ICMS20 | 20 | 01 | `51260900083503463100559210000010981728185986` |
| G7 | 1 | 5949 OUTRAS SAIDAS DE MERCADORIAS OU PRESTACAO DE SERVICOS | 5949 | 1 | ICMS40 | 41 | 90 | `51260600083503463100559210000010601100116727` |

Counts sum to 115 ✓. Group boundaries are: natOp/CFOP pair + finNFe + ICMS tag. Everything else is a sub-variant.

## Sub-variant matrix

| Variant | Distribution | Notes |
|---------|--------------|-------|
| `modFrete` | 1 ×84, 0 ×21, 9 ×10 | Always `transporta` + `veicTransp`, except nNF 1063 (no transporta/veic) |
| Destination doc | CNPJ ×114, **CPF ×1** (nNF 1063, July folder) | All industrial/agro buyers with IE; the CPF case is the only non-CNPJ |
| Discounts | nNF 1064, 1065 (both G3 devoluções) | `vItem = vProd − vDesc` there; everywhere else `vItem = vProd` |
| Units | KG ×109, PC ×3, UN ×1, UND ×2 | No uCom/uTrib conversion in any file |
| `infAdProd` | nNF 1091 (G4), 1098 (G6) | |
| `indPag` | 0 only in 3 files | |
| `infCpl` markers | FUNRURAL, DIFERIDO, FETHAB, CND, IBPT | G1/G2 grain-sale boilerplate incl. CND authentication codes, motorista name, contract numbers |

## Value invariants (verified across all 115 files)

- `vItem == vProd − vDesc` — **the IBS/CBS amounts are NOT added to vItem** in this corpus (NT 2025.002 formula would add `vIBSUF + vIBSMun + vCBS`; the accepted-by-SEFAZ-MT real XMLs do not). Decided in `docs/sdd/nfe_builder_replay_gaps/`: the Armazenda builder uses the **full NT formula** (`vItem = vProd − vDesc + vIBSUF + vIBSMun + vCBS`), so built `det/vItem`/`total/vNFTot` deliberately differ from these fixtures — allowed registry key `vItemFormula`.
- `vNFTot == vNF` and `vNF == Σ vItem` (single item).
- `vBC == vIBSUF / 0.001 == vCBS / 0.009` per item (symbolic rates).
- `dhRecbto ≈ dhEmi + ~1 min`; `verAplic = MT_A2_SRV240225-4.00`.
- No `<cobr>` block in any file (payments without installment/cobrança).

## Not present in this corpus (export gaps — cannot be replayed from real data)

- Cancellation (`procEventoNFe`, tpEvento 110111)
- Contingency (tpEmis 6/7 with `dhCont`/`xJust`)
- Denial (`cStat` 110/301/302), inutilização, async authorization (`cStat` 105→100 polling)
- Multi-item NF-e, NFS-e, homologação documents

## Use-case → Armazenda mapping

Full pass results (parser round-trip over all 115 files + builder goldens per group) are locked by the gap registry in `pkg/nfe/xml/realworld_test.go` (`pkg/nfe/xml/testdata/realworld` fixtures). If a gap below is fixed, update that registry together with this table.

| Use case / behavior | Status | Evidence |
|----------|--------|----------|
| ICMS51 (deferral) — G1, G2 (93% of corpus) | **Supported** (`nfe_builder_replay_gaps`) — `buildICMS` switches per CST: ICMS00 (unchanged default), ICMS20/40/41/50/51/90 with the corpus-verified child sequences; `ICMSData` gained `PRedBC` (+ ICMS90 ST terms). `modBC`/`modBCST` are 0-1 and only emitted when set. Former registry key `ICMSTag` (retired). | builder.go `buildICMS` |
| ICMS40 / ICMS90 / ICMS20 — G3–G7 | **Supported** — same switch (see row above). | builder.go `buildICMS` |
| `finNFe=4` (devolução) — G3 | **Supported** — `InvoiceInput.FinNFe` (empty defaults to `"1"`), emitted as-is. Former registry key `finNFe` (retired). | builder.go `buildIDE` |
| `<vItem>` per item — all files | **Supported (full NT formula — deliberate divergence)** — 1-1 mandatory per reform layout (rejeição 1137 `vItem não informado`, res_1.md H01); emitted as the LAST `det` child with `vItem = vProd − vDesc + vIBSUF + vIBSMun + vCBS` (user decision `nfe_builder_replay_gaps`; the corpus' flat `vProd − vDesc` variant appears as allowed key `vItemFormula`). Remaining NT terms (vICMSST, vFCPST, vFrete, vSeg, vOutro, vII, vIPI, vIPIDevol, vServ, PISST/COFINSST, vIS) are not representable in the current entity and are 0 by construction. | builder.go `buildDet`/`itemVItem` |
| `<vNFTot>` in totals — all files | **Supported (deliberate divergence)** — `vNFTot = Σ vItem` (soma do total dos itens, res_4.md:2074 W60), last child of `<total>` after `<IBSCBSTot>`. Deliberately ≠ `vNF` (vNF stays sourced from `input.TotalValue`). Allowed under `vItemFormula`. | builder.go `buildTotal` |
| `PISNT`/`COFINSNT` (CST 08) — all files | **Supported** — CST 04–09 emit `PISNT`/`COFINSNT` (single `<CST>` child); CST 01/02 keep `PISAliq`/`COFINSAliq`. Former registry keys `pisTag`/`cofinsTag` (retired). | builder.go `buildPIS`/`buildCOFINS` |
| `dhSaiEnt` — all files | **Supported** — `InvoiceInput.DhSaiEnt *time.Time` (RFC3339, after `dhEmi`); nil omits; caller owns the ≥ dhEmi validation. Former registry key `dhSaiEnt` (retired). | builder.go `buildIDE` |
| `infRespTec` — all files | **Supported** — `InvoiceInput.RespTec *RespTecData` (CNPJ/xContato/email/fone), last child of `infNFe`; nil omits the block. Former registry key `infRespTec` (retired). | builder.go `buildRespTec` |
| `NFref` document references — G3 (devolução) | **Supported** — `InvoiceInput.NFRefs []string` → one `NFref`/`refNFe` per 44-digit key, inside `ide` after `verProc`; invalid entries are skipped. Former registry key `nfref` (retired). | builder.go `buildIDE` |
| `<entrega>` delivery-site group — G2 (vendas a ordem) | **Supported** — `InvoiceInput.Entrega *EntregaData` (flat group: doc-choice, xNome, xLgr, nro, xBairro, cMun, xMun, UF, CEP, IE), sibling of `<dest>`. Former registry key `entrega` (retired). | builder.go `buildEntrega` |
| Item discounts — prod `vDesc` and `ICMSTot/vDesc` — G3 | **Supported** — `ProdutoData.VDesc`: `prod/vDesc` emitted between `vUnTrib`/`indTot` when non-zero (also feeds the untouched NT vItem term); `ICMSTot/vDesc` = Σ item discounts (no longer hardcoded 0.00). Former keys `prodVDesc`/`icmsTotDesc` (retired). | builder.go `buildDet`/`buildTotal` |
| Per-item `pIBSUF` derivation — G3 | **Divergence** — real third-party writes the flat rate (`0.1000`); builder derives it from rounded `vIBSUF/vBC` → `0.1001-0.1002` on the 3 devolução files. Registry key `pIBSUFDrift`. | builder.go `perItemIBSRate` |
| `pIBSMun` decimal places | Cosmetic — builder 4dp (`0.0000`), real corpus 2dp (`0.00`); both SEFAZ-accepted. Registry key `ibsRateDP`. | builder.go `buildIBSCBS` |
| `<cobr>/<fat>` — real corpus never emits | Divergence — builder emits when TotalValue ≠ 0 (builder.go:532-541); schema-valid but divergent from the reference emitter. Registry key `cobr`. | builder.go |
| `cPais`/`xPais` (emit + dest) — all files | **Supported** — always emitted (`1058`/`Brasil`) after CEP in enderEmit and enderDest. Former registry key `paisFields` (retired). | builder.go `buildEmit`/`buildDest` |
| `dest/email` — real files have it | **Supported** — `RecipientData.Email`, emitted after `IE`. Former registry key `destEmail` (retired). | builder.go `buildDest` |
| `xPag` in detPag — most files | **Supported** — `PagamentoDetalhe.XPag`, emitted between `tPag`/`vPag` when set. Former registry key `xPag` (retired). | builder.go `buildPag` |
| `indPag` always emitted | **Supported as omission** — builder now emits `indPag` only when ≠ 0, matching the corpus' default; the 3 corpus files carrying `indPag=0` are fixture-only diffs. Registry key `indPag` (allowed). | builder.go `buildPag` |
| `veicTransp/UF` empty element emitted | **Supported as omission** — UF skipped when unknown (placa-only when empty). Former registry key `veicTranspUFEmpty` (retired). | builder.go `buildTransp` |
| `<vTotTrib>`, `<II>` zero group | Cosmetic gaps — optional fields, never emitted. Registry keys `vTotTrib`/`iiGroup` (required in the golden registry — the last intentionally-unemitted optional groups). | builder.go |
| `<gMono>` in totals | Intentional skip (AGENTS.md: 0-1 optional, "skipped for grain") — real corpus includes it zero-valued. Registry key `gMonoTot`. | builder.go |
| tPag 90 (`vPag=0.00`), tPag 01/99 | **Supported** — `PaymentData` incl. tPag 90 special case (builder.go:543-556). | builder.go |
| `modFrete` + transporta + veicTransp + vol | **Supported** — `TransportData` (builder.go:484-530); only the empty-`UF` cosmetic above. | builder.go |
| CPF emitter (personal farm producer) / CPF destination (nNF 1063) | **Supported** — `EmitterData.Type=2` / `RecipientData.Type=2` (builder.go:153,187). | builder.go |
| `infAdProd` (nNF 1091, 1098) | **Supported** (builder.go:251-253). | builder.go |
| IBS/CBS per-item block + totals (values) | **Supported** — adapter-driven goldens match on all values except the rate-formatting notes above. | builder.go:358-482 |
| `infCpl` free text (FUNRURAL/FETHAB/CND boilerplate) | **Supported** — passes through `SanitizeSchemaString` unchanged for ASCII text; auto-generated agrotax boilerplate lives in the service layer (out of scope). | builder.go:558-561 |
| `natOp`/`CFOP` free values, `serie`, `nNF`, `cNF` | **Supported** — pass-through. | builder.go:90,230 |
| `dhEmi` = now | Builder behavior (time.Now) — expected, not a corpus gap. Registry key `dhEmiNow`. | builder.go:94 |

## Replay strategy

Fixtures: all 115 files deterministically anonymized into `pkg/nfe/xml/testdata/realworld/` (people/companies synthesized, business data preserved) + `manifest.json` (group per file). Tests: parser round-trip over all files; builder golden per group with an explicit allowlist of the gaps above. Details in `docs/sdd/nfe_use_case_replay/plan.md`.
