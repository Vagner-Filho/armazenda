# sdd:departure_nfe_user_inputs

## Status
WIP

## Goal

Let a user review and provide the payment, billing inclusion, and transport
details for a departure-linked NF-e before preview/confirmation, and replace
free-text tax-code fields with regime-appropriate selectors. Keep this work
scoped to the romaneio/departure flow; `templates/pages/nfe-emit.html` is the
UI reference for the separate Detached NF-e flow, not a target for this SDD.

## Requirements

### Functional

#### F1 — Payment and billing (`pag` / `cobr`)

1. The departure modal gains a `tPag` selector consistent with the Detached
   NF-e form: `90` (Sem pagamento, default), `01` (Dinheiro), and `99`
   (Outros). Existing requests with no payment field retain `tPag=90`.
2. The form gains an `indPag` selector for `0` (à vista, default), `1`
   (a prazo), and `2` (outros). The default `0` is omitted from XML. A
   non-zero value is written to the corresponding `<detPag>` only; the XML
   builder must use `PagamentoDetalhe.IndPag`, not the payment-container
   field. Invalid posted values are rejected server-side with a pt-br warning.
3. Departure emissions keep one payment detail in this phase. `vPag` is
   derived as the NF-e total for `tPag` other than `90`, and is `0.00` for
   `tPag=90`. For `tPag=99`, retain the builder's default `xPag=Outros`.
   Partial/multiple payments and installment schedules are out of scope.
4. The modal gains an explicit option to emit `<cobr>/<fat>`, default enabled
   to preserve the current XML. When enabled, `nFat` uses the allocated NF-e
   number, `vOrig` and `vLiq` use the current NF-e total, and `vDesc` remains
   zero. When disabled, omit `<cobr>` without changing `vNF` or `<pag>`.
   Manually entered billing amounts, discount, due dates, and `<dup>` entries
   are out of scope for this increment; the toggle controls inclusion only.

#### F2 — Transport (`transp`)

5. Keep the existing `modFrete` selector and its farm-config default. Add
   optional inputs matching the transport groups already used by the Detached
   NF-e form and `entity.TransportData`:
   - Transportadora: document (CPF/CNPJ), name, IE, address/number, city, UF.
   - Vehicle: plate, plate UF, and ANTT/RNTRC.
   - Volumes: quantity, species, brand, net weight, and gross weight.
6. Defaults preserve current departure behavior: use the departure vehicle
   when available; derive one volume (`qVol=1`, `esp=Granel`) and weights from
   the departure. Show these defaults in the form where available. User edits
   override derived values. Explicit controls must let the user omit the
   auto-derived vehicle or volume group; blank fields alone must not silently
   restore a value the user intentionally disabled.
7. Optional transportadora data maps to `<transporta>`; vehicle values map to
   `<veicTransp>`; volumes map to `<vol>`. Empty optional sub-groups are
   omitted. Existing behavior for legacy submissions with no new transport
   fields remains unchanged.
8. Validate submitted data before XML construction: valid CPF/CNPJ check
   digits; schema limits for names/address/IE/RNTRC; IE requires UF; valid
   plate shape and uppercase normalization; positive integer `qVol` and
   non-negative decimal weights. If a plate is informed in an interstate
   operation, reject with the pt-br warning for the Rejeição 868 condition.
   Follow the rules in `docs/sdd/nfe_detached_transport_fields/sdd.md` rather
   than creating a second, looser validation policy.

#### F3 — Tax-code selectors

9. Replace the free-text `icmsCST`, `pisCST`, and `cofinsCST` controls in the
   departure modal with selects, using the labels and normal-regime choices
   from `templates/pages/nfe-emit.html`:
   - ICMS CST: `00`, `20`, `40`, `41`, `50`, `51`, `90`.
   - PIS and COFINS CST: `01`, `02`, `04`, `05`, `06`, `07`, `08`, `09`.
     CST `03` is excluded because the current builder does not implement its
     correct XML branch.
10. The farm-configured value remains selected by default. Render an
    ICMS-regime-appropriate control: normal-regime farms select CST; Simples
    Nacional farms select CSOSN `{101, 102, 103, 201, 202, 203, 300, 400,
    500, 900}`. Keep the existing `icmsCST` override carrier if useful, but
    validate against the correct list for the farm's tax regime and map the
    selected code to CST or CSOSN accordingly.
