# sdd:nfe_rate_snapshot

## Status
Backlog

> **Sequencing:** implement after `docs/sdd/nfe_dead_schema_cleanup/` (Backlog).
> Both efforts touch `model/nfe_model/detached_model.go` and the
> `service/nfe_service` files; this SDD assumes the cleanup edits are already
> in the tree (e.g. `CreateDetachedInvoice` already carries `dhCont`/`xJust`
> params, dead helpers already deleted). Line numbers below refer to the
> pre-cleanup tree as of 2026-10-09; re-locate by symbol name after cleanup.

## Goal

Freeze the **effective** tax rates of each emission in the
`nfe_invoice_tax_rates` / `detached_nfe_tax_rates` rows, so draft/SVC
rebuilds and any DB consumer reproduce the exact emitted XML regardless of
later farm-config changes.

**Root cause of the drift:** `MergeRates(userRates, farmConfig)` resolves
each axis at emission time (user-typed → farm config → hardcoded fallback:
ICMS 0.17, PIS 0.0165, COFINS 0.076, IBS/CBS 2026) — but persistence records
only the **raw user-typed rates**, and only when the user typed at least one
(`hasAnyUserRate`). Farm-sourced and fallback rates are never snapshotted, so
a farm-config edit between the original emission and a draft/SVC rebuild
silently changes the rebuilt XML's rates.

**Design decisions (user-confirmed):**
1. **Snapshot for every emission** — not only when the user typed rates.
2. **Hoist** the detached merge into the single-source preparation step
   (`prepareDetachedBuildData`) instead of merging inside
   `buildDetachedItems`.
3. **Flip the row semantics** — `*_tax_rates` rows record the effective rates
   the XML actually used, not the user's raw input. Provenance
   (user-overridden vs farm-sourced) is not recorded; a hypothetical
   `rate_source` column is parked for later if ever needed.

## Requirements

### Functional

- **R1** — `prepareInvoiceBuildData` (`service/nfe_service/service.go:46`)
  returns the merged `entity.TaxRates` (computed at `:181`) as an additional
  return value, so callers can persist exactly what `buildItems` consumed.
- **R2** — Departure flow persists the snapshot: `BuildInvoiceFromDeparture`
  (persist site at `:306-329`) and `attemptSVCContingency` (call `:361`,
  create at `:449-453`) persist the merged struct instead of the raw or nil
  input. The worker's draft rebuild (`worker.go:227` caller) persists the
  merged rates **returned by `prepareInvoiceBuildData`** (not its own pre-merge
  at `:219`) — for complete snapshots the two are identical; for legacy partial
  rows the prepared one is the exact struct the rebuilt XML used.
- **R3** — Detached flow hoists the merge: `prepareDetachedBuildData` computes
  `MergeRates(input.TaxRates, farmNFeConfig)` once and carries it on
  `detachedBuildData` (new field, e.g. `RatesSnapshot`); `buildDetachedItems`
  takes the pre-merged struct as its parameter (its internal
  `MergeRates` at `detached_service.go:562` is removed);
  `BuildDetachedInvoice` and `attemptDetachedSVCContingency` persist
  `&build.RatesSnapshot`.
- **R4** — Every emission path persists the full snapshot:
  `nfe_invoice_tax_rates` / `detached_nfe_tax_rates` rows from this point on
  have all five non-null columns (icms, pis, cofins, ibs, cbs). The existing
  skip-if-all-nil guard in the model insert helpers stays as a defensive
  legacy-shape guard; it simply never triggers for new emissions.
- **R5** — `hasAnyUserRate` (`service.go:984-987`; call sites `:307`,
  `worker.go:246`, `detached_service.go:300`) is deleted with its call sites
  rewritten per R2/R3.
- **R6** — No schema migration, no XML/DANFE output change: the persisted
  struct is, by construction, the same merged struct the item XML already used
  (`MergeRates` is pure; identical arguments → identical result). Backward
  compatibility for legacy partial tax-rates rows is preserved: the worker's
  per-rate nil → farm-fallback merge (`worker.go:212-219`) stays untouched.

### Non-Functional

- **N1** — Rebuild safety invariant locked by unit test:
  `MergeRates(fullSnapshot, farmConfig) == fullSnapshot` (all five rates
  non-nil) — the property that makes the worker's re-merge a no-op on
  snapshotted rows.
- **N2** — Smoke/testing stays SEFAZ-offline: `NFE_SEFAZ_GUARD=block` for any
  running app, throwaway locally-generated certificate for the smoke farm
  config, no direct SEFAZ calls (AGENTS.md "SEFAZ Network Isolation"). The
  smoke config row is inserted only while the guard is active and removed
  afterwards (documented deviation from Rule 1, guarded by the network guard).
- **N3** — No user-visible behavior change: same forms, toasts, statuses.
  All comments/messages in pt-br where user-facing (none introduced here).

## Acceptance Criteria

| # | Criterion | Verification |
|---|-----------|--------------|
| AC1 | Rebuild invariant: `MergeRates` with a complete snapshot returns it unchanged; with a partial (legacy) snapshot the missing axes fall back to farm/fallback values | `go test -run TestMergeRates ./service/nfe_service/` — new unit tests in `service/nfe_service/rates_test.go` |
| AC2 | The detached flow's persisted snapshot equals the struct `buildDetachedItems` used for the XML (single computation, hoisted) | `go test ./service/nfe_service/...` + grep checklist below (only one `MergeRates` call per prepare step; none inside `buildDetachedItems`) |
| AC3 | Normal + SVC attempt + worker-rebuild paths persist the full snapshot for both flows | Code-inspection grep checklist (5 persist sites) + guard-blocked smoke (AC4). The SVC/worker sites cannot run offline; accepted by inspection |
| AC4 | Smoke: a guard-blocked detached emission (throwaway local certificate in a temporary `nfe_farm_config`, removed after) persists a `detached_nfe_tax_rates` row with all five columns non-null and values matching user>farm>fallback expectations; departure-flow smoke optional, accepted by inspection | Manual smoke on the e2e DB, exact commands in plan.md Phase 5 |
| AC5 | `hasAnyUserRate` fully removed | `grep -rn "hasAnyUserRate" ` → 0 hits |
| AC6 | Existing suites green; the 3 `buildDetachedItems` test call sites updated for the hoisted signature | `make test-go` (PASS), `make test-e2e` (PASS) |
| AC7 | Doc comments flipped to snapshot semantics (persist-site and worker comments) | Manual inspection (grep checklist) |

## Verification Results

<!-- To be filled during implementation: exact commands and PASS/FAIL/NOT RUN with reasons. -->
