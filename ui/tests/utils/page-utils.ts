import { Page, Locator, expect } from "@playwright/test"

import {
	SOURCE_CONNECTOR_TEST_ID_MAP,
	DESTINATION_CONNECTOR_TEST_ID_MAP,
} from "./constants"
import { TIMEOUTS } from "../../playwright.config"
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

/**
 * Polls for specific text by repeatedly clicking a given button.
 * Default polling interval is 1 seconds. Max timeout uses `TIMEOUTS.LONG`.
 */
export const pollToClickAndVerifyText = async (
	page: Page,
	refreshButton: Locator,
	expected: string | Locator,
	options: { timeout?: number; interval?: number; expectTimeout?: number } = {},
) => {
	const {
		timeout = TIMEOUTS.LONG,
		interval = 1000,
		expectTimeout = 1000,
	} = options

	await expect(async () => {
		await refreshButton.click()

		// If string is passed, grab by text. Otherwise, use the locator directly.
		const target =
			typeof expected === "string" ? page.getByText(expected) : expected

		await expect(target).toBeVisible({
			timeout: expectTimeout,
		})
	}).toPass({
		timeout,
		intervals: [interval],
	})
}
