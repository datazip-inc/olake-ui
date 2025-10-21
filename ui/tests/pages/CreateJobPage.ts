import { Page, Locator } from "@playwright/test"
import { BasePage } from "./BasePage"
import { JobFormConfig } from "../types/PageConfig.types"
import { DestinationConnector, SourceConnector } from "../enums"
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
		this.createJobButton = page.getByRole("button", { name: "Create Job" })
		this.cancelButton = page.getByRole("button", { name: "Cancel" })
		this.syncAllCheckbox = page.getByRole("checkbox", { name: "Sync all" })
		this.fullRefreshIncrementalRadio = page.getByRole("radio", {
			name: "Full Refresh + Incremental",
		})
		this.frequencyDropdown = page.getByText("Every Minute")
		this.pageTitle = page.locator("text=Create Job")
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
		await this.useExistingSourceRadio.check()
		await selectConnector(this.page, this.sourceConnectorSelect, connector)
		await this.existingSourceSelect.click()

		await this.page.getByText(sourceName).click()
		await this.nextButton.click()
	}

	async selectExistingDestination(
		destinationName: string,
		connector: DestinationConnector,
	) {
		await this.useExistingDestinationRadio.click()
		await selectConnector(this.page, this.destinationConnectorSelect, connector)

		await this.existingDestinationSelect.click()
		await this.page.getByText(destinationName, { exact: true }).click()
		await this.nextButton.click()
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
	) {
		await this.jobNameInput.click()
		await this.jobNameInput.fill(jobName)

		// Change frequency
		await this.frequencyDropdown.click()
		await this.page.getByText(frequency).click()

		await this.nextButton.click()
	}

	async goToJobsPage() {
		await this.jobsArrowButton.click()
	}

	async fillJobCreationForm(data: JobFormConfig) {
		// Step 1: Configure job settings
		await this.configureJobSettings(data.jobName, data.frequency)

		// Step 2: Select source
		await this.selectExistingSource(data.sourceName, data.sourceConnector)

		// Step 3: Select destination
		await this.selectExistingDestination(
			data.destinationName,
			data.destinationConnector,
		)

		// Step 4: Configure streams
		await this.configureStreams(data.streamName)
	}
}
