# sdd:nfe_detached_transport_fields

## Status
Done

## Goal

Add the full NF-e transport block (MOC 7.0, Grupo X) to the detached NF-e
emission form (`templates/pages/nfe-emit.html`). Today the form only offers
"Modalidade do Frete"; the transporter identity, vehicle and volume inputs are
missing, volumes are hard-coded (`qVol=1`, `esp="Granel"`) and the net weight
is silently derived as the sum of item quantities (wrong for non-KG units such
as sacas). The feature adds validated inputs for transporter, vehicle (placa,
UF, ANTT/RNTRC) and volumes (quantidade, espécie, marca, peso líquido, peso
bruto), persists them in a normalized table, round-trips them through the
preview, and renders them on the DANFE.

The XML builder (`pkg/nfe/xml/builder.go`, `buildTransp`) already supports the
whole Grupo X — this feature wires form → service → persistence → XML/DANFE.

## Requirements

### Functional

**F1 — Form fields (Transporte fieldset, three sub-groups):**

| Sub-group | Form field | NF-e tag (MOC id) | Type |
|---|---|---|---|
| Transportadora | `transportNome` | `transporta/xNome` (X06) | text, ≤60 |
| Transportadora | `transportDocumento` | `transporta/CPF` (X05) or `CNPJ` (X04) | text; type auto-detected by digit count (11 → CPF, 14 → CNPJ) |
| Transportadora | `transportIE` | `transporta/IE` (X07) | text, ≤14 |
| Transportadora | `transportLogradouro` | `transporta/xEnder` (X08) | text |
| Transportadora | `transportNumero` | — (folded into `xEnder`) | text |
| Transportadora | `transportMunicipio` | `transporta/xMun` (X09) | text, ≤60 |
| Transportadora | `transportUF` | `transporta/UF` (X10) | text, 2 |
| Veículo | `transportPlaca` | `veicTransp/placa` (X19) | text, ≤7 |
| Veículo | `transportPlacaUF` | `veicTransp/UF` (X20) | text, 2 |
| Veículo | `transportRNTC` | `veicTransp/RNTC` (X21) | text, ≤20 |
| Volumes | `volQtd` | `vol/qVol` (X27) | integer |
| Volumes | `volEsp` | `vol/esp` (X28) | text, ≤60 |
| Volumes | `volMarca` | `vol/marca` (X29) | text, ≤60 |
| Volumes | `volPesoL` | `vol/pesoL` (X31) | decimal (kg) |
| Volumes | `volPesoB` | `vol/pesoB` (X32) | decimal (kg) |

- "Número" has no dedicated schema tag: the service concatenates
  `Logradouro, Número` into `xEnder` (schema limit 1–60 chars).
- RNTC (label "ANTT (RNTRC)") belongs to the vehicle group; the UI keeps it
  beside placa and the vehicle sub-group is visually disabled until placa is
  filled. Disabled-state clearing is visual only — the service ignores vehicle
  data whenever placa is blank.
- All fields are optional; `modFrete` keeps its current select and semantics
  (including `9 - Sem frete`; no extra coupling rule between modFrete=9 and
  transporta — the MOC imposes none for the NF-e model).

**F2 — Volume prefill (editable):** the volume inputs are prefilled from the
item rows (current auto-derivation): `volQtd = 1`, `volEsp = "Granel"`,
`volPesoL = Σ item quantities`, `volPesoB = Σ item gross weights`. Prefill is
recomputed client-side when items change, but a volume field the user has
edited ("dirty") is never overwritten. At submit, blank volume fields fall
back per-field to the auto-derived value (qVol→1, esp→"Granel", pesoL→Σ
quantities, pesoB→Σ gross weights, marca→omitted). If the user clears the
whole group, the current auto block is produced; if the user provides any
value, the block is built from user values with per-field fallback as above.

**F3 — Validation (service layer, warning toasts in pt-br, no XML built on failure):**
- CPF/CNPJ: digits-only after stripping punctuation, 11 or 14 digits, valid
  check digit (MOC Rej. 542/543). Type mismatch (digits not 11 or 14) rejected.
- Placa: `XXX9999`, `XXX999`, `XX9999`, `XXXX999` (MOC) or MERCOSUL `AAA9A99`,
  case-insensitive, uppercased before persistence/XML.
- IE informed ⇒ transporter UF must be informed (Rej. 559); IE "ISENTO"
  allowed and exempt from UF pairing? No — Rej. 559 applies to any informed
  IE; keep the pairing requirement for "ISENTO" too (defensive; SEFAZ
  validates per UF and would reject).
- RNTC ≤ 20 chars; xNome/xMun/esp/marca ≤ 60; `xEnder` concatenation ≤ 60.
- `volQtd`, if informed, must be a non-negative integer; `volPesoL`/`volPesoB`
  valid decimals ≥ 0.

