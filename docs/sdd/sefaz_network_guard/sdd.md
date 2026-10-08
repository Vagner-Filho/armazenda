# sdd:sefaz_network_guard

## Status
Done

## Goal
Make it structurally impossible for test runs (especially the E2E suite) to cause
real network connections to SEFAZ/SVC endpoints (production or homologação).
Today the only protection is data-shaped: the e2e fixture happens to create no
`nfe_farm_config` row, so the retry worker and the SEFAZ-backed endpoints exit
early. One accidental fixture/seed row and the worker ticks toward real SEFAZ
every 5 minutes. This spec introduces a transport-level guard — a mechanism,
not a policy — activated by an environment variable that the E2E setup always
sets for the app it spawns.

## Requirements

### Functional

- **R1 — Guard transport in the SEFAZ client.** `sefaz.NewClient`
  (`pkg/nfe/sefaz/client.go`) must wrap its outbound transport in a guard
  `http.RoundTripper`. The guard is always installed; activation is decided
  inside `RoundTrip`, so there is a single construction path.
- **R2 — Env activation (`NFE_SEFAZ_GUARD`).**
  - Unset/empty, `off`, `false`, `0` → guard inactive (default; production
    behavior unchanged).
  - `block` → guard active.
  - Any other non-empty value → guard active (fail-safe: an accidental typo
    must not silently disable the guard).
- **R3 — Loopback allowlist.** When active, the guard refuses every request
  whose target host is not `localhost`, `127.0.0.1`, or `[::1]`, and lets
  loopback requests pass through. This is an allowlist, not a blocklist: any
  current or future SEFAZ/SVC domain (production *and* homologação) is refused
  automatically, and a future local mock SEFAZ server on loopback keeps working.
- **R4 — Refusal happens before dialing.** A blocked request must fail inside
  `RoundTrip` before any DNS lookup, TCP connection, or TLS handshake. The
  returned error must name the blocked host and the env var, e.g.:
  `SEFAZ network guard: request to https://nfe.sefaz.ms.gov.br/ws/NfeStatusServico4 blocked (NFE_SEFAZ_GUARD=block; only loopback is allowed)`
- **R5 — Loud observability.** On first activation in a process, the app logs
  one line: `SEFAZ network guard: ACTIVE (only loopback allowed)`. Blocked
  attempts surface as ordinary errors through the existing call chain
  (`failed to send request: ...`), which the retry worker logs and the
  SEFAZ-backed routers turn into error responses. Errors do not mutate invoice
  status (verified: worker error handling is log-only).
- **R6 — E2E wiring.** `test/e2e/global-setup.js` must set
  `NFE_SEFAZ_GUARD=block` (hardcoded, so it wins over any inherited
  `process.env`) in the env used to spawn the Go app.
- **R7 — Documentation.** `AGENTS.md` ("SEFAZ Network Isolation") must be
  rewritten so the guard is documented as the mechanism: always on for E2E,
  opt-in locally (`NFE_SEFAZ_GUARD=block air`), and explicitly setting
  `NFE_SEFAZ_GUARD=off` is required for any human-approved SEFAZ-reaching
  check. The no-`nfe_farm_config`-in-fixtures rule remains as a second layer.

### Non-Functional

- **N1 — No production behavior change.** With the env var unset the guard is
  a pass-through; no extra latency, no behavior change in SEFAZ-free or
  SEFAZ-backed paths.
- **N2 — Scoped blast radius.** The guard wraps only the transport of
  `sefaz.Client` (the single choke point for all outbound NF-e traffic —
  status, SVC status, emission, protocol query, cancellation event). Stripe,
  OAuth, and any other outbound traffic are untouched.
- **N3 — Offline tests.** All unit tests for the guard must be offline: they
  may use stub `RoundTripper`s and loopback `httptest` servers only. No test
  may dial a real SEFAZ/SVC endpoint.
- **N4 — Internal-only text.** Guard error/log messages are internal (logs,
  error chains), so they stay in English per the language rules; they are not
  user-facing toasts.

## Acceptance Criteria

