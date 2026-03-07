# Progressive Humidity Discount Design Document

## Overview

This document describes the implementation of progressive humidity discounts, replacing the current single-value discount system with a tiered approach where discount rates vary based on humidity levels.

### Current System
- Single `humidity_discount` value stored in `person_config`, `default_person_config`, and `farm_config`
- Formula: `discount = (humidity - 14%) * humidity_discount`
- Default values: 1.7 for persons, 1.15 for farms

### New System
- **Progressive tiers**: Different discount rates for different humidity ranges
- **System default**: 14%→1.7, 16%→1.8, 18%→2.0, >20%→2.2
- **User-defined progressions**: Users can create custom progression tables
- **Fallback chain**: Person's progression → Farm's progression → System default

## Database Schema Changes

### New Tables

#### 1. `humidity_progression`
Stores progression definitions. Farm_id is NULL for system default.

```sql
CREATE TABLE IF NOT EXISTS humidity_progression (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL,
    farm_id INTEGER,  -- NULL means system-level (not tied to a specific farm)
    is_system_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (farm_id) REFERENCES farm(id),
    CONSTRAINT single_system_default CHECK (
        (is_system_default = TRUE AND farm_id IS NULL) OR 
        (is_system_default = FALSE)
    )
);
```

#### 2. `humidity_progression_tier`
Stores the tier data with threshold values (Option B - threshold-based).

```sql
CREATE TABLE IF NOT EXISTS humidity_progression_tier (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    progression_id INTEGER NOT NULL,
    threshold_humidity NUMERIC(5, 2) NOT NULL,  -- e.g., 14, 16, 18, 20
    discount_value NUMERIC(5, 2) NOT NULL,      -- e.g., 1.7, 1.8, 2.0, 2.2
    FOREIGN KEY (progression_id) REFERENCES humidity_progression(id) ON DELETE CASCADE,
    UNIQUE(progression_id, threshold_humidity)
);
```

### Modified Tables (Schema Changes in database.go)

#### 1. `person_config`
Remove `humidity_discount` column, add reference to progression.

```sql
CREATE TABLE IF NOT EXISTS person_config (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    person_id INTEGER UNIQUE NOT NULL,
    ie TEXT NOT NULL,
    farm INTEGER NOT NULL,
    -- REMOVED: humidity_discount NUMERIC(5, 2),
    humidity_progression_id INTEGER,  -- NEW: reference to progression
    entry_soy_discount NUMERIC (5, 2),
    entry_corn_discount NUMERIC (5, 2),
    FOREIGN KEY (farm) REFERENCES farm(id),
    FOREIGN KEY (humidity_progression_id) REFERENCES humidity_progression(id)
);
```

#### 2. `default_person_config`
Remove `humidity_discount` column, add reference to system default progression.

```sql
CREATE TABLE IF NOT EXISTS default_person_config (
    id INTEGER PRIMARY KEY DEFAULT 1,
    -- REMOVED: humidity_discount NUMERIC(5, 2) NOT NULL DEFAULT 1.7,
    humidity_progression_id INTEGER DEFAULT 1,  -- NEW: points to system default progression
    entry_soy_discount NUMERIC (5, 2) NOT NULL DEFAULT 3.5,
    entry_corn_discount NUMERIC (5, 2) NOT NULL DEFAULT 5.5,
    FOREIGN KEY (humidity_progression_id) REFERENCES humidity_progression(id),
    CONSTRAINT single_row CHECK (id = 1)
);
```

#### 3. `farm_config`
Remove `humidity_discount` column, add reference to progression.

```sql
CREATE TABLE IF NOT EXISTS farm_config (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    farm_id INTEGER NOT NULL UNIQUE,
    name TEXT NOT NULL,
    -- REMOVED: humidity_discount NUMERIC(6, 3) DEFAULT 1.15,
    humidity_progression_id INTEGER,  -- NEW: reference to progression
    storage_name TEXT NOT NULL,
    FOREIGN KEY (farm_id) REFERENCES farm(id),
    FOREIGN KEY (humidity_progression_id) REFERENCES humidity_progression(id)
);
```

