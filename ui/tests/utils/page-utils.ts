import { Page, Locator, expect } from "@playwright/test"
import {
	SOURCE_CONNECTOR_TEST_ID_MAP,
	DESTINATION_CONNECTOR_TEST_ID_MAP,
} from "./constants"
import { DestinationConnector, SourceConnector } from "../enums"

// Shared utility to select a connector from a dropdown for both source and destination
export const selectConnector = async (
	page: Page,
	connectorSelect: Locator,
	connector: SourceConnector | DestinationConnector,
) => {
	await connectorSelect.click()

	await page.waitForSelector(".ant-select-dropdown:visible")

	const testId =
		connector in SOURCE_CONNECTOR_TEST_ID_MAP
			? SOURCE_CONNECTOR_TEST_ID_MAP[connector]
			: DESTINATION_CONNECTOR_TEST_ID_MAP[connector]

	if (testId) {
		await page.getByTestId(testId).click()
	} else {
		await page
			.locator(".ant-select-dropdown:visible")
			.getByText(connector, { exact: true })
			.click()
	}

	await expect(connectorSelect).toContainText(connector)
}
