# E2E Tests for Armazenda

End-to-end tests using Playwright for the Armazenda grain warehouse management system.

## Setup

### Prerequisites

- Node.js (>=18)
- pnpm
- Docker (for test database)
- PostgreSQL client tools (for seeding, optional)

### Installation

```bash
cd e2e
pnpm install
```

Playwright browsers will be installed automatically during `pnpm install`.

## Test Database

Tests run against an isolated PostgreSQL database in Docker. The database is automatically started and stopped by Playwright's global setup/teardown hooks.

### Automatic (Recommended)

Just run the tests - the database starts automatically:

```bash
pnpm test
```

### Manual Control

If you need to manage the database manually:

```bash
pnpm run db:start    # Start test database
pnpm run db:seed     # Seed test data
pnpm run db:stop     # Stop and remove test database
```

### Test Database Configuration

| Setting | Value |
|---------|-------|
| Host | localhost |
| Port | 5433 |
| User | test |
| Password | test |
| Database | armazenda_test |

### PostgreSQL Version

The test database uses PostgreSQL 16. Keep this in sync with your production/development database version in `docker-compose.test.yml`.

## Test Admin Credentials

- **CPF**: `52998224725`
- **Password**: `TestAdmin123!`
- **Role**: admin
- **Email**: test-admin@armazenda.test

## Running Tests

### Run all tests
```bash
pnpm test
```

### Run tests in headed mode (see browser)
```bash
pnpm run test:headed
```

### Run tests in UI mode (interactive debugging)
```bash
pnpm run test:ui
```

### Run specific test file
```bash
npx playwright test tests/login.spec.js
```

### Run tests in specific browser
```bash
npx playwright test --project=chromium
```

### Run without Docker (use existing database)

If you want to run tests against an existing database instead of the Docker container:

```bash
BASE_URL=http://localhost:8100 \
WEB_SERVER_CMD="cd .. && go run ." \
pnpm test --no-deps
```

Note: You'll need to ensure the test admin user exists in your database.

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `BASE_URL` | `http://localhost:8100` | Application URL |
| `WEB_SERVER_CMD` | See config | Command to start the server |
| `CI` | - | Enable CI mode |

### Playwright Config

See `playwright.config.js` for browser configurations and test settings.

## Project Structure

```
e2e/
├── docker-compose.test.yml    # Test database definition
├── global-setup.js            # Start DB, seed data
├── global-teardown.js         # Stop DB
├── package.json               # Dependencies and scripts
├── playwright.config.js       # Playwright configuration
├── tests/
│   └── login.spec.js          # Login flow tests
├── fixtures/
│   └── test-user.sql          # Database seed
└── utils/
    └── auth.js                # Authentication helpers
```

## Writing New Tests

1. Create a new file in `tests/` with the `.spec.js` extension
2. Import the test utilities you need
3. Use the `login()` helper from `utils/auth.js` for authenticated tests

Example:

```javascript
const { test, expect } = require('@playwright/test');
const { login } = require('../utils/auth');

test('my new test', async ({ page, browserName }) => {
  await login(page, browserName);
  // Your test code here
});
```

Note: Always pass `browserName` to `login()` for WebKit compatibility.

## WebKit Compatibility

WebKit (Safari) has strict cookie handling and doesn't accept secure cookies over HTTP. The `login()` helper in `utils/auth.js` handles this automatically by:

1. For Chromium/Firefox: Standard form-based login
2. For WebKit: API-based login with manual cookie injection

## Troubleshooting

### Docker not running

```
Error: Cannot connect to the Docker daemon
```
Ensure Docker is running: `docker info`

### Port 5433 already in use

```
Error: port is already allocated
```
Stop any existing PostgreSQL on port 5433, or modify the port in `docker-compose.test.yml`.

### Browsers not installed

```bash
npx playwright install
```

### Database seed fails

If seeding fails, the container is automatically cleaned up. Check the SQL file for syntax errors:

```bash
pnpm run db:start
pnpm run db:seed
```

## Generating Tests

Use Playwright's code generator to create new tests:

```bash
pnpm run codegen
```

This will open a browser where you can interact with the app and Playwright will generate test code for your actions.
