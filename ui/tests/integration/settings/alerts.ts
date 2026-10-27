/** System settings: Alerts and Notifications. */

import { Page } from "@playwright/test"

import { expect } from "../../fixtures"

/**
 * The webhook URL on a new OLake: save url, edit and cancel, clear, and save url again, which
 * leaves it set for the failure alert.
 */
export const expectWebhookAlerts = async (page: Page, url: string) => {
	const open = async () => {
		await page.goto("/settings")
		await page.getByRole("tab", { name: "Alerts and Notifications" }).click()
	}
	const input = page.getByPlaceholder("Enter your webhook URL")
	const save = page.getByRole("button", { name: "Save changes" })

	await open()
	// A new OLake has no URL, so the field starts editable and Save is disabled while empty.
	await expect(input).toBeEditable()
	await expect(input).toHaveValue("")
	await expect(save).toBeDisabled()
	await input.fill(url)
	await save.click()
	await expect(input).toBeDisabled()
	await expect(input).toHaveValue(url)
	await open()
	await expect(input).toHaveValue(url)
	await expect(input).toBeDisabled()

	// Edit, then Cancel, keeps the saved URL.
	await page.getByRole("button", { name: "Edit" }).click()
	await expect(input).toBeEditable()
	await input.fill("http://changed.invalid/alert")
	await page.getByRole("button", { name: "Cancel", exact: true }).click()
	await expect(input).toHaveValue(url)
	await expect(input).toBeDisabled()

	// Clear removes it.
	await page.getByRole("button", { name: "Clear webhook URL" }).click()
	await expect(input).toHaveValue("")
	await expect(input).toBeEditable()
	await open()
	await expect(input).toHaveValue("")

	// Save again: the URL the failure alert goes to.
	await input.fill(url)
	await save.click()
	await expect(input).toBeDisabled()
	await open()
	await expect(input).toHaveValue(url)
}
