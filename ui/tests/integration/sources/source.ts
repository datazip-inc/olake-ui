/** Steps any source connector uses: creating one, and its test connection. */

import { Locator, Page } from "@playwright/test"

import { expect } from "../../fixtures"
import { CreateSourcePage } from "../../pages/CreateSourcePage"
import { SourcesPage } from "../../pages/SourcesPage"
import { SourceFormConfig } from "../../types/PageConfig.types"
import { expectTestConnectionModalVisible } from "../../utils"
import { TIMEOUTS } from "../e2e"

/**
 * Waits for the source test connection to finish and fails unless it succeeded. With CDC on the
 * prerequisite checks must all pass: "Continue anyway" means one did not.
 */
export const expectSourceTestPassed = async (page: Page) => {
	const wait = (locator: Locator, outcome: string) =>
		locator
			.waitFor({ state: "visible", timeout: TIMEOUTS.LONG })
			.then(() => outcome)
			.catch(() => "timeout")
	const outcome = await Promise.race([
		wait(page.getByText("Source test connection is successful"), "success"),
		wait(page.getByText("Source test connection has failed"), "failed"),
		wait(
			page.getByRole("button", { name: "Continue anyway" }),
			"checks failed",
		),
	])
	// On failure, name the checks: the prerequisite dialog lists each one with its current value.
	const details =
		outcome === "success"
			? ""
			: await page
					.getByRole("dialog", { name: "Testing Source Configuration" })
					.innerText({ timeout: 2000 })
					.catch(() => "")
	expect(
		outcome,
		`source test connection and its CDC prerequisite checks must all pass\n${details}`,
	).toBe("success")
}

/**
 * Creates a source: fills the Create Source form from config, then fill sets what the form config
 * does not, and Create tests the connection, which must pass.
 */
export const createSource = async (
	page: Page,
	sourcesPage: SourcesPage,
	createSourcePage: CreateSourcePage,
	config: SourceFormConfig,
	fill: () => Promise<void> = async () => {},
) => {
	await sourcesPage.navigateToSources()
	await sourcesPage.expectSourcesPageVisible()
	await sourcesPage.clickCreateSource()
	await createSourcePage.expectCreateSourcePageVisible()
	await createSourcePage.fillSourceForm(config)
	await fill()
	await createSourcePage.clickCreate()
	await expectTestConnectionModalVisible(page, "Source")
	await expectSourceTestPassed(page)
	await createSourcePage.expectEntitySavedModal()
	await sourcesPage.expectSourcesPageVisible()
	await sourcesPage.expectSourceExists(config.name)
}
