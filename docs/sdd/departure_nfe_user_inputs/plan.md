# plan — departure_nfe_user_inputs

## Scope and current architecture

This SDD is for NF-e emitted from a departure/romaneio. The departure UI is
`templates/nfe/nfe-emit-modal.html`, distinct from the Detached NF-e page
`templates/pages/nfe-emit.html`. The latter is the design reference for payment,
transport, and CST selectors and is not changed by this plan.

Current flow:

1. `router/nfe_router/router.go:getNFeEmitModal` renders the modal;
   `previewNFe` and `buildNFe` parse the same form values.
2. `templates/nfe/nfe-preview.html` carries values to final confirmation as
   hidden inputs.
3. `service/nfe_service/service.go:prepareInvoiceBuildData` builds a single
   `entity.InvoiceInput` item, hardcodes `tPag=90`, derives the vehicle and
   volume data, and accepts tax-code overrides.
4. `pkg/nfe/xml/builder.go` already supports payment, `cobr/fat`, and the full
   `TransportData` XML groups. Its `buildPag` currently checks the
   container-level `PaymentData.IndPag`; implementation must use each
   `PagamentoDetalhe.IndPag` for the `detPag/indPag` element.
5. `model/nfe_model/model.go` persists scalar overrides and tax-rate snapshots.
   `service/nfe_service/worker.go` reconstructs overrides for SVC rebuilds;
   payment and detailed transport choices are not currently part of that
   snapshot.

The current migration sequence includes an in-flight
`000029_detached_nfe_wiring.sql`; confirm the next available version when the
implementation starts. The departure feature should not edit that file.

## Decisions/proposed defaults to review before WIP

- `tPag` options initially mirror the existing detached UI (`90`, `01`, `99`);
  keep one payment detail, derive `vPag`, and do not implement split payments.
- Add a `cobr/fat` include toggle, default on. Keep `nFat` and amounts derived;
  manual fatura amounts, discounts, and duplicatas are not in this slice.
- Show departure-derived vehicle/volume defaults, but provide explicit controls
  to omit them; an empty field must not override an explicit omit choice.
- Use the detached page's finite normal-regime CST option lists and add the
  builder-supported CSOSN choices for Simples Nacional. Keep IBS/CBS CST and
  cClassTrib text inputs as they are today.

## Test strategy

| Behavior | Layer/location | Test data / environment |
|---|---|---|
| `detPag` defaults, tPag cases, `indPag` omission/nonzero handling, optional `cobr` | `pkg/nfe/xml/builder_test.go` | `minimalInvoiceInput`; offline |
| Transport validation and mapping, defaults vs explicit omit, regime-specific tax codes | `service/nfe_service/*_test.go` | Pure inputs and mockable validation helpers; offline |
| Form parsing rejects invalid enum/code/transport values; modal renders selected farm defaults | `router/nfe_router/*_test.go`, `nfe_templates_test.go` | Gin form contexts and isolated template rendering |
| Preview-to-confirm hidden-field round-trip | `router/nfe_router/*_test.go` and `templates/nfe/nfe-preview.html` render test | No DB/SEFAZ |
| Snapshot model JSONB create/read and legacy empty snapshot | `model/nfe_model/*_test.go` | E2E PostgreSQL only; full `DB_*` plus explicit `RUN_DB_TESTS=1` |
| Worker rebuild uses original snapshot and does not read changed departure vehicle/config | `service/nfe_service/worker_test.go` or pure snapshot reconstruction test | Fake stored invoice; no DB/SEFAZ |
| User flow through modal and preview only | `test/e2e/tests/departure_nfe_inputs.spec.js` | Playwright app; global setup keeps `NFE_SEFAZ_GUARD=block`; do not submit final emission |

Baseline before implementation: run `go test ./pkg/nfe/xml ./service/nfe_service
./router/nfe_router ./model/nfe_model` and `make test-go` when practical; record
any pre-existing failures, then change Status to WIP.

DB-backed test setup is opt-in only: start the dedicated e2e DB with
`cd test/e2e && bun run db:start`, apply migrations through the test setup, and
set the full test `DB_HOST/DB_PORT/DB_USER/DB_PASS/DB_NAME` plus
`RUN_DB_TESTS=1`. Never point this test at the local development database.

## Phases

### Phase 1 — Entity and XML builder behavior

1. `pkg/nfe/entity/invoice.go`: add a typed billing option to `InvoiceInput`;
   reuse the existing `PaymentData`, `PagamentoDetalhe`, and `TransportData`.
   Add typed input/options fields for payment and explicit suppression of
   departure-derived vehicle/volume data without changing Detached NF-e
   defaults.
2. `pkg/nfe/xml/builder.go`: make `buildPag` read per-detail `IndPag`; make
   `buildCobr` respect the per-emission include option while keeping legacy
   nil/default behavior; keep the existing element order and value formatting.
