# Development Guidelines for Armazenda

This document provides guidelines for AI agents working on the Armazenda codebase.

## SDD Pattern (Spec-Driven Development)

Before implementing any feature or fix, agents must create the SDD artifacts under `docs/`. Specs first, code second.

### Folder & File Conventions

1. Create a folder `docs/sdd/<feature-or-fix>/` named after the feature or fix being worked on (use snake_case)
2. Inside it, create two files:
   - **`sdd.md`** — the specification (what and why)
   - **`plan.md`** — the implementation plan (how)

```
docs/
└── sdd/
    └── nfe_cancellation/       # folder = feature or fix name (snake_case)
        ├── sdd.md              # specification
        └── plan.md             # implementation plan
```

### sdd.md — Specification

The specification file must contain, in this order: **Status**, **Goal**, **Requirements**, and **Acceptance Criteria**.

````markdown
# sdd:<feature-or-fix>

## Status
<!-- One of: Backlog | WIP | Done -->

## Goal
<!-- Brief description of what this feature/fix delivers and why -->

## Requirements

### Functional
<!-- What the system must do -->

### Non-Functional
<!-- Performance, security, accessibility, observability, compatibility, etc. -->

## Acceptance Criteria
<!-- Verifiable, ideally testable conditions; each one should map to how the work is validated -->
````

### plan.md — Implementation Plan

