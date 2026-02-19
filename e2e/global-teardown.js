/**
 * Playwright Global Teardown
 * 
 * Stops the PostgreSQL test database and removes volumes.
 */

const { execSync } = require('child_process');
const path = require('path');

async function globalTeardown() {
  const composeFile = path.join(__dirname, 'docker-compose.test.yml');
  
  console.log('\n🧹 Stopping test database...');
  
  try {
    execSync(`docker compose -f "${composeFile}" down -v`, {
      stdio: 'inherit',
    });
    console.log('✅ Test database stopped\n');
  } catch (error) {
    console.error('⚠️ Failed to stop test database cleanly');
    // Don't throw - we don't want to fail the tests because cleanup failed
  }
}

module.exports = globalTeardown;
