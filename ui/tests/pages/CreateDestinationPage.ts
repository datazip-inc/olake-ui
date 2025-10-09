import { Page, Locator, expect } from "@playwright/test"
import { TIMEOUTS } from "../../playwright.config"

/**
 * Generic configuration for destination connector forms
 */
export interface DestinationFormConfig {
	name: string
	connector: string
	version?: string
	catalogType?: "glue" | "jdbc" | "hive" | "rest" // For Iceberg only
	fields: Record<string, any>
}

export class CreateDestinationPage {
	readonly page: Page
	readonly destinationNameInput: Locator
	readonly connectorSelect: Locator
	readonly versionSelect: Locator
	readonly createButton: Locator
	readonly cancelButton: Locator
	readonly backToDestinationsLink: Locator
	readonly pageTitle: Locator
	readonly testConnectionButton: Locator
	readonly setupTypeNew: Locator
	readonly setupTypeExisting: Locator
	readonly icebergCatalogInput: Locator

	constructor(page: Page) {
		this.page = page
		this.destinationNameInput = page.getByPlaceholder(
			"Enter the name of your destination",
		)
		this.connectorSelect = page.getByTestId("destination-connector-select")
		this.versionSelect = page.getByTestId("destination-version-select")
		this.icebergCatalogInput = page.locator(
			'div[name="root_writer_catalog_type"]',
		)
		this.createButton = page.getByRole("button", { name: "Create" })
		this.cancelButton = page.getByRole("button", { name: "Cancel" })
		this.backToDestinationsLink = page.getByRole("link").first()
		this.pageTitle = page.locator("text=Create destination")
		this.testConnectionButton = page.getByRole("button", {
			name: "Test Connection",
		})
		this.setupTypeNew = page.getByText("Set up a new destination")
		this.setupTypeExisting = page.getByText("Use an existing destination")
	}

	async goto() {
		await this.page.goto("/destinations/new")
	}

	async expectCreateDestinationPageVisible() {
		await expect(this.pageTitle).toBeVisible()
		await expect(this.destinationNameInput).toBeVisible()
		await expect(this.createButton).toBeVisible()
	}

	/**
	 * Generic method to get a form field by its ID
	 * Destination fields are under 'writer' prefix: #root_writer_fieldname
	 *
	 * @param fieldId - The field ID (e.g., 's3_bucket', 'catalog_type')
	 * @returns Locator for the field
	 */
	getFieldById(fieldId: string): Locator {
		return this.page.locator(`#root_writer_${fieldId}`)
	}

	/**
	 * Generic method to fill any text/number input field
	 *
	 * @param fieldId - The field ID under writer (e.g., 's3_bucket', 'aws_region')
	 * @param value - The value to fill
	 */
	async fillField(fieldId: string, value: string) {
		const field = this.getFieldById(fieldId)
		await field.click()
		await field.fill(value)
	}

	/**
	 * Generic method to toggle a switch/checkbox field
	 *
	 * @param fieldId - The field ID under writer
	 */
	async toggleSwitch(fieldId: string) {
		const field = this.getFieldById(fieldId)
		await field.click()
	}

	/**
	 * Select catalog type for Iceberg destinations
	 * Maps friendly names to schema values
	 *
	 * @param catalogType - One of: "glue", "jdbc", "hive", "rest"
	 */
	async selectCatalogType(
		catalogType: "glue" | "jdbc" | "hive" | "rest",
	): Promise<void> {
		await this.icebergCatalogInput.click()

		// Map catalog type to display name in UI
		const catalogDisplayNames: Record<string, string> = {
			glue: "AWS Glue",
			jdbc: "JDBC",
			hive: "Hive",
			rest: "REST",
		}

		const displayName = catalogDisplayNames[catalogType]
		await this.page.getByText(displayName, { exact: true }).click()

		// Wait for form to update based on catalog type
		await this.page.waitForTimeout(500)
	}

	async fillDestinationName(name: string) {
		await this.destinationNameInput.click()
		await this.destinationNameInput.fill(name)
	}

	async selectConnector(connector: string) {
		await this.connectorSelect.click()
		await this.page
			.locator("div")
			.filter({ hasText: new RegExp(`^${connector}$`) })
			.nth(1)
			.click()
	}

	async selectVersion(version: string) {
		await this.versionSelect.click()
		await this.page.getByTitle(version).click()
	}

	/**
	 * Main method to fill destination form - works for ANY connector!
	 *
	 * @param config - Destination form configuration
	 *
	 */
	async fillDestinationForm(config: DestinationFormConfig) {
		// Select connector
		await this.selectConnector(config.connector)

		// Select version if provided
		if (config.version) {
			await this.selectVersion(config.version)
		}

		// Fill destination name
		await this.fillDestinationName(config.name)

		// For Iceberg, select catalog type first if provided
		if (config.connector.toLowerCase() === "apache iceberg") {
			if (config.catalogType) {
				await this.selectCatalogType(config.catalogType)
			}
		}

		// Fill dynamic fields (all under writer prefix)
		await this.fillDynamicFields(config.fields)
	}

	/**
	 * Generic method to fill dynamic form fields
	 * All destination fields are under 'writer' prefix in RJSF
	 */
	async fillDynamicFields(fields: Record<string, any>) {
		for (const [fieldId, value] of Object.entries(fields)) {
			if (value === undefined || value === null) continue

			// Handle boolean fields (switches/toggles)
			if (typeof value === "boolean") {
				if (value === true) {
					await this.toggleSwitch(fieldId)
				}
			}
			// Handle string/number fields
			else if (typeof value === "string" || typeof value === "number") {
				await this.fillField(fieldId, value.toString())
			}
		}
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

	async assertTestConnectionSucceeded() {
		const failure = this.page
			.waitForSelector("text=Your test connection has failed", {
				state: "visible",
				timeout: TIMEOUTS.LONG,
			})
			.then(() => "failure")
		const success = this.page
			.waitForSelector("text=Connection successful", {
				state: "visible",
				timeout: TIMEOUTS.LONG,
			})
			.then(() => "success")

		const outcome = await Promise.race([failure, success])
		expect(outcome, "Test connection failed").toBe("success")
	}

	async expectEntitySavedModal() {
		await expect(
			this.page.getByText("Destination is connected and saved successfully"),
		).toBeVisible()
		await this.page.getByRole("button", { name: "Destinations" }).click()
	}
}
