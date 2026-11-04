# plan:transport_person_selector

Implementation plan for the Transporte source selector and the two Destinatário fixes. Spec: [sdd.md](./sdd.md).

## Patterns to mirror

- Source switch: the `recipientType` radios and their handler in `templates/pages/nfe-emit.html`; `parseDetachedRecipient` (`router/nfe_router/detached_router.go`) and `resolveOrCreateRecipient` (`service/nfe_service/detached_service.go`).
- Disabled means not submitted: `applyTaxRateState` in `templates/pages/nfe-emit.html`.
- Farm-scoped reference check: `ValidateDetachedProfileReferences` (`model/nfe_model/detached_profile_model.go`). Mirror the behavior only, and use named arguments.
- Pure-first transport logic: `validateDetachedTransport` and `mapDetachedTransport` in `service/nfe_service/detached_transport.go`, with their tests in `detached_transport_test.go`.
- DB-gated tests: `TestDetachedInvoiceTransport_DBRoundTrip` in `model/nfe_model/detached_transport_model_test.go`.

## Layers touched

| Layer | File | Change |
|---|---|---|
| Entity | `entity/public/person.go` | none; reuse `FullPerson` |
| Model | `model/person_model/model.go` | `PersonBelongsToFarm` |
| Service | `service/nfe_service/detached_transport.go` | `DetachedTransportInput.PersonID`, `transportInputFromPerson` |
| Service | `service/nfe_service/detached_service.go` | `resolveTransportContact`; recipient farm check |
| Router | `router/nfe_router/detached_router.go` | parsing, page view, preview view, summary label |
| Template | `templates/pages/nfe-emit.html` | source switch, disabled state, recipient fix |
| Template | `templates/nfe/nfe-detached-preview.html` | `transportPersonId` hidden input, summary line |
| JS | `assets/js/detachedNFeModes.js` (new) | pure `modeState`, DOM `applySourceMode` |
| CSS | `assets/static/css/output.css` | regenerate only if new utilities are used |

Not touched: migrations; `assets/js/detachedNFeTabs.js` and `assets/js/detachedNFeValidation.js` (they already skip disabled controls, because `stepControls` filters on `willValidate`); the romaneio modal; rascunho code.

## Test strategy

| Behavior | Layer | Location |
|---|---|---|
| Source switch, default state, labels | Go template test | `router/nfe_router/nfe_templates_test.go` |
| Mode helper | Bun unit | `test/unit/detachedNFeModes.test.js` |
| Contact mapping, precedence, resolved validation | Go unit, offline | `service/nfe_service/detached_transport_test.go` |
| Parsing `transportPersonId`, round-trip, preview summary | Go unit, offline | `router/nfe_router/detached_router_test.go`, `nfe_templates_test.go` |
| Farm-scoped lookup | Go DB-gated | `model/person_model/person_farm_test.go` |
| Check-digit document helpers | Bun unit | `test/unit/document-helpers.test.js` |
| Switching, stale selection, farm scope, no contact created | Playwright E2E | `test/e2e/tests/detached_nfe_transport.spec.js`, `detached_nfe_validation.spec.js` |

Commands:

```bash
go test ./router/nfe_router -run TestNFeEmitPage
go test ./service/nfe_service -run 'TestTransportInputFromPerson|TestResolvedTransportValidation'
go test ./router/nfe_router -run 'TestParseDetachedTransport_PersonID|TestDetachedTransportView_PersonIDRoundTrip|TestDetachedPreviewTransportPerson'
cd test && bun test unit/detachedNFeModes.test.js unit/document-helpers.test.js
```

The DB-gated test is skipped by default under `make test-go`, and the test asserts the skip, as the previous transport change did. It runs only with `RUN_DB_TESTS=1` and `DB_*` pointing at the e2e test database (`localhost:5433`, database `armazenda_test`). It must never point at the development database. An earlier change did, and its self-cleanup removed rows.

## Phase 0 — Baseline (no code)

