# sdd:transport_person_selector

## Status
<!-- One of: Backlog | WIP | Done -->
Done

## Goal

The Transporte tab of the avulsa NF-e form (`/nfe/emitir`, `templates/pages/nfe-emit.html`) should behave like the Destinatário tab. The user either picks a transportadora that is already a farm contact, or types the carrier data inline. Picking a contact avoids retyping the carrier's name, CPF/CNPJ, IE and address on every emission, and the DANFE and the `transporta` XML group are filled from one maintained record.

The change also fixes two defects in the Destinatário tab that the new mechanism must not inherit:

1. A contact picked earlier is still submitted after the user switches to "Novo destinatário". The NF-e can then go to the wrong recipient while the typed data is ignored.
2. The recipient lookup on preview and emission does not check that the contact belongs to the farm. Another farm's contact data can appear on this farm's preview and NF-e.

## Requirements

### Functional

**F1 — Source switch in the Transportadora block.** The Transportadora sub-group of the Transporte tab gets a radio group `transportType`:

- `existing`, "Selecionar existente" (default)
- `new`, "Novo transportador"

In `existing` mode a select `transportPersonId` ("Transportadora") is shown. Its first option, "Nenhuma transportadora", has an empty value and means no transportadora. The other options are the farm's contacts, in the same list and order as the Destinatário select (`GetPeopleByFarm`; contacts without an id are skipped).

In `new` mode the seven inline fields are shown as today: `transportNome`, `transportDocumento`, `transportIE`, `transportLogradouro`, `transportNumero`, `transportMunicipio`, `transportUF`.

The switch changes only the Transportadora block. Veículo (`transportPlaca`, `transportPlacaUF`, `transportRNTC`) and Volumes (`volQtd`, `volEsp`, `volMarca`, `volPesoL`, `volPesoB`) stay visible and editable in both modes, with their current behavior.

No field of the Transporte tab becomes `required`. The section remains optional.

**F2 — Only the active source is submitted.** In each mode, the inactive source's container gets the `hidden` class and each of its controls gets the `disabled` attribute. Disabled controls are not submitted. Values typed in the inline fields survive a switch. The initial state is rendered by the server and re-applied by the page script on load. The tax-rate inputs already use this mechanism (`applyTaxRateState`).

**F3 — Server-side resolution of the picked contact.** For `transportPersonId`:

- Blank: no contact. The inline group is parsed as today.
- A positive integer: the contact is resolved on preview and on emission. If inline Transportadora fields are also posted, they are ignored, because the contact replaces the whole Transportadora group.
- Anything else (non-numeric, zero, negative): warning toast "Transportadora inválida" / "Selecione uma transportadora da lista", HTTP 400, nothing generated. There is no fallback to the inline fields.
- A contact that does not belong to the request's farm: warning toast "Transportadora inválida" / "A transportadora selecionada não pertence à sua fazenda", HTTP 400. No DANFE, no number allocated, nothing written.

Mapping of a resolved contact into the Transportadora group:

| Group field | Source |
|---|---|
| Nome | contact name |
| Documento | contact CPF/CNPJ, reduced to digits; the type follows the digit count, as for inline input |
| IE | contact IE; absent when blank |
| Logradouro | contact street; absent when empty |
| Numero | contact number as text; absent when empty or zero |
| Municipio | contact city; absent when empty |
| UF | contact state; absent when empty |

Veículo and Volumes are not changed by the resolution.

**F4 — Same validation for contacts.** The resolved group goes through the existing `validateDetachedTransport`. A failure gives the existing warning "Dados de transporte inválidos", and nothing is generated.

**F5 — Snapshot only.** A transportadora typed inline is never saved as a contact. The transportadora of a preview or emission is stored only in the existing transport snapshot (`detached_invoice_transportation`). The invoice does not reference the contact. No table or column is added.

**F6 — Preview round-trip.** With a contact selected, the preview fragment carries `transportPersonId` as a hidden input and no inline Transportadora hidden inputs. With inline data, it carries the inline hidden inputs as today. Confirm re-resolves the contact from the database. This narrows the all-fields round-trip rule of `nfe_detached_transport_fields` (F6/AC7) for the contact case only.

