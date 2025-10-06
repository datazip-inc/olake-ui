import { Page, Locator, expect } from "@playwright/test"
import { TIMEOUTS } from "../../playwright.config"

export class CreateDestinationPage {
	readonly page: Page
	readonly destinationNameInput: Locator
	readonly connectorSelect: Locator
	readonly catalogSelect: Locator
	readonly versionSelect: Locator
	readonly bucketNameInput: Locator
	readonly regionInput: Locator
	readonly pathInput: Locator
	readonly createButton: Locator
	readonly cancelButton: Locator
	readonly backToDestinationsLink: Locator
	readonly pageTitle: Locator
	readonly testConnectionButton: Locator
	readonly setupTypeNew: Locator
	readonly setupTypeExisting: Locator

	// Iceberg JDBC Catalog
	readonly icebergJdbcCatalogSelect: Locator
	readonly icebergJdbcUrlInput: Locator
	readonly icebergJdbcUsernameInput: Locator
	readonly icebergJdbcPasswordInput: Locator
	readonly icebergJdbcDatabaseInput: Locator
	readonly icebergJdbcS3EndpointInput: Locator
	readonly icebergJdbcS3AccessKeyInput: Locator
	readonly icebergJdbcS3SecretKeyInput: Locator
	readonly icebergJdbcS3RegionInput: Locator
	readonly icebergJdbcS3PathInput: Locator
	readonly icebergUsePathStyleForS3Switch: Locator
	readonly icebergJdbcUseSSLForS3Switch: Locator

	constructor(page: Page) {
		this.page = page
		this.destinationNameInput = page.getByPlaceholder(
			"Enter the name of your destination",
		)
		this.connectorSelect = page.locator(".ant-select").first()
		this.catalogSelect = page.locator(".ant-select").nth(1)
		this.versionSelect = page.locator(".ant-select").nth(2)
		this.bucketNameInput = page.getByRole("textbox", { name: "S3 Bucket *" })
		this.regionInput = page.getByRole("textbox", { name: "S3 Region *" })
		this.pathInput = page.getByRole("textbox", { name: "S3 Path *" })
		this.createButton = page.getByRole("button", { name: "Create" })
		this.cancelButton = page.getByRole("button", { name: "Cancel" })
		this.backToDestinationsLink = page.getByRole("link").first()
		this.pageTitle = page.locator("text=Create destination")
		this.testConnectionButton = page.getByRole("button", {
			name: "Test Connection",
		})
		this.setupTypeNew = page.getByText("Set up a new destination")
		this.setupTypeExisting = page.getByText("Use an existing destination")

		// Iceberg JDBC Catalog
		this.icebergJdbcUrlInput = page.getByRole("textbox", { name: "JDBC URL *" })
		this.icebergJdbcUsernameInput = page.getByRole("textbox", {
			name: "JDBC Username *",
		})
		this.icebergJdbcPasswordInput = page.getByRole("textbox", {
			name: "JDBC Password *",
		})
		this.icebergJdbcS3PathInput = page.getByRole("textbox", {
			name: "S3 Path *",
		})
		this.icebergJdbcS3EndpointInput = page.getByRole("textbox", {
			name: "S3 Endpoint",
		})
		this.icebergJdbcDatabaseInput = page.getByRole("textbox", {
			name: "Database *",
		})
		this.icebergJdbcS3AccessKeyInput = page.getByRole("textbox", {
			name: "AWS Access Key",
		})
		this.icebergJdbcS3SecretKeyInput = page.getByRole("textbox", {
			name: "AWS Secret Key",
		})
		this.icebergJdbcS3RegionInput = page.getByRole("textbox", {
			name: "AWS Region *",
		})
		this.icebergUsePathStyleForS3Switch = page.getByRole("switch", {
			name: "Use Path Style for S3",
		})
		this.icebergJdbcCatalogSelect = page
			.locator("div")
			.filter({ hasText: /^AWS Glue$/ })
			.nth(2)

		this.icebergJdbcUseSSLForS3Switch = page.getByRole("switch", {
			name: "Use SSL for S3",
		})
	}

	async goto() {
		await this.page.goto("/destinations/new")
	}

	async expectCreateDestinationPageVisible() {
		await expect(this.pageTitle).toBeVisible()
		await expect(this.destinationNameInput).toBeVisible()
		await expect(this.createButton).toBeVisible()
	}

	async fillDestinationName(name: string) {
		await this.destinationNameInput.click()
		await this.destinationNameInput.fill(name)
	}

	async selectConnector(connector: string) {
		await this.connectorSelect.click()
		await this.page.getByText(connector).click()
	}

	async selectCatalog(catalog: string) {
		await this.catalogSelect.click()
		await this.page.getByText(catalog).click()
	}

