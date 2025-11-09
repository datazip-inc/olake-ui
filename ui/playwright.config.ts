import { defineConfig, devices } from "@playwright/test"

//TODO: Decide on Timeouts and adjust them as needed

// Constants
export const TIMEOUTS = {
	LONG: 5 * 60 * 1000, // 5 minutes
	SHORT: 10 * 1000, // 10 seconds
} as const

export default defineConfig({
	testDir: "./tests",
	timeout: 10 * 60 * 1000, // 10 minutes global test timeout
	expect: {
		timeout: 5 * 60 * 1000, // 5 minutes for expect assertions
	},
	use: {
		baseURL: "http://localhost:8000",
		screenshot: "only-on-failure",
		actionTimeout: 5 * 60 * 1000, // 5 minutes for actions
		navigationTimeout: 30 * 1000, // 30 seconds for navigation
	},
	projects: [
		// Setup project - runs first to create authenticated state
		{
			name: "setup",
			testMatch: /.*\.setup\.ts/,
		},
		// Main test project - depends on setup
		{
			name: "chromium",
			use: { ...devices["Desktop Chrome"] },
			dependencies: ["setup"],
		},
	],
})