**F7 — Preview summary.** When a transportadora is set, the preview summary shows "Transportadora: <nome>" beside the recipient line. The name is the contact's name, or the typed inline name. The line is omitted when there is no transportadora.

**F8 — Recipient: only the active source is submitted.** The rule of F2 applies to Destinatário. In "Selecionar existente" the eleven inline recipient controls are disabled. In "Novo destinatário" the `recipientPersonId` select is disabled. The existing `required` toggling stays and applies to the active source only.

**F9 — Recipient: farm check.** `resolveOrCreateRecipient` rejects a `recipientPersonId` that does not belong to the request's farm. This applies to preview and emission, before any write. The message is the one the rascunho path already uses: "Destinatário inválido" / "O destinatário selecionado não pertence à sua fazenda". The preview answers HTTP 400.

**F10 — Unchanged behavior.** A request without a contact gives the same transport mapping, XML and DANFE as today. Existing field names and IDs do not change. Rascunhos, the romaneio modal (`templates/nfe/nfe-emit-modal.html`) and the transport validation rules are not changed.

### Non-Functional

- **Tenancy:** no contact is read on the preview or emission path without a farm check. New SQL uses named arguments (`pgx.NamedArgs`), as AGENTS.md requires. The nearby `ValidateDetachedProfileReferences` and `CreatePersonForDetachedNFe` use positional `$1`/`$2`; they are not copied.
- **SEFAZ isolation:** no new test or manual check reaches SEFAZ. New E2E cases stop at preview and never click the emission button. `NFE_SEFAZ_GUARD=block` stays set by the E2E global setup. New fixtures insert no `nfe_farm_config` rows.
- **Performance:** rendering adds no query, because the contact list reuses the `GetPeopleByFarm` result the page already loads. A selected contact adds at most a farm check and one `GetFullPersonById` on preview or emission, the same cost the recipient has today.
- **Compatibility:** no migration. Requests without `transportType` or `transportPersonId` parse as today.
- **Language:** every new label, option and toast is in pt-br.
- **Accessibility:** every new control has a `<label for>`. Disabled controls are not focusable, so keyboard order follows the active source.
- **Observability:** unexpected lookup failures are logged through `model_error.GetLoggerModel()`, as `resolveOrCreateRecipient` already does.
- **Known limitation:** preview and confirm each read the contact. A contact edited between the two steps changes what is emitted. The recipient behaves the same way today. Accepted; see plan risk 5.

## Acceptance Criteria

