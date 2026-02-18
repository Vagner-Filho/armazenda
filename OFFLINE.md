# Armazenda Offline-First Architecture

## Overview

This document describes the offline-first architecture implemented in Armazenda, enabling users to work with entry, departure, and person management features without an internet connection.

## Architecture Components

### 1. WASM Calculator (`wasm/calculator/`)

**Purpose**: Provides exact same calculation logic on client and server

**Files**:
- `main.go` - WASM entry point, exports JS-callable functions
- `calc/calculator.go` - Pure calculation logic (no DB dependencies)
- `calc/calculator_test.go` - Unit tests

**Exported Functions**:
- `calculateEntry(entry, personConfig)` - Calculates net weight and all discounts
- `calculateDeparture(departure)` - Calculates departure net weight
- `calculateDiscounts(...)` - Calculates individual discount components
- `validateEntry(entry)` - Validates entry data

**Why WASM?**
- Exact same code runs on client and server
- Uses `shopspring/decimal` for precise calculations (banking-grade precision)
- All existing Go tests remain valid
- No risk of JS/Go logic divergence

### 2. PWA Infrastructure

**Service Worker** (`assets/sw.js`):
- Caches static assets (JS, CSS, WASM, templates)
- Intercepts HTMX requests when offline
- Queues mutating operations (POST/PUT/DELETE) for background sync
- Serves offline pages when network fails

**Manifest** (`assets/manifest.json`):
- PWA configuration
- Icons, theme colors, display mode
- Enables "Add to Home Screen" functionality

### 3. IndexedDB Layer (`assets/js/db/`)

**Database Schema**:
```javascript
{
  entries: { id, data, synced, modifiedAt, farm },
  departures: { id, data, synced, modifiedAt, farm },
  people: { id, data, synced, modifiedAt, farm },
  crops: { id, data, farm },
  fields: { id, data, farm },
  vehicles: { id, data, farm },
  pendingChanges: { id, operation, entity, data, timestamp, retries },
  syncMetadata: { key, value, updatedAt },
  templates: { name, html, cachedAt }
}
```

**Sync Engine** (`syncEngine.js`):
- Monitors online/offline status
- Uploads pending changes to server
- Downloads updates from server
- Handles conflict resolution (server wins)
- Retries failed operations with 1-minute delay

### 4. Offline Manager (`assets/js/offlineManager.js`)

Main coordinator that:
- Initializes all offline components
- Manages service worker registration
- Shows offline/sync indicators
- Handles HTMX requests when offline
- Integrates WASM calculator with forms

### 5. Template Renderer (`assets/js/templateRenderer.js`)

Renders Go HTML templates client-side using cached templates and IndexedDB data:
- Parses Go template syntax
- Supports `{{ .Field }}`, `{{ range }}`, `{{ if }}`
- Handles nested field access
- Caches templates from server

## Data Flow

### Online Operation (Normal)
1. User performs action (create/edit/delete)
2. HTMX sends request to server
3. Server processes and returns HTML
4. HTMX swaps DOM
5. Background: updates IndexedDB with server response

### Offline Operation
1. Intercept HTMX request in Service Worker
2. Queue change in `pendingChanges` store
3. Update IndexedDB immediately (optimistic UI)
4. Render response using cached template
5. Show "offline" indicator
6. Retry sync when connection restored (after 1 min delay)

### Sync When Online
1. Detect connection restored
2. Upload all pending changes
3. Server re-calculates and returns authoritative values
4. Update IndexedDB with server response
5. Notify user of any discrepancies
6. Download any updates from server

## Build Integration

### Air Configuration
Updated `.air.toml` to include WASM build:
```toml
cmd = "go build -o ./tmp/main . && GOOS=js GOARCH=wasm go build -o ./assets/wasm/calculator.wasm ./wasm/calculator"
```

This ensures the WASM module is rebuilt whenever Go files change.

### Required Files
- `assets/wasm/calculator.wasm` - Compiled WASM module
- `assets/wasm/wasm_exec.js` - Go WASM runtime helper (copied from Go installation)

## Pages with Offline Support

Updated pages include:
- `/templates/pages/romaneio.html` - Entry/Departure management
- `/templates/pages/person.html` - Person management
- `/templates/pages/login.html` - Service worker registration

Each page includes:
```html
<meta name="theme-color" content="#1e3a5f">
<link rel="manifest" href="/public/assets/manifest.json">
<script type="module" src="/public/assets/js/offlineManager.js"></script>
```

## Testing

### WASM Tests
```bash
go test ./wasm/calculator/calc/... -v
```

### Build Test
```bash
# Test main build
go build -o ./tmp/main .

# Test WASM build
GOOS=js GOARCH=wasm go build -o ./assets/wasm/calculator.wasm ./wasm/calculator
```

## Sync API Reference

### Authentication
All sync endpoints require a valid session cookie (`session_id`).