### Database Initialization Functions

Add new initialization functions in `database.go`:

```go
func initHumidityProgression(c *pgx.Conn) {
    // Create humidity_progression table
    stmt, err := c.Prepare(context.Background(), "init humidity progression table", `
        CREATE TABLE IF NOT EXISTS humidity_progression (
            id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
            name TEXT NOT NULL,
            farm_id INTEGER,
            is_system_default BOOLEAN DEFAULT FALSE,
            created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (farm_id) REFERENCES farm(id),
            CONSTRAINT single_system_default CHECK (
                (is_system_default = TRUE AND farm_id IS NULL) OR 
                (is_system_default = FALSE)
            )
        );
    `)
    handleStmtExec(c, stmt, err, "create humidity_progression")

    // Create humidity_progression_tier table
    stmt, err = c.Prepare(context.Background(), "init humidity progression tier table", `
        CREATE TABLE IF NOT EXISTS humidity_progression_tier (
            id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
            progression_id INTEGER NOT NULL,
            threshold_humidity NUMERIC(5, 2) NOT NULL,
            discount_value NUMERIC(5, 2) NOT NULL,
            FOREIGN KEY (progression_id) REFERENCES humidity_progression(id) ON DELETE CASCADE,
            UNIQUE(progression_id, threshold_humidity)
        );
    `)
    handleStmtExec(c, stmt, err, "create humidity_progression_tier")
}

func initDefaultHumidityProgression(c *pgx.Conn) {
    // Insert system default progression if not exists
    var count int
    err := c.QueryRow(context.Background(), 
        "SELECT COUNT(*) FROM humidity_progression WHERE is_system_default = TRUE").Scan(&count)
    
    if err == nil && count == 0 {
        // Insert system default progression
        _, insertErr := c.Exec(context.Background(), `
            INSERT INTO humidity_progression (name, farm_id, is_system_default) 
            VALUES ('System Default', NULL, TRUE)
        `)
        if insertErr != nil {
            fmt.Printf("error inserting system default humidity progression: %v\n", insertErr.Error())
            return
        }

        // Insert default tiers
        _, insertErr = c.Exec(context.Background(), `
            INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value) 
            SELECT id, 14, 1.7 FROM humidity_progression WHERE is_system_default = TRUE;
            INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value) 
            SELECT id, 16, 1.8 FROM humidity_progression WHERE is_system_default = TRUE;
            INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value) 
            SELECT id, 18, 2.0 FROM humidity_progression WHERE is_system_default = TRUE;
            INSERT INTO humidity_progression_tier (progression_id, threshold_humidity, discount_value) 
            SELECT id, 20, 2.2 FROM humidity_progression WHERE is_system_default = TRUE;
        `)
        if insertErr != nil {
            fmt.Printf("error inserting default humidity progression tiers: %v\n", insertErr.Error())
        }
    }
}
```

Update `InitDb()` to call these new functions.

## Calculator Changes

The calculator must remain pure with no dependencies. **No changes to the calculation logic itself** - it will continue to receive a single `HumidityModifier` value.

The change is in how that value is determined:
- **Before**: Service layer fetches single `humidity_discount` from config
- **After**: Service layer fetches progression, finds appropriate tier based on humidity value, passes that tier's `discount_value` as `HumidityModifier`

### EntryCalculationInput (No Change)
```go
type EntryCalculationInput struct {
    GrossWeight        decimal.Decimal
    Tare               decimal.Decimal
    Humidity           *decimal.Decimal
    Damage             *decimal.Decimal
    Impurity           *decimal.Decimal
    HumidityModifier   *decimal.Decimal  // Still receives single value
    StorageTaxModifier *decimal.Decimal
}
```

### Calculator Functions (No Change)
All calculator functions remain unchanged:
- `CalculateEntry()`
- `CalculateDiscounts()`
- `DiscountHumidity()`
- `CalculateDeparture()`

