/** The test connection a source or destination runs when it is saved, or in the job wizard. */

import { Locator, Page } from "@playwright/test"

import { expect } from "../../fixtures"
import { TIMEOUTS } from "../e2e"

/**
 * Waits for the test connection of type to finish and fails unless it passed.
 * - The success message shows for only a second, so next, what the page shows once the test has
 *   passed, counts as success too.
 * - A source with CDC must also pass its prerequisite checks: "Continue anyway" means one did not.
 */
export const expectConnectionTestPassed = async (
	page: Page,
	type: "Source" | "Destination",
	next: Locator,
) => {
	const wait = (locator: Locator, outcome: string) =>
		locator
			.first()
			.waitFor({ state: "visible", timeout: TIMEOUTS.LONG })
			.then(() => outcome)
			.catch(() => "timeout")
	const outcome = await Promise.race([
		wait(page.getByText(`${type} test connection is successful`), "success"),
		wait(next, "success"),
		wait(page.getByText(`${type} test connection has failed`), "failed"),
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
					.getByRole("dialog", { name: `Testing ${type} Configuration` })
					.innerText({ timeout: 2000 })
					.catch(() => "")
	expect(
		outcome,
		`${type} test connection and its prerequisite checks must all pass\n${details}`,
	).toBe("success")
}
