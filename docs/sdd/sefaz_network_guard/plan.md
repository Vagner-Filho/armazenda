# plan:sefaz_network_guard

Implementation plan for the SEFAZ network guard. See `sdd.md` for the
specification and acceptance criteria.

## Background / chosen design

All outbound SEFAZ HTTP traffic funnels through `sefaz.Client.Post()`
(`c.httpClient.Do()`, `pkg/nfe/sefaz/client.go:65`), and `sefaz.NewClient` is
constructed only inside `pkg/nfe/service/invoice_service.go` (emission, status,
SVC status, protocol query, cancellation event). Endpoints are hardcoded
`https://*.sefaz.*.gov.br` / SVC hosts in `pkg/nfe/sefaz/endpoints.go`
(`StateRegistry` + `SVCRegistry`). The retry worker starts unconditionally at
`main.go:257`; the E2E app is spawned with a fixed env in
`test/e2e/global-setup.js` (only `DB_*` today); playwright has no `webServer`,
so `global-setup.js` is the only spawn point.

Chosen option (from the options discussion): **transport-level guard with a
loopback allowlist, activated by env var, wired by the E2E setup**. Rejected
alternatives: scattered call-site kill switch (default-open, easy to forget in
new paths), endpoint override/mock SEFAZ (valuable later, out of scope here),
infrastructure-level egress blocking (environment-specific, silent failures),
fail-closed default everywhere (changes production deployment needs),
build-tag test binary (forks code paths).

## Test strategy

- **Layer / location:** `pkg/nfe/sefaz/guard_test.go` (package
  `sefaz_test`), next to the guard code. Follows repo conventions
  (`_test` package suffix, PascalCase `TestX` names, no mocking framework —
  hand-written stubs).
- **Behaviors covered:**
  1. `TestGuard_BlocksNonLoopbackHost` — guard on + non-allowlisted host →
     error mentions `SEFAZ network guard`, the target, and `NFE_SEFAZ_GUARD`;
     no dial occurs. **Test URL note:** instead of a real SEFAZ hostname
     (which would attempt a real connection if the guard itself failed to
     activate — the very bug under test), the block tests use
     `http://127.0.0.2:1/...`: loopback-range IP (not allowlisted) whose
     failure mode, if the guard is broken, is an instant local "connection
     refused" — never a packet towards SEFAZ. Real SEFAZ hostnames are
     deliberately not dialed from tests.
  2. `TestGuard_AllowsLoopback` — guard on + `httptest.NewServer` (loopback,
     `127.0.0.1`) → request passes through and returns the served body.
  3. `TestGuard_AllowsLoopbackHostname` — guard on + same server addressed
     as `localhost:<port>` → exercises the `localhost` allowlist branch.
  4. `TestGuard_InactiveDelegates` — guard unset + closed loopback port →
     error is a local connection error, **not** a guard error (proves the
     guard delegated to the base transport; a stub `RoundTripper` is not
     injectable from the external `sefaz_test` package, so delegation is
     verified through the public API).
  5. `TestGuard_UnknownValueFailsClosed` — `NFE_SEFAZ_GUARD=blok` → blocked
     (typo must not disable the guard).
  6. `TestGuard_ExplicitOff` — `off` / `OFF` / `false` / `0` → identical
     delegation behavior to 4.
- **Environment needs:** `t.Setenv` for `NFE_SEFAZ_GUARD` manipulation;
  `httptest` loopback servers and closed-loopback-port dials only; no
  fixtures, no Docker, no network.
- **E2E:** existing suite must keep passing with the guard env set (the app's
  SEFAZ-free paths are unaffected). The guard activation line in the app
  output (stdio inherited in `global-setup.js`) is the manual wiring check.
- **Commands:**
  - During development: `go test ./pkg/nfe/sefaz/` (and focused `-run`).
  - Before completion: `make test-go`; E2E: `make test-e2e` (or a focused
    Playwright run when iterating, full suite before Done).

## Phases

### Phase 1 — SDD artifacts (this pair of files)
- [x] `docs/sdd/sefaz_network_guard/sdd.md`
- [x] `docs/sdd/sefaz_network_guard/plan.md`

### Phase 2 — Guard implementation + unit tests
- [x] Run existing tests first and note pre-existing failures:
      `go test ./pkg/nfe/sefaz/`. Then set `Status: WIP` in `sdd.md`.
      (Result: PASS before implementation.)
- [x] Create `pkg/nfe/sefaz/guard.go`.
- [x] Wire in `pkg/nfe/sefaz/client.go` `NewClient`: guard always installed.
- [x] Add `pkg/nfe/sefaz/guard_test.go` (6 tests + explicit-off subtests; safe
      URLs per the test-strategy note — no real SEFAZ hostname dialed).
- [x] Run: `go test -v ./pkg/nfe/sefaz/` — all green, no network.
- [x] `gofmt`/`goimports` on touched files (clean; `go vet` clean).

### Phase 3 — E2E wiring
- [x] `test/e2e/global-setup.js`: `NFE_SEFAZ_GUARD: 'block'` hardcoded in the
      spawn env + console line noting the guard is on.

### Phase 4 — Docs + full verification
- [x] Rewrite the "SEFAZ Network Isolation" section of `AGENTS.md`.
- [x] Run `go test ./...` and record results (14 packages ok, 0 FAIL).
- [x] Run `make test-e2e` and record results (pre-existing parallel-load
      flakiness documented; baseline comparison + focused re-runs prove no
      regression; guard banner observed in all guarded runs).
- [x] Fill "Verification Results" in `sdd.md`; `Status: Done`.

## Risks / notes

- **Logger dependency:** `pkg/nfe` packages must not depend on
  `model/error`'s DB-backed logger (initialization order in `main.go`); use
  stdlib `log` in the guard.
- **Redirects:** `http.Client` follows redirects with the same transport, so
  the guard still applies to redirected requests.
- **mTLS/TLS config:** irrelevant to the guard — refusal happens before any
  dial; homologação's `InsecureSkipVerify` is unaffected.
- **Worker behavior under guard:** SEFAZ calls fail instantly with a logged
  error; invoices keep their status (`pending`/`draft`). In E2E there are no
  NF-e rows today, so the worker paths are effectively inert regardless.
- **New outbound NF-e code:** any future SEFAZ call must go through
  `sefaz.Client` (already the pattern — all 5 construction sites live in
  `pkg/nfe/service/invoice_service.go`). Code review rule, not automated:
  don't create ad-hoc `http.Client`s in `pkg/nfe` for SEFAZ traffic.

## Out of scope

- Endpoint override + local mock SEFAZ server (future follow-up; the loopback
  allowlist deliberately leaves the door open for it).
- Worker early-exit check on the env var (redundant: the transport refuses
  instantly, no timeout cost).
- Guarding non-SEFAZ outbound traffic (Stripe, OAuth) or infrastructure-level
  egress blocking.
- Changing default behavior of production deployments.
