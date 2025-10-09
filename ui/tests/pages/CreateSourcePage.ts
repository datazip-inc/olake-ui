import { Page, Locator, expect } from "@playwright/test"
import { TIMEOUTS } from "../../playwright.config"

/**
 * Generic configuration for source connector forms
 */
export interface SourceFormConfig {
	name: string
	connector: string
	version?: string
	fields: Record<string, any>
}

export class CreateSourcePage {
	readonly page: Page
	readonly sourceNameInput: Locator
	readonly connectorSelect: Locator
	readonly versionSelect: Locator
	readonly createButton: Locator
	readonly cancelButton: Locator
	readonly backToSourcesLink: Locator
	readonly pageTitle: Locator
	readonly testConnectionButton: Locator
	readonly setupTypeNew: Locator
	readonly setupTypeExisting: Locator

	constructor(page: Page) {
		this.page = page
		this.sourceNameInput = page.getByPlaceholder(
			"Enter the name of your source",
		)
		this.connectorSelect = page.getByTestId("source-connector-select")
		this.versionSelect = page.getByTestId("source-version-select")
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

	/**
	 * Generic method to get a form field by its ID
	 * RJSF generates IDs in the format: #root_fieldname
	 * For nested fields: #root_parent_child (e.g., #root_ssl_mode)
	 *
	 * @param fieldId - The field ID or path (e.g., 'host', 'ssl_mode', 'ssh_config_host')
	 * @returns Locator for the field
	 *
	 */
	getFieldById(fieldId: string): Locator {
		return this.page.locator(`#root_${fieldId}`)
	}

	/**
	 * Generic method to get an array field by its index
	 * RJSF generates array IDs in the format: #root_fieldname_index
	 *
	 * @param fieldId - The field ID (e.g., 'hosts')
	 * @param index - The array index
	 * @returns Locator for the array field
	 *
	 */
	getArrayFieldById(fieldId: string, index: number = 0): Locator {
		return this.page.locator(`#root_${fieldId}_${index}`)
	}

	/**
	 * Generic method to fill any text/number input field
	 * Handles both simple and nested fields
	 *
	 * @param fieldId - The field ID or path (use underscore for nested: 'ssl_mode', 'ssh_config_host')
	 * @param value - The value to fill
	 *
	 */
	async fillField(fieldId: string, value: string) {
		const field = this.getFieldById(fieldId)
		await field.click()
		await field.fill(value)
	}

	/**
	 * Generic method to fill an array field
	 *
	 * @param fieldId - The field ID (e.g., 'hosts')
	 * @param value - The value to fill
	 * @param index - The array index (default: 0)
	 *
	 */
	async fillArrayField(fieldId: string, value: string, index: number = 0) {
		const field = this.getArrayFieldById(fieldId, index)
		await field.click()
		await field.fill(value)
	}

	/**
	 * Generic method to toggle a switch/checkbox field
	 *
	 * @param fieldId - The field ID or path (use underscore for nested fields)
	 *
	 */
	async toggleSwitch(fieldId: string) {
		const field = this.getFieldById(fieldId)
		await field.click()
	}

	async fillSourceName(name: string) {
		await this.sourceNameInput.click()
		await this.sourceNameInput.fill(name)
	}

	async selectConnector(connector: string) {
		await this.connectorSelect.click()
		await this.page
			.locator("div")
			.filter({ hasText: new RegExp(`^${connector}$`) })
			.nth(1)
			.click()
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

	async selectVersion(version: string) {
		await this.versionSelect.click()
		await this.page.getByTitle(version).click()
	}

	/**
	 * Generic method to fill a complete source form
	 * This is the main method to use for any connector type
	 *
	 * @param config - Source form configuration containing connector, name, version, and fields
	 *
	 */
	async fillSourceForm(config: SourceFormConfig) {
		// Select connector
		await this.selectConnector(config.connector)

		// Select version if provided
		if (config.version) {
			await this.selectVersion(config.version)
		}

		// Fill source name
		await this.fillSourceName(config.name)

		// Fill dynamic fields
		await this.fillDynamicFields(config.fields)
	}

	/**
	 * Generic method to fill dynamic form fields based on configuration
	 * Automatically handles arrays, nested objects, and primitive values
	 *
	 * @param fields - Object containing field IDs and their values
	 * @param parentPath - Internal parameter for nested object handling
	 *
	 */
	async fillDynamicFields(
		fields: Record<string, any>,
		parentPath: string = "",
	) {
		for (const [fieldId, value] of Object.entries(fields)) {
			if (value === undefined || value === null) continue

			// Construct the full field path
			const fullPath = parentPath ? `${parentPath}_${fieldId}` : fieldId

			// Handle nested objects
			if (
				typeof value === "object" &&
				!Array.isArray(value) &&
				value !== null &&
				typeof value !== "boolean"
			) {
				// Recursively handle nested object
				await this.fillDynamicFields(value, fullPath)
			}
			// Handle array fields
			else if (Array.isArray(value)) {
				for (let i = 0; i < value.length; i++) {
					await this.fillArrayField(fullPath, value[i], i)
				}
			}
			// Handle boolean fields (switches/toggles)
			else if (typeof value === "boolean") {
				// Only toggle if we need to change from default state
				if (value === true) {
					await this.toggleSwitch(fullPath)
				}
			}
			// Handle string/number fields
			else if (typeof value === "string" || typeof value === "number") {
				await this.fillField(fullPath, value.toString())
			}
		}
	}
}
