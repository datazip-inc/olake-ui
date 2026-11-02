/** Steps any destination connector uses. */

import { Page } from "@playwright/test"

import { CreateDestinationPage } from "../../pages/CreateDestinationPage"
import { DestinationsPage } from "../../pages/DestinationsPage"
import { DestinationFormConfig } from "../../types/PageConfig.types"
import { expectTestConnectionModalVisible } from "../../utils"
import { expectConnectionTestPassed } from "../common/connection"

/**
 * Creates a destination: fills the Create Destination form from config, and Create tests the
 * connection, which must pass.
 */
export const createDestination = async (
	page: Page,
	destinationsPage: DestinationsPage,
	createDestinationPage: CreateDestinationPage,
	config: DestinationFormConfig,
) => {
	await destinationsPage.navigateToDestinations()
	await destinationsPage.expectDestinationsPageVisible()
	await destinationsPage.clickCreateDestination()
	await createDestinationPage.expectCreateDestinationPageVisible()
	await createDestinationPage.fillDestinationForm(config)
	await createDestinationPage.clickCreate()
	await expectTestConnectionModalVisible(page, "Destination")
	await expectConnectionTestPassed(
		page,
		"Destination",
		page.getByText("Destination is connected and saved successfully"),
	)
	await createDestinationPage.expectEntitySavedModal()
	await destinationsPage.expectDestinationsPageVisible()
	await destinationsPage.expectDestinationExists(config.name)
}