1. Set `Status: WIP` in `sdd.md` when implementation starts.
2. Run these and record the results in `sdd.md` Verification Results:
   - `go test ./service/nfe_service ./router/nfe_router ./model/person_model ./model/nfe_model ./pkg/nfe/xml`
   - `cd test && bun test unit/`
   - `cd test/e2e && bunx playwright test tests/detached_nfe_transport.spec.js tests/detached_nfe_validation.spec.js --project=chromium --workers=1`
3. Run the E2E serially. The previous transport change found parallel runs stalled at login on this machine.

## Phase 1 — Recipient farm check (security fix, first)

Covers AC11.

1. **Regression test first.** In `detached_nfe_validation.spec.js`, add "a recipient from another farm is rejected at preview":
   - create a second farm with `INSERT INTO farm (inscricao_estadual, uf) VALUES (…, 'MT') ON CONFLICT (inscricao_estadual) DO NOTHING`, as `detached_nfe_rascunho.spec.js` does, and a natural person in it;
   - add an option with that person's id to `#recipientPersonId`, select it, and preview;
   - expect HTTP 400, no `#nfe-detached-preview-panel`, and no preview.
   This must fail before the fix.
2. **Model.** Add `func (pm *PersonModel) PersonBelongsToFarm(personID, farmID uint32) (bool, *model_error.ModelError)` in `model/person_model/model.go`. The SQL is `SELECT EXISTS (SELECT 1 FROM person WHERE id = @id AND farm = @farm)`, run with `pgx.NamedArgs`.
3. **Service.** In `resolveOrCreateRecipient` (`detached_service.go`), in the `input.PersonID != nil` branch and before `GetFullPersonById`:
   - not in the farm: `entity_public.GetWarningToast("Destinatário inválido", "O destinatário selecionado não pertence à sua fazenda")`;
   - lookup error: log through `model_error.GetLoggerModel()`, then `entity_public.GetErrorToast("Falha ao buscar destinatário", "")`.
   Preview and emission both reach this through `prepareDetachedBuildData`, so one check covers both.
4. **DB-gated test.** `model/person_model/person_farm_test.go`, `TestPersonBelongsToFarm`: true for a contact of the farm, false for a contact of another farm, and false for a missing id. Follow the package and gating of `model/nfe_model/detached_transport_model_test.go`.
5. Run the focused E2E specs and `go test ./service/nfe_service`. The existing detached suites must pass unchanged.

## Phase 2 — Service: contact resolution (pure first)

Covers AC6, AC7, AC8 and AC10.

1. `service/nfe_service/detached_transport.go`:
   - Add `PersonID *uint32` to `DetachedTransportInput`, with a comment: the contact that supplies the Transportadora group, resolved before validation.
   - Add `transportInputFromPerson(in DetachedTransportInput, person entity_public.FullPerson) DetachedTransportInput`. It returns `in` with the Transportadora fields (`Nome`, `Document`, `IE`, `Logradouro`, `Numero`, `Municipio`, `UF`) replaced from the contact, and copies the Veículo and Volumes fields unchanged. The mapping is the F3 table. `Numero` comes from `FullPerson.Number` through `strconv.FormatUint`, and is nil when the number is nil or 0.
2. `service/nfe_service/detached_service.go`:
   - Add `resolveTransportContact(input DetachedInvoiceInput) (DetachedTransportInput, entity_public.Toast)`. It is read-only. With `PersonID == nil` it returns `input.Transport` unchanged. Otherwise it runs `PersonBelongsToFarm`, which gives the warning "Transportadora inválida" / "A transportadora selecionada não pertence à sua fazenda" when the check fails. It then runs `GetFullPersonById`, which gives the error toast "Falha ao buscar transportadora" on failure, and finally `transportInputFromPerson`. Mirror the person lookup style of `resolveOrCreateRecipient`.
   - Call it in `prepareDetachedBuildData` right before the recipient resolution (`resolveOrCreateRecipient`, around line 214). A rejected contact is then refused before any write. Assign the result to the local `input.Transport`.
   - Leave `validateDetachedTransport` and `mapDetachedTransport` where they are, around lines 283–290. They validate the resolved data without change.
