/**
 * Playwright Global Setup
 * 
 * 1. Starts PostgreSQL test database via Docker Compose
 * 2. Waits for database to be ready (healthcheck)
 * 3. Starts the Go application (creates schema)
 * 4. Waits for Go app to be ready
 * 5. Seeds test data from fixtures/test-user.sql
 */

const { execSync, spawn } = require('child_process');
const { writeFileSync, unlinkSync, existsSync } = require('fs');
const http = require('http');
const path = require('path');

const DB_CONFIG = {
  host: 'localhost',
  port: '5433',
  user: 'test',
  password: 'test',
  database: 'armazenda_test',
};

const APP_URL = process.env.BASE_URL || 'http://localhost:8100';
const APP_TIMEOUT = 30000;
const PID_FILE = path.join(__dirname, '.go-server.pid');

function waitForApp(url, timeout) {
  return new Promise((resolve, reject) => {
    const startTime = Date.now();

    function check() {
      if (Date.now() - startTime > timeout) {
        reject(new Error(`Go app did not start within ${timeout}ms`));
        return;
      }

      const req = http.get(url, (res) => {
        if (res.statusCode >= 200 && res.statusCode < 500) {
          resolve();
        } else {
          setTimeout(check, 500);
        }
      });

      req.on('error', () => {
        setTimeout(check, 500);
      });

      req.setTimeout(2000, () => {
        req.destroy();
        setTimeout(check, 500);
      });
    }

    check();
  });
}

function waitForTable(dbConfig, timeout) {
  return new Promise((resolve, reject) => {
    const startTime = Date.now();
    const checkQuery = `SELECT 1 FROM pg_tables WHERE schemaname = 'public' AND tablename = 'farm' LIMIT 1;`;

    function check() {
      if (Date.now() - startTime > timeout) {
        reject(new Error(`Database schema was not ready within ${timeout}ms`));
        return;
      }

      try {
        execSync(
          `PGPASSWORD=${dbConfig.password} psql ` +
          `-h ${dbConfig.host} ` +
          `-p ${dbConfig.port} ` +
          `-U ${dbConfig.user} ` +
          `-d ${dbConfig.database} ` +
          `-c "${checkQuery}"`,
          { stdio: 'pipe' }
        );
        resolve();
      } catch {
        setTimeout(check, 500);
      }
    }

    check();
  });
}

async function globalSetup() {
  const composeFile = path.join(__dirname, 'docker-compose.test.yml');
  const projectRoot = path.join(__dirname, '../..');
  const binaryPath = path.join(__dirname, '.test-server');

  console.log('\n📦 Starting test database...');

  try {
    execSync(`docker compose -f "${composeFile}" up -d --wait`, {
      stdio: 'inherit',
    });
  } catch (error) {
    console.error('❌ Failed to start test database');
    throw error;
  }

  console.log('\n🔨 Building Go application...');

  try {
    execSync(`go build -o "${binaryPath}" .`, {
      cwd: projectRoot,
      stdio: 'inherit',
    });
  } catch (error) {
    console.error('❌ Failed to build Go application');
    execSync(`docker compose -f "${composeFile}" down -v`, { stdio: 'inherit' });
    throw error;
  }

  console.log('\n🚀 Starting Go application...');

  const goApp = spawn(binaryPath, [], {
    cwd: projectRoot,
    env: {
      ...process.env,
      DB_HOST: DB_CONFIG.host,
      DB_PORT: DB_CONFIG.port,
      DB_USER: DB_CONFIG.user,
      DB_PASS: DB_CONFIG.password,
      DB_NAME: DB_CONFIG.database,
    },
    stdio: ['ignore', 'inherit', 'inherit'],
    detached: true,
  });

  writeFileSync(PID_FILE, goApp.pid.toString());

  goApp.on('error', (error) => {
    console.error('❌ Failed to start Go app:', error.message);
  });

  try {
    await waitForApp(APP_URL, APP_TIMEOUT);
    console.log('✅ Go application responding');
  } catch (error) {
    console.error('❌ Go app failed to start:', error.message);
    try { process.kill(-goApp.pid); } catch (e) { }
    if (existsSync(PID_FILE)) unlinkSync(PID_FILE);
    if (existsSync(binaryPath)) unlinkSync(binaryPath);
    execSync(`docker compose -f "${composeFile}" down -v`, { stdio: 'inherit' });
    throw error;
  }

  console.log('\n⏳ Waiting for database schema...');

  try {
    await waitForTable(DB_CONFIG, APP_TIMEOUT);
    console.log('✅ Database schema ready');
  } catch (error) {
    console.error('❌ Database schema not ready:', error.message);
    try { process.kill(-goApp.pid); } catch (e) { }
    if (existsSync(PID_FILE)) unlinkSync(PID_FILE);
    if (existsSync(binaryPath)) unlinkSync(binaryPath);
    execSync(`docker compose -f "${composeFile}" down -v`, { stdio: 'inherit' });
    throw error;
  }

  console.log('\n🌱 Seeding test data...');

  const seedFile = path.join(__dirname, 'fixtures', 'test-user.sql');

  try {
    execSync(
      `PGPASSWORD=${DB_CONFIG.password} psql ` +
      `-h ${DB_CONFIG.host} ` +
      `-p ${DB_CONFIG.port} ` +
      `-U ${DB_CONFIG.user} ` +
      `-d ${DB_CONFIG.database} ` +
      `-f "${seedFile}"`,
      { stdio: 'inherit' }
    );
  } catch (error) {
    console.error('❌ Failed to seed test data');
    try { process.kill(-goApp.pid); } catch (e) { }
    if (existsSync(PID_FILE)) unlinkSync(PID_FILE);
    if (existsSync(binaryPath)) unlinkSync(binaryPath);
    execSync(`docker compose -f "${composeFile}" down -v`, { stdio: 'inherit' });
    throw error;
  }

  console.log('\n✅ Test environment ready!\n');
}

module.exports = globalSetup;
