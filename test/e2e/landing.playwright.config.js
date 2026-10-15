const { defineConfig, devices } = require('@playwright/test');

const port = process.env.LANDING_TEST_PORT || '18180';
const baseURL = `http://127.0.0.1:${port}`;

module.exports = defineConfig({
  testDir: __dirname,
  testMatch: 'landing.spec.js',
  fullyParallel: true,
  reporter: 'list',
  timeout: 30_000,
  use: {
    baseURL,
    trace: 'off',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },
  ],
  webServer: {
    command: `python3 -m http.server ${port} --bind 127.0.0.1 --directory ../../landing`,
    cwd: __dirname,
    url: `${baseURL}/`,
    reuseExistingServer: false,
    timeout: 10_000,
    stdout: 'ignore',
    stderr: 'pipe',
  },
});
