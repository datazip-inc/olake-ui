import { Page, Locator, expect } from "@playwright/test"

import { BasePage } from "./BasePage"
import { DestinationConnector, SourceConnector } from "../enums"
import { JobFormConfig } from "../types/PageConfig.types"
import {
	verifyEntityTestConnectionSuccessModal,
	expectTestConnectionModalVisible,
	assertTestConnectionOutcome,
} from "../utils"
import { selectConnector } from "../utils/page-utils"

export class CreateJobPage extends BasePage {
	readonly jobNameInput: Locator
	readonly useExistingSourceRadio: Locator
	readonly useExistingDestinationRadio: Locator
	readonly existingSourceSelect: Locator
	readonly existingDestinationSelect: Locator
	readonly nextButton: Locator
	readonly createJobButton: Locator
	readonly cancelButton: Locator
	readonly syncAllCheckbox: Locator
	readonly fullRefreshIncrementalRadio: Locator
	readonly frequencyDropdown: Locator
	readonly pageTitle: Locator
	readonly jobsArrowButton: Locator
	readonly sourceConnectorSelect: Locator
	readonly destinationConnectorSelect: Locator

	constructor(page: Page) {
		super(page)
		this.jobNameInput = page.getByRole("textbox")
		this.useExistingSourceRadio = page.getByRole("radio", {
			name: "Use an existing source",
		})
		this.useExistingDestinationRadio = page.getByText(
			"Use an existing destination",
		)
		this.existingSourceSelect = page.getByTestId("existing-source")
		this.existingDestinationSelect = page.getByTestId("existing-destination")
		this.nextButton = page.getByRole("button", { name: "Next" })
		this.createJobButton = page.getByTestId("create-job-wizard-submit")
		this.cancelButton = page.getByRole("button", { name: "Cancel" })
		this.syncAllCheckbox = page.getByRole("checkbox", { name: "Sync all" })
		this.fullRefreshIncrementalRadio = page.getByRole("radio", {
			name: "Full Refresh + Incremental",
		})
		this.frequencyDropdown = page.getByText("Every Minute")
		this.pageTitle = page.getByTestId("create-job-page-title")
		this.jobsArrowButton = page.getByRole("button", { name: "Jobs →" })
		this.sourceConnectorSelect = page.getByTestId("source-connector-select")
		this.destinationConnectorSelect = page.getByTestId(
			"destination-connector-select",
		)
	}

	async goto() {
		await super.goto("/jobs/new")
	}

	async expectCreateJobPageVisible() {
		await this.expectVisible(this.pageTitle)
		await this.expectVisible(this.nextButton)
	}

	async selectExistingSource(sourceName: string, connector: SourceConnector) {
		await selectConnector(this.page, this.sourceConnectorSelect, connector)
		await this.existingSourceSelect.click()

		await this.page.getByText(sourceName).click()
	}

	async selectExistingDestination(
		destinationName: string,
		connector: DestinationConnector,
	) {
		await selectConnector(this.page, this.destinationConnectorSelect, connector)

		await this.existingDestinationSelect.click()
		await this.page.getByText(destinationName, { exact: true }).click()
	}

	async configureStreams(streamName: string) {
		// Uncheck sync all first
		await this.syncAllCheckbox.uncheck()

		// Select specific stream
		await this.page
			.getByRole("button", { name: streamName })
			.getByLabel("")
			.check()

		// Set sync mode
		await this.fullRefreshIncrementalRadio.check()

		// Enable the stream switch
		await this.page.getByRole("switch").first().click()

		await this.createJobButton.click()
	}

	async configureJobSettings(
		jobName: string,
		frequency: string = "Every Week",
		sourceName: string,
		sourceConnector: SourceConnector,
		destinationName: string,
		destinationConnector: DestinationConnector,
	) {
		await this.jobNameInput.click()
		await this.jobNameInput.fill(jobName)

		// Change frequency
		await this.frequencyDropdown.click()
		await this.page.getByText(frequency).click()

		await this.selectExistingSource(sourceName, sourceConnector)
		await this.selectExistingDestination(destinationName, destinationConnector)

		await this.nextButton.click()
	}

