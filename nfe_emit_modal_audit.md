# Audit: `templates/nfe/nfe-emit-modal.html` + its routes

Static analysis of the departure-linked NF-e emission modal, the routes that
serve and consume it, and the client-side chain that wires the form to the
server. Audit performed on 2026-11-10 (commit `07d6611`).

**Scope reviewed:**

- Fragment: `templates/nfe/nfe-emit-modal.html`
- Mounting route: `GET /nfe/modal/:departureId` → `getNFeEmitModal` (`router/nfe_router/router.go`)
- Submission routes: `POST /nfe/preview/:departureId` → `previewNFe`, `POST /nfe/build/:departureId` → `buildNFe` (`router/nfe_router/router.go`)
- Form parsers: `router/nfe_router/departure_input_router.go` (`parseDepartureFormOverrides` et al.)
- Swapped fragment: `templates/nfe/nfe-preview.html`
- JS chain: `assets/js/form/nfe-form.js`, `assets/js/decimalMask.js`, `assets/js/utils.js` (close button), `templates/components/close-dialog-btn.html`
- Service gates: `service/nfe_service/service.go` (`prepareInvoiceBuildData`, `BuildInvoiceFromDeparture`, `GeneratePreviewDANFE`), `departure_inputs.go`, `detached_transport.go`, `cert_gate.go`
- Guarding tests: `router/nfe_router/departure_input_router_test.go`, `departure_input_template_test.go`, `nfe_templates_test.go`, e2e `test/e2e/tests/departure_nfe_inputs.spec.js`
- DB constraints: `model/armazenda_database/migrations/000002_add_nfe_tables.sql`

---

## Findings

### Critical

#### F1. IDOR — no farm-ownership check on the departure-linked emission routes

`getNFeEmitModal` (router.go:658), `previewNFe` (router.go:87) and `buildNFe`
(router.go:32) load the departure by raw ID and **never verify
`departure.Farm == farmID`**. Contrast with `downloadNFeXML` (router.go:565),
`downloadNFeDANFE` (router.go:599) and `getNFeCancelModal` (router.go:830),
which all enforce `departure.Farm != farmID → 403`.
`prepareInvoiceBuildData` (service/nfe_service/service.go:47) doesn't check
either — it merges the *caller's* farm config/emitter with the *victim's*
departure data.

Impact for any authenticated user of another farm (the `/nfe` group only
requires an active subscription + `fiscal` tier):

- `GET /nfe/modal/:id` — renders another farm's departure summary (product,
  net weight, vehicle plate, recipient-derived `infCpl` via
  `BuildDefaultInfCpl`).
- `POST /nfe/preview/:id` — leaks the victim departure's recipient
  (name/document/address) and weights into a DANFE PDF.
- `POST /nfe/build/:id` — **emits a real NF-e**: farm B's departure +
  recipient under farm A's certificate/emitter, allocating a number in farm
  A's series (`AllocateNumber(farmID, …)`) and persisting a row attached to
  farm B's departure. Cross-tenant data corruption.

Additionally, the `nfe-existing-modal` branch (router.go:683) serves
`GetInvoiceByDeparture` results unscoped — access key, series/number and
total value leak. (Its XML/DANFE download links *do* enforce the farm check,
so they would 403 — an inconsistency that follows from this same gap.)

#### F2. Payment/billing fields missing from the modal — parser drift (confirmed failing tests)

The service layer fully supports `indPag` and `IncludeCobranca`
(`pkg/nfe/entity/invoice.go:218` `DepartureInvoiceOverrides`; consumed by
`departurePayment`/`departureCobrancaEnabled` in
`service/nfe_service/departure_inputs.go:118–155`), and the tests expect
them — but none of the three layers wire them:

- `nfe-emit-modal.html` has **no** `indPag` select and no `includeCobranca`
  checkbox.
- `parseDepartureInvoiceOverrides`
  (`router/nfe_router/departure_input_router.go:33`) never parses `indPag` /
  `includeCobrancaSpecified` / `includeCobranca`.
- `nfe-preview.html` has no hidden round-trip inputs for them.

Consequence: every departure NF-e emits with `indPag=0` and cobrança always
enabled — users cannot mark "a prazo" or suppress `<cobr>/<fat>`.

This drift is documented as "pre-existing" in several SDDs
(`docs/sdd/detached_nfe_tab_validation/sdd.md`,
`docs/sdd/decimal_comma_mask/sdd.md`,
`docs/sdd/transport_person_selector/sdd.md`) — the departure flow was left
behind when the detached (avulsa) flow gained these fields. See
Verification below for the live failures.

