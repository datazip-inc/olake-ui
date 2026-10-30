import { test as base } from "@playwright/test"

import { CreateDestinationPage } from "../pages/CreateDestinationPage"
import { CreateJobPage } from "../pages/CreateJobPage"
import { CreateSourcePage } from "../pages/CreateSourcePage"
import { DestinationsPage } from "../pages/DestinationsPage"
import { JobsPage } from "../pages/JobsPage"
import { LoginPage } from "../pages/LoginPage"
import { SourcesPage } from "../pages/SourcesPage"

/**
 * Base fixture providing all page objects
 */
export type BaseFixtures = {
	loginPage: LoginPage
	sourcesPage: SourcesPage
	createSourcePage: CreateSourcePage
	destinationsPage: DestinationsPage
	createDestinationPage: CreateDestinationPage
	jobsPage: JobsPage
	createJobPage: CreateJobPage
}

export const test = base.extend<BaseFixtures>({
	loginPage: async ({ page }, use) => {
		await use(new LoginPage(page))
	},

	sourcesPage: async ({ page }, use) => {
		await use(new SourcesPage(page))
	},

	createSourcePage: async ({ page }, use) => {
		await use(new CreateSourcePage(page))
	},

	destinationsPage: async ({ page }, use) => {
		await use(new DestinationsPage(page))
	},

	createDestinationPage: async ({ page }, use) => {
		await use(new CreateDestinationPage(page))
	},

	jobsPage: async ({ page }, use) => {
		await use(new JobsPage(page))
	},

	createJobPage: async ({ page }, use) => {
		await use(new CreateJobPage(page))
	},
})

export { expect } from "@playwright/test"
