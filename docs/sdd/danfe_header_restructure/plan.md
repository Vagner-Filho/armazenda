# Implementation Plan: danfe_header_restructure

## Phases

### Phase 1: Database migration
- Create `model/armazenda_database/migrations/000033_create_farm_logo.sql` with `farm_logo` table.

### Phase 2: Model layer
- Create `model/farm_logo_model/model.go` with singleton + Get/Upsert.

### Phase 3: Entity
- Add `LogoData []byte` to `entity.DANFEData`.

### Phase 4: Router + template
- Update `getNFeConfigForm`, `saveNFeConfig` for logo upload.
- Update download endpoints to fetch logo.
- Add template input.

### Phase 5: Service wiring
- Preview paths fetch logo and set `LogoData`.

### Phase 6: DANFE generator
- Restructure `drawHeaderBlock` and `drawAccessKeyBlock`.
- Embed default logo.

### Phase 7: Tests
- Add focused generator tests.
- Run targeted and full Go tests.

## Test strategy
- Focused: `go test ./pkg/nfe/service/...`
- Completion: `make test-go`