/** Steps any source connector uses. */

import { Page } from "@playwright/test"

import { CreateSourcePage } from "../../pages/CreateSourcePage"
import { SourcesPage } from "../../pages/SourcesPage"
import { SourceFormConfig } from "../../types/PageConfig.types"
import { expectTestConnectionModalVisible } from "../../utils"
import { expectConnectionTestPassed } from "../common/connection"

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
	await expectConnectionTestPassed(
		page,
		"Source",
		page.getByText("Source is connected and saved successfully"),
	)
	await createSourcePage.expectEntitySavedModal()
	await sourcesPage.expectSourcesPageVisible()
	await sourcesPage.expectSourceExists(config.name)
}
