import { Page, Locator, expect } from "@playwright/test"

export class CreateSourcePage {
	readonly page: Page
	readonly sourceNameInput: Locator
	readonly connectorSelect: Locator
	readonly versionSelect: Locator
	readonly hostsInput: Locator
	readonly postgresHostsInput: Locator
	readonly addHostButton: Locator
	readonly databaseInput: Locator
	readonly usernameInput: Locator
	readonly passwordInput: Locator
	readonly sslToggle: Locator
	readonly createButton: Locator
	readonly cancelButton: Locator
	readonly backToSourcesLink: Locator
	readonly pageTitle: Locator
	readonly testConnectionButton: Locator
	readonly setupTypeNew: Locator
	readonly setupTypeExisting: Locator
	readonly port: Locator

	constructor(page: Page) {
		this.page = page
		this.sourceNameInput = page.getByPlaceholder(
			"Enter the name of your source",
		)
		this.connectorSelect = page.locator(".ant-select").first()
		this.versionSelect = page.locator(".ant-select").nth(1)
		this.hostsInput = page.getByRole("textbox", { name: "Hosts-1*" })
		this.postgresHostsInput = page.getByRole("textbox", {
			name: "Postgres Host *",
		})
		this.databaseInput = page.getByRole("textbox", { name: "Database Name *" })
		this.usernameInput = page.getByRole("textbox", { name: "Username *" })
		this.port = page.getByRole("spinbutton", { name: "Postgres Port *" })
		this.passwordInput = page.getByRole("textbox", { name: "Password *" })
		this.createButton = page.getByRole("button", { name: "Create" })
		this.cancelButton = page.getByRole("button", { name: "Cancel" })
		this.backToSourcesLink = page.getByRole("link").first()
		this.pageTitle = page.locator("text=Create source")
		this.testConnectionButton = page.getByRole("button", {
			name: "Test Connection",
		})
		this.setupTypeNew = page.getByText("Set up a new source")
		this.setupTypeExisting = page.getByText("Use an existing source")
	}

	async goto() {
		await this.page.goto("/sources/new")
	}

	async expectCreateSourcePageVisible() {
		await expect(this.pageTitle).toBeVisible()
		await expect(this.sourceNameInput).toBeVisible()
		await expect(this.createButton).toBeVisible()
	}

	async fillSourceName(name: string) {
		await this.sourceNameInput.click()
		await this.sourceNameInput.fill(name)
	}

	async selectConnector(connector: string) {
		await this.connectorSelect.click()
		await this.page.getByText(connector).click()
	}

	async addHost(host: string, type: "mongodb" | "postgres") {
		switch (type) {
			case "mongodb":
				await this.hostsInput.click()
				await this.hostsInput.fill(host)
				break
			case "postgres":
				await this.postgresHostsInput.click()
				await this.postgresHostsInput.fill(host)
				break
		}
	}

	async fillDatabaseName(database: string) {
		await this.databaseInput.click()
		await this.databaseInput.fill(database)
	}

	async fillUsername(username: string) {
		await this.usernameInput.click()
		await this.usernameInput.fill(username)
	}

	async fillPort(port: string) {
		await this.port.click()
		await this.port.fill(port)
	}

	async fillCredentials(username: string, password: string) {
		await this.usernameInput.click()
		await this.usernameInput.fill(username)
		await this.passwordInput.click()
		await this.passwordInput.fill(password)
	}

	async toggleSSL() {
		await this.sslToggle.click()
	}

	async clickCreate() {
		await this.createButton.click()
	}

	async clickCancel() {
		await this.cancelButton.click()
	}

	async goBackToSources() {
		await this.backToSourcesLink.click()
	}

	async selectSetupType(type: "new" | "existing") {
		if (type === "new") {
			await this.setupTypeNew.click()
		} else {
			await this.setupTypeExisting.click()
		}
	}

	async expectValidationError(message: string) {
		await expect(this.page.locator(`text=${message}`)).toBeVisible()
	}

	async expectTestConnectionModal() {
		// Wait for the modal with "Testing your connection" text to appear
		await this.page.waitForSelector("text=Testing your connection", {
			state: "visible",
		})

		// Check if the text exists (more reliable than checking modal visibility)
		await expect(this.page.getByText("Testing your connection")).toHaveCount(1)
	}

	async expectSuccessModal() {
		await this.page.waitForSelector("text=Connection successful", {
			state: "visible",
		})
		await expect(this.page.getByText("Connection successful")).toBeVisible()
	}

	async expectEntitySavedModal() {
		await this.page.waitForSelector(
			"text=Source is connected and saved successfully",
			{
				state: "visible",
			},
		)
		await expect(
			this.page.getByText("Source is connected and saved successfully"),
		).toBeVisible()
		await this.page.getByRole("button", { name: "Sources" }).click()
	}

	async selectPostgresFillPostgresCreds(data: {
		name: string
		host: string
		database: string
		username: string
		password: string
		port: string
	}) {
		await this.page.getByText("MongoDB").click()
		await this.page
			.locator("div")
			.filter({ hasText: /^Postgres$/ })
			.nth(3)
			.click()
		await this.fillSourceName(data.name)
		await this.addHost(data.host, "postgres")
		await this.fillDatabaseName(data.database)
		await this.fillPort(data.port)
		await this.fillCredentials(data.username, data.password)
	}

	async fillMongoDBForm(data: {
		name: string
		host: string
		database: string
		username: string
		password: string
		useSSL?: boolean
	}) {
		await this.fillSourceName(data.name)
		await this.addHost(data.host, "mongodb")
		await this.fillDatabaseName(data.database)
		await this.fillCredentials(data.username, data.password)

		// if (data.useSSL) {
		// 	await this.toggleSSL()
		// }
	}
}