## Model Layer Changes

### New Model: `humidity_progression_model`

Create `model/humidity_progression_model/model.go`:

```go
package humidity_progression_model

import (
    "context"
    "errors"
    
    entity_public "armazenda/entity/public"
    model_error "armazenda/model/error"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/shopspring/decimal"
)

type HumidityProgressionModel struct {
    pool *pgxpool.Pool
}

var humidityProgressionModelImpl *HumidityProgressionModel

func InitHumidityProgressionModel(pool *pgxpool.Pool) (*HumidityProgressionModel, error) {
    if pool == nil {
        return nil, errors.New("pool cant be null")
    }
    if humidityProgressionModelImpl == nil {
        humidityProgressionModelImpl = &HumidityProgressionModel{pool: pool}
    }
    return humidityProgressionModelImpl, nil
}

func GetHumidityProgressionModel() *HumidityProgressionModel {
    if humidityProgressionModelImpl == nil {
        panic("humidity progression model hasn't been initialized")
    }
    return humidityProgressionModelImpl
}

// GetProgression retrieves a progression by ID with all its tiers
func (hm *HumidityProgressionModel) GetProgression(id uint32) (entity_public.HumidityProgression, *model_error.ModelError) {
    // Query progression and tiers
    // Return struct with ordered tiers
}

// GetDiscountForHumidity finds the appropriate discount value for a given humidity
// Returns the discount_value from the tier with highest threshold <= humidity
func (hm *HumidityProgressionModel) GetDiscountForHumidity(progressionId *uint32, humidity decimal.Decimal) (decimal.Decimal, *model_error.ModelError) {
    // If progressionId is nil, use system default
    // Query: SELECT discount_value FROM humidity_progression_tier 
    //        WHERE progression_id = $1 AND threshold_humidity <= $2
    //        ORDER BY threshold_humidity DESC LIMIT 1
    // If no tier found (humidity < 14), return 0 or default
}

// GetSystemDefaultProgression returns the system default progression ID
func (hm *HumidityProgressionModel) GetSystemDefaultProgression() (uint32, *model_error.ModelError) {
    // Query: SELECT id FROM humidity_progression WHERE is_system_default = TRUE
}

// CRUD operations for progressions
func (hm *HumidityProgressionModel) AddProgression(name string, farmId uint32, tiers []entity_public.HumidityProgressionTier) (uint32, *model_error.ModelError) {
    // Insert progression, then insert all tiers
    // Validate: min 1 tier, max 30 tiers
}

func (hm *HumidityProgressionModel) UpdateProgression(id uint32, name string, tiers []entity_public.HumidityProgressionTier) *model_error.ModelError {
    // Update progression name, delete old tiers, insert new tiers
    // Validate: min 1 tier, max 30 tiers
}

func (hm *HumidityProgressionModel) DeleteProgression(id uint32) *model_error.ModelError {
    // Delete progression (cascades to tiers)
    // Prevent deletion of system default
}

func (hm *HumidityProgressionModel) ListProgressions(farmId uint32) ([]entity_public.HumidityProgression, *model_error.ModelError) {
    // List all progressions for a farm (including system default)
}
```

### Modified Model: `person_model`

Update `GetHumidityDiscount()` to use progression lookup:

