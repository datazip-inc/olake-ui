import { Page, Locator, expect } from "@playwright/test"

// Shared utility to select a connector from a dropdown for both source and destination
export const selectConnector = async (
	page: Page,
	connectorSelect: Locator,
	connector: string,
	connectorTestIdMap: Record<string, string>,
) => {
	await connectorSelect.click()

	await page.waitForSelector(".ant-select-dropdown:visible")

	const testId = connectorTestIdMap[connector]
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
