import { test as base } from "@playwright/test"
import { LoginPage } from "../pages/LoginPage"
import { SourcesPage } from "../pages/SourcesPage"
import { CreateSourcePage } from "../pages/CreateSourcePage"
import { EditSourcePage } from "../pages/EditSourcePage"
import { DestinationsPage } from "../pages/DestinationsPage"
import { CreateDestinationPage } from "../pages/CreateDestinationPage"
import { EditDestinationPage } from "../pages/EditDestinationPage"
import { JobsPage } from "../pages/JobsPage"
import { CreateJobPage } from "../pages/CreateJobPage"

/**
 * Base fixture providing all page objects
 */
export type BaseFixtures = {
	loginPage: LoginPage
	sourcesPage: SourcesPage
	createSourcePage: CreateSourcePage
	editSourcePage: EditSourcePage
	destinationsPage: DestinationsPage
	createDestinationPage: CreateDestinationPage
	editDestinationPage: EditDestinationPage
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

	editSourcePage: async ({ page }, use) => {
		await use(new EditSourcePage(page))
	},

	destinationsPage: async ({ page }, use) => {
		await use(new DestinationsPage(page))
	},

	createDestinationPage: async ({ page }, use) => {
		await use(new CreateDestinationPage(page))
	},

	editDestinationPage: async ({ page }, use) => {
		await use(new EditDestinationPage(page))
	},

	jobsPage: async ({ page }, use) => {
		await use(new JobsPage(page))
	},

	createJobPage: async ({ page }, use) => {
		await use(new CreateJobPage(page))
	},
})

export { expect } from "@playwright/test"
