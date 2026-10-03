# sdd:testing_workflow

## Status
Done

## Goal
Make testing an explicit part of Armazenda's spec-driven development workflow and provide reliable Make targets for the Go, JavaScript unit, and Playwright E2E suites.

## Requirements

### Functional
- Each SDD acceptance criterion must identify an automated test and command, or a manual verification with a reason automation is not appropriate.
- Test strategy and test tasks must be planned before implementation and included in the relevant implementation phases, rather than deferred to a final testing phase.
- Behavioral bug fixes should add a regression test unless that is not reasonably automatable; skipping SDD for a trivial change does not waive relevant testing.
- The development guide must specify proportionate test selection, completion/reporting expectations, and the distinction between focused, unit, E2E, and full-suite checks.
- Make targets must expose the full Go suite, JavaScript unit suite, their combined unit suites, E2E tests, and a full test target.

### Non-Functional
- Existing build targets and application behavior must remain unchanged.
- Test targets must use the existing `GO` and `BUN` Make variables.
- The full test target must include Go tests, JavaScript unit tests, and Playwright E2E tests; the unit target must not imply it runs E2E.

## Acceptance Criteria
- `AGENTS.md` makes test planning, regression tests, proportionate verification, and transparent reporting part of the workflow.
- `AGENTS.md` documents the Make targets accurately and explains that Playwright auto-setup runs the E2E database and app.
- `Makefile` provides `test-go`, `test-js`, `test-unit`, `test-e2e`, and `test`, with `test` covering all three suites.
- `make test-unit` passes, and Make dry-runs show the expected Go, JavaScript, and E2E commands.
- Existing unrelated working-tree changes remain untouched.

## Verification Results
- `make test-unit` — PASS (`go test ./...` and `cd test && bun test unit/`).
- `make -n test` — PASS; expands to Go, JavaScript unit, and E2E commands.
- `make -n test-unit` — PASS; expands to Go and JavaScript unit commands.
- `make -n test-e2e` — PASS; expands to the existing E2E runner.
- `git diff --check` — PASS.
- `make test-e2e` — NOT RUN; E2E behavior is unchanged and the target composition was verified with a dry-run.
