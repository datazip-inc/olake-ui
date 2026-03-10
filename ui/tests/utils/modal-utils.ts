import { Page, expect } from "@playwright/test"
import { TIMEOUTS } from "../../playwright.config"

export interface ModalTestablePage {
	expectTestConnectionModal: () => Promise<void>
	assertTestConnectionSucceeded: () => Promise<void>
}

export interface EntityCreationPage extends ModalTestablePage {
	expectEntitySavedModal: () => Promise<void>
}

/**
 * Verifies entity creation success by checking:
 * 1. Test connection modal appears
 * 2. Test connection succeeds
 * 3. Entity saved modal appears and navigates back
 */
export const verifyEntityCreationSuccessModal = async (
	pageObject: EntityCreationPage,
) => {
	await pageObject.expectTestConnectionModal()
	await pageObject.assertTestConnectionSucceeded()
	await pageObject.expectEntitySavedModal()
}

/** Verifies entity test connection success by checking:
 * 1. Test connection modal appears
 * 2. Test connection succeeds
 */
export const verifyEntityTestConnectionSuccessModal = async (
	pageObject: ModalTestablePage,
) => {
	await pageObject.expectTestConnectionModal()
	await pageObject.assertTestConnectionSucceeded()
}

/**
 * Reusable utility to assert the testing connection modal appears.
 * @param page Playwright Page
 * @param type "Source" or "Destination"
 */
export const expectTestConnectionModalVisible = async (
	page: Page,
	type: "Source" | "Destination",
) => {
	await page.waitForSelector(`text=Testing ${type} connection`, {
		state: "visible",
	})
	await expect(page.getByText(`Testing ${type} connection`)).toHaveCount(1)
}

/**
 * Reusable utility to assert the connection testing outcome.
 * @param page Playwright Page
 * @param type "Source" or "Destination"
 */
export const assertTestConnectionOutcome = async (
	page: Page,
	type: "Source" | "Destination",
) => {
	const failure = page
		.waitForSelector(`text=${type} test connection has failed`, {
			state: "visible",
			timeout: TIMEOUTS.LONG,
		})
		.then(() => "failure")
	const success = page
		.waitForSelector(`text=${type} test connection is successful`, {
			state: "visible",
			timeout: TIMEOUTS.LONG,
		})
		.then(() => "success")

	const outcome = await Promise.race([failure, success])
	expect(outcome, `${type} test connection failed`).toBe("success")
}
