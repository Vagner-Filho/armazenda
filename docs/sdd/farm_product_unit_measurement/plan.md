# Implementation plan: farm_product_unit_measurement

## Phases

### Phase 1 — Schema
- Create `model/armazenda_database/migrations/000032_add_farm_product_unit_measurement.sql`.

### Phase 2 — Entity + enum
- Add `entity/public/unit_measurement.go` with `UnitMeasurement` type, constants, options slice, label helper.
- Add `UnitMeasurement string` field to `FarmProduct` (form tag `unit_measurement`).

### Phase 3 — Model
- Update `model/farm_product_model/model.go` queries to include `unit_measurement`.

### Phase 4 — Service
- Normalize empty to `KG` in `AddFarmProduct`/`UpdateFarmProduct`.

### Phase 5 — Router
- Bind `FarmProduct` instead of `Product` in product router.
- Pass `UnitMeasurementOptions` to product form and detached page view.

### Phase 6 — Templates + JS
- Add unit select to `product-form.html`.
- Add `data-unit` to farm product option in `nfe-emit.html`.
- Update JS change handler to populate unit.

### Phase 7 — Verification
- Run focused tests and full Go suite.