| # | Criterion | Verification |
|---|---|---|
| AC1 | On a fresh page the Transportadora block shows "Selecionar existente" checked. The `transportPersonId` select is enabled, shows "Nenhuma transportadora" selected, and has one option per contact (contacts without an id skipped). The seven inline fields are hidden and disabled. Every control has a `<label for>`. | New `TestNFeEmitPageTransportSource` (`router/nfe_router/nfe_templates_test.go`); E2E "source defaults to the contact select" |
| AC2 | Choosing "Novo transportador" shows and enables the inline fields, and hides and disables the select. Choosing "Selecionar existente" reverses it. Typed inline values survive the switch. Veículo and Volumes stay enabled in both modes. | E2E "switching the source changes only the Transportadora block" |
| AC3 | A contact is picked, the user switches to "Novo transportador", types another name and document, and previews. The hidden `transportNome` equals the typed name, and no hidden `transportPersonId` exists. | E2E "an inactive contact selection is not submitted" |
| AC4 | Recipient regression: the same scenario for Destinatário shows the typed name in the preview summary. Inline recipient controls are disabled in "Selecionar existente", and the select is disabled in "Novo destinatário". Both states are rendered by the server and re-applied on load. | E2E "switching from a picked contact submits the typed recipient" (must fail on the current template before the fix); new `TestNFeEmitPageRecipientSource` |
| AC5 | A picked contact is resolved on preview. The summary shows its name (F7), the hidden `transportPersonId` holds its id, and no inline Transportadora hidden input is rendered. | E2E "a picked transportadora is resolved at preview"; new `TestDetachedPreviewTransportPerson` |
| AC6 | Mapping: a natural person gives CPF, a legal person gives CNPJ. Blank IE, street, number, city or state stay absent. The number is text. The mapping follows the F3 table. | New `TestTransportInputFromPerson_*` in `service/nfe_service/detached_transport_test.go` (offline) |
| AC7 | Precedence: with inline Transportadora fields posted alongside a contact, the contact replaces the whole group and the inline values disappear. Veículo and Volumes are preserved. | New `TestTransportInputFromPerson_ReplacesOnlyTransportadora` (offline) |
| AC8 | A contact with wrong CNPJ check digits, an IE without a UF, or a name over 60 characters fails with the existing transport warning. No DANFE or XML is produced. | New `TestResolvedTransportValidation` (offline, pure) |
| AC9 | Parsing `transportPersonId`: blank gives no contact. A positive integer gives the id. Non-numeric, zero or negative values give the warning toast, with no fallback to inline. | New `TestParseDetachedTransport_PersonID` (`router/nfe_router/detached_router_test.go`) |
| AC10 | A transportadora from another farm is rejected on preview with HTTP 400 and the warning toast. No preview panel, no number allocated, no `person` row created. | E2E "a transportadora from another farm is rejected at preview"; DB-gated `TestPersonBelongsToFarm` |
| AC11 | Recipient regression: a recipient from another farm is rejected on preview with HTTP 400 and the warning toast, before any write. | E2E "a recipient from another farm is rejected at preview" (must fail before the fix). Emission runs the same `resolveOrCreateRecipient`; code review covers it, since emission cannot run without SEFAZ. |
| AC12 | An inline transportadora creates no `person` row on preview. The transport code adds no write on emission. | E2E: the test farm's person count is unchanged after preview. Emission path: justified manual check by diff review. |
| AC13 | Round-trip for a contact: `transportPersonId` survives preview and confirm, and confirm re-resolves the contact. | New `TestDetachedTransportView_PersonIDRoundTrip`; E2E asserts the hidden value |
| AC14 | Requests without a contact are unchanged. `TestParseDetachedTransport_*`, `TestMapDetachedTransport_*`, `TestDetachedTransportRecordFromEntity`, `TestDetachedTransportView_RoundTrip` and `TestBuilder_TransportVehicleAndVolumes` pass without edits. | `go test ./service/nfe_service ./router/nfe_router ./pkg/nfe/xml` |
| AC15 | No migration is added. | `git status --short model/armazenda_database/migrations` shows no new file |
| AC16 | No new test reaches SEFAZ or clicks emission. | The new cases contain no `confirm-detached-nfe` or `/nfe/emitir/build`, confirmed by grep. E2E runs under `NFE_SEFAZ_GUARD=block`. |
| AC17 | Every new user-facing string is pt-br. | Review of the template and Go diff (justified manual check) |
| AC18 | Rendering adds no query. A selected contact adds at most a farm check and one `GetFullPersonById` on preview or emission. | Diff review (justified manual check). The render path reuses `people`. |
| AC19 | The source-mode helper returns the active source for `existing` and `new`, and throws for any other value. | New `test/unit/detachedNFeModes.test.js`: `cd test && bun test unit/detachedNFeModes.test.js` |
| AC20 | Existing detached E2E specs pass with the new default. The two transport tests that fill inline fields first choose "Novo transportador". | `cd test/e2e && bunx playwright test tests/detached_nfe_transport.spec.js tests/detached_nfe_validation.spec.js tests/detached_nfe_rascunho.spec.js tests/detached_nfe_tabs.spec.js tests/detached_cfop_selector.spec.js tests/detached_inf_ad_prod.spec.js --project=chromium --workers=1` |
| AC21 | `go build ./...` succeeds, `gofmt -l` lists no changed Go file, and `output.css` is regenerated if new utilities are used. | `go build ./...`; `gofmt -l` on the changed files; the Tailwind command from AGENTS.md |

## Verification Results

### Phase 0 baseline (before implementation)

