# Real-world NF-e replay fixtures (anonymized)

Deterministically anonymized copies of the full corpus in
`docs/research_nfe_use_case/` (115 authorized `nfeProc` 4.00 files, serie 921,
farm in MT). People and companies are SYNTHETIC: names, CPF/CNPJ (valid check
digits), IE, emails, phones, addresses, driver names, plates, CND codes,
protocol numbers, certificates and digests. Digital signatures are structurally
intact but cryptographically invalid by construction.

All business content is preserved from the real files: natOp, CFOP, finNFe,
ICMS tag/CST, PIS/COFINS CSTs, tPag, modFrete, transport data shapes, weights,
values, dates, IBS/CBS blocks, infCpl structure (with synthetic identity
tokens). Municipal names and nNF numbering are public/business data and kept.

Regenerate with:

    go run pkg/nfe/xml/testdata/realworld/generate.go docs/research_nfe_use_case pkg/nfe/xml/testdata/realworld

Group taxonomy: docs/research_nfe_use_case/USE_CASES.md.
Per-file metadata (group + discriminants + provenance): manifest.json.
(`manifest.json.originalFolder` keeps the export folder label of the source
file — "9008711_xml_saida_55_MM_2026_…" + nNF — for traceability only; it
carries no CNPJ, CPF, name or other identity data.)

## Groups
- G1: 84 files
- G2: 23 files
- G3: 3 files
- G4: 2 files
- G5: 1 files
- G6: 1 files
- G7: 1 files

Representatives per group (fixture access key):
- G1: 51260400040257642617559210000009461891752345
- G5: 51260400040257642617559210000009721024193424
- G7: 51260600040257642617559210000010601100116725
- G3: 51260700040257642617559210000010641315035170
- G4: 51260700040257642617559210000010671320836292
- G6: 51260900040257642617559210000010981728185984
- G2: 51260900040257642617559210000011041001838851