3. Add builder tests in `pkg/nfe/xml/builder_test.go` for nonzero detail
   `indPag`, default omission, tPag 90 zero `vPag`, tPag 01/99 value handling,
   and `cobr` included/omitted. Run:
   `go test ./pkg/nfe/xml -run 'TestBuilder_(Payment|Cobranca)'`.

### Phase 2 — Persistence and retry snapshot

1. Add an idempotent migration after the current latest migration (expected
   `model/armazenda_database/migrations/000030_departure_nfe_inputs.sql`): add
   a JSONB per-invoice snapshot for payment, billing inclusion, and transport
   details with an empty-object default for legacy rows.
2. `model/nfe_model/model.go`: extend `Invoice`, create/scan paths, and the
   `CreateInvoiceWithEmission` call to persist/load the snapshot. Use named SQL
   placeholders and `pgx.NamedArgs` in any modified SQL.
3. `service/nfe_service/worker.go`: rebuild the typed input options from the
   snapshot; missing snapshot fields on old rows must reproduce the existing
   hardcoded/derived defaults.
4. Add a DB-gated JSONB round-trip test (including legacy empty JSON) and a pure
   worker reconstruction test. Run focused model tests against the dedicated
   e2e DB with the explicit opt-in described above; no app/SEFAZ emission.

### Phase 3 — Service mapping and validation

1. `service/nfe_service/service.go`: resolve form choices into `InvoiceInput`;
   explicit values take precedence, while absent legacy values preserve the
   current `tPag=90`, enabled fatura, departure vehicle, and auto-volume
   behavior. Keep the same resolved values for preview, initial emission,
   synchronous SVC contingency, and worker retry.
2. Map transportadora, vehicle, and volume fields into existing entity types.
   Reuse/centralize the validation rules from
   `nfe_detached_transport_fields`; keep all validation messages pt-br.
3. Ensure preview `DANFEData` receives transport fields already supported by
   `pkg/nfe/entity/danfe.go` and `GeneratePreviewDANFE`; add any missing mapping
   tests. Do not add SEFAZ calls.
4. Add service tests for default compatibility, explicit payment/billing
   choices, transport override/omit semantics, interstate plate rejection,
   and invalid tax values. Run:
   `go test ./service/nfe_service -run 'TestDeparture(Payment|Billing|Transport|TaxCode)'`.

### Phase 4 — Router, templates, and browser flow

1. `router/nfe_router/router.go`: load tax regime/options and transport defaults
   for the modal; parse and validate payment, billing, and transport inputs for
   both preview and build; reject invalid codes even if a client bypasses the
   `<select>`.
2. `templates/nfe/nfe-emit-modal.html`: add payment/billing and transport
   controls using the existing `templates/pages/nfe-emit.html` field styling;
   convert ICMS to a CST or CSOSN selector based on tax regime and PIS/COFINS
   to the finite CST selectors. Leave IBS/CBS CST, cClassTrib, rates, and
   natural-language fields unchanged.
3. `router/nfe_router/router.go` and `templates/nfe/nfe-preview.html`: carry
   every choice in the preview view and hidden confirmation form. Add rendered
   template and pure round-trip tests in `router/nfe_router`.
4. Add `test/e2e/tests/departure_nfe_inputs.spec.js` to exercise modal values
   through preview and assert the returned preview retains them. Stop before
   confirm; the project guard remains on. Run the focused spec, then `make
   test-e2e`.
5. `test/e2e/global-setup.js:waitForTable` must verify the query returned the
   `farm` row, not merely that `psql` exited successfully; the previous helper
   accepted an empty result and seeded before migrations finished. Keep this
   fix limited to test readiness. The existing login E2E expectations must
   also derive their origin from `BASE_URL` so the suite can use an alternate
   port when a developer app is already listening on 8100.

### Phase 5 — Completion verification

Run focused Go tests as listed above, `make test-go`, the model snapshot test
with its explicit test-DB opt-in, and `make test-e2e`. Record exact outcomes in
the SDD; set Status to Done only when all applicable acceptance criteria pass.

## Risks and out of scope

- **Fiscal policy:** selecting a CST does not determine whether its rates or
  legal treatment are correct. The user remains responsible for the tax
  classification; defaults remain farm-configured. This SDD only makes the
  finite code choices safer to enter.
- **Billing semantics:** this increment does not support manually changing
  fatura amounts, discounts, due dates, or `<dup>` installments. If those are
  required, revise the SDD before implementation rather than inferring amounts
  from payment method.
- **Payment shape:** multiple `detPag` entries and partial payments are out of
  scope; current departure flow is one item/one payment detail.
- **Shared UI:** do not broaden implementation to Detached NF-e templates or
  rascunho profiles; their input work is separate. Reuse option definitions
  only if it does not make their behavior drift.
- **Tax-code catalog:** IBS/CBS CST and cClassTrib remain text until an
  authoritative, versioned option catalog is selected.
- **Migration numbering:** confirm the next version after the in-flight
  `000029_detached_nfe_wiring.sql` is integrated.
- **SEFAZ:** all checks stay offline. If a future verification appears to
  require homologação/production network access, stop and ask for explicit
  human approval.
