# sdd:nfe_cst_zero_values

## Status
WIP

## Goal
Make the per-item tax entity values mirror exactly what the XML declares for each CST, so the totals block (`ICMSTot`) always equals the sum of the item-declared values. This fixes SEFAZ-MT's homologação rejection **531 – "Total da BC ICMS difere do somatório dos itens"** seen on the G4 replay (NF-e Avulsa, CFOP 5913, ICMS CST 50, PIS/COFINS CST 08) and pre-empts the PIS/COFINS totals rules that would fire next.

## Evidence (homolog, 2026-10-15)

- G4 attempt (DB row `detached_nfe_invoice` id 29, `512610000835…`): item XML = `<ICMS40><orig>0</orig><CST>50</CST></ICMS40>` (declares **nothing**), but `ICMSTot` = `vBC 600000.00`, `vICMS 102000.00` (= farm default 17% × vProd), `vPIS 6000.00`, `vCOFINS 42000.00` despite `PISNT/COFINSNT CST 08` declaring nothing. SEFAZ: cStat 531.
- G3 attempts (DB rows 26–28): items used ICMS90, which **does** declare `vBC` at item level, so totals were self-consistent — 531 never fired (rule 321 rejected first for other reasons).
- Root cause chain: `service/nfe_service/detached_service.go` `buildDetachedItems` fills `Imposto.ICMS.VBC/VICMS/PICMS` and `Imposto.PIS.VBC/PPIS/VPIS` / `COFINS` **unconditionally** from the farm default rates (lines ~729–734 and ~787–798), regardless of the selected CST. The XML builder correctly omits those values for CST 40/41/50 and PIS/COFINS 04–09 (`pkg/nfe/xml/builder.go:408-411`, `505-518`, `524-537`), but `buildTotal` (`builder.go:603-642`) aggregates the **entity** values unconditionally → totals contain values no item declared.

## Requirements

### Functional
1. For items with ICMS CST in **{40, 41, 50}**, the item tax data must carry `VBC = 0`, `VICMS = 0`, `PICMS = 0` (regardless of farm default rates).
2. For items with ICMS CST in **{00, 20, 51, 90}**, the current behavior is preserved (these groups declare vBC/pICMS/vICMS and are self-consistent).
3. For items with PIS CST in **{04, 05, 06, 07, 08, 09}**, the item tax data must carry `VBC = 0`, `PPIS = 0`, `VPIS = 0` (the XML group `PISNT` declares only `<CST>`).
4. Same rule for COFINS CST {04–09} → `COFINS.VBC = PCOFINS = VCOFINS = 0`.
5. PIS/COFINS CST **01/02** (and any other `PISAliq`-routed code) keep current behavior.
6. The same rules apply to the **departure-flow** item mapper (`service/nfe_service/service.go`, item-imposto mapping around lines 900–930), which currently has the same unconditional fill.
7. The `ICMSTot` aggregation (`buildTotal`) itself is NOT changed — it remains a plain sum of entity values; correctness is restored at the mapping layer.
8. Values persisted per emission (invoice row totals derived from the same entities) become coherent as a side effect; no schema or migration change.

### Non-Functional
- No XML shape changes: the emitted groups stay exactly as today (corpus fidelity preserved; G-set registry unaffected).
- All tests offline: no test may dial SEFAZ (network guard rules apply).
- No new user-facing text; no pt-br message changes.
- Formatting: `gofmt` clean, except the documented pre-existing `builder.go` flag (leave untouched).

## Acceptance Criteria
Each criterion maps to an automated test or a justified manual check. (Regexes below are indicative; final test names may differ, keep the coverage.)

1. **AC1** — Detached item with ICMS CST 50 (farm rates configured, Lucro Real) produces `Imposto.ICMS` with `VBC=0, VICMS=0, PICMS=0`.
   `go test -run 'TestBuildDetachedItems.*beZeroesValues' ./service/nfe_service/`
2. **AC2** — Same for ICMS CST 40 and 41 (table-driven in the same test).
3. **AC3** — Same items with PIS CST 08 and COFINS CST 08 produce `VBC=0, PPIS=0, VPIS=0` and `VBC=0, PCOFINS=0, VCOFINS=0`. At least one more NT code (e.g. 04) parameterized.
4. **AC4** — CST 00 (and 90) keep non-zero values equal to today's behavior (existing coverage stays green).
   `go test ./service/nfe_service/`