3. Tests, offline, in `service/nfe_service/detached_transport_test.go`:
   - `TestTransportInputFromPerson_NaturalPersonMapsCPF`: CPF, name, street, number as text, city, state.
   - `TestTransportInputFromPerson_LegalPersonBlankFieldsStayAbsent`: CNPJ; nil IE, street, number, city and state stay nil.
   - `TestTransportInputFromPerson_ReplacesOnlyTransportadora`: inline values are replaced; vehicle and volume fields are preserved.
   - `TestResolvedTransportValidation`: a mapped contact with a wrong CNPJ check digit, an IE without a UF, or a name over 60 characters returns the expected messages from `validateDetachedTransport`.
   Build `entity_public.FullPerson` literals in the tests. No database.
4. Run `go test ./service/nfe_service`.

## Phase 3 — Router: parsing, view models, preview round-trip

Covers AC5 (view side), AC9, AC13 and AC14.

1. `parseDetachedTransport` (`router/nfe_router/detached_router.go`) reads `transportPersonId`:
   - blank: nil;
   - otherwise `strconv.ParseUint(value, 10, 32)` must succeed and be greater than zero, or the result is `entity_public.GetWarningToast("Transportadora inválida", "Selecione uma transportadora da lista")`.
   `parseDetachedForm` already returns on transport toasts.
2. `detachedPageView` gets `TransportType string` and `TransportPeople []personOptionDisplay`. `buildDetachedPageView` sets `TransportType: "existing"` next to `RecipientType`, and `TransportPeople: personOptionDisplays(people, nil)`. It reuses the `people` slice the page already loads, so there is no new query.
3. `detachedTransportFields` gets `PersonID string`. `detachedTransportView` maps it with `uint32PtrString(transport.PersonID)`.
4. `detachedPreviewView` gets `TransportLabel string`. `buildDetachedPreviewViewData` sets it to the contact's name when `PersonID` is set. That lookup runs after the service has accepted the contact, the same way `RecipientLabel` works. For inline data it is the typed `Nome`, and empty otherwise.
5. `templates/nfe/nfe-detached-preview.html`:
   - summary: `Transportadora: {{ .TransportLabel }}`, only when set (F7);
   - with `.Transport.PersonID`: a hidden `transportPersonId`, and no `transportNome` to `transportUF` hidden inputs;
   - otherwise: the inline Transportadora hidden inputs, as today;
   - vehicle and volume hidden inputs unchanged.
6. Tests:
   - `detached_router_test.go`, `TestParseDetachedTransport_PersonID`: blank gives nil; "42" gives 42; "abc", "0" and "-1" give the warning toast and no transport.
   - `detached_router_test.go`, `TestDetachedTransportView_PersonIDRoundTrip`: PersonID becomes "42", Nome is empty, and vehicle and volume fields are unchanged.
   - `nfe_templates_test.go`, `TestDetachedPreviewTransportPerson`: a contact renders the hidden id and the summary label; no contact renders the inline hidden inputs. Follow `TestNFeDetachedPreviewRendersInfAdProd`.
7. Run `go test ./router/nfe_router`.

## Phase 4 — Template and mode control

Covers AC1, AC2, AC3, AC4, AC19 and AC20.

1. **Regression and feature tests first.** These run before any template change, and the regression must fail on the current template:
   - `detached_nfe_validation.spec.js`, "switching from a picked contact submits the typed recipient" (AC4). Pick a contact, switch to "Novo destinatário", type a different name and document, and preview. The summary must show the typed name.
   - `detached_nfe_transport.spec.js`, "source defaults to the contact select", "switching the source changes only the Transportadora block" and "an inactive contact selection is not submitted" (AC1 to AC3).
