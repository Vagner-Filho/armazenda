/**
 * Playwright Global Teardown
 * 
 * 1. Stops the Go application
 * 2. Stops the PostgreSQL test database and removes volumes.
 */

const { execSync } = require('child_process');
const { readFileSync, unlinkSync, existsSync } = require('fs');
const path = require('path');

const PID_FILE = path.join(__dirname, '.go-server.pid');
const BINARY_FILE = path.join(__dirname, '.test-server');

async function globalTeardown() {
  const composeFile = path.join(__dirname, 'docker-compose.test.yml');

  console.log('\n🧹 Stopping Go application...');

  if (existsSync(PID_FILE)) {
    try {
      const pid = parseInt(readFileSync(PID_FILE, 'utf8').trim(), 10);
      process.kill(-pid);
      console.log('✅ Go application stopped');
    } catch (error) {
      console.warn('⚠️ Failed to stop Go application:', error.message);
    }
    try {
      unlinkSync(PID_FILE);
    } catch (e) { }
  }

  if (existsSync(BINARY_FILE)) {
    try {
      unlinkSync(BINARY_FILE);
    } catch (e) { }
  }

  console.log('\n🧹 Stopping test database...');

  try {
    execSync(`docker compose -f "${composeFile}" down -v`, {
      stdio: 'inherit',
    });
    console.log('✅ Test database stopped\n');
  } catch (error) {
    console.error('⚠️ Failed to stop test database cleanly');
  }
}

module.exports = globalTeardown;
