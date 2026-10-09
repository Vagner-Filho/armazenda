# plan:nfe_rate_snapshot

Implementation plan for `sdd.md` in this folder. Assumes
`docs/sdd/nfe_dead_schema_cleanup/` has already been implemented (its edits to
`detached_model.go`, the service files, and migration 000027 are present in the
tree). Line numbers refer to the pre-cleanup tree as of 2026-10-09 — re-locate
by symbol name.

## Test Strategy

Behaviors, layers, fixtures, and commands:

1. **Rebuild invariant (pure functions — the core regression guard).** New
   `service/nfe_service/rates_test.go`:
   - `TestMergeRates_SnapshotIsIdempotent` — `MergeRates(merged, cfg) ==
     merged` when all five rates are non-nil (the property that makes the
     worker's re-merge a no-op on snapshotted rows).
   - `TestMergeRates_PartialSnapshotFallsBack` — nil axes fall back to farm
     config, then to the hardened fallbacks (ICMS 0.17 / PIS 0.0165 /
     COFINS 0.076 / IBS·CBS 2026); locks legacy-row behavior.
   - `TestMergeRates_UserOverFarmOverFallback` — explicit precedence; add only
     if no equivalent exists (current coverage is indirect through
     `TestBuildDetachedItems_UserRatesAndOverrides`, which moves to the hoisted
     signature in Phase 4).
   Both `MergeRates` (`service.go:958`) and `pickRate` are pure — no fixtures
   needed.
2. **Hoisted single-source merge (AC2).** Structural: `grep` checklist —
   exactly one `MergeRates` call per preparation function, none inside
   `buildDetachedItems`. Plus the existing `buildDetachedItems` value
   assertions, which continue to pass after their call sites are rewritten to
   pass the pre-merged struct.
3. **Persist-the-snapshot behavior (AC3).** XML output does not change (the
   persisted struct is identical to the one items consumed, by construction),
   so there is no new unit-testable observable beyond the invariant; the
   five persist sites are verified by the grep checklist plus the guard-blocked
   smoke for the one path reachable offline (detached normal emission). The
   worker-rebuild and both SVC attempt sites require SEFAZ/SVC and are accepted
   by focused code inspection.
4. **No schema change** — no migration edits in this SDD.
5. **Environment needs:** e2e DB at `localhost:5433` (`test/e2e`:
   `bun run db:start`, `bun run db:seed`), app run with `NFE_SEFAZ_GUARD=block`
   and `DB_*` env, a **throwaway locally generated P12** (openssl, offline)
   inserted into a temporary `nfe_farm_config` for farm 1, and `psql` for the
   smoke SQL. The temporary config row is removed after the smoke (see Risks).

Pre-implementation: baseline suites in Phase 0, note pre-existing failures,
then set `Status: WIP`.

## Phase 0 — Baseline

- `go build ./...`
- `go test ./service/nfe_service/... ./model/nfe_model/`
- Record results (incl. any pre-existing failures) in `sdd.md`; set
  `Status: WIP`.

## Phase 1 — Departure flow: return + persist the merged struct

**`service/nfe_service/service.go`**

1. `prepareInvoiceBuildData` (`:46`): add a 6th return value
   `entity.TaxRates` carrying the merged struct computed at `:181`; update the
   doc comment.
2. Rewrite the `:178-180` comment ("the raw user input is returned separately
   so the caller can persist a partial-override row") → effective-rate
   snapshot semantics.
3. `BuildInvoiceFromDeparture`:
   - destructure the new return at `:262`;
   - replace the raw-or-nil block `:306-309` with
     `ratesToPersist := &rates` (naming: keep `ratesToPersist` to minimize
     churn, or rename to `ratesSnapshot` — pick once and use consistently);
   - rewrite its comment to say the row freezes the effective rates.
4. `attemptSVCContingency`: pass `&rates` at the call site `:361`; adjust the
   parameter name/doc inside the function (`taxRates *entity.TaxRates` →
   snapshot semantics) and the create call `:449-453` (signature itself does
   not change).
5. `:638` caller (preview path): destructure the new return as `_`.

**`model/nfe_model/model.go`** — doc comment only: `CreateInvoiceWithEmission`
(`:273`) parameter doc flips from user-provided rates to effective-rate
snapshot; no signature/logic changes.

## Phase 2 — Worker rebuild persists the prepared snapshot

**`service/nfe_service/worker.go`**

1. `:227` — destructure the new 6th return of `prepareInvoiceBuildData`.
2. `:243-248` — replace the `hasAnyUserRate` branch with
   `ratesToPersist := &returnedRates` (not the `:219` pre-merge: for complete
   snapshots they are equal — idempotence — and for legacy partial rows the
   prepared struct is exactly what the rebuilt XML used).
3. Update comments `:212-218` (legacy rows-only note) and `:243-244`.

## Phase 3 — Detached flow: hoist the merge

**`service/nfe_service/detached_service.go`**

1. `detachedBuildData` — new field `RatesSnapshot entity.TaxRates` (doc: the
   exact struct buildDetachedItems used for the XML).
2. `prepareDetachedBuildData` — compute
   `ratesSnapshot := MergeRates(input.TaxRates, farmNFeConfig)` before item
   building; pass it into `buildDetachedItems` in place of
   `input.TaxRates`; attach to the returned struct.
3. `buildDetachedItems` — rename the third parameter to `rates
   entity.TaxRates` (pre-merged; five non-nil pointers by contract); delete
   the internal `rates := MergeRates(userRates, farmConfig)` at `:562`; adjust
   its doc comment ("caller owns the merge; rates are authoritative").
4. `BuildDetachedInvoice`: replace `:299-303` with
   `ratesToPersist := &build.RatesSnapshot` at the create call `:314-316`; the
   SVC call site `:354` passes the same snapshot.
5. `attemptDetachedSVCContingency`: parameter `taxRates *entity.TaxRates`
   (called at `:1001` post-cleanup) receives the snapshot; doc comment
   updated; forwarded into `CreateDetachedInvoice` unchanged.
6. `CreateDetachedInvoice` doc comment in
   `model/nfe_model/detached_model.go` — same flip as Phase 1 item 6.

## Phase 4 — Dead code + test updates + new unit tests

1. Delete `hasAnyUserRate` (`service.go:984-987`) — all call sites were
   rewritten in Phases 1-3. `grep -rn "hasAnyUserRate" ` must return nothing
   (AC5).
2. Rewrite the 3 `buildDetachedItems` test call sites to pass the pre-merged
   struct — `MergeRates(entity.TaxRates{}, cfg)`, qualified as
   `nfe_service.MergeRates(...)` if the test file is the external test
   package (check each file's package declaration when editing):
   - `service/nfe_service/detached_preview_test.go` ~`:224`, ~`:298`
   - `service/nfe_service/detached_profile_service_test.go` ~`:102`
   Value assertions stay valid: `MergeRates(entity.TaxRates{}, cfg)` yields
   the fallback/farm rates, which is what the old internal merge produced.
3. Add `service/nfe_service/rates_test.go` with the three tests from the Test
   Strategy.
4. `gofmt` the touched files.

## Phase 5 — Verification

Order: build → focused → full suites → smoke → bookkeeping.

1. `go build ./...`
2. `go test ./service/nfe_service/ ./model/nfe_model/` (focused; include
   `-run TestMergeRates` and `-run TestBuildDetachedItems` when iterating)
3. `make test-go` (AC6)
4. `make test-e2e` (AC6; Docker; auto-starts DB/app/fixtures)
5. Guard-blocked smoke (AC4; commands mirrored from the cleanup SDD):

```bash
cd test/e2e && bun run db:start && bun run db:seed
# throwaway local certificate (offline):
openssl req -x509 -newkey rsa:2048 -nodes -keyout /tmp/opencode/nfe-smoke.key \
  -out /tmp/opencode/nfe-smoke.crt -days 2 -subj "/CN=NF-e SMOKE"
openssl pkcs12 -export -out /tmp/opencode/nfe-smoke.p12 -inkey /tmp/opencode/nfe-smoke.key \
  -in /tmp/opencode/nfe-smoke.crt -passout pass:smoke
# temporary farm config for farm 1 while the guard is ACTIVE, then delete after:
psql "host=localhost port=5433 …" -c "INSERT INTO nfe_farm_config (…) …"
NFE_SEFAZ_GUARD=block DB_HOST=localhost DB_PORT=5433 DB_USER=… DB_PASS=… \
  DB_NAME=… air
# as the seeded fixture user, via UI: NF-e Avulsa → nível de alíquota ICMS
# digitada (ex.: 7) → Emitir  → toast "SEFAZ e SVC indisponiveis" (guard), row draft
psql … -c "SELECT t.icms_rate, t.pis_rate, t.cofins_rate, t.ibs_rate, t.cbs_rate
  FROM detached_nfe_tax_rates t JOIN detached_nfe_invoice i ON i.id = t.invoice_id
  ORDER BY i.id DESC LIMIT 1"
# expected: five non-null values; icms_rate = 0.07 (user-typed), others =
# temp-farm-config values or the hardcoded fallbacks
psql … -c "DELETE FROM nfe_farm_config WHERE farm_id = 1"   # cleanup
```

6. **Stop the app process** (worker keeps ticking otherwise).
7. Record every command + PASS/FAIL/NOT RUN in `sdd.md` → Verification
   Results; flip `Status: Done` only when all mapped ACs pass; keep `WIP`
   with notes otherwise.

## Risks & Notes

- **Cross-SDD file contention:** cleanup and this SDD both edit
  `detached_model.go`, `service.go`, `worker.go`, `detached_service.go`.
  Implement cleanup first and rebase this plan's hunks by symbol; do not
  parallelize the two before Phase 0 is done for tree consistency.
- **Worker persist change** means a superseded contingency row for a *legacy*
  partial-rates invoice now gets its own complete snapshot row (persisting the
  prepared merged rates instead of raw-or-nil) — deliberate: every row becomes
  self-describing. No backfill of historical partial rows (out of scope).
- **Fallback constants freeze** into each row from its emission date — that is
  the point (per-invoice reproducibility), not a regression.
- **Provenance is intentionally not persisted** (user-typed vs farm-sourced);
  parked: a future `rate_source` column if ever formally needed.
- **Preview→emit staleness stays open by design:** if the farm config changes
  between preview and emission, the preview's numbers can go stale; the
  emission snapshots what it actually used. Out of scope here.
- **Smoke touches `nfe_farm_config`** — normally forbidden on test DBs
  (AGENTS.md Rule 1). The deviation is explicit and doubly safe here: the row
  exists only while `NFE_SEFAZ_GUARD=block` is active (the guard refuses every
  non-loopback SEFAZ host before any I/O — the enforced layer), the
  certificate is a throwaway local P12, and the row is deleted immediately
  after. If signing proves impractical in the environment, record the smoke as
  NOT RUN with the reason in `sdd.md` and accept AC3 for the normal-detached
  path via the most available equivalent (unit invariants + inspection), keep
  `Status: WIP` — do not mark Done with a skipped smoke and an unverified
  claim.