**F4 — Interstate guard (Rej. 868):** if `transportPlaca` is informed and the
recipient UF differs from the emitter UF, the request is rejected with a
warning toast ("Dados do veículo não podem ser informados em operação
interestadual …") and no XML is built, signed or sent. Destination UF: from
the existing person record (existing-recipient path) or the inline
`recipientState` field (new-recipient path).

**F5 — Persistence (normalized, JSONB payload):** new table
`detached_invoice_transportation` (migration `000026_detached_nfe_transport.sql`),
1:1 with `detached_nfe_invoice` (`UNIQUE(invoice_id)`). Instead of one column
per field, the transport document lives in a single JSONB column (same
serialize/deserialize pattern as `items_json` on `detached_nfe_invoice`):

- `id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY`
- `invoice_id INTEGER NOT NULL UNIQUE REFERENCES detached_nfe_invoice(id)`
- `transport JSONB NOT NULL` — one JSON document mirroring the Grupo X
  sub-groups:
  ```json
  {
    "transportadora": { "tpDoc": 1|2, "docNumber": "", "xNome": "",
                        "ie": "", "xEnder": "", "xMun": "", "uf": "" },
    "veiculo":        { "placa": "", "uf": "", "rntc": "" },
    "vol":            { "qtd": 1, "esp": "Granel", "marca": "",
                        "pesoL": 0.000, "pesoB": 0.000 }
  }
  ```
- A row is created on emission only when at least one transport field is
  informed; `mod_frete` stays on `detached_nfe_invoice` (unchanged).
- Persisted on both emission paths: normal (`BuildDetachedInvoice` after
  `CreateDetachedInvoice`) and SVC contingency supersede
  (`attemptDetachedSVCContingency` after the second `CreateDetachedInvoice`),
  so the contingency invoice carries the same transport data.
- Read access: `GetDetachedInvoiceTransport(invoiceID)` model function.

**F6 — Preview round-trip:** all 15 fields survive preview → confirm via
hidden inputs in `templates/nfe/nfe-detached-preview.html`, fed by
`detachedPreviewView`; confirm emits with exactly the previewed transport
data.

**F7 — DANFE:** the transport block renders placa, UF do veículo and
RNTRC/ANTT (per `pkg/nfe/resources/DANFE_RULES.md` §4.10) in addition to the
existing rows. `entity.DANFEData` gains `VeicRNTC`;
`danfe_parser.go` parses `veicTransp/RNTC` (the signed-XML download path is
already correct once parsed).

### Non-Functional

- **SEFAZ isolation**: no new SEFAZ calls anywhere; all unit tests offline
  (guard rules of AGENTS.md respected).
- **Performance**: transport persistence is a single extra INSERT inside the
  existing emission flow; no N+1 on list pages (transport is not joined into
  the invoice list query).
- **pt-br**: all validation messages and UI labels in Brazilian Portuguese.
- **Compatibility**: pre-existing behavior for forms without transport fields
  must be unchanged (the new block is purely additive; empty fields produce
  today's XML).
- **Single source**: volume weights stay derived in one place (JS prefill
  mirrors the service fallback; no third computation in a template).

## Acceptance Criteria

| # | Criterion | Verification |
|---|---|---|
| AC1 | Migration `000026` creates `detached_invoice_transportation` idempotently with the UNIQUE FK and the `transport JSONB` column of F5; runner applies it on startup | `go build ./...` + e2e global setup (Docker DB) applies migrations without error; column list asserted in `model/nfe_model/detached_transport_model_test.go` (DB-backed) |
| AC2 | Transport round-trip: `CreateDetachedInvoiceTransport` + `GetDetachedInvoiceTransport` persist and return every field through the JSONB document (transportadora CNPJ and CPF branches, vehicle, volumes) | `RUN_DB_TESTS=1 DB_*… go test ./model/nfe_model -run TestDetachedInvoiceTransport -v` (explicit opt-in against the e2e test DB; recorded in Verification Results — PASS) |
| AC3 | Validation: invalid CPF/CNPJ check digit, malformed placa, IE without UF, RNTC > 20, `xEnder` > 60 each produce a pt-br warning toast and no XML | `go test ./service/nfe_service -run TestDetachedTransportValidation` (offline, pure-function tests) |
| AC4 | Transport data maps into `entity.TransportData`: transportadora (Type 1/2), vehicle (placa uppercased, UF, RNTC) and the volume block with per-field fallback; blanks keep today's auto-derivation | `go test ./service/nfe_service -run TestDetachedTransportMapping` (offline) |
| AC5 | Interstate guard: placa informed + destination UF ≠ emitter UF → warning toast, no emission; destination UF resolved for both existing-person and inline recipient paths | `go test ./service/nfe_service -run TestDetachedTransportInterstateGuard` (offline, pure-function) |
| AC6 | XML output contains `transp` with `modFrete`, `transporta`, `veicTransp` (incl. `RNTC`) and `vol` exactly as mapped; existing XML without transport data unchanged | `go test ./pkg/nfe/xml -run TestBuilder` (extends existing transport test) |
| AC7 | Preview round-trip: all 15 fields re-appear as hidden inputs and confirm emits identical transport data | `go test ./router/nfe_router -run TestDetachedPreviewTransportRoundTrip` (pure view-builder test) + focused Playwright check of the rendered preview (manual/e2e, SEFAZ-free preview path) |
| AC8 | DANFE renders PLACA / UF / RNTRC row in the transport block and `VeicRNTC` flows through both the detached mapping and `danfe_parser.go` | `go test ./pkg/nfe/service -run TestDANFE` + `go test ./pkg/nfe/xml -run TestDanfeParser` |
| AC9 | Form: 15 inputs grouped in 3 sub-groups; vehicle sub-group visually disabled without placa; volume prefill mirrors item sums and never overwrites user-edited fields | Focused Playwright run (`make test-e2e` full suite before completion) |
| AC10 | Rascunho unaffected: save/load profile ignores transport fields; profile pages unchanged | `make test-js` + existing e2e rascunho specs stay green |

## Verification Results

- **Baseline (pre-implementation)**: `go test ./pkg/nfe/xml ./pkg/nfe/service
  ./service/nfe_service ./router/nfe_router ./model/nfe_model` — PASS (all
  ok, no pre-existing failures).
- **`make test-go` (`go test ./...`)** — PASS (14 packages ok, 0 FAIL;
  final run after all fixes). The DB-backed model test **skips** by default
  (see gating below).
- **`go test -count=1 ./service/nfe_service ./pkg/nfe/service ./pkg/nfe/xml
  ./router/nfe_router ./model/nfe_model`** — PASS (all new/changed packages).
- **`make test-js` (`bun test unit/`)** — PASS (33 pass / 0 fail).
- **Focused e2e (Playwright), `--project=chromium --workers=1`**:
  - `tests/detached_nfe_transport.spec.js` — PASS 4/4. Re-run per browser:
    PASS 4/4 on firefox and PASS 4/4 on webkit as well.
  - `tests/login.spec.js` (isolated) — PASS 6/6.
- **Full serial e2e, `--project=chromium --workers=1 --reporter=line` (all
  spec files)** — PASS **41/41** (departure, entry, entry_discount, login,
  detached_cfop_selector, detached_nfe_rascunho, detached_nfe_transport);
  2.6m, exercising the pre-existing detached specs against the modified form
  and preview templates.
- **Baseline note (environment, not this change)**: the e2e suite with the
  default parallel workers repeatedly stalls at login on this machine (page
  snapshots show tests stuck on the login page across all browsers, before
  any detached-NF-e page is reached); the log also shows a pre-existing
  duplicate-key retry inside `detached_nfe_rascunho` (uniqueSuffix re-insert).
  Serial (`--workers=1`) execution passes everything; record as a known
  environment limitation.
- **DB-level round trip** — env-gated
  `TestDetachedInvoiceTransport_DBRoundTrip` is **double-gated**: the full
  `DB_*` set plus an explicit `RUN_DB_TESTS=1` opt-in (an inherited
  dev-shell `DB_*` environment alone must never cause an implicit DB write).
  With the gate: PASS against the e2e Docker DB (JSON groups, FK,
  absent-row nil). Default `make test-go` runs: SKIP (asserted).
- **Migration application** — `RunMigrations()` applied `000026` on every
  e2e boot; the DB round-trip test read/wrote the created table.
- **SEFAZ isolation** — no SEFAZ contact: e2e runs with
  `NFE_SEFAZ_GUARD=block` (global-setup), all unit tests offline, and the
  one-time app spawn used for bootstrapping the schema ran with
  `NFE_SEFAZ_GUARD=block` and was stopped before the DB round-trip test.
- **Incident note (transparent)**: one early run of the DB round-trip test,
  before the double gate existed, connected to the *development* database
  (inherited `DB_*` env pointing at localhost:5432). The test's self-cleanup
  removed all created rows; residue was re-verified afterwards as zero
  (transport rows 0, `dbroundtrip%` invoices 0, IE-less persons 0). The gate
  now prevents this; no SEFAZ endpoints were involved.
- **`gofmt`** — clean on all changed files.
- **CSS build** — `tailwind -i assets/static/css/input.css -o
  assets/static/css/output.css --minify` regenerated; new utilities present
  (`md:grid-cols-5`, `disabled:opacity-40`).
