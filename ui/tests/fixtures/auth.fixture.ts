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

type TestFixtures = {
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

export const test = base.extend<TestFixtures>({
	loginPage: async ({ page }, use) => {
		const loginPage = new LoginPage(page)
		await use(loginPage)
	},

	sourcesPage: async ({ page }, use) => {
		const sourcesPage = new SourcesPage(page)
		await use(sourcesPage)
	},

	createSourcePage: async ({ page }, use) => {
		const createSourcePage = new CreateSourcePage(page)
		await use(createSourcePage)
	},

	editSourcePage: async ({ page }, use) => {
		const editSourcePage = new EditSourcePage(page)
		await use(editSourcePage)
	},

	destinationsPage: async ({ page }, use) => {
		const destinationsPage = new DestinationsPage(page)
		await use(destinationsPage)
	},

	createDestinationPage: async ({ page }, use) => {
		const createDestinationPage = new CreateDestinationPage(page)
		await use(createDestinationPage)
	},

	editDestinationPage: async ({ page }, use) => {
		const editDestinationPage = new EditDestinationPage(page)
		await use(editDestinationPage)
	},

	jobsPage: async ({ page }, use) => {
		const jobsPage = new JobsPage(page)
		await use(jobsPage)
	},

	createJobPage: async ({ page }, use) => {
		const createJobPage = new CreateJobPage(page)
		await use(createJobPage)
	},
})

export { expect } from "@playwright/test"
