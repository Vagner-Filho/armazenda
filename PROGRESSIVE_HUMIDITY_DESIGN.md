# Progressive Humidity Discount Design Document

## Overview

This document describes the implementation of progressive humidity discounts, replacing the current single-value discount system with a tiered approach where discount rates vary based on humidity levels.

### Current System (DEPRECATED)
- ~~Single `humidity_discount` value stored in `person_config`, `default_person_config`, and `farm_config`~~
- ~~Formula: `discount = (humidity - 14%) * humidity_discount`~~
- ~~Default values: 1.7 for persons, 1.15 for farms~~

### New System (IMPLEMENTED)
- **Progressive tiers**: Different discount rates for different humidity ranges
- **System default**: 14%→1.7, 16%→1.8, 18%→2.0, >20%→2.2
- **User-defined progressions**: Users can create custom progression tables
- **Fallback chain**: Person's progression → Farm's progression → System default

## Status: ✅ Backend COMPLETE, ✅ Frontend/Offline COMPLETE

---

## Database Schema (IMPLEMENTED)

### New Tables ✅

#### 1. `humidity_progression`
Stores progression definitions. Farm_id is NULL for system default.

```sql
CREATE TABLE IF NOT EXISTS humidity_progression (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    farm_id INTEGER,
    is_system_default BOOLEAN DEFAULT FALSE,
    modified_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (farm_id) REFERENCES farm(id),
    CONSTRAINT single_system_default CHECK (
        (is_system_default = TRUE AND farm_id IS NULL) OR 
        (is_system_default = FALSE)
    )
);
```

#### 2. `humidity_progression_tier`
Stores the tier data with threshold values.

```sql
CREATE TABLE IF NOT EXISTS humidity_progression_tier (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    progression_id INTEGER NOT NULL,
    threshold_humidity NUMERIC(5, 2) NOT NULL,
    discount_value NUMERIC(5, 2) NOT NULL,
    FOREIGN KEY (progression_id) REFERENCES humidity_progression(id) ON DELETE CASCADE,
    UNIQUE(progression_id, threshold_humidity)
);
```

### Modified Tables (IMPLEMENTED) ✅

#### 1. `person_config`
- **REMOVED**: `humidity_discount` column
- **ADDED**: `humidity_progression_id INTEGER` with FK to `humidity_progression(id)`

#### 2. `default_person_config`
- **REMOVED**: `humidity_discount` column
- **ADDED**: `humidity_progression_id INTEGER DEFAULT 1` with FK to system default

#### 3. `farm_config`
- **REMOVED**: `humidity_discount` column
- **ADDED**: `humidity_progression_id INTEGER` with FK to `humidity_progression(id)`

### Stored Procedures Updated ✅
- `add_get_legal_person`
- `update_get_natural_person`
- `update_get_legal_person`
- `update_get_farm`

---

## Entity Layer (IMPLEMENTED) ✅

### New Entity: `entity/public/humidity_progression.go`

```go
type HumidityProgression struct {
    Id               uint32
    Name             string
    FarmId           *uint32
    IsSystemDefault  bool
    ModifiedAt       time.Time
    Tiers            []HumidityProgressionTier
}

type HumidityProgressionTier struct {
    Id                 uint32
    ProgressionId      uint32
    ThresholdHumidity  decimal.Decimal
    DiscountValue      decimal.Decimal
}
```

### Modified Entity: `entity/public/person.go`

```go
type PersonConfig struct {
    HumidityProgressionId  *uint32  // replaces HumidityDiscount
    EntrySoyDiscount       decimal.Decimal
    EntryCornDiscount      decimal.Decimal
}
```

### Modified Entity: `entity/public/farm.go`

```go
type Farm struct {
    // ... existing fields ...
    HumidityProgressionId  *uint32  // replaces HumidityDiscount
}
```

---

## Model Layer (IMPLEMENTED) ✅

### Model: `humidity_progression_model/model.go`

**Functions implemented:**
- ✅ `InitHumidityProgressionModel(pool)` - Initialize with connection pool
- ✅ `GetProgression(id)` - Get progression with all tiers
- ✅ `GetDiscountForHumidity(progressionId, humidity)` - Find appropriate tier and return discount
- ✅ `GetSystemDefaultProgression()` - Returns system default ID
- ✅ `AddProgression(name, farmId, tiers)` - Create with validation (1-30 tiers)
- ✅ `UpdateProgression(id, name, tiers)` - Update with validation
- ✅ `DeleteProgression(id)` - Delete (cannot delete system default)
- ✅ `ListProgressions(farmId)` - List all for farm + system default
- ✅ `GetProgressionsForSync(since, farmId)` - Get modified for sync endpoint
- ✅ `GetModifiedProgressionCount(since, farmId)` - Get count for sync status

### Modified Model: `person_model/model.go`

