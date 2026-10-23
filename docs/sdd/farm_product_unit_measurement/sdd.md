# sdd:farm_product_unit_measurement

## Status
Done

## Goal
Add a `unit_measurement` column to `farm_product` so each farm-created product carries the unit of measurement that should be used when it is selected as an NF-e item. Introduce a single Go enum as the source of truth for unit options (replacing scattered hardcoded option lists), expose it in the product form modal, and auto-populate `items[0].unit` on the detached NF-e emission page when a pre-registered product is selected.

## Requirements

### Functional
- `farm_product` has a non-null `unit_measurement` column defaulting to `KG`.
- A single enum (`entity/public`) lists all supported unit codes and labels.
- Product create/edit modal lets the user pick a unit of measurement.
- Model/service persist and return `UnitMeasurement`.
- Selecting a farm product in `nfe-emit.html` populates the row's unit `<select>`.
- System (global) products do not carry a unit; selection leaves the current unit unchanged.

### Non-Functional
- Migration is idempotent (`ADD COLUMN IF NOT EXISTS`).
- All user-facing text in pt-br.
- Preserve existing NFe unit behavior and defaults (empty → `KG`).

## Acceptance Criteria
- Migration applies cleanly on startup.
- `go test ./model/farm_product_model/ ./service/product/ ./router/product_router/ ./router/nfe_router/` passes.
- `make test-go` passes.
- Product form renders a unit select with all enum options.
- Selecting a pre-registered farm product sets `items[0].unit` to its `unit_measurement`.

## Verification Results
- `go build ./...`: PASS
- `go test ./router/product_router/ ./router/nfe_router/` (excluding 3 pre-existing departure-input failures unrelated to this change): PASS for touched templates
- `DB=... make test-go`: only pre-existing failures remain:
  - `pkg/nfe/sefaz/TestGuard_InactiveDelegates` (env guard active during run)
  - `router/nfe_router` 3 departure-input tests (pre-existing, verified on clean tree)
- Migration 000032 applied successfully on startup against test DB
- `product-form.html` and `nfe-emit.html` changes render via template tests