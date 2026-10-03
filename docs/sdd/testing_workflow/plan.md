# Implementation Plan: Testing Workflow

## Phase 1: Define the workflow
- Update `AGENTS.md` SDD guidance so acceptance criteria map to automated tests/commands or a justified manual check.
- Require test strategy before coding, regression coverage for behavioral fixes, tests alongside implementation phases, and transparent reporting of failed or skipped checks.
- Document which suites are appropriate for Go, JavaScript, and user-facing browser changes, without requiring E2E for unrelated work.

## Phase 2: Align Make targets
- Update `Makefile` with `test-go` (`go test ./...`), `test-js` (`cd test && bun test unit/`), `test-unit` (both unit suites), `test-e2e` (existing Playwright command), and `test` (all suites).
- Preserve `all: test build` and existing build targets.
- Update the test command section in `AGENTS.md` to match these targets and clarify the narrower scope of `test:all` under `test/package.json` if it is retained as a JavaScript-only convenience.

## Phase 3: Validate
- Make dry-runs for `test`, `test-unit`, and `test-e2e` confirmed the expected target composition.
- `make test-unit` passed the full Go and JavaScript unit suites.
- E2E was not run because its runner is unchanged; its Make target was verified with a dry-run and is included in `make test`.
- `git diff --check` passed.

## Risks and out of scope
- `make test` includes Playwright E2E and therefore requires Docker and may take longer than the unit-only target.
- No CI pipeline or application test coverage is added as part of this change.