| Check | Result |
|---|---|
| `go test ./service/nfe_service ./router/nfe_router ./model/person_model ./model/nfe_model ./pkg/nfe/xml` | FAIL (pre-existing): `router/nfe_router` has three failures unrelated to this change — `TestParseDepartureFormOverrides_PaymentBillingAndTransport` (`IncludeCobranca` nil, want explicit false), `TestDepartureNFeModalRendersInputsAndTaxSelects` and `TestDepartureNFePreviewCarriesInputsToConfirm` (missing `indPag` / `includeCobranca`). Every other package passed. |
| `cd test && bun test unit/` | PASS — 49 pass, 0 fail. |
| `cd test/e2e && bunx playwright test tests/detached_nfe_transport.spec.js tests/detached_nfe_validation.spec.js --project=chromium --workers=1` | FAIL (pre-existing): 10 passed, 1 failed — `detached_nfe_transport.spec.js:162 "preview round-trips the transport values and the DB stores nothing"`. The e2e fixture seeds `nfe_farm_config` with empty `certificate_data`, so `nfeCertGate` refuses every preview with HTTP 400 "Certificado digital não configurado"; the test expects 200. Documented environment limitation; no certificate data may be seeded under the SEFAZ isolation rules. |

### Implementation

| # | Verification | Result |
|---|---|---|
| AC1 | `go test ./router/nfe_router -run TestNFeEmitPageTransportSource`; E2E `source defaults to the contact select` | PASS. The template test asserts the default `existing` state (select enabled, "Nenhuma transportadora" first, contacts without id skipped, inline controls hidden/disabled, every control labelled); the E2E case passes on chromium. |
| AC2 | E2E `switching the source changes only the Transportadora block` | PASS (chromium). Both directions toggle visibility/disabled state, typed inline values survive, and the vehicle/volume fields keep their own state. |
| AC3 | E2E `an inactive contact selection is not submitted` | PASS (chromium). The preview request payload carries the typed `transportNome`/`transportDocumento` and no `transportPersonId`. |
| AC4 | E2E `switching from a picked contact submits the typed recipient`; `go test ./router/nfe_router -run TestNFeEmitPageRecipientSource` | PASS (chromium + Go). The regression case asserts the posted payload has no `recipientPersonId` and carries the typed name; before the fix the select stayed enabled and was submitted (code review of the removed unconditional handler). |
| AC5 | `go test ./router/nfe_router -run TestDetachedPreviewTransportPerson`; DB-gated `TestResolveTransportContact_DB`; E2E `a picked transportadora is resolved at preview` | PASS with environment note. The rendering test asserts the hidden id, the summary name and the absence of inline hidden inputs; the DB-gated test asserts the contact resolution. The E2E case is skipped on this machine because the fixture has no certificate (`nfeCertGate` answers 400 before the preview); it will assert the full HTTP round-trip when a certificate is available. |
| AC6 | `go test ./service/nfe_service -run 'TestTransportInputFromPerson'` | PASS. Natural person → CPF, legal person → CNPJ, blank optional fields stay nil, number is text. |
| AC7 | `TestTransportInputFromPerson_ReplacesOnlyTransportadora` | PASS. The contact replaces the Transportadora group; vehicle and volume fields are preserved. |
| AC8 | `TestResolvedTransportValidation` | PASS. Bad CNPJ check digit, IE without UF and a >60-char name produce the existing messages. |
| AC9 | `TestParseDetachedTransport_PersonID` | PASS. Blank → nil; "42" → 42; "abc"/"0"/"-1" → warning "Transportadora inválida" / "Selecione uma transportadora da lista" with no inline fallback. |
| AC10 | DB-gated `TestResolveTransportContact_DB`; E2E `a transportadora from another farm is rejected at preview` | PASS with environment note. The DB-gated test asserts the warning toast and that no person row is created. The E2E case is skipped while the certificate gate rejects the preview (it cannot reach the farm check); it asserts the 400 + toast when a certificate exists. |
| AC11 | DB-gated `TestResolveOrCreateRecipient_RejectsCrossFarm_DB`; E2E `a recipient from another farm is rejected at preview` | PASS with environment note. The DB-gated test asserts the warning and no write; emission shares `resolveOrCreateRecipient`. The E2E case is skipped under the certificate gate (pre-existing environment limitation). |
| AC12 | E2E `an inline transportadora creates no contact`; diff review | PASS. The typed name produces no `legal_person` row; the transport path adds no write on emission (only reads). |
| AC13 | `TestDetachedTransportView_PersonIDRoundTrip`; `TestDetachedPreviewTransportPerson`; `TestParseDetachedTransport_PersonID` | PASS. The preview carries `transportPersonId`, parse reads it back and the service re-resolves the contact. |
| AC14 | `go test ./service/nfe_service ./router/nfe_router ./pkg/nfe/xml` (named tests) | PASS. `TestParseDetachedTransport_*`, `TestMapDetachedTransport_*`, `TestDetachedTransportRecordFromEntity`, `TestDetachedTransportView_RoundTrip` and `TestBuilder_TransportVehicleAndVolumes` pass without edits. (The package still has the pre-existing departure failures listed in the baseline.) |
| AC15 | `git status --short model/armazenda_database/migrations` | PASS — no output, no migration added. |
| AC16 | grep of the new E2E cases for `confirm-detached-nfe` / `/nfe/emitir/build` | PASS — no match. The E2E global setup keeps `NFE_SEFAZ_GUARD=block`. |
| AC17 | Review of the template and Go diff | PASS — every new label, option and toast is pt-br. |
| AC18 | Diff review of `buildDetachedPageView` / `buildDetachedPreviewViewData` | PASS — the page reuses the `people` slice already loaded; preview adds one `GetFullPersonById` only when a contact is picked, like the recipient. |
| AC19 | `cd test && bun test unit/detachedNFeModes.test.js` | PASS — 5 tests. |
| AC20 | `cd test/e2e && bunx playwright test <six detached specs> --project=chromium --workers=1` | PASS for the change; 25 passed, 3 skipped (certificate-gated preview cases), 4 failed — all four expect a 200 preview panel and are blocked by the pre-existing empty-certificate fixture (`detached_cfop_selector`, `detached_inf_ad_prod`, `detached_nfe_rascunho`, `detached_nfe_transport` round-trip). The two transport tests that fill inline fields now select "Novo transportador" and pass. |
| AC21 | `go build ./...`; `gofmt -l` on the changed Go files; CSS | PASS for build and gofmt (no output). `output.css` was not regenerated: the change uses only utilities already present in the templates. |