### High

#### F3. `POST /nfe/build/:id` has no double-emission guard

Only the modal GET checks `GetInvoiceByDeparture` (router.go:675); the build
POST never re-checks, and `nfe_invoice.departure_id` has **no unique
constraint** (migration `000002_add_nfe_tables.sql:88` — only `access_key`
is unique). A double-click on "Confirmar e Emitir" (the client-side
`confirmBtn.remove()` in `nfe-form.js:21` only runs after the response), or
a re-submitted stale preview page, produces two SEFAZ numbers and two rows
for the same departure. `GetInvoiceByDeparture` itself returns
`ORDER BY i.id DESC LIMIT 1`, so the older duplicate becomes invisible in
the UI.

#### F4. Modal↔preview certificate gate mismatch (form renders, preview always 400s)

`getNFeEmitModal` gates only on `hasConfig` (a `nfe_farm_config` row
exists), but `prepareInvoiceBuildData` runs `nfeCertGate`
(`service/nfe_service/cert_gate.go:11` — requires
`len(CertificateData) > 0`) for **both** preview and build. A farm with
config but no certificate gets a fully usable form that always fails on
"Gerar Pré-visualização". The e2e fixture (`test/e2e/fixtures/test-user.sql:90`,
empty `certificate_data` for SEFAZ isolation) demonstrates exactly this
trap. The modal could pre-check cert presence and disable/relabel the submit
button.

### Medium

#### F5. `nfe_NaN.xml` download filename — `assets/js/form/nfe-form.js:12`

```js
const departureId = Number(reqPath.slice(reqPath.lastIndexOf('/')));
```

`slice(lastIndexOf('/'))` keeps the leading slash → `Number("/24")` = `NaN`
→ after a successful emission the downloaded file is named `nfe_NaN.xml`
instead of `nfe_24.xml`. Off-by-one: should be
`reqPath.lastIndexOf('/') + 1`.

#### F6. Null-deref before the null check — `nfe-form.js:58–71`

`setupNFeForm()` executes `dialogEl._nfeAbortSignal = signal` (line 62)
**before** `if (dialogEl)` (line 71). If `#nfe-emit-dialog` is not in the
DOM when the module script runs, this throws
`TypeError: Cannot set properties of null`. Latent today (the fragment
always includes the dialog), but the guard is in the wrong place — the
modal's own inline script defensively uses `dialog?.querySelector` while
this function does not.

#### F7. Duplicated, conflicting tax-rate toggle logic

