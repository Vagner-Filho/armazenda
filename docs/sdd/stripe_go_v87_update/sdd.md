# sdd:stripe_go_v87_update

## Status
Done

## Goal
Update the codebase to use stripe-go v87 instead of v85 (the user upgraded the dependency from v86 to v87).

## Requirements

### Functional
- All Stripe imports must reference `github.com/stripe/stripe-go/v87`.
- `go.mod`/`go.sum` must declare only stripe-go v87 (no v85/v86).

### Non-Functional
- Project must build (`go build ./...`) and unit tests pass.

## Acceptance Criteria
- No references to `stripe-go/v85` or `stripe-go/v86` remain in Go source.
- `go build ./...` succeeds.
- `go mod tidy` leaves only v87 in go.sum.

## Verification Results
- `go build ./...`: PASS
- `go mod tidy`: PASS
- `go test ./...`: PASS except pre-existing unrelated `router/nfe_router` failures (not caused by this change).
`