2. **Helper** `assets/js/detachedNFeModes.js` (new):
   - `modeState(mode)` returns `{ existing: boolean, new: boolean }`. It accepts `'existing'` and `'new'`. Any other value throws, so a typo cannot disable both sources.
   - `applySourceMode(mode, sources)` takes `sources = { existing: { section, controls }, new: { section, controls } }`. For each source it sets `section.classList.toggle('hidden', !active)` and `control.disabled = !active` for every control.
   - Tests in `test/unit/detachedNFeModes.test.js`. The Bun unit tests have no DOM library, so the tests use plain objects for the elements. The module must not touch the DOM when it is imported.
3. **Template** `templates/pages/nfe-emit.html`:
   - Transportadora block: radios `transportType` (`existing` and `new`), checked from `.TransportType`; `#existing-transport-section` holds the `transportPersonId` select with the empty option "Nenhuma transportadora" and options from `.TransportPeople`; `#new-transport-section` wraps the existing seven-field grid unchanged. Initial `hidden` and `disabled` attributes are rendered from `.TransportType`. Labels: "Selecionar existente", "Novo transportador", "Transportadora".
   - Recipient block: the existing radio handler calls `applySourceMode` for `#existing-recipient-section` (the select) and `#new-recipient-section` (the eleven inline controls), and keeps the `required` toggling. The initial state comes from `.RecipientType` and is re-applied on load, as the recipient's `dispatchEvent` does today.
   - Transport handler: the same, for `transportType`, re-applied on load.
4. **Template tests** in `router/nfe_router/nfe_templates_test.go`:
   - `TestNFeEmitPageTransportSource`: with the default `existing`, the select is enabled and the inline inputs are disabled and hidden. The empty option comes first, contacts without an id are skipped, and every control has a `for` label. With `TransportType: "new"` it is the inverse.
   - `TestNFeEmitPageRecipientSource`: the states for `RecipientType` `existing` and `new`, matching F8.
5. **Existing transport tests** in `detached_nfe_transport.spec.js`: "renders the three sub-groups…" and "preview round-trips the transport values…" click "Novo transportador" before touching the inline fields (AC20). Make this change in the same commit as the default change, so the suite stays green.
6. **CSS:** reuse existing utility classes where possible. If new ones are needed, regenerate `assets/static/css/output.css` with the Tailwind command from AGENTS.md.

## Phase 5 — End-to-end cases

Covers AC5, AC10, AC11, AC12, AC13 and AC16.

**Fixtures** (helpers in the spec files, or in `test/e2e/utils/`):

- `createContact({ farmId, kind, name, document, ie, address })` inserts `person` (with `farm`), then `legal_person` or `natural_person`, then `address`, and returns the person id. `legal_person` needs `companyName`, `cnpj` and `personId`. `natural_person` needs `name`, `cpf` and `personId`. `address` needs `street`, `cep` (8 digits), `neighborhood`, `city`, `state` (2 letters) and `person_id`; `number` is an optional integer.
- `deleteContact(id)` deletes in dependency order: `address`, then `legal_person` or `natural_person`, then `person`.
- **Documents** must pass check-digit validation, because the transport validation runs on resolved data. Faker's `faker.string.numeric` produces random digits that fail. Add check-digit generators in `test/e2e/utils/document-helpers.js` (new), with Bun unit tests in `test/unit/document-helpers.test.js`.
- **IE:** leave it blank for transport contacts. `unique_person_ie_in_farm` (migration 000024) applies only to non-empty IEs.
- **Foreign farm:** as in Phase 1.
- Do not insert `nfe_farm_config` rows. Use the farm the existing detached specs already use (farm 1).

**Cases**

`detached_nfe_transport.spec.js`:

1. "a picked transportadora is resolved at preview" (AC5): pick a contact that has an address; preview; the summary shows its name (F7); the hidden `transportPersonId` equals its id; no hidden `transportNome` is rendered.
2. "no transportadora selected previews without carrier data": leave "Nenhuma transportadora"; the preview succeeds; no `transportPersonId` is submitted.
3. "a transportadora from another farm is rejected at preview" (AC10): inject a foreign contact's id into the select; preview returns 400 with the warning; no preview panel; the test farm's person count is unchanged.
4. "an inline transportadora creates no contact" (AC12): the person count for farm 1 is unchanged after a preview with inline data.

