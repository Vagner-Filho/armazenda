# plan:nfe_detached_transport_fields

Implements `docs/sdd/nfe_detached_transport_fields/sdd.md`.

Architecture follows the project's Layered pattern: Entity → Model → Service →
Router. **No XML builder changes are required.**

## Test strategy

- **Go unit, service (offline, pure functions)** — `service/nfe_service/`
  (`detached_transport_test.go`): transport input validation (CPF/CNPJ check
  digits, placa formats, IE⇒UF, lengths, interstate guard), transport→
  `entity.TransportData` mapping with per-field volume fallback, and xEnder
  concatenation. Pure functions over `DetachedTransportInput`, mirroring the
  existing pattern in `detached_preview_test.go` (`buildDetachedDANFEData` is
  tested without DB). Commands: `go test ./service/nfe_service -run
  TestDetachedTransport`.
- **Go unit, model** — `model/nfe_model/detached_transport_model_test.go`:
  JSON marshal/unmarshal round-trip of `TransportRecord` (DB-free, runs in
  every `make test-go`) plus `TestDetachedInvoiceTransport_DBRoundTrip`,
  double-gated on the full `DB_*` set **and** `RUN_DB_TESTS=1` so an
  inherited dev-shell `DB_*` environment never causes an implicit write.
  Command: `go test ./model/nfe_model -run TestDetachedInvoiceTransport`.
- **Go unit, XML builder** — extend the existing transport case in
  `pkg/nfe/xml/builder_test.go` (line ~206) to assert `RNTC`, volume
  values and the CPF branch. Command: `go test ./pkg/nfe/xml`.
- **Go unit, DANFE** — renderer assertions for the new placa/UF/RNTRC cells in
  `pkg/nfe/service/danfe*` and parser assertions in
  `pkg/nfe/xml/danfe_parser_test.go`. Commands: `go test ./pkg/nfe/service`,
  `go test ./pkg/nfe/xml`.
- **Go unit, router** — extend `router/nfe_router/detached_router_test.go`(s):
  parse transport form fields and assert `buildDetachedPreviewViewData` emits
  the 15 hidden fields. Command: `go test ./router/nfe_router`.
- **E2E (Playwright, SEFAZ-free paths only)** — form rendering, sub-group
  toggle, volume prefill/dirty behavior, preview hidden fields; the e2e app
  runs with `NFE_SEFAZ_GUARD=block` (global-setup), so emission endpoints are
  exercised only via the preview route which never touches SEFAZ. Command:
  `make test-e2e` before completion.
- **Baseline**: before coding, run `go test ./pkg/nfe/xml
  ./pkg/nfe/service ./service/nfe_service ./router/nfe_router
  ./model/nfe_model` and record pre-existing failures in the SDD.

## Phase 1 — Migration + model

1. Create `model/armazenda_database/migrations/000026_detached_nfe_transport.sql`
   per SDD F5: `CREATE TABLE IF NOT EXISTS detached_invoice_transportation` —
   identity PK, `invoice_id INTEGER NOT NULL UNIQUE REFERENCES
   detached_nfe_invoice(id)`, payload column `transport JSONB NOT NULL`,
   `created_at TIMESTAMPTZ NOT NULL DEFAULT now()`. Field detail lives in the
   JSON document (transportadora / veiculo / vol keys), not in columns.
2. Extend `model/nfe_model/detached_model.go` (or new
   `detached_transport_model.go` in the same package, mirroring
   `detached_profile_model.go`), reusing the `items_json` pattern
   (`json.Marshal` → named-args INSERT, scan → `json.Unmarshal`):
   - `CreateDetachedInvoiceTransport(invoiceID int, t TransportRecord) error`
     — named args (`pgx.NamedArgs`) only; delete+reinsert not needed (single
     create per invoice, UNIQUE constraint guards).
   - `GetDetachedInvoiceTransport(invoiceID int) (*TransportRecord, error)`
     — `nil` when no row.
   - `TransportRecord` struct with the three nested groups of F5
     (exported pointer-optional style consistent with `DetachedInvoice`);
     add `json:"..."` tags for stable JSON keys.
3. Tests: `detached_transport_model_test.go` per test strategy (AC1, AC2).

**Deliverable check**: `go build ./... && go test ./model/nfe_model`.

## Phase 2 — Service: input struct, validation, mapping

Files: `service/nfe_service/detached_service.go` (+ new
`detached_transport.go` in the same package for cohesion).

1. Add `DetachedTransportInput` (`*string` / `*decimal.Decimal` / `*int`
   optionals, one field per form input; provider-style comments like
   `DetachedRecipient`).
2. Extend `DetachedInvoiceInput` with `Transport DetachedTransportInput`
   (`VehicleID *uint16` placeholder stays untouched).
3. `validateDetachedTransport(in, destUF, emitterUF string) *toast` — pure
   function implementing F3 + F4 (pt-br messages). Returns warning toast on
   first violation. Reuse/extend existing digit-validation helpers (check how
   `validateRecipient` validates `recipientDocument`).
4. `mapDetachedTransport(in, itemsForDB) entity.TransportData` — pure:
   - transportadora present (name/document/…) → `TransportadoraData{Type:
     1\|2 by digit count, CNPJ/CPF, XNome, IE, Endereco: "logradouro, número"
     (trim; omit suffix when número blank), Municipio, UF}`.
   - placa informed → `VeiculoData{Placa: uppercased, UF, RNTC}`; placa blank
     → nil (current behavior).
   - volume resolution per F2: any provided value switches to user block with
     blank-field fallback; all blank → exactly today's auto-derived block.
5. Wire both into `prepareDetachedBuildData`: call validation after the
   recipient is resolved (destination UF needs the person record for the
   existing-recipient path), map into `invoiceInput.Transport`. Emission and
   preview both flow through it — F4 and F3 hold on both endpoints.
