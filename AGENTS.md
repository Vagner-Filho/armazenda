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

### Database Patterns

- Use `pgx` driver with connection pooling
- Use `pgx.NamedArgs` for parameterized queries: `pgx.NamedArgs{"field_id": id}`
- Use `pgx.CollectRows` for scanning multiple rows into structs
- Mock database operations using `pgxmock` in tests

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