### Endpoints

#### GET /api/entries
Download entries modified since a specific timestamp.

**Query Parameters:**
- `since` (required) - ISO 8601 timestamp (e.g., "2024-01-15T10:30:00Z")
- `farm` (required) - Farm ID

**Response:**
```json
[
  {
    "id": 123,
    "field": 1,
    "crop": 2,
    "vehicle": 5,
    "cargoWeight": {
      "grossWeight": 1000.50,
      "tare": 100.00,
      "netWeight": 900.50
    },
    "analysis": {
      "humidity": 15.5,
      "damage": 7.0,
      "impurity": 1.5
    },
    "arrivalDate": "2024-01-15T10:30:00Z",
    "farm": 1,
    "origin": 5,
    "modifiedAt": "2024-01-15T10:30:00Z"
  }
]
```

#### GET /api/departures
Download departures modified since a specific timestamp.

**Query Parameters:** Same as /api/entries

**Response:** Same structure as entries with `departureDate` instead of `arrivalDate`

#### GET /api/people
Download people modified since a specific timestamp.

**Query Parameters:** Same as /api/entries

**Response:**
```json
[
  {
    "id": 1,
    "ie": "123456789",
    "farm": 1,
    "humidityDiscount": 1.15,
    "entrySoyDiscount": 1.50,
    "entryCornDiscount": 1.25,
    "modifiedAt": "2024-01-15T10:30:00Z"
  }
]
```

#### POST /api/sync/status
Check sync status and get counts of pending updates.

**Request Body:**
```json
{
  "farm": 1,
  "lastSync": "2024-01-15T10:30:00Z"
}
```

**Response:**
```json
{
  "hasUpdates": true,
  "entriesCount": 5,
  "departuresCount": 3,
  "peopleCount": 2,
  "serverTimestamp": "2024-01-15T11:00:00Z"
}
```

### Testing with cURL

```bash
# Get entries since timestamp
curl "http://localhost:8080/api/entries?since=2024-01-01T00:00:00Z&farm=1" \
  -H "Cookie: session_id=your_session_cookie"

# Check sync status
curl -X POST http://localhost:8080/api/sync/status \
  -H "Content-Type: application/json" \
  -H "Cookie: session_id=your_session_cookie" \
  -d '{"farm": 1, "lastSync": "2024-01-01T00:00:00Z"}'
```

### Error Responses

All endpoints return JSON error responses:
```json
{
  "error": "Description of what went wrong"
}
```

Common HTTP status codes:
- `200` - Success
- `400` - Bad Request (invalid timestamp)
- `401` - Unauthorized (no valid session)
- `403` - Forbidden (access to different farm)
- `500` - Internal Server Error

## Conflict Resolution

**Strategy**: Server Wins

1. Client sends operation with client timestamp
2. Server processes and calculates authoritative values
3. If server calculation differs from client:
   - Server stores both values in response
   - UI shows: "Server adjusted weight from X to Y"
4. Client updates IndexedDB with server values

This ensures data integrity while giving users accurate previews offline.

## Offline Limitations

1. **Initial Login Required**: User must be online for initial authentication
2. **Reference Data**: Crops, fields, vehicles are cached but won't update until online
3. **PDF Generation**: Works offline using cached templates
4. **File Uploads**: Not supported offline
5. **Session Duration**: Works until connection is re-established

## Future Enhancements

1. **Background Sync API**: Use native Background Sync for better reliability
2. **Conflict UI**: Better visualization of conflicts with user resolution
3. **Partial Sync**: Only sync recent data for large datasets
4. **Multi-Tab Sync**: Use BroadcastChannel for cross-tab synchronization
5. **Compression**: Compress IndexedDB data for large datasets

## Debugging

### Check Service Worker
```javascript
// In browser console
navigator.serviceWorker.ready.then(reg => console.log('SW ready', reg));
```

### Check IndexedDB
```javascript
// View all data
const db = await offlineManager.db.db;
const entries = await db.getAll('entries');
console.log(entries);
```

### Check Sync Status
```javascript
await offlineManager.getSyncStatus();
```

### Force Sync
```javascript
await offlineManager.forceSync();
```

## Security Considerations

1. **Data Privacy**: All data stored locally in browser
2. **HTTPS Required**: Service workers require HTTPS (except localhost)
3. **Session Management**: Sessions persist offline but validate on sync
4. **Data Integrity**: Server re-calculates all values on sync

## Browser Support

- **Chrome/Edge**: Full support
- **Firefox**: Full support
- **Safari**: Full support (iOS 11.3+)
- **Requirements**: Service Worker, IndexedDB, WebAssembly

## Performance

- **WASM Loading**: ~500KB (one-time load)
- **IndexedDB**: Depends on data size, handles thousands of records efficiently
- **Template Caching**: Reduces server requests when online
- **Sync**: Batched operations, minimal network overhead