	// Fills the job config step for existing entities and goes to the streams step.
	async openStreamsStep(config: {
		jobName: string
		source: string
		sourceConnector: SourceConnector
		destination: string
		destinationConnector: DestinationConnector
		queryEngine?: string
	}) {
		await this.goto()
		await this.expectCreateJobPageVisible()
		await this.jobNameInput.fill(config.jobName)
		// the default runs every minute, which would sync on top of the syncs the test starts
		await this.frequencyDropdown.click()
		await this.page.getByText("Every Week").click()
		await this.selectExistingSource(config.source, config.sourceConnector)
		await this.selectExistingDestination(
			config.destination,
			config.destinationConnector,
		)
		if (config.queryEngine) {
			await this.chooseOption(
				this.page
					.locator(".ant-select")
					.filter({ hasText: "Select Query Engines" }),
				config.queryEngine,
			)
			await this.page.keyboard.press("Escape")
		}
		await this.nextButton.click()
		await verifyEntityTestConnectionSuccessModal(this)
		await expect(this.syncAllCheckbox).toBeVisible()
	}

	// In edit mode: goes from the streams step back to the job config step.
	async backToJobConfig() {
		await this.page.getByRole("button", { name: "Back" }).click()
		await this.page.getByRole("button", { name: "Yes, Leave" }).click()
		await expect(this.page.getByText("Job Configuration")).toBeVisible()
	}

	async changeFrequency(from: string, to: string) {
		await this.page.getByText(from, { exact: true }).click()
		await this.page
			.locator(".ant-select-dropdown:visible")
			.getByText(to, { exact: true })
			.click()
	}

	// In edit mode: finishes the edit and confirms the warning about the streams it affects.
	async finishEdit() {
		await this.page.getByRole("button", { name: "Finish" }).click()
		await this.page.getByRole("button", { name: "Confirm and Finish" }).click()
		// the edit is saved by the time the page leaves the edit URL
		await this.page.waitForURL(url => !url.pathname.endsWith("/edit"))
	}

	// --- Streams step ---

	streamRow(streamName: string): Locator {
		return this.page.getByRole("button", { name: streamName, exact: true })
	}

	// Opens the stream in the right-hand panel without changing whether it is selected.
	async openStream(streamName: string) {
		await this.streamRow(streamName).click()
	}

	// Syncs exactly the given streams.
	async selectOnlyStreams(streamNames: string[]) {
		await this.syncAllCheckbox.uncheck()
		for (const name of streamNames) {
			await this.streamRow(name).getByRole("checkbox").check()
		}
	}

	async chooseOption(select: Locator, optionText: string) {
		await select.click()
		await this.page
			.locator(".ant-select-dropdown:visible .ant-select-item-option")
			.filter({ has: this.page.getByText(optionText, { exact: true }) })
			// a dropdown that was just closed can still be animating out, so take the newest
			.last()
			.click()
	}

	async selectSyncMode(label: string) {
		await this.page.getByRole("radio", { name: label, exact: true }).check()
	}

	async selectIngestionMode(label: "Upsert" | "Append") {
		await this.page.getByRole("radio", { name: label, exact: true }).check()
	}

	upsertTypeSelect(): Locator {
		return this.page.locator(".ant-select").filter({
			hasText: /Equality Deletes|Positional Deletes|Deletion Vector/,
		})
	}

	async selectUpsertType(label: string) {
		await this.chooseOption(this.upsertTypeSelect(), label)
	}

	// Section toggles of the Config tab (Normalization, Data Filter) share one layout: a label and a switch.
	sectionSwitch(label: string, root: Page | Locator = this.page): Locator {
		return root
			.getByText(label, { exact: true })
			.locator("xpath=ancestor::div[.//button[@role='switch']][1]")
			.getByRole("switch")
	}