```go
func (bm *PersonModel) GetHumidityDiscount(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, *model_error.ModelError) {
    hpm := humidity_progression_model.GetHumidityProgressionModel()
    
    var progressionId *uint32
    
    if person != nil {
        // Try to get person's progression
        err := bm.pool.QueryRow(context.Background(), `
            SELECT humidity_progression_id FROM person_config WHERE person_id = @person
        `, pgx.NamedArgs{"person": person}).Scan(&progressionId)
        
        if err != nil && !errors.Is(err, pgx.ErrNoRows) {
            return decimal.Zero, &model_error.ModelError{Message: err.Error()}
        }
        
        // If no person progression, try default_person_config
        if progressionId == nil {
            err = bm.pool.QueryRow(context.Background(), `
                SELECT humidity_progression_id FROM default_person_config WHERE id = 1
            `).Scan(&progressionId)
            
            if err != nil {
                // Fall back to system default
                sysDefault, err := hpm.GetSystemDefaultProgression()
                if err != nil {
                    return decimal.Zero, err
                }
                progressionId = &sysDefault
            }
        }
    } else {
        // No origin (Própria) - use farm's progression
        err := bm.pool.QueryRow(context.Background(), `
            SELECT humidity_progression_id FROM farm_config WHERE farm_id = @farm
        `, pgx.NamedArgs{"farm": farm}).Scan(&progressionId)
        
        if err != nil || progressionId == nil {
            // Fall back to system default
            sysDefault, err := hpm.GetSystemDefaultProgression()
            if err != nil {
                return decimal.Zero, err
            }
            progressionId = &sysDefault
        }
    }
    
    // Get the discount value for this humidity from the progression
    return hpm.GetDiscountForHumidity(progressionId, humidity)
}
```

## Service Layer Changes

### Modified Service: `entry_service`

Update `AddEntry()` to pass humidity value to `GetHumidityDiscount()`:

```go
func AddEntry(ge entity_public.Entry, em EntryModelInterface, pm PersonModelInterface, prod_m ProductModelInterface, cm CropModelInterface) (entity_public.DisplayEntry, entity_public.Toast) {
    // ... existing code for storage tax ...
    
    var discountModifier decimal.Decimal
    if ge.Humidity != nil && ge.Humidity.GreaterThan(calculator.HumidityThreshold) {
        // Pass humidity value to GetHumidityDiscount for tier lookup
        discountModifierTmp, humErr := pm.GetHumidityDiscount(ge.Origin, ge.Farm, *ge.Humidity)
        if humErr != nil {
            return entity_public.DisplayEntry{}, entity_public.GetErrorToast("Falha ao calcular desconto de humidade", "")
        }
        discountModifier = discountModifierTmp
    }
    
    // Rest of function unchanged...
    result := calculator.CalculateEntry(calculator.EntryCalculationInput{
        GrossWeight:        ge.GrossWeight,
        Tare:               ge.Tare,
        Humidity:           ge.Humidity,
        Damage:             ge.Damage,
        Impurity:           ge.Impurity,
        HumidityModifier:   &discountModifier,
        StorageTaxModifier: &storageTaxModifier,
    })
    // ... rest unchanged
}
```

Similar updates needed for `UpdateEntry()` and departure functions.

## Entity Changes

### New Entity: `entity/public/humidity_progression.go`

```go
package entity_public

import "github.com/shopspring/decimal"

type HumidityProgression struct {
    Id               uint32
    Name             string
    FarmId           *uint32  // NULL for system default
    IsSystemDefault  bool
    Tiers            []HumidityProgressionTier
}

type HumidityProgressionTier struct {
    Id                 uint32
    ProgressionId    uint32
    ThresholdHumidity  decimal.Decimal
    DiscountValue      decimal.Decimal
}
```

### Modified Entity: `entity/public/person.go`

Update `PersonConfig` struct:

```go
type PersonConfig struct {
    HumidityProgressionId    *uint32  // NEW: replaces HumidityDiscount
    EntrySoyDiscount         *decimal.Decimal
    EntryCornDiscount        *decimal.Decimal
}
```

## Router Layer

### New Router: `humidity_progression_router`

Create `router/humidity_progression_router/router.go` with endpoints:
- `GET /api/humidity-progressions` - List all progressions for user's farm
- `GET /api/humidity-progressions/:id` - Get single progression with tiers
- `POST /api/humidity-progressions` - Create new progression (min 1, max 30 tiers)
- `PUT /api/humidity-progressions/:id` - Update progression (min 1, max 30 tiers)
- `DELETE /api/humidity-progressions/:id` - Delete progression (cannot delete system default)

## Test Updates

### Calculator Tests (`pkg/calculator/calculator_test.go`)