The plan file describes how the implementation shall be done and may be divided into **phases** (e.g. Phase 1: schema/migration, Phase 2: service layer, Phase 3: UI, Phase 4: tests). Reference the concrete files, packages, and commands involved (following the project's Layered architecture: Entity → Model → Service → Router), and note risks or out-of-scope items when relevant.

### Workflow

1. Write `docs/sdd/<name>/sdd.md` — set `Status: Backlog` initially
2. Write the phased `docs/sdd/<name>/plan.md`
3. When starting implementation, update `Status: WIP`
4. Implement, following the acceptance criteria in `sdd.md`
5. When all acceptance criteria are met (tests passing), update `Status: Done`

### Rules

- Do **not** start coding before `sdd.md` and `plan.md` exist
- Keep the specs updated if scope changes during implementation
- Small fixes do not need SDD when the change is trivial and self-explanatory (`Status: Done` on completion is not required for those)

## Development Environment

- **Go**: 1.25.4
- **Database**: PostgreSQL with pgx driver
- **CSS**: TailwindCSS
- **Architecture**: Layered (Entity → Model → Service → Router)

## Build & Run Commands

`make` shortcuts exist (`make build` = CSS + WASM + Go, `make test` = test-unit + e2e) — but note `make test-unit` only covers `service/entry_service/test/`, `pkg/calculator/` and `pkg/nfe/...`; use `go test ./...` for the full run.

```bash
# Build the Go application
go build -o ./tmp/main .

# Run development server with live reload (builds both app and WASM)
air

# Build CSS
tailwindcss -i assets/static/css/input.css -o assets/static/css/output.css

# Build WASM module manually
GOOS=js GOARCH=wasm go build -o ./assets/wasm/calculator.wasm ./pkg/calculator/wasm
```

## Testing Commands

### Go Tests (Backend)
```bash
# Run all tests
go test ./...

# Run tests in a specific package
go test ./model/field_model/

# Run a single test by exact name
go test -run ^TestFieldModel_AddField$ ./model/field_model/

# Run tests matching a pattern
go test -run ^TestCalculateEntry ./pkg/calculator/

# Run tests with verbose output
go test -v ./pkg/calculator/
```

### Unit Tests (JavaScript)
```bash
# Run unit tests (Bun test runner)
cd test
bun test

# Run with watch mode
bun run test:watch

# Run from project root
cd test && bun test unit/
```

### E2E Tests (Playwright)
```bash
# Run all E2E tests
cd test/e2e
bun run test

# Run with UI mode
bun run test:ui

# Run headed (see browser)
bun run test:headed

# Run in debug mode
bun run test:debug

# Database management
bun run db:start    # Start test database
bun run db:seed     # Seed test data
bun run db:stop     # Stop test database
```

**E2E auto-setup:** Playwright's `global-setup.js` starts the Docker test DB (localhost:5433), starts the Go app (http://localhost:8100), and seeds fixtures automatically — manual `db:start`/`db:seed` is only needed for debugging.

**App requires DB env vars to run:** `DB_HOST`, `DB_USER`, `DB_PASS`, `DB_NAME`, `DB_PORT` (set with no defaults in code).

### Run All Tests
```bash
cd test
bun run test:all    # Runs both unit and E2E tests
```

## Code Style Guidelines

### Imports Organization

Group imports in this order with blank lines between groups:

1. **Standard library** packages
2. **Third-party** packages
3. **Internal** packages (grouped by layer)

```go
import (
    // Standard library
    "context"
    "fmt"
    "net/http"

    // Third-party
    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
    "github.com/shopspring/decimal"

    // Internal - Models
    "armazenda/model/field_model"
    "armazenda/model/error"

    // Internal - Services
    "armazenda/service/entry_service"

    // Internal - Routers
    "armazenda/router/field_router"
)
```

### Naming Conventions

| Element | Convention | Examples |
|---------|------------|----------|
| Exported identifiers | PascalCase | `GetFieldModel`, `AddField`, `EntryCalculationInput` |
| Local variables/functions | camelCase | `newField`, `calculateNetWeight`, `rawNetWeight` |
| Packages | snake_case | `field_model`, `entry_router`, `person_service` |
| Error variables | Suffix with Err | `scanErr`, `queryErr`, `addErr` |
| Interfaces | PascalCase with Interface suffix | `EntryModelInterface`, `FieldServiceInterface` |

### Type Conventions

- **Monetary/Weight values**: Always use `decimal.Decimal` (from `github.com/shopspring/decimal`)
- **Optional fields**: Use pointer types (`*uint32`, `*decimal.Decimal`) for nullable database columns
- **IDs**: Use unsigned integers (`uint8`, `uint16`, `uint32`) for database IDs
- **Timestamps**: Use `time.Time` from standard library

### Error Handling

Use the custom `ModelError` type for all model operations:

```go
// In model_error package
package model_error

type ModelError struct {
    IsServerErr bool
    Message     string
}

func (e *ModelError) Error() string {
    return e.Message
}
```

**Guidelines:**
- Return `*model_error.ModelError` from model functions
- Set `IsServerErr: true` for internal/server errors (log these, show generic message to user)
- Set `IsServerErr: false` for user-facing validation errors (show specific message)
- Log internal errors via the logger model: `model_error.GetLoggerModel().Log(...)`
- Service functions typically return the entity plus `*entity_public.Toast` (`GetWarningToast`/`GetErrorToast` on failure, message in pt-br) instead of errors — see `service/field/service.go`

### Language Requirements

All user-facing text must be written in **Brazilian Portuguese (pt-br)**. This includes:
- Toast notifications and UI alerts
- HTTP response messages
- Error messages returned to the client (`ModelError.Message`)
- Any other text displayed to the end user

Internal-only strings (log messages, comments, variable names) may remain in English.

### Database Patterns

- Use `pgx` driver with connection pooling
- Use `pgx.NamedArgs` for parameterized queries: `pgx.NamedArgs{"field_id": id}`
- Use `pgx.CollectRows` for scanning multiple rows into structs
- Each model is a singleton: `Init<X>Model(pool)` at startup, then `Get<X>Model()` from services; services depend on model interfaces (see `service/entry_service/interfaces.go`) so tests can inject hand-written mocks (see `service/entry_service/test/mocks.go`)

### Authentication

The system uses stateful JWT authentication with session validation:

1. **Login Flow**:
   - User submits credentials (CPF/password, Google OAuth, Microsoft OAuth)
   - Server creates a session record in `user_session` table
   - Server issues JWT token with `session_id` claim (20-hour expiration)
   - JWT stored in `session_id` cookie (httpOnly, secure)

2. **Request Validation**:
   - Middleware validates JWT signature and expiration
   - Middleware validates session exists in database and is active
   - Middleware checks user is not deactivated (not in `inactive_user` table)

3. **Session Management**:
   - Sessions are deleted on logout, user deactivation, or role change
   - Expired sessions can be cleaned up with `user_service.CleanupExpiredSessions()`
   - Each login creates a new session (multiple devices allowed)

4. **Key Functions**:
   - `user_service.CreateSession()`: Creates new session
   - `user_service.ValidateSession()`: Validates session exists and user is active
   - `user_service.DeleteSession()`: Deletes specific session
   - `user_service.DeleteUserSessions()`: Deletes all sessions for a user
   - `user_service.ValidateTokenAndSession()`: Validates JWT and session together

5. **Security Considerations**:
   - Immediate session invalidation on logout
   - Immediate permission revocation on role change
   - Immediate access denial on user deactivation
   - Audit trail with IP address and user agent tracking

### Project Structure

```
armazenda/
├── entity/public/          # Domain entities (Entry, Field, Person, Toast, etc.)
├── model/                  # Data access layer
│   ├── field_model/        # Package: snake_case + _model suffix
│   ├── entry_model/
│   ├── armazenda_database/ # Pool, migrations, stored procedures
│   └── error/              # Custom ModelError type + logger model
├── service/                # Business logic
│   ├── entry_service/
│   └── person_service/     # dir omits suffix, package name keeps it
├── view/                   # Display/render layer (DTOs & queries feeding templates)
├── router/                 # HTTP handlers (Gin framework)
├── pkg/calculator/         # Shared calculation logic (also compiled to WASM)
├── pkg/nfe/                # NF-e domain: xml/, sefaz/, danfe service, sign, config
├── templates/              # HTML templates (embedded via //go:embed in main.go)
└── assets/                 # Static files, JS, CSS, WASM, service worker
```

> **Directory name vs package name:** directories drop the layer suffix (`service/person/`, `model/field_model/` excepted) but the package name keeps it (`package person_service`, `package field_router`). Create new layers following this split.

**Domain reference docs** live in `pkg/nfe/resources/` (MOC layout, DANFE rules, adding-a-state guide, maintenance guide) and `docs/` (feature summaries, research notes, SDD specs).

### Database Migrations

The project uses a lightweight, custom migration system with raw SQL files.

**How it works:**
- Migration files live in `model/armazenda_database/migrations/`
- Files are named `000001_description.sql` (6-digit version + snake_case description)
- Migrations are embedded into the binary via `//go:embed` and applied on startup by `RunMigrations()`
- Applied versions are tracked in the `schema_migrations` table
- Each migration runs inside a transaction

**Convention:**
- **All new schema changes** (tables, columns, indexes, constraints) go into a new `.sql` migration file
- **Existing `InitDb` in `database.go`** remains as the baseline for bootstrapping fresh databases — do not add new schema there
- Stored procedures currently use `DROP IF EXISTS` + `CREATE OR REPLACE` in `InitDb`, which is acceptable since they refresh on every deploy

**Creating a new migration:**
1. Create `model/armazenda_database/migrations/000XXX_description.sql`
2. Write idempotent SQL (`CREATE TABLE IF NOT EXISTS`, `ADD COLUMN IF NOT EXISTS`, etc.)
3. The runner will automatically detect and apply it on the next startup

### Testing Patterns

- Test files use `_test` suffix in package name: `package calculator_test`
- Place tests in same directory as code being tested
- Mock models with hand-written mocks implementing the model interfaces (`service/entry_service/test/mocks.go` is the reference pattern) — no mocking framework is used
- Test functions use PascalCase starting with Test: `TestCalculateEntry`
- Use descriptive test names with underscores: `TestEntry_WithExceedingHumidity`

### Code Formatting

- Run `gofmt` on all Go files
- Use `goimports` for import management (stdlib, third-party, internal grouping)
- Follow standard Go conventions from Effective Go

## Additional Notes

- **WASM**: Calculator package compiled to WebAssembly for client-side use
- **Offline Support**: Offline-first architecture (see OFFLINE.md)
- **Air**: Auto-rebuilds Go app and WASM; excludes `_test.go` files

### NF-e Contingency Architecture

The NF-e system implements **automatic SVC (SEFAZ Virtual de Contingência)** contingency per MOC 7.0 Anexo III. FS-DA and EPEC are reserved in the data model but not actively wired.

**State → SVC mapping** (Ato COTEPE 39/2012):

| SVC | States |
|-----|--------|
| SVC-AN (tpEmis=6) | AC, AL, AP, MG, PB, RJ, RS, RO, RR, SC, SE, SP, TO, DF |
| SVC-RS (tpEmis=7) | AM, BA, CE, ES, GO, MA, **MT**, MS, PA, PE, PI, PR, RN |

**Emission flow:**

1. User clicks **"Emitir NF-e"**
2. System builds normal XML (`tpEmis=1`) and sends to origin SEFAZ
3. **If authorized/processing** → save with matching status, DANFE available
4. **If network error** → check SVC status
   - SVC returns **107** → allocate **new number**, rebuild XML with `tpEmis=6|7`, add `<dhCont>` + `<xJust>`, send to SVC, supersede old draft
   - SVC not **107** → save as **draft**, return **error to user** — no DANFE

**Status lifecycle:**

| Status | DANFE? |
|--------|--------|
| `draft` | No |
| `pending` | No |
| `authorized` | **Yes** |
| `cancelled` | Yes — with "NF-e CANCELADA" banner |
| `denied` | No |
| `superseded` | No |

**Worker behavior:**
- `pending` invoices → query status via the endpoint matching their `tp_emis` (normal or SVC)
- `draft` invoices (≤ 24h old) → check SVC status; if active, auto-rebuild and send
- `superseded` invoices → deferred cleanup after SVC deactivation (Phase 2)

**Cancellation flow (evento 110111):**

1. User clicks the cancel button on an `authorized` NF-e in the list, types a justification (15–256 chars) in the modal
2. `CancelInvoice()` builds and signs a cancellation event (`<envEvento>` 1.00, `Id="ID110111"+chNFe+"01"`, `detEvento` with `nProt` + `xJust`)
3. The event is sent to `RecepcaoEvento4` of the **same environment that authorized the NF-e** (origin SEFAZ for `tpEmis=1`, SVC for `tpEmis=6|7`) — per MOC Anexo III, SVC-authorized NF-e can only be cancelled at the SVC
4. Success is `cStat=135` (event registered and linked); `218` (already cancelled at SEFAZ) is treated as idempotent success to reconcile local state
5. On success → `status='cancelled'`, `cancellation_reason`, `cancelled_at`, and the signed event XML in `xml_cancel_event` (legal proof, keep ≥ 5 years)
6. DANFE of a cancelled NF-e is still downloadable, rendered with a red "NF-e CANCELADA" banner (`GenerateCancelled()`)

**Key implementation files:**
- `pkg/nfe/defaults/agriculture.go` — `TpEmis` enum, `SVCForState()`
- `pkg/nfe/sefaz/endpoints.go` — SVC-AN and SVC-RS endpoint sets
- `pkg/nfe/xml/builder.go` — dynamic `tpEmis`, `<dhCont>`, `<xJust>`
- `pkg/nfe/xml/sanitize.go` — `SanitizeSchemaString()`: all free-text XML fields (infCpl, xNome, xProd, xJust, ...) must pass through it — newlines/control chars are normalized and accented Latin-1 characters are folded to their ASCII base letter (`ç`→`c`, `ã`→`a`). SEFAZ-MT rejects accented code points in the área de dados with cStat 402 ("codificação diferente de UTF-8"), so the output is a strict ASCII subset of UTF-8. Characters above U+007E that are not folded are dropped (the schema `TString` pattern would otherwise reject with `cvc-type.3.1.3`)
- `pkg/nfe/xml/event.go` — cancellation event builder (`BuildCancellationEvent`)
- `pkg/nfe/sefaz/response.go` — `EventoResponse` / `ParseEventoResponse`
- `service/nfe_service/service.go` — `BuildInvoiceFromDeparture()` flow, `CancelInvoice()`

### Detached NF-e (NF-e Avulsa)

A parallel emission flow independent of departures (romaneios), with its **own tables and service** — do not mix it with the departure-linked `nfe_invoice` flow:

- Migration `000021_detached_nfe.sql`: `detached_nfe_invoice` + `detached_nfe_tax_rates` (multi-item via JSONB)
- Model: `model/nfe_model/detached_model.go`; service: `service/nfe_service/detached_service.go`
- Feature summary: `docs/DETACHED_NFE_IMPLEMENTATION.md`

### Tax Reform (IBS / CBS)

Armazenda emits the per-item `<IBSCBS>` group and the per-NF-e `<IBSCBSTot>` block required by the indirect tax reform (EC 132/2023, NT 2025.002-RTC / MOC 7.0). Mandatory in the NF-e layout from August 2026.

**Tax axes:**
- **IBS** — Imposto sobre Bens e Serviços, state + municipal (replaces ICMS + ISS over time)
- **CBS** — Contribuição sobre Bens e Serviços, federal (replaces PIS + COFINS over time)

**2026 symbolic rates** (per Ato Conjunto RFB/CGIBS):
- CBS = 0.9 %, IBS = 0.1 % (informational; no financial effect in 2026)
- Stored as decimal *rates* (CBS=0.009, IBS=0.001); XML `pIBSUF` / `pCBS` multiply by 100

**XML structure** (`pkg/nfe/xml/builder.go`):
- Per-item `<IBSCBS>` is emitted after `<COFINS>` inside `<imposto>` (always, even pre-reform — keeps schema stable)
- Per-NF-e `<IBSCBSTot>` is emitted as a sibling of `<ICMSTot>` inside `<total>`
- The 2026 phase allocates the full IBS rate to the state share (`<gIBSUF>`); municipal (`<gIBSMun>`) stays zero — replace when state/municipal split rates are published

**Canonical element order** (per NT 2025.002-RTC v1.51 §6.7.4 + NFe_Util 2Gv5.02b reference). **SEFAZ rejects the NF-e with status 215 / `cvc-complex-type.2.4.a` when children are missing or out of order, and `cvc-complex-type.2.4.d` ("no child element is expected at this point") when extras are present.** The constants in `pkg/nfe/defaults/reform.go` are the source of truth; `TestBuilder_IBSCBS_ElementOrder` is the regression guard for both the per-item structure and the totals ordering.

Per-item (`<IBSCBS>` / `<gIBSCBS>`, Group UB, NT 2025.002-RTC v1.51):

```
IBSCBS:    CST, cClassTrib, [indDoacao 0-1], gIBSCBS
gIBSCBS:   vBC, gIBSUF, gIBSMun, vIBS, gCBS       ← FLAT sequence (no <gIBS> wrapper)
gIBSUF:    pIBSUF, [optional inner seq], vIBSUF
gIBSMun:   pIBSMun, [optional inner seq], vIBSMun
gCBS:      pCBS, [optional inner seq], vCBS         ← NO vCredPres/vCredPresCondSus (totals-only)
```

> **The per-item `<gIBSCBS>` has NO `<gIBS>` wrapper.** `gIBSUF`, `gIBSMun`, `vIBS`, `gCBS` are siblings of each other inside `gIBSCBS`. This is asymmetric with the totals block (which DOES use a `<gIBS>` wrapper) and is by design per NT 2025.002-RTC v1.51 §6.7.4 UB15. Adding a `<gIBS>` wrapper here produces SEFAZ 215 with `cvc-complex-type.2.4.a` ("Invalid content starting with element 'gIBS'. One of 'gIBSUF' is expected").

> **`<vCredPres>` / `<vCredPresCondSus>` are TOTALS-ONLY fields.** They must NOT appear as direct children of per-item `<gIBS>` or `<gCBS>`. Per-item credit-presumption fields belong inside the optional `<gIBSCredPres>` (UB78, 0-1) and `<gCBSCredPres>` (UB120, 0-1) subgroups, which are siblings of `<gIBSUF>`/`<gIBSMun>`/`<gCBS>` inside `<gIBSCBS>`. Emitting them per-item produces SEFAZ 215 with `cvc-complex-type.2.4.d` ("Invalid content starting with element 'vCredPres'. No child element is expected at this point").

> **`<vIBS>` (UB54a) is the per-item IBS total = `<vIBSUF>` + `<vIBSMun>`.** Mandatory 1-1 as a direct sibling of `<gIBSUF>`/`<gIBSMun>`/`<gCBS>` inside `<gIBSCBS>`. Added to the schema in NT 2025.002 v1.20 (not present in v1.00).

Totals (`<IBSCBSTot>` / `<gIBS>` / `<gCBS>`, Group W):

```
IBSCBSTot: vBCIBSCBS, gIBS, gCBS                 ← gMono optional 0-1 (skipped for grain)
gIBS:      gIBSUF, gIBSMun, vIBS, vCredPres, vCredPresCondSus
gIBSUF:    vDif, vDevTrib, vIBSUF                 ← all 1-1 (W38, W39, W41)
gIBSMun:   vDif, vDevTrib, vIBSMun                ← all 1-1 (W43, W44, W46)
gCBS:      vDif, vDevTrib, vCBS, vCredPres, vCredPresCondSus
                                                ← (W53, W54, W56, W51, W52)
```

Optional fields skipped for the 2026 grain-sale phase (0-1 in schema): `pDif`, `vDif`, `vDevTrib`, `pRedAliq`, `pAliqEfet`, `gIBSCredPres`, `gCBSCredPres` per item; `<gMono>` block in totals. Re-add them when Armazenda starts supporting deferral / credit / monofásico regimes.

**Gate:**
- `defaults.IsTaxReformActive(t)` returns true from 2026-01-01 onwards
- Pre-reform: the XML group is still emitted but with zero rates/values so consumers see a stable schema
- 2026+: rates come from `MergeRates(userRates, farmNFeConfig)` (`service/nfe_service/service.go`) which resolves `user > farm > 2026 fallback`

**Persistence:**
- `nfe_farm_config`: `cbs_rate`, `ibs_rate`, `cbs_cst`, `ibs_cst`, `c_class_trib` (per-farm defaults)
- `nfe_invoice_tax_rates`: `cbs_rate`, `ibs_rate` (per-emission user overrides)
- `nfe_invoice`: `cbs_value`, `ibs_value`, `cbs_cst`, `ibs_cst`, `c_class_trib` (totals + CST persisted per invoice for retry / SVC rebuild)

**Migration:** `model/armazenda_database/migrations/000019_add_ibs_cbs.sql` (idempotent `ADD COLUMN IF NOT EXISTS`).

**Out of scope (explicit):**
- Bank-mediated split payment collection at Pix/card settlement (the MOC's "split payment" feature for separating taxes at the moment of financial settlement)
- NFS-e (services) layout changes — only NF-e (goods) is implemented
- Credit balance tracking for "Superinteligente" mode
- `service/nfe_service/worker.go` — draft auto-send logic

## Other Subsystems

- **Billing/Subscriptions** (`service/billing_service/`, Stripe): webhook endpoint `POST /stripe/webhook` validates the `Stripe-Signature` header — never weaken that check
- **Offline/PWA** (`assets/sw.js`, `assets/manifest.json`, WASM calculator): architecture doc in `OFFLINE.md`; calculation logic must stay single-source in `pkg/calculator/` (shared server + WASM), never forked into JS
- **UI/design system**: `.agents/skills/ui-design/SKILL.md` + its `references/REFERENCE.md` define the project's GUI patterns — consult when touching templates