5. **AC5** — A built XML whose item tax data is zeroed (CST 40-family + PISNT) renders `ICMSTot` with `vBC=0.00`, `vICMS=0.00`, `vPIS=0.00`, `vCOFINS=0.00` while `vProd`/`vNF` remain the item totals.
   `go test ./pkg/nfe/xml/ -run 'TestBuilder.*Total'`
6. **AC6** — Departure-flow mapping zeroes the same values for 40/41/50 + 04–09.
   `go test -run 'Test.*Imposto.*Zero|TestBuild.*Items.*' ./service/nfe_service/`
7. **AC7** — Full Go suite green.
   `make test-go`
8. **AC8 (manual)** — Homolog re-run of the G4 Avulsa payload (CFOP 5913, CST 50, PIS/COFINS 08) is accepted by SEFAZ-MT (or fails on a genuinely different rule). **Why manual:** requires touching SEFAZ with `NFE_SEFAZ_GUARD=off`, human-driven per house rules. Record the result in Verification Results when performed.

## Verification Results
(implementation record — 2026-10-15)

- **Baseline (pre-change):** `go test ./pkg/nfe/xml/ ./service/nfe_service/` — ok (cached, no failures).
- **AC1–AC3** — detached mapper zeroing: `go test -run TestBuildDetachedItems_NonDeclaringCSTValuesAreZero -v ./service/nfe_service/` — **PASS** (table-driven ICMS 40/41/50 with PIS/COFINS 08 zeroed; IBS vBC intact; declaring CST 00 + 01/01 keeps farm rates).
- **AC6** — departure mapper zeroing: `go test -run TestBuildItems_NonDeclaringCSTValuesAreZero -v ./service/nfe_service/` — **PASS** (same families; item total 1000; IBS vBC intact; CST 00/01/01 keeps rates at 170/10/70).
- **AC5** — totals contract: `go test -run TestBuilder_ICMSTot_OnlyDeclaredValues -v ./pkg/nfe/xml/` — **PASS** (`ICMSTot vBC/vICMS/vPIS/vCOFINS = 0.00` with `vProd/vNF = 600.00`; `<ICMS40>` children limited to orig+CST).
- **AC1–AC4 predicates** — `go test -run 'TestICMSDeclaresValues|TestPISAndCOFINSDeclaresValues' ./pkg/nfe/defaults/` — **PASS**.
- **AC7** — full suite: `make test-go` — **PASS** (no FAIL across all packages, e.g. `ok service/nfe_service`, `ok pkg/nfe/xml`, `ok pkg/nfe/defaults`).
- **Formatting:** `gofmt -l` clean for all files touched by this SDD. Remaining pre-existing flags, left untouched: `pkg/nfe/xml/builder.go` (documented at HEAD), `pkg/nfe/defaults/agriculture_test.go` (uncommitted earlier-session state), `pkg/nfe/defaults/reform.go`, `pkg/nfe/defaults/reform_test.go`.
- **AC8 (manual, pending):** homolog re-run of the unchanged G4 Avulsa payload (CFOP 5913, ICMS CST 50, PIS/COFINS 08) with `NFE_SEFAZ_GUARD=off air` — human-driven; recorded here when performed. Status stays **WIP** until then.

## Changed files
- `pkg/nfe/defaults/agriculture.go` — `ICMSDeclaresValues`, `PISDeclaresValues`, `COFINSDeclaresValues` (single source of truth)
- `service/nfe_service/detached_service.go` — `buildDetachedItems` gating (ICMS + PIS/COFINS)
- `service/nfe_service/service.go` — `buildItems` (departure flow) same gating
- Tests: `pkg/nfe/defaults/cst_declares_test.go`, `service/nfe_service/detached_items_cst_zero_test.go`, `service/nfe_service/build_items_cst_zero_test.go`, `pkg/nfe/xml/totals_declared_values_test.go`
- Not touched: `pkg/nfe/xml/builder.go` (`buildTotal` unchanged, emission shapes unchanged), schema/migrations.
