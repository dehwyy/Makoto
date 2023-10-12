import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
	testDir: './tests/e2e',
	fullyParallel: true,
	/* Fail the build on CI if you accidentally left test.only in the source code. */
	forbidOnly: !!process.env.CI,
	/* Retry on CI only */
	retries: process.env.CI ? 2 : 0,
	/* Opt out of parallel tests on CI. */
	workers: process.env.CI ? 1 : undefined,
	timeout: 30 * 1000,
	reporter: 'html',
	expect: {
		timeout: 30 * 1000
	},
	use: {
		baseURL: 'http://localhost:3001',

		/* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
		trace: 'on-first-retry'
	},

	/* Configure projects for major browsers */
	projects: [
		{
			name: 'chromium',
			use: { ...devices['Desktop Chrome'] }
		},

		{
			name: 'firefox',
			use: { ...devices['Desktop Firefox'] }
		},

		{
			name: 'webkit',
			use: { ...devices['Desktop Safari'] }
		}
	],

	webServer: {
		command: 'npm run dev',
		timeout: 30 * 1000
	}
})