6. Persistence hooks: after `CreateDetachedInvoice` in `BuildDetachedInvoice`
   and after the second create in `attemptDetachedSVCContingency`, call
   `CreateDetachedInvoiceTransport` (log+continue on error, matching the
   `UpdateDetachedInvoiceSignedXML` error-tolerance pattern — a transport
   insert failure must not fail an authorized emission).
7. Tests per test strategy (AC3, AC4, AC5).

**Deliverable check**: `go test ./service/nfe_service`.

## Phase 3 — Router: parse + preview round-trip

Files: `router/nfe_router/detached_router.go`,
`templates/nfe/nfe-detached-preview.html`.

1. `detachedFormData` gains `Transport nfe_service.DetachedTransportInput`;
   `parseDetachedTransport(c)` reads the 15 fields (strings trimmed, blanks →
   nil). Widen the existing numeric parse helpers for `volQtd`.
2. `buildDetachedPreviewViewData` + `detachedPreviewView` carry the 15 fields;
   `nfe-detached-preview.html` gains the hidden inputs next to `modFrete`
   (line ~55). Keep the display string format consistent with other decimal
   hidden fields (raw `decimal` string, see `items[].quantity`).
3. Tests: parse + `buildDetachedPreviewViewData` round-trip assertions (AC7).

**Deliverable check**: `go test ./router/nfe_router`.

## Phase 4 — Templates: nfe-emit.html transport UI

Files: `templates/pages/nfe-emit.html`.

1. Extend the "Transporte" fieldset: three sub-groups (Transportadora, Veículo,
   Volumes) using the existing `flex flex-col-reverse` + `peer` label pattern
   and the grid classes already used in Destinatário/Itens.
2. Inline script (same `<script nonce>` block):
   - Vehicle enablement: placa empty → transportPlacaUF/transportRNTC get
     `disabled` + cleared (visual only; service rule is authoritative); placa
     filled → enabled. Prefill `transportPlacaUF` from `{{ .EmitterUF }}`-style
     farm default (add the view field if missing — verify `DetachedEmitView`).
   - Volume prefill with dirty tracking: recompute `volQtd/volPesoL/volPesoB`
     from the live `.item-row` values and set `volEsp` default "Granel" only
     while untouched (per-field `data-dirty` set on user input; the existing
     `reindexItems()`/item-change listeners are the hook points).
3. Labels exactly per SDD F1 (pt-br): "Nome / Razão Social", "CPF / CNPJ",
   "Inscrição Estadual", "Logradouro", "Número", "Município", "UF", "Placa do
   Veículo", "UF da Placa", "ANTT (RNTRC)", "Quantidade", "Espécie", "Marca",
   "Peso Líquido (kg)", "Peso Bruto (kg)".
4. Manual Playwright check per AC9.

**Deliverable check**: `make test-js` (regression) + focused e2e spec.

## Phase 5 — DANFE: placa/UF/RNTRC rendering

Files: `pkg/nfe/entity/danfe.go`, `pkg/nfe/service/danfe.go`,
`pkg/nfe/xml/danfe_parser.go`,
`service/nfe_service/detached_service.go` (`buildDetachedDANFEData`).

1. `DANFEData`: add `VeicRNTC string`.
2. `danfe_parser.go`: parse `veicTransp/RNTC` next to the existing placa/UF
   parse (line ~248).
3. `danfe.go drawTransport`: add row 4 with cells PLACA / UF / RNTRC
   (`data.VeicPlate`, `data.VeicUF`, `data.VeicRNTC`); increase the block
   height `h` (72.0) accordingly. Render only the standard NF-e layout — the
   release-path decision is renderer-internal and covered by unit tests.
4. `buildDetachedDANFEData`: map `data.VeicRNTC` (and keep existing
   placa/UF/volume mappings).
5. Tests per test strategy (AC8).

**Deliverable check**: `go test ./pkg/nfe/service ./pkg/nfe/xml`.

## Phase 6 — Verification & SDD close-out

1. Run `make test-go` (all Go packages) — record PASS/FAIL + exact commands in
   the SDD *Verification Results*.
2. Run `make test-js` and `make test-e2e` — record results; e2e asserts the
   form/preview behaviors of AC7/AC9/AC10.
3. Confirm no SEFAZ contact occurred (guard stayed active throughout; no
   `nfe_farm_config` seeded in test DBs — second layer per AGENTS.md).
4. Update SDD `Status: Done` only if acceptance criteria are met and checks
   pass; otherwise document gaps and keep WIP.
5. `gofmt`/`goimports` over touched files.

## Risks & notes

- **Destination UF resolution for existing persons**: person row's UF must be
  available in the service (already loaded via `GetFullPersonById` during
  recipient resolution — reuse it). Inline path uses `recipientState`.
- **xEnder concatenation ambiguity**: "Logradouro, Número" uses a comma
  separator (DANFE displays the combined string); net weight semantics stay
  byte-identical to today when the volume group is blank (no behavior change
  for existing flows).
- **The `Vol` note**: NF-e `vol` group allows 0–5000 occurrences; we persist
  and emit exactly one block (1:1 table), consistent with the single volume
  group the form offers.
- **Persistence failure tolerance**: transport INSERT failure logs and
  continues (matches `UpdateDetachedInvoiceSignedXML` handling) — the
  authorized NF-e and its XML are the source of truth; the table serves audit
  and the future detached SVC rebuild (worker.go:495 TODO).
- **Out of scope (explicit)**: reboque (X22), `retTransp` (ICMS retenção,
  X11–X17), lacres (X33), `nVol`, rascunho/profile support, departure-flow
  (romaneio) transport inputs, `modFrete` semantics changes, DB read at list
  pages.