**No changes needed** - Calculator still receives single `HumidityModifier` value.

However, add new test cases for progressive scenarios:

```go
{
    name: "Success - Entry with high humidity (19%) using progressive discount 2.0",
    input: calculator.EntryCalculationInput{
        GrossWeight:      decimal.NewFromFloat(50.000),
        Tare:             decimal.NewFromFloat(25.000),
        Humidity:         decimalPtr(decimal.NewFromInt(19)),
        HumidityModifier: decimalPtr(decimal.NewFromFloat(2.0)), // 2.0 for humidity >= 18
    },
    expectedNetWeight: decimal.NewFromFloat(23.0), // 25 - (25 * 5 * 2.0 / 100) = 25 - 2.5 = 22.5
    expectedValid:     true,
},
```

### Service Tests (`service/entry_service/test/service_test.go`)

Update mock interfaces to support new `GetHumidityDiscount` signature:

```go
type MockPersonModel struct {
    // Update signature to include humidity parameter
    GetHumidityDiscountFunc      func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, *model_error.ModelError)
    GetHumidityDiscountCalled      bool
}

func (m *MockPersonModel) GetHumidityDiscount(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, *model_error.ModelError) {
    m.GetHumidityDiscountCalled = true
    if m.GetHumidityDiscountFunc != nil {
        return m.GetHumidityDiscountFunc(person, farm, humidity)
    }
    return decimal.Zero, nil
}
```

Update all test cases to use new signature with humidity parameter:

```go
{
    name: "Success - Entry with exceeding humidity limit by 2%",
    entry: func() entity_public.Entry {
        entry := CreateBasicTestEntry()
        entry.GrossWeight = decimal.NewFromFloat(50.000)
        entry.Tare = decimal.NewFromFloat(25.000)
        hum := decimal.NewFromInt(16)
        entry.Humidity = &hum
        return entry
    }(),
    expectedNetWeight: decimal.NewFromFloat(24.425),
    setupMocks: func(em *MockEntryModel, pm *MockPersonModel, prodM *MockProductModel, cm *MockCropModel) {
        // Mock returns different values based on humidity
        pm.GetHumidityDiscountFunc = func(person *uint32, farm uint32, humidity decimal.Decimal) (decimal.Decimal, *model_error.ModelError) {
            // Return 1.7 for humidity 16 (from default progression tier)
            return decimal.NewFromFloat(1.7), nil
        }
        // ... rest of setup
    },
}
```

### New Model Tests

Create `model/humidity_progression_model/model_test.go` with tests for:
- `GetDiscountForHumidity()` - various humidity values and progression scenarios
- `GetSystemDefaultProgression()` - returns correct ID
- CRUD operations
- Validation: min 1 tier, max 30 tiers
- Fallback chain behavior

## Implementation Order

1. **Database schema** - Add new tables, modify existing tables (drop old columns immediately)
2. **Entity layer** - Create new structs, modify existing
3. **Model layer** - Create progression model, modify person/farm models
4. **Calculator tests** - Add new test cases for progressive scenarios
5. **Service layer** - Update entry/departure services
6. **Service tests** - Update mocks and test cases
7. **Router layer** - Create progression router
8. **Frontend** - Will be addressed in the next step

## Key Points

### Validation Rules
- Minimum 1 tier per progression
- Maximum 30 tiers per progression
- Cannot delete system default progression
- Threshold values must be >= 14 (base threshold)

### Real-time Calculation
The client-side will need to be updated to support real-time discount calculation with the new progressive system. This will be addressed in the next design step.

### Fallback Chain
1. Person has specific progression → use it
2. Person has no progression → use default_person_config's progression
3. default_person_config has no progression → use system default
4. For entries without origin (Própria):
   - Use farm_config's progression
   - If farm has no progression → use system default

### Data Integrity
- Existing entries continue to work - they store the actual `humidity_discount_modifier` used
- The `entry_analysis.humidity_discount_modifier` column is unchanged
- New entries will use the progressive lookup but store the computed value the same way