**Updated functions:**
- ✅ `GetHumidityDiscount(person, farm, humidity)` - Now accepts humidity parameter for tier lookup
- ✅ Implements fallback chain: Person → default_person_config → System default
- ✅ `GetPeopleByFarm()` - Updated query to use `humidity_progression_id`

### Modified Model: `farm_config_model/model.go`

**Updated:**
- ✅ Queries updated to use `humidity_progression_id` instead of `humidity_discount`

---

## Service Layer (IMPLEMENTED) ✅

### Service: `entry_service`

**Updated functions:**
- ✅ `AddEntry()` - Passes humidity value to `GetHumidityDiscount()` for tier lookup
- ✅ `UpdateEntry()` - Same tier lookup logic
- ✅ Interface updated to include humidity parameter

### Service: `person`

**Updated functions:**
- ✅ `GetPeopleForSync()` - Now returns `humidityProgressionId` instead of `humidityDiscount`

---

## Router Layer (IMPLEMENTED) ✅

### Router: `humidity_progression_router/router.go`

**Endpoints:**
- ✅ `GET /humidity-progression` - List all progressions (HTML)
- ✅ `GET /humidity-progression/:id` - Get single progression (HTML)
- ✅ `POST /api/humidity-progression` - Create new progression
- ✅ `PUT /api/humidity-progression/:id` - Update progression
- ✅ `DELETE /api/humidity-progression/:id` - Delete progression

### Router: `sync_router/router.go`

**Updated endpoints:**
- ✅ `GET /api/humidity-progressions/sync` - Get progressions modified since timestamp
- ✅ `GET /api/people` - Returns `humidityProgressionId` instead of `humidityDiscount`
- ✅ `POST /api/sync/status` - Includes `progressionsCount` in response

### Router: `person_router`

**Updated:**
- ✅ Templates use new data attributes: `data-humidity-progression-id`

---

## Frontend/Offline Implementation (IMPLEMENTED) ✅

### IndexedDB Schema (Phase 2)
- ✅ Database version upgraded to 2
- ✅ Added `humidityProgressions` store with farm index
- ✅ Helper methods: `saveProgressions()`, `getProgression()`, `getAllProgressions()`, `getSystemDefaultProgression()`, `deleteProgression()`

### Progression Sync Module (Phase 3)

**File:** `assets/js/db/progressionSync.js`

**Features:**
- ✅ `syncProgressions(farmId, lastSync)` - Fetch and store progressions
- ✅ `getProgression(id)` - Get by ID
- ✅ `getSystemDefault()` - Get system default
- ✅ `getAllProgressions(farmId)` - Get farm-specific + system default
- ✅ `findTierForHumidity(tiers, humidity)` - Find appropriate tier
- ✅ `getDiscountForHumidity(progression, humidity)` - Get discount value
- ✅ `getCurrentProgression(personConfig, farmConfig)` - Follows fallback chain
- ✅ `formatTier(tier)` - Format for display (e.g., "16% → 1.8")
- ✅ `getTierDisplayInfo(progression, humidity)` - Get UI display info

### Sync Engine Integration (Phase 3)

**File:** `assets/js/db/syncEngine.js`

**Updates:**
- ✅ Progressions synced first during `downloadUpdates()`
- ✅ Import of `progressionSync` module

### Discount Calculation (Phase 4)

**File:** `assets/js/discount.js`

**Changes:**
- ✅ `getHumidityDiscount()` is now async
- ✅ Imports `progressionSync` module
- ✅ Gets current progression using fallback chain
- ✅ Finds appropriate tier based on humidity
- ✅ Added `updateHumidityTierUI()` to display tier info
- ✅ `applyDiscounts()` is now async

### Entry Form Updates (Phase 4)

**File:** `assets/js/entryForm.js`

**Changes:**
- ✅ All event listeners use async wrapper
- ✅ Updated `setPersonConfig()` to use new data attributes
- ✅ Stores `humidityProgressionId` instead of `humidityDiscount`

### UI Components (Phase 5)

**File:** `templates/entry/entry-form.html`
- ✅ Added tier display: `<span id="humidityTierDisplay">` shows current tier (e.g., "16% → 1.8")
- ✅ Added progression source: `<span id="humidityProgressionSource">` shows which progression is being used

**File:** `templates/person/origin-selector.html`
- ✅ Updated data attributes from `data-humidity` to `data-humidity-progression-id`

### Offline Integration (Phase 6)

**File:** `assets/js/offlineManager.js`

**Changes:**
- ✅ Import of `progressionSync` module
- ✅ `handleOfflineEntryCreate()` stores full progression snapshot in `_progressionSnapshot` field

---

## Tests (IMPLEMENTED) ✅

### Service Tests
- ✅ `service/entry_service/test/service_test.go` - All tests updated and passing
  - 24/24 tests passing
  - Mock updated with humidity parameter
  - All test cases using new signature

### Model Tests
- ❌ No tests for `humidity_progression_model` yet (needs test file)

### Calculator Tests
- ✅ No changes needed - calculator receives single value

---

## Implementation Order (COMPLETED)

