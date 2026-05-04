# armazenda

## Database Naming Conventions

### Legal Person Name Display

The system uses different COALESCE patterns for legal person names depending on the context:

#### Pattern 1: Fantasy Name First (UI/Listings)
```sql
COALESCE(lp.fantasyname, lp.companyname)
```

**Usage**: Used in UI displays, dropdowns, listings, and general user interfaces.

**Files using this pattern**:
- `model/person_model/model.go` - Person dropdowns (`GetPeopleByFarm()`)
- `model/entry_model/model.go` - Entry listings (`GetEntries()`, `GetPaginatedEntries()`)
- `model/departure_model/model.go` - Departure listings (`GetDepartures()`, `GetAllDepartureDrafts()`)
- `model/armazenda_database/database.go` - Stored procedures for entry/departure display updates

**Rationale**: Users prefer seeing familiar fantasy names (business/trade names) in the interface for easier recognition.

#### Pattern 2: Company Name First (PDF/Official Documents)
```sql
COALESCE(lp.companyname, lp.fantasyname)
```

**Usage**: Used exclusively for PDF generation and official documents.

**Files using this pattern**:
- `model/entry_model/model.go` - PDF generation (`GetEntryForPdf()`)
- `model/departure_model/model.go` - PDF generation (`GetDepartureForPdf()` - both recipient and origin)

**Rationale**: Official documents should display the formal/legal company name rather than the trade name.

#### Pattern 3: Natural Person + Legal Person
```sql
COALESCE(np.name, lp.fantasyname, lp.companyname)
```

**Usage**: Used when querying across both natural and legal persons in the same context.

**Files using this pattern**:
- `model/armazenda_database/database.go` - Entry/departure draft stored procedures
- `model/entry_model/model.go` - Entry drafts (`GetEntryDraftsByFarm()`)
- `model/departure_model/model.go` - Departure listings with fallback (`GetPaginatedDepartures()`, `GetAllDepartureDrafts()`)

**Rationale**: Natural person names take precedence, followed by legal person's fantasy name, then company name. Some contexts also include `'Própria'` (Own) as a final fallback for internal/self-owned records.

### Field Naming Conventions

- **Lowercase**: Most SQL queries use lowercase field names (`fantasyname`, `companyname`)
- **CamelCase**: One exception exists in `model/departure_model/model.go` (line 103) where `fantasyName` and `companyName` are used

This distinction between UI and document contexts is intentional and should be maintained when modifying queries.
