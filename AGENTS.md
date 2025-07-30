## Development Environment

This project uses Go and TailwindCSS.

### Build & Run

- To build the Go application: `go build -o ./tmp/main .`
- To run the development server with live reload: `air`
- To build the CSS: `tailwindcss -i assets/static/css/input.css -o assets/static/css/output.css`

### Testing

- Run all tests: `go test ./...`
- Run tests in a specific package: `go test ./model/field_model/`
- Run a single test: `go test -run ^TestFieldModel_AddField$ ./model/field_model/`

### Code Style

- **Go:** Follow standard Go conventions (`gofmt`).
- **Imports:** Group standard library, third-party, and internal packages.
- **Error Handling:** Use custom error types from the `model/error` package where applicable. Return `model.Error` for user-facing errors and log internal errors.
- **Naming:** Use camelCase for local variables and function names. Use PascalCase for exported identifiers.
- **Types:** Use `decimal.Decimal` for monetary values.
- **Database:** Use `pgx` for database interaction. Mocks are generated using `pgxmock`.
