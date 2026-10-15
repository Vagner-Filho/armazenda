# Homologação replay — accepted XMLs vs corpus representatives

**Dates:** 2026-10-14 / 15 (MT homologação, `tpAmb=2`; farm config = corpus producer identity with test names)
**Sources:**
- Accepted UI emissions: `pkg/nfe/xml/testdata/from_ui_to_sefaz_manual/nfe_*_g{N}.xml` (G1, G2, G4, G5, G6, G7 — G3 deferred, see below)
- Corpus representatives: `docs/research_nfe_use_case/9008711_xml_saida_55_*/…` (original, real-party documents)

## Coverage matrix

| Group | Corpus rep key | Homolog identity shape | ICMS group emitted | PIS/COFINS | tPag | Result |
|-------|----------------|------------------------|--------------------|------------|------|--------|
| G1 | …9461891752347 | dest **CPF** (corpus-farm person), IE present, indIEDest 1 | **ICMS90** + empty ST block | Aliq 01 | 99 (+ `xPag Outros`) | **Authorized** |
| G2 | …11041001838853 | dest **CPF**, IE present, indIEDest 1 | **ICMS51** pICMS 17 | Aliq 01 | 01 | **Authorized** |
| G4 | …10671320836294 | dest CNPJ, **no IE → indIEDest 9** | ICMS40 CST 50 | NT 08 | 01 | **Authorized** (after Rej. 531 → `docs/sdd/nfe_cst_zero_values` fix) |
| G5 | …09721024193426 | dest CNPJ + real IE (indIEDest 1) | ICMS40 CST 41 | NT 08 | 90, vPag 0.00 | **Authorized** |
| G6 | …10981728185986 | dest CNPJ, no IE (indIEDest 9) | **ICMS20** pRedBC 0, full declared base | NT 08 | 01 | **Authorized** |
| G7 | …10601100116727 | dest CNPJ + real IE; **transporta = dest itself + placa QHM0709 + (our extra) veh UF** | ICMS40 CST 41 | NT 08 | 90, vPag 0.00 | **Authorized** |