`detached_nfe_validation.spec.js`:

5. "a recipient from another farm is rejected at preview" (AC11, written in Phase 1).
6. "switching from a picked contact submits the typed recipient" (AC4, written in Phase 4).

**Commands**

```bash
cd test/e2e && bunx playwright test tests/detached_nfe_transport.spec.js tests/detached_nfe_validation.spec.js --project=chromium --workers=1
```

Then repeat for firefox and webkit, as the previous transport change did.

**SEFAZ:** no case clicks `confirm-detached-nfe`. Every case stops at the preview.

## Phase 6 — Verification and close-out

1. `gofmt -l` on the changed Go files, then `go build ./...`.
2. `make test`, since this change crosses layers (AGENTS.md test selection). If the full run is impractical, run `make test-go` and `make test-js` plus the focused E2E commands above, and say so in Verification Results.
3. `make test-e2e` serially (`--workers=1`) if the parallel run stalls at login again. Record any stall as an environment limitation.
4. Regenerate `output.css` if new utilities were used.
5. `git status --short model/armazenda_database/migrations` shows no addition (AC15).
6. Grep the new E2E cases for `confirm-detached-nfe` and `/nfe/emitir/build` (AC16).
7. Review checks for AC17 (pt-br strings) and AC18 (no extra query on render).
8. Record every AC in `sdd.md` Verification Results. Set `Status: Done` only when all ACs pass. If an AC cannot be verified, document why and keep the spec WIP (AGENTS.md).

## Risks and pre-existing issues

1. **Orphan contacts when emission fails (pre-existing).** On emission, `resolveOrCreateRecipient` creates the inline recipient before the emitter, recipient and transport validations run (`prepareDetachedBuildData`, around lines 214–290). A later validation failure leaves a new contact behind, and a retry can create a duplicate. This change keeps the transport contact check before the write, but the structural checks on inline transport data still run after it. Follow-up: move the structural checks ahead of the write, and check the interstate rule separately, because it needs the recipient UF. Not fixed here unless the review asks for it.
2. **Positional SQL nearby.** `ValidateDetachedProfileReferences` and `CreatePersonForDetachedNFe` use `$1`/`$2`. They are not changed here. The new code uses `pgx.NamedArgs`.
3. **No carrier flag.** Every contact appears in the transportadora select. A carrier flag would be a schema change, so it is out of scope.
4. **Legacy contacts can fail transport validation.** Contacts are saved without document check-digit validation. A contact with a bad CNPJ, a name over 60 characters, or an IE without a state fails on preview with the generic transport warning, which names the block rather than the contact. Open for review: include the contact's name in that toast.
5. **Preview and confirm each read the contact.** Editing the contact in between changes the emitted data. The recipient behaves the same way. Accepted.
6. **Romaneio modal.** `templates/nfe/nfe-emit-modal.html` keeps inline-only transport. Out of scope.
7. **`modFrete = 9` with a transportadora.** No coupling, unchanged from `nfe_detached_transport_fields`.
8. **Browser coverage.** Chromium is the gate. Firefox and webkit are recommended for the transport and validation specs, as in the previous change.

## Out of scope

- The romaneio modal.
- Storing transport in rascunhos.
- A carrier flag on persons.
- Saving inline transportadoras as contacts.
- The ordering fix in risk 1, unless the review includes it.
- Positional SQL cleanup (risk 2).

## Open points for review

1. **F7, the preview summary line "Transportadora: …".** It also gives the E2E cases a visible check of the resolved name (AC5). If it is dropped, AC5 keeps only the hidden-input check.
2. **Risk 4:** whether the transport warning should name the contact.
3. **Risk 1:** whether the ordering fix belongs in this change or in a follow-up.