11. Enforce the same option sets server-side; a client-submitted value outside
    the applicable list is rejected with a pt-br warning before signing or
    persistence. Keep `ibsCST`, `cbsCST`, `cClassTrib`, tax rates, and free-text
    fields unchanged in this SDD: the reference page also leaves the reform
    codes as text, and their maintained option catalog is a separate concern.

#### F4 — Preview and retry fidelity

12. All payment, billing, and transport choices survive the preview-to-confirm
    round-trip through hidden inputs in `templates/nfe/nfe-preview.html`; the
    final signed XML uses the same choices shown in the preview.
13. Persist a per-invoice JSONB snapshot of the newly user-supplied payment,
    billing-toggle, and transport inputs. Draft retry/SVC reconstruction must
    use this snapshot rather than current farm settings or mutable departure
    vehicle data. Legacy rows without the snapshot retain current defaults.
14. Preview DANFE maps the transport values that it supports (including
    transporter, vehicle, and volume data); the transport shown in preview
    must not contradict the XML input. The builder already emits these entity
    groups; changes are in input wiring, validation, and persistence.

### Non-Functional

- All user-facing labels and validation messages are in Brazilian Portuguese.
- No test or preview path may contact SEFAZ. E2E uses the existing
  `NFE_SEFAZ_GUARD=block` setup; do not confirm/emit in UI tests.
- Defaults preserve the current output when the user accepts the modal without
  changes: `tPag=90`, `indPag` omitted, `<cobr>/<fat>` enabled, existing
  `modFrete`/departure vehicle/volume derivation, and farm-config tax codes.
- Persisted transport documents and payment settings are only exposed through
  the existing farm-scoped invoice access paths; do not add them to public or
  cross-farm queries.
- Database changes are idempotent migrations and SQL parameters use named
  placeholders with `pgx.NamedArgs`.

## Acceptance Criteria

1. **AC1 — Backward-compatible defaults:** no new fields (legacy request) and
   an untouched modal produce the existing departure defaults, including
   `tPag=90`, `vPag=0.00`, omitted `indPag`, and a `<cobr>/<fat>` based on the
   invoice total. Verify with a service mapping test and
   `go test ./pkg/nfe/xml -run 'TestBuilder_(Payment|Cobranca)'`.
2. **AC2 — Payment choice:** valid `tPag`/`indPag` values map to one `<detPag>`;
   `tPag=90` writes zero `vPag`, other supported choices write the NF-e total,
   `indPag=0` is omitted, and non-zero `indPag` is emitted from the detail.
   Invalid values are rejected before build. Verify with router/service tests
   and `go test ./pkg/nfe/xml -run 'TestBuilder_Payment'`.
3. **AC3 — Billing control:** the enabled default emits the existing `<fat>`;
   disabling it omits `<cobr>` while preserving totals and payment. Verify
   with XML-builder tests and a service-level regression test.
4. **AC4 — Transport mapping and validation:** each supplied transporter,
   vehicle, and volume field is mapped to its corresponding XML tag; invalid
    identifiers, plate/UF combinations, lengths, zero quantity, or negative weights produce a
   pt-br warning and no XML. Explicitly disabled derived groups stay omitted.
   Verify with `go test ./service/nfe_service -run 'TestDepartureTransport'`
   and `go test ./pkg/nfe/xml -run 'TestBuilder_Transport'`.
5. **AC5 — Tax selectors:** modal renders regime-appropriate selectors with
   configured defaults selected; values outside the allowed CST/CSOSN lists
   are rejected server-side. Verify with `go test ./router/nfe_router -run
   'TestDepartureTaxCode'` and a rendered-template assertion.
6. **AC6 — Preview round-trip:** payment, billing, transport, and tax values
   submitted in preview are present in the confirm form's hidden inputs and
   are identical when the service builds the final input. Verify with
   `go test ./router/nfe_router -run 'TestDeparturePreviewRoundTrip'` and a
   focused Playwright test that stops at preview.
7. **AC7 — Retry snapshot:** JSONB persistence and reload reproduce the same
   input for a draft/SVC rebuild; an old row with an empty/default snapshot
   remains readable and uses legacy behavior. Verify with an explicitly
   gated DB round-trip test and a pure worker reconstruction test.
8. **AC8 — Full verification:** `make test-go` passes; `make test-e2e` passes
   with the project guard active. No test performs an external SEFAZ request.

## Verification Results

- **Baseline before implementation** —
  `go test ./pkg/nfe/xml ./service/nfe_service ./router/nfe_router ./model/nfe_model`
  — PASS (all four packages; no pre-existing failures observed).
- Feature acceptance checks — NOT RUN yet; implementation in progress.