**G3** (CFOP 5556 + NFref): **not authorized — environment boundary.** The corpus refNFe exists only in MT's production base; homologação counts zero resolvable references → Rej. 321 ("NF-e de devolução não possui documento fiscal referenciado", rule B25-70). Code-side G3's XML shape was proven valid locally (refNFe emitted and totals consistent — verified in the DB copy of the signed XML). A homolog replay needs a referenced key that exists in the homolog base (e.g. one of this batch's own authorized keys, paired with a coherent CFOP like 1202). The two SEFAZ gates reached with G3 (321, plus 531 on an earlier attempt) validated the devolução shape up to the business rule.

## Structural divergence matrix

Findings from a path-set diff of each accepted XML vs its corpus representative (protocol envelope and signature excluded).

### Corpus-only (reference emitter carries, Armazenda omits — MT accepted the omission every time)

| Element | Why we omit | Status |
|---------|-------------|--------|
| `det/imposto/II` (all-zero group) | builder never emits II | accepted |
| `det/imposto/vTotTrib` (0.00) | not collected | accepted |
| `infAdic/infCpl` | form left blank → group omitted | accepted |
| `infRespTec` | RespTec gate disabled (`pkg/nfe/defaults/resptec.go`) | accepted — RespTec is optional at MT homolog |
| `IBSCBSTot/gMono` (+ 6 mono totals) | documented skip for the 2026 grain phase | accepted |

### Homolog-only (Armazenda carries, corpus omits — MT accepted the addition every time)

| Element | Why we add | Status |
|---------|-----------|--------|
| `pag/detPag/indPag` (always) | builder emits IndPag | accepted — cleanup candidate (Findings #3) |
| `transp/vol/esp` ("Granel", qVol 1 vs corpus 0) | UI volume prefill | cosmetic |
| `transp/veicTransp/UF` (G7) | template auto-fills emitter UF on plate entry | harmless (0-1 element) |
| `cobr/fat` (G1/G2/G4/G6 — exactly the tPag ≠ 90 documents) | builder emits the cobrança/fatura group for paying docs | accepted — corpus never carries it |
| `IBSCBS/gIBSMun/pIBSMun` scale (0.0000 vs 0.00) | decimal-formatter scale | cosmetic |

### Deliberate value divergences (test-episode choices, not engine gaps)

- **dest identity**: G1/G2 replayed with the corpus farm's own **CPF person** (IE present, indIEDest 1); G4/G6 reused the SEMEAR person **without IE** → `indIEDest 9`; G5/G7 carry the real CNPJ+IE pairs. All four variants were accepted, including no-IE CNPJ destinations.
- **ICMS family shifts**: G1 was intentionally run with **CST 90 + empty ST block** to close the `modBCST`-positional defect (cvc 2.4.a) — authorized. G2 ran **CST 51 with real pICMS 17** (corpus had 0.0000) — authorized. G6 ran CST 20 with pRedBC 0 and a full declared base (corpus declared all zeros) — authorized, arithmetically consistent.
- **masking**: dest/xNome is masked to the homolog phrase at emission (corpus carries real names).
- **city names emitted unaccented** (`Cuiaba`, `Colider`) — IBGE codes resolved correctly; both spellings work against the seed.
- **units**: `UND/PC → UN` (form select constraint), auto-numbered `cProd` (`1`), farm-config identity (`TESTE NOME CORRETO`, `Teste Arm.`) — test-environment artifacts.
- **`vItem` ≠ corpus**: Armazenda implements NT 2025.002 (`vItem = vProd − vDesc + vIBSUF + vIBSMun + vCBS`); the corpus uses flat `vItem = vProd` — registered allowed key `vItemFormula`.
- **`orig` 5→0** (G4 ICMS40) and `uTrib` mirrors `uCom` — cosmetic.

## Findings (MT homolog behaviors observed)

1. **Optional-group tolerance is total.** Every structural addition (`indPag`, `cobr`, `vol/esp`, `veicTransp/UF`) and every omission (`II`, `vTotTrib`, `infAdic`, `infRespTec`, `gMono`) was accepted. The gates that fired were business-value rules only: **321** (devolução without a resolvable ref) and **531** (totals ≠ Σ declared items).
2. **The totals rule is the hard gate.** SEFAZ compares totals against what the items actually declare — not against rate-derived expectations. The mapping-layer predicates (`ICMSDeclaresValues` / `PISDeclaresValues` / `COFINSDeclaresValues`, `docs/sdd/nfe_cst_zero_values`) now guarantee that coherence structurally.
3. **`indPag=1` is emitted on every detPag — including à-vista money payments (tPag 01).** Accepted by MT, but semantically wrong ("a prazo" on a cash payment). Cleanup candidate: omit `indPag` by default and emit it only when explicitly scheduled.
4. **`cobr` appears only on paying documents (tPag ≠ 90)** — an accidental mirror of the corpus behavior; the corpus never carries `cobr`, and MT accepted both. Worth a registry key so future golden-gap comparisons treat it as allowed.
5. **RespTec stays optional** for MT homolog; the disabled system gate can remain until a ≤60-char `xContato` is decided.
6. **ICMS90 with the emptied ST block is now proven against MT** (positional `modBCST` fix verified in G1).

## Follow-up candidates (not started — user decision required)

1. **`realworld_test.go` registry additions:** allowed keys for `cobr/fat` (paying docs), `detPag/indPag`, `vol/esp`, `veicTransp/UF`, `dest/indIEDest=9` without IE, unaccented `xMun`, ICMS51-with-values — so golden-gap comparisons accept the real-world shapes Armazenda emits.
2. **`indPag` emission cleanup** (omit by default; emit only when explicitly set).
3. **G3 homolog variant:** devolução referencing one of this batch's own authorized keys (coherent CFOP 1202), or the production replay once a real supplier devolução exists.
4. **RespTec re-enable** (≤60-char `xContato` decision).
5. **G6 semantic alignment** (optional): the corpus's zero-valued ICMS20 may deserve dedicated support (pRedBC input or ICMS40-family fallback) if Armazenda will actually emit asset sales.
