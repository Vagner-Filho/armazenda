# plan:nfe_use_case_replay

Implementation plan for replaying the real-world NF-e corpus against Armazenda.
Spec: `docs/sdd/nfe_use_case_replay/sdd.md`.

## Test strategy (defined before implementation)

- **Layers:** pure Go unit tests in `pkg/nfe/xml` (no DB, no HTTP, no SEFAZ).
- **Fixtures:** all 115 anonymized XMLs + `manifest.json` in `pkg/nfe/xml/testdata/realworld/` — self-contained, independent of `docs/`.
- **Behaviors to cover:**
  1. Corpus integrity: 115 files, all classified, counts sum to 115 (AC1).
  2. Parser round-trip: every file parses via `ParseDANFEData`; key ↔ filename ↔ `infNFe Id` ↔ `protNFe chNFe` consistency; item/totals coherence; prot cStat=100 (AC3, AC4).
  3. Builder goldens per group G1–G7: entity inputs derived from representative real values; DANFE-level field comparison + tag-structure comparison against the real anonymized XML; per-group allowlist of known gaps (AC5).
  4. Identity-leak check: known real tokens absent from fixtures (AC3).
- **Commands:** `go test -run '^TestRealWorld' ./pkg/nfe/xml/` during work; `make test-go` before completion.
- **Environment needs:** none (no DB, no network); guard untouched.

## Phase 1 — Classification (spec + USE_CASES.md)

1. One-off extraction script in `/tmp/opencode/` walks all 115 XMLs and emits a per-file CSV with every discriminant (natOp, CFOP, finNFe, ICMS group+CST, PIS/COFINS CST, cClassTrib, tPag, indPag, modFrete, dest doc type, discounts, carrier/vehicle/vol, infCpl markers, prot data).
2. Verify every file maps to exactly one group key; counts sum to 115.
3. Write `docs/research_nfe_use_case/USE_CASES.md`: taxonomy table, representative access key per group, sub-variant matrix, use-case → Armazenda mapping (supported / gap / reject-risk) with concrete file/function references.

## Phase 2 — Anonymized fixtures

1. Write `pkg/nfe/xml/testdata/realworld/generate.go` (single self-contained generator; `testdata/` is ignored by Go builds):
   - Pass A: collect real identities (emit/dest/transporta/infRespTec names, CPF/CNPJ, IE, emails, phones, addresses, motorista names from infCpl) across all files → deterministic synthetic map (FNV hash → valid-check-digit CNPJ/CPF; synthetic ASCII names/addresses/emails/phones).
   - Pass B: rewrite each XML with etree: entities, infCpl via longest-first token replacement + `MOTORISTA:` regex, rebuild 44-char access key with synthetic emit document + recomputed cDV, sync `infNFe Id` / `protNFe chNFe`, synthetic nProt/digVal/DigestValue/X509 (documented as invalid signatures).
   - Emit `manifest.json` (per-file group + discriminants + provenance) and `README.md` (rules, provenance, per-group counts).
   - Sanity check each output: parses, key consistency, no real tokens.
2. Run generator: `go run pkg/nfe/xml/testdata/realworld/generate.go <in> <out>` and eyeball one output per group against its original.

## Phase 3 — Tests

1. `pkg/nfe/xml/realworld_test.go`:
   - `TestRealWorld_CorpusInventory` — manifest counts, sum, group set (AC1).
   - `TestRealWorld_ParseDANFEData` — table over all 115 files (AC3, AC4).
   - `TestRealWorldBuilder_*` — per-group goldens with allowlists (AC5); representative-per-group fixture selection from the manifest.
2. Run `go test -run '^TestRealWorld' -v ./pkg/nfe/xml/`; fix issues found by the tests (they may reveal parser weaknesses — fix the parser only if it is genuinely wrong, otherwise document).
3. Identity-leak grep for AC3.

## Phase 4 — Wrap-up

1. Fill SDD Verification Results with the exact commands run and PASS/FAIL.
2. Run `make test-go`; run `gofmt -l pkg/nfe/xml/`.
3. Update USE_CASES.md gap table to mirror test allowlists (AC6).
4. Set SDD Status: Done when ACs met.

## Out of scope / follow-up SDDs (user prioritizes)

- ICMS variants beyond ICMS00 (CST 51 deferral = 93% of real traffic — top priority).
- `finNFe=4` devolução support.
- PIS/COFINS non-taxed variants (PISNT/COFINSNT, Outr).
- `dhSaiEnt` emission, `tPag` 90 emission path, CFOP 5118 flow.
- Optional human-run SEFAZ homolog smoke (explicit `NFE_SEFAZ_GUARD=off` required).

## Risks

- Free-text infCpl synthesis: replacement must not mangle business-relevant markers (FETHAB/SENAR lines); keep structure (`;`/`|` segments) intact.
- Access-key rebuild must keep cNF (8 digits) and recompute only cDV; CPF emit documents are zero-padded to 14 inside the key.
- Manifest must stay in sync with files (generator writes both).