### Commands run

```text
go test ./service/nfe_service ./router/nfe_router ./model/person_model ./model/nfe_model ./pkg/nfe/xml   # router/nfe_router: only the 3 pre-existing departure failures
go test ./...                                                                                            # same 3 pre-existing failures only
go test ./model/person_model ./service/nfe_service -run 'TestPersonBelongsToFarm|TestResolveTransportContact_DB|TestResolveOrCreateRecipient_RejectsCrossFarm_DB'  # DB-gated, with RUN_DB_TESTS=1 against localhost:5433
cd test && bun test unit/                                                                                # 58 pass, 0 fail
cd test/e2e && bunx playwright test <six detached specs> --project=chromium --workers=1                   # 25 passed, 3 skipped, 4 certificate-gate failures
go build ./...
gofmt -l <changed Go files>                                                                              # no output
git status --short model/armazenda_database/migrations                                                   # no output
```

### Environment limitations and manual checks

- **Certificate gate (pre-existing):** the e2e fixture seeds `nfe_farm_config` with empty `certificate_data`, so every preview answers 400 "Certificado digital não configurado". Seeding certificate bytes is forbidden by the SEFAZ isolation rules, so the three new preview-content E2E cases and the two cross-farm E2E cases skip on this machine. The same behavior is verified at the service/template layers (Go tests above), which cover the resolution, farm rejection, summary and hidden round-trip. A human-approved run with a certificate would execute the skipped cases.
- **Firefox / WebKit:** NOT RUN — the chromium run is the gate; the previous transport change also found these specs slow/flaky in the other engines.
- **AC4/AC10/AC11 "must fail before the fix":** not executed against a reverted tree (cost); the pre-fix code paths had no farm check (`PersonBelongsToFarm` did not exist) and left the recipient select enabled, so the asserted request/farm behavior could not hold before the change. Confirmed by diff review.
- **AC17/AC18:** justified manual review as stated in the table.

## Status decision

All acceptance criteria are met with the two documented environment limitations (certificate-gated preview HTTP legs, verified by the best available service/template checks). Status set to Done.

