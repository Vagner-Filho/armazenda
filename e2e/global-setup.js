/**
 * Playwright Global Setup
 * 
 * 1. Starts PostgreSQL test database via Docker Compose
 * 2. Waits for database to be ready (healthcheck)
 * 3. Seeds test data from fixtures/test-user.sql
 */

const { execSync } = require('child_process');
const path = require('path');

const DB_CONFIG = {
  host: 'localhost',
  port: '5433',
  user: 'test',
  password: 'test',
  database: 'armazenda_test',
};

async function globalSetup() {
  const composeFile = path.join(__dirname, 'docker-compose.test.yml');
  
  console.log('\n📦 Starting test database...');
  
  try {
    execSync(`docker compose -f "${composeFile}" up -d --wait`, {
      stdio: 'inherit',
    });
  } catch (error) {
    console.error('❌ Failed to start test database');
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
    // Attempt to clean up
    execSync(`docker compose -f "${composeFile}" down -v`, { stdio: 'inherit' });
    throw error;
  }
  
  console.log('\n✅ Test database ready!\n');
}

module.exports = globalSetup;