The modal's inline script (`nfe-emit-modal.html:445–455`) toggles
**`disabled`** on the five rate inputs, while `setupNFeForm()` →
`setupTaxRateToggle()` (`nfe-form.js:30–51`) toggles **`readOnly`** on the
same inputs, with a comment describing the opposite behavior ("readonly…
so the user can override"). Both listeners fire on every change. The
effective contract is "disabled ⇒ not submitted ⇒ farm defaults" (the
disabled attribute wins for serialization), making the `readOnly` mechanism
dead weight and its comment wrong. If either half is changed
independently, the submission contract silently breaks.

#### F8. "Voltar" from preview discards all user input

`nfe-preview.html:61` re-`GET`s `/nfe/modal/:id`, which re-renders the
form **from farm-config defaults** (router.go:698+). Everything the user
typed (price, transport, rates, CSTs, notes) is lost when going back to
edit. The preview carries the values as hidden inputs only for the confirm
form, not for back navigation.

### Low / minor

#### F9. `format-weight` not wired

`<span class="format-weight">{{ .NetWeight }}</span>`
(`nfe-emit-modal.html:10`): only `full-report.html` has a handler for the
`.format-weight` class; the modal shows the raw decimal ("25000") while the
preview page formats via `formatWeight` (`nfe-preview.html:94–97`). Cosmetic
inconsistency.

#### F10. `DefaultTPag: "99"` hardcoded (router.go:787)

Three divergent defaults coexist: the route hardcodes "99" (matching the
service default in `departurePayment`), the select lists "90 - Sem pagamento"
first, and `formValueOr(c, "tpag", "90")` (router.go:177) is dead fallback
code. It is also the one default not driven by farm config.

#### F11. English user-facing error strings (violates the pt-br rule)

`getNFeEmitModal`: "Invalid departure ID" (router.go:662), "Failed to get
departure" (:669), "Failed to check existing invoice" (:677);
`previewNFe`: "Failed to get departure" (:139). AGENTS.md requires pt-br for
all user-facing text.

#### F12. No server-side format validation for free-text tax fields

`validateDepartureTaxCodes` (`departure_inputs.go:71`) only validates
ICMS/PIS/COFINS code lists; `ibsCST`, `cbsCST`, `cClassTrib`, `cest`, and
`naturezaOp` (natOp ≤ 60 chars per MOC) go to the XML unvalidated → late
SEFAZ rejection at emission time. `NCM` is `readonly` in the UI, but that
is not a server-side boundary.

#### F13. No upper bound on rates, no positive-price check

`parsePercentRateOrNil` (router.go:895) rejects negatives but accepts e.g.
"500" → 5.0 ICMS rate. There is also no `unitPrice > 0` check (a zero-price
NF-e with the sales CFOPs offered in the select).

#### F14. Plate UF guessed as emitter UF

`defaultTransportPlateUF = farmNFeConfig.EmitterUF` (router.go:761): a
vehicle registered in another state emits with the wrong plate UF. The
vehicle table has no UF field, so this is a heuristic — worth surfacing or
storing.

#### F15. `previewNFe` re-queries the departure after generating the PDF

router.go:136–141 — a failure there 500s after the expensive work, with a
plain string body instead of a toast (and an English message, see F11).

#### F16. Silent modFrete "fix"

If the farm config holds a `DefaultModFrete` outside 0/1/2/3/4/9, no
`<option>` matches and the browser silently selects the first ("0"), so the
submitted value masks the bad config instead of surfacing it (the service
would otherwise reject it via `validDepartureModFrete`).

---

## Checked and found sound

- **Template escaping / XSS:** all interpolated values go through Go
  `html/template` auto-escaping (text, attribute, and textarea contexts);
  `PDFBase64` is server-generated base64.
- **Decimal mask contract:** `wireDecimalParamOverride`
  (`assets/js/decimalMask.js:311`) only rewrites parameters whose serialized
  value exactly matches the input's visible value; disabled inputs (never
  serialized) and unrelated forms are untouched. Comma display ↔ dot raw
  translation is consistent with the server parsers
  (`parsePercentRateOrNil` dot-only, `parseDepartureDecimal` tolerant of
  both, mixed separators rejected).
- **`includeVehicle`/`includeVolumes` round-trip:** disabled fields are not
  serialized; the `*Specified` hidden markers let the server distinguish
  "absent" from "explicitly false" — modal, preview and service agree.
- **Transport validation** (`validateDetachedTransport`): transporter
  document digits/CPF/CNPJ checks, name/address/municipality length caps,
  plate format incl. Mercosul, interstate vehicle rule (Rejeição 868),
  RNTRC cap, volume quantity/weight rules.
- **Duplicate-dialog risk** from `hx-target="body" hx-swap="beforeend"`
  (departure-list-item.html:19) is mitigated in practice by the native
  `<dialog>` blocking the underlying page while open.
- **`superseded` status constraint** was widened by migration
  `000005_nfe_contingency.sql` (the original `000002` CHECK predates it).
- **Dialog lifecycle:** `close-dialog-btn` → `closeModal` and the `close`
  event → `closeNFeFormDialog` both close+remove; removal is idempotent and
  the AbortController cleans up mask/toggle listeners.

## Verification results

| Command | Result |
|---|---|
| `go test ./router/nfe_router/` | **FAIL (3)** — all attributable to F2: `TestParseDepartureFormOverrides_PaymentBillingAndTransport` (indPag 0, IncludeCobranca nil), `TestDepartureNFeModalRendersInputsAndTaxSelects` (missing `name="indPag"`/`name="includeCobranca"`), `TestDepartureNFePreviewCarriesInputsToConfirm` (missing hidden inputs). Pre-existing and documented in the SDDs referenced above. |
| Other suites | NOT RUN — read-only audit; no SEFAZ contact, no DB started, no behavior changed. |

## Recommended fix order

1. **F1** (authorization) — add `departure.Farm != farmID → 403` to
   `getNFeEmitModal`, `previewNFe`, `buildNFe` (route level or in
   `prepareInvoiceBuildData`), and scope the existing-invoice branch.
2. **F2** — wire `indPag` + `includeCobranca` through modal, preview
   round-trip and parser; the three failing tests already spec the expected
   behavior.
3. **F3** (idempotency) and **F4** (cert gate surfaced in the modal).
4. Medium items F5–F8, then the minor list.

Per the repo's SDD workflow, fixes should be implemented via
`docs/sdd/<name>/sdd.md` + `plan.md` before coding.