	async setDataFilter(
		conditions: { column: string; operator: string; value: string }[],
		logicalOperator: "AND" | "OR" = "AND",
	) {
		await this.sectionSwitch("Data Filter").click()
		for (const [i, condition] of conditions.entries()) {
			if (i > 0) {
				await this.page
					.getByRole("button", { name: "New Column filter" })
					.click()
				await this.page
					.getByRole("button", { name: logicalOperator, exact: true })
					.click()
			}
			// each condition renders a label followed by its control
			const control = (label: string, controlXpath: string) =>
				this.page
					.getByText(label, { exact: true })
					.nth(i)
					.locator(`xpath=following-sibling::${controlXpath}`)
			await this.chooseOption(
				control("Column Name", "div[contains(@class,'ant-select')][1]"),
				condition.column,
			)
			await this.chooseOption(
				control("Operator", "div[contains(@class,'ant-select')][1]"),
				condition.operator,
			)
			await control("Value", "input[1]").fill(condition.value)
		}
	}

	async openSchemaTab() {
		await this.page.getByRole("button", { name: "Schema", exact: true }).click()
	}

	async openConfigTab() {
		await this.page.getByRole("button", { name: "Config", exact: true }).click()
	}

	syncNewColumnsSwitch(): Locator {
		return this.sectionSwitch("Sync new columns automatically")
	}

	// Names the destination database (or S3 folder). Dynamic appends the source schema to the name,
	// custom stores every table directly in the one folder.
	async setDestinationDatabase(
		name: string,
		format: "Dynamic (Default)" | "Custom" = "Dynamic (Default)",
	) {
		await this.page.getByTestId("edit-destination-database").click()
		const modal = this.page.getByRole("dialog")
		await modal.getByText(format).click()
		await modal.getByRole("textbox").fill(name)
		await modal.getByRole("button", { name: "Save Changes" }).click()
	}

	// The cursor selects of an incremental stream, found by their labels.
	private cursorSelect(label: "Cursor field:" | "Fallback Cursor:"): Locator {
		return this.page
			.getByText(label, { exact: true })
			.locator("xpath=following-sibling::div[contains(@class,'ant-select')][1]")
	}

	async setCursor(cursor: string, fallback?: string) {
		await this.chooseOption(this.cursorSelect("Cursor field:"), cursor)
		if (fallback) {
			await this.page
				.getByRole("button", { name: "Add Fallback Cursor" })
				.click()
			await this.chooseOption(this.cursorSelect("Fallback Cursor:"), fallback)
		}
	}

	async setPartitionRegex(regex: string) {
		await this.page
			.getByRole("button", { name: "Partitioning", exact: true })
			.click()
		await this.page.getByPlaceholder("Enter your partition regex").fill(regex)
		await this.page.getByRole("button", { name: "Set Partition" }).click()
		await expect(this.page.getByText("Active partition regex:")).toBeVisible()
	}

	async createJob() {
		await this.createJobButton.click()
	}

	// Confirms the warning about indexes that the next sync has to build.
	async confirmIndexBuildWarning() {
		const modal = this.page
			.getByRole("dialog")
			.filter({ hasText: "Please Confirm" })
		await expect(modal).toContainText("Indexes might get built")
		await modal.getByRole("button", { name: "Confirm" }).click()
	}

	async goToJobsPage() {
		await this.jobsArrowButton.click()
	}

	async expectTestConnectionModal() {
		await expectTestConnectionModalVisible(this.page, "Source")
	}

	async assertTestConnectionSucceeded() {
		await assertTestConnectionOutcome(this.page, "Source")
		await expectTestConnectionModalVisible(this.page, "Destination")
		await assertTestConnectionOutcome(this.page, "Destination")
	}

	async fillJobCreationForm(data: JobFormConfig) {
		// Step 1: Configure job settings
		await this.configureJobSettings(
			data.jobName,
			data.frequency,
			data.sourceName,
			data.sourceConnector,
			data.destinationName,
			data.destinationConnector,
		)

		await verifyEntityTestConnectionSuccessModal(this)

		// Step 2: Configure streams
		await this.configureStreams(data.streamName)
	}
}