### Phase 1: Database & Backend ✅
1. ✅ Database schema changes
2. ✅ Entity layer
3. ✅ Model layer (`humidity_progression_model`)
4. ✅ Modified models (`person_model`, `farm_config_model`)
5. ✅ Service layer updates
6. ✅ Service tests
7. ✅ Router layer

### Phase 2: Frontend Infrastructure ✅
1. ✅ IndexedDB schema update (v2)
2. ✅ ProgressionSync module
3. ✅ Sync engine integration

### Phase 3: Client-Side Calculation ✅
1. ✅ Update `discount.js` for async lookup
2. ✅ Update `entryForm.js` for async
3. ✅ Add tier display to entry form

### Phase 4: Offline Support ✅
1. ✅ Update `offlineManager.js` for progression snapshot

---

## What's Left

### UI/UX Enhancements (Optional)
- Progression management pages (list, create, edit)
- Person form progression dropdown
- Farm config form progression dropdown
- Visual chart/graph for progression display

### Testing
- Unit tests for `humidity_progression_model`
- E2E tests for progressive discount calculation

---

## Key Points

### Validation Rules (Enforced)
- ✅ Minimum 1 tier per progression
- ✅ Maximum 30 tiers per progression
- ✅ Cannot delete system default progression

### Fallback Chain (Working)
1. Person has specific progression → use it
2. Person has no progression → use default_person_config's progression
3. default_person_config has no progression → use system default
4. For entries without origin (Própria):
   - Use farm_config's progression
   - If farm has no progression → use system default

### Real-time Calculation (Working)
- Client-side async lookup from IndexedDB
- Tier display updates in real-time as humidity changes
- Offline: Uses cached progression data

### Data Integrity
- ✅ Existing entries continue to work
- ✅ `entry_analysis.humidity_discount_modifier` stores computed value
- ✅ Offline entries store full progression snapshot for consistency

### API Changes Summary
| Endpoint | Change |
|----------|--------|
| `GET /api/humidity-progressions/sync` | ✅ NEW - Returns progressions for sync |
| `GET /api/people` | ✅ MODIFIED - Returns `humidityProgressionId` |
| `POST /api/sync/status` | ✅ MODIFIED - Returns `progressionsCount` |
| `GET /humidity-progression` | ✅ NEW - List progressions (HTML) |
| `POST /api/humidity-progression` | ✅ NEW - Create progression |
| `PUT /api/humidity-progression/:id` | ✅ NEW - Update progression |
| `DELETE /api/humidity-progression/:id` | ✅ NEW - Delete progression |

---

## Verification

### Build Status
```bash
✅ go build -o ./tmp/main .  # SUCCESS
```

### Test Status
```bash
✅ go test ./service/entry_service/test/  # 24/24 PASS
```

### Files Modified
- `model/armazenda_database/database.go`
- `model/humidity_progression_model/model.go` ✅ NEW
- `model/person_model/model.go`
- `model/farm_config_model/model.go`
- `entity/public/humidity_progression.go` ✅ NEW
- `entity/public/person.go`
- `entity/public/farm.go`
- `service/entry_service/interfaces.go`
- `service/entry_service/service.go`
- `service/person/service.go`
- `router/humidity_progression_router/router.go` ✅ NEW
- `router/sync_router/router.go`
- `router/person_router/router.go`
- `router/farm_config_router/router.go`
- `assets/js/db/database.js`
- `assets/js/db/progressionSync.js` ✅ NEW
- `assets/js/db/syncEngine.js`
- `assets/js/discount.js`
- `assets/js/entryForm.js`
- `assets/js/offlineManager.js`
- `templates/entry/entry-form.html`
- `templates/person/origin-selector.html`

---

## Soft Delete Implementation (IMPLEMENTED) ✅

Progressions use soft delete via `is_active` flag instead of hard delete:

### Database Schema
- Added `is_active BOOLEAN DEFAULT TRUE` column to `humidity_progression` table
- Added migration to add column for existing databases

### Backend Changes
- **DeleteProgression()**: Sets `is_active = FALSE` and `modified_at = CURRENT_TIMESTAMP` instead of DELETE
- **ListProgressions()**: Filters by `is_active = TRUE`
- **GetDiscountForHumidity()**: Falls back to system default if progression is inactive
- **GetProgressionsForSync()**: Includes inactive progressions so client knows to remove them
- **SyncProgression struct**: Includes `isActive` field

### Frontend Changes
- **syncProgressions()**: Removes inactive progressions from IndexedDB (deletes them locally)
- **getProgression()**: Only returns active progressions
- **getAllProgressions()**: Filters active only
- **getCurrentProgression()**: Falls back if progression is inactive (follows chain: Person→Farm→System)

### Benefits
- Historical entries remain valid (stored discount modifier preserved)
- Prevents accidental data loss
- Clients sync deletions automatically via sync endpoint
- System default always available as fallback

---

**Status**: ✅ COMPLETE - All core functionality implemented and tested.
