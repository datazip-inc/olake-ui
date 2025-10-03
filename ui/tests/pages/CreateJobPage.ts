import { Page, Locator, expect } from "@playwright/test"

export class CreateJobPage {
	readonly page: Page
	readonly jobNameInput: Locator
	readonly useExistingSourceRadio: Locator
	readonly useExistingDestinationRadio: Locator
	readonly sourceSelect: Locator
	readonly destinationSelect: Locator
	readonly nextButton: Locator
	readonly createJobButton: Locator
	readonly cancelButton: Locator
	readonly syncAllCheckbox: Locator
	readonly fullRefreshIncrementalRadio: Locator
	readonly frequencyDropdown: Locator
	readonly pageTitle: Locator
	readonly jobsArrowButton: Locator

	constructor(page: Page) {
		this.page = page
		this.jobNameInput = page.getByRole("textbox")
		this.useExistingSourceRadio = page.getByRole("radio", {
			name: "Use an existing source",
		})
		this.useExistingDestinationRadio = page.getByText(
			"Use an existing destination",
		)
		this.sourceSelect = page.getByTestId("existing-source")
		this.destinationSelect = page.getByTestId("existing-destination")
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
	}

	async goto() {
		await this.page.goto("/jobs/new")
	}

	async expectCreateJobPageVisible() {
		await expect(this.pageTitle).toBeVisible()
		await expect(this.nextButton).toBeVisible()
	}

	async selectExistingSource(sourceName: string) {
		await this.useExistingSourceRadio.check()
		console.log(this.sourceSelect, "source select elem")
		await this.sourceSelect.click()

		await this.page.getByText(sourceName).click()
		await this.nextButton.click()
	}

	async selectExistingDestination(destinationName: string) {
		await this.useExistingDestinationRadio.click()
		await this.destinationSelect.click()
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

	async expectValidationError(message: string) {
		await expect(this.page.locator(`text=${message}`)).toBeVisible()
	}

	async fillJobCreationForm(data: {
		sourceName: string
		destinationName: string
		streamName: string
		jobName: string
		frequency?: string
	}) {
		// Step 1: Configure job settings
		await this.configureJobSettings(data.jobName, data.frequency)

		// Step 2: Select source
		await this.selectExistingSource(data.sourceName)

		// Step 3: Select destination
		await this.selectExistingDestination(data.destinationName)

		// Step 4: Configure streams
		await this.configureStreams(data.streamName)
	}
}
