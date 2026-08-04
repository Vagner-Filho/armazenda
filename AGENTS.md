# Development Guidelines for Armazenda

This document provides guidelines for AI agents working on the Armazenda codebase.

## Development Environment

- **Go**: 1.25.4
- **Database**: PostgreSQL with pgx driver
- **CSS**: TailwindCSS
- **Architecture**: Layered (Entity → Model → Service → Router)

## Build & Run Commands

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
- Use `pgxmock` for mocking database operations in tests

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
- Mock database operations using `pgxmock` in tests

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
├── entity/public/          # Domain entities (Entry, Field, Person, etc.)
├── model/                  # Data access layer
│   ├── field_model/        # Package: snake_case + _model suffix
│   ├── entry_model/
│   └── error/              # Custom error types
├── service/                # Business logic
│   ├── entry_service/
│   └── person_service/
├── router/                 # HTTP handlers (Gin framework)
│   ├── entry_router/
│   └── person_router/
├── pkg/calculator/         # Shared calculation logic (also WASM)
├── templates/              # HTML templates
└── assets/                 # Static files, JS, CSS, WASM
```

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
- Use `pgxmock` for database mocking
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
- `pkg/nfe/xml/sanitize.go` — `SanitizeSchemaString()`: all free-text XML fields (infCpl, xNome, xProd, xJust, ...) must pass through it — the schema `TString` pattern forbids newlines, control chars, and characters above U+00FF (SEFAZ rejects with `cvc-type.3.1.3` otherwise)
- `pkg/nfe/xml/event.go` — cancellation event builder (`BuildCancellationEvent`)
- `pkg/nfe/sefaz/response.go` — `EventoResponse` / `ParseEventoResponse`
- `service/nfe_service/service.go` — `BuildInvoiceFromDeparture()` flow, `CancelInvoice()`
- `service/nfe_service/worker.go` — draft auto-send logic