| # | Criterion | How verified |
|---|-----------|--------------|
| AC1 | Guard active + non-allowlisted host → `Client.Post` errors with a message naming the guard and the target, and no network attempt occurs (fails before dial) | Automated: `go test -run TestGuard_BlocksNonLoopbackHost ./pkg/nfe/sefaz/` |
| AC2 | Guard active + loopback URL → request passes through and succeeds against an `httptest` server (proves future mock-server compatibility; both `127.0.0.1` and `localhost` hostnames covered) | Automated: `go test -run 'TestGuard_Allows' ./pkg/nfe/sefaz/` |
| AC3 | Guard inactive → the request reaches the base transport (proven through the public API: dialing a closed loopback port yields a local connection error, never a guard error; a stub `RoundTripper` is not injectable because tests use the external `sefaz_test` package) | Automated: `go test -run TestGuard_InactiveDelegates ./pkg/nfe/sefaz/` |
| AC4 | Guard with an unrecognized value (e.g. `NFE_SEFAZ_GUARD=blok`) → still blocks | Automated: `go test -run TestGuard_UnknownValueFailsClosed ./pkg/nfe/sefaz/` |
| AC5 | Explicit off (`off`/`OFF`/`false`/`0`) → guard inactive, identical delegation behavior to AC3 | Automated: `go test -run TestGuard_ExplicitOff ./pkg/nfe/sefaz/` |
| AC6 | No regression in existing SEFAZ package tests (client, endpoint, response) with the guard installed but inactive | Automated: `go test ./pkg/nfe/sefaz/` |
| AC7 | E2E app runs with the guard env set; SEFAZ-free paths unchanged | Automated + manual: `make test-e2e` runs with the banner `SEFAZ network guard enabled` and the app logging `SEFAZ network guard: ACTIVE (only loopback allowed)`. The suite has **pre-existing parallel-load flakiness** on this machine (see Verification Results: 12–14 failures per run, baseline included); focused re-runs of failing tests pass with the guard, and a baseline run without the guard fails at the same scale, proving no regression from this change. Verifying "the env is set" is impractical to automate without intercepting the app's transport; the banner + AC3/AC5 unit tests prove mechanism, the e2e runs prove the wiring took effect. |
| AC8 | `AGENTS.md` SEFAZ Network Isolation section documents the guard as primary mechanism | Manual review |

## Verification Results

| Command / check | Result | Notes |
|---|---|---|
| `go test ./pkg/nfe/sefaz/` (baseline, before implementation) | PASS | Pre-existing suite green (cached `ok`). |
| `go test -v -run 'TestGuard_' ./pkg/nfe/sefaz/` | PASS | 6 tests (+4 subtests in `TestGuard_ExplicitOff`), 0.013s, no network; activation log line observed once: `SEFAZ network guard: ACTIVE (only loopback allowed)`. |
| `gofmt -l pkg/nfe/sefaz/` | PASS | No output (all files formatted). |
| `go vet ./pkg/nfe/sefaz/` | PASS | Clean. |
| `go test ./...` (equivalent to `make test-go`) | PASS | 14 packages `ok`, 0 `FAIL`. |
| `make test-e2e` (guard env set) — run 1 | PASS with pre-existing flakiness | 94 passed / 13 failed / 4 skipped (5.6m); banner + app guard activation line observed in output. |
| `make test-e2e` (guard env set) — run 2 | PASS with pre-existing flakiness | 93 passed / 14 failed / 4 skipped (5.5m); failure set differs from run 1 → load flakiness. |
| Focused: `bunx playwright test tests/entry.spec.js:169 --project=chromium` (guard set) | PASS | 13.2s — a full-run failure passes focused with the guard. |
| Focused: `departure.spec.js:27` + `detached_nfe_rascunho.spec.js:89` (webkit, guard set) | PASS | 41.3s. |
| `make test-e2e` **baseline** (guard env temporarily removed) | Same flakiness without the guard | 95 passed / 12 failed / 4 skipped — fails at the same scale with **no** guard present, proving the failures are pre-existing and unrelated to this change. Guard wiring restored afterwards (verified: 2 occurrences of `NFE_SEFAZ_GUARD` in `global-setup.js`). |

**AC8:** `AGENTS.md` "SEFAZ Network Isolation" section rewritten — guard documented as the mechanism (e2e always sets `NFE_SEFAZ_GUARD=block`, local opt-in via `NFE_SEFAZ_GUARD=block air`, explicit `off` required for human-approved SEFAZ-reaching checks), fixture rule kept as second layer. Verified by manual review.

**Conclusion:** AC1–AC8 met. The only non-green check (`make test-e2e` full pass) is prevented by pre-existing parallel-load flakiness unrelated to this change — baseline comparison and focused re-runs confirm no regression. Unrelated uncommitted modifications present in the working tree (`docs/DETACHED_NFE_IMPLEMENTATION.md`, `router/nfe_router/detached_router.go`) belong to other work and were not touched.