	async fillS3Configuration(config: {
		bucketName: string
		region: string
		path: string
	}) {
		await this.bucketNameInput.click({ timeout: TIMEOUTS.LONG })
		await this.bucketNameInput.fill(config.bucketName)

		await this.regionInput.click({ timeout: TIMEOUTS.LONG })
		await this.regionInput.fill(config.region)

		await this.pathInput.click({ timeout: TIMEOUTS.LONG })
		await this.pathInput.fill(config.path)
	}

	async fillIcebergJdbcConfiguration(config: {
		jdbcUrl: string
		jdbcUsername: string
		jdbcPassword: string
		jdbcDatabase: string
		jdbcS3Endpoint: string
		jdbcS3AccessKey: string
		jdbcS3SecretKey: string
		jdbcS3Region: string
		jdbcS3Path: string
		jdbcUsePathStyleForS3: boolean
		jdbcUseSSLForS3: boolean
	}) {
		await this.connectorSelect.click({ timeout: TIMEOUTS.LONG })
		await this.page
			.locator("div")
			.filter({ hasText: /^Apache Iceberg$/ })
			.nth(1)
			.click()

		await this.icebergJdbcCatalogSelect.click({ timeout: TIMEOUTS.LONG })
		await this.page.getByText("JDBC").click()

		await this.icebergJdbcUrlInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcUrlInput.fill(config.jdbcUrl)

		await this.icebergJdbcDatabaseInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcDatabaseInput.fill(config.jdbcDatabase)

		await this.icebergJdbcUsernameInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcUsernameInput.fill(config.jdbcUsername)

		await this.icebergJdbcPasswordInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcPasswordInput.fill(config.jdbcPassword)

		await this.icebergJdbcS3EndpointInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcS3EndpointInput.fill(config.jdbcS3Endpoint)

		await this.icebergJdbcS3AccessKeyInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcS3AccessKeyInput.fill(config.jdbcS3AccessKey)

		await this.icebergJdbcS3SecretKeyInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcS3SecretKeyInput.fill(config.jdbcS3SecretKey)

		await this.icebergJdbcS3RegionInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcS3RegionInput.fill(config.jdbcS3Region)

		await this.icebergJdbcS3PathInput.click({ timeout: TIMEOUTS.LONG })
		await this.icebergJdbcS3PathInput.fill(config.jdbcS3Path)

		if (!config.jdbcUsePathStyleForS3)
			await this.icebergUsePathStyleForS3Switch.click({
				timeout: TIMEOUTS.LONG,
			})

		if (config.jdbcUseSSLForS3)
			await this.icebergJdbcUseSSLForS3Switch.click({
				timeout: TIMEOUTS.LONG,
			})
	}
	async clickCreate() {
		await this.createButton.click()
	}

	async clickCancel() {
		await this.cancelButton.click()
	}

	async goBackToDestinations() {
		await this.backToDestinationsLink.click()
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
		await expect(this.page.locator(".ant-modal")).toBeVisible()
	}

	async expectSuccessModal() {
		await expect(this.page.getByText("Connection successful")).toBeVisible({
			timeout: TIMEOUTS.LONG,
		})
	}

	async expectEntitySavedModal() {
		await expect(
			this.page.getByText("Destination is connected and saved successfully"),
		).toBeVisible()
		await this.page.getByRole("button", { name: "Destinations" }).click()
	}

	async fillAmazonS3Form(data: {
		name: string
		bucketName: string
		region: string
		path: string
	}) {
		await this.fillDestinationName(data.name)
		// await this.selectConnector("Amazon S3")
		await this.fillS3Configuration({
			bucketName: data.bucketName,
			region: data.region,
			path: data.path,
		})
	}

	async fillIcebergJdbcForm(data: {
		name: string
		jdbcUrl: string
		jdbcUsername: string
		jdbcPassword: string
		jdbcDatabase: string
		jdbcS3Endpoint: string
		jdbcS3AccessKey: string
		jdbcS3SecretKey: string
		jdbcS3Region: string
		jdbcS3Path: string
		jdbcUsePathStyleForS3: boolean
		jdbcUseSSLForS3: boolean
	}) {
		await this.fillDestinationName(data.name)
		await this.fillIcebergJdbcConfiguration({
			jdbcUrl: data.jdbcUrl,
			jdbcUsername: data.jdbcUsername,
			jdbcPassword: data.jdbcPassword,
			jdbcDatabase: data.jdbcDatabase,
			jdbcS3Endpoint: data.jdbcS3Endpoint,
			jdbcS3AccessKey: data.jdbcS3AccessKey,
			jdbcS3SecretKey: data.jdbcS3SecretKey,
			jdbcS3Region: data.jdbcS3Region,
			jdbcS3Path: data.jdbcS3Path,
			jdbcUsePathStyleForS3: data.jdbcUsePathStyleForS3,
			jdbcUseSSLForS3: data.jdbcUseSSLForS3,
		})
	}
}
