/** antd controls the steps use: switches and selects. */

import { Locator, Page } from "@playwright/test"

import { expect } from "../../fixtures"

/** The switch in the closest section around a label. */
export const switchNear = (page: Page, label: string) =>
	page
		.getByText(label, { exact: true })
		.first()
		.locator("xpath=ancestor::div[.//button[@role='switch']][1]")
		.getByRole("switch")
		.first()

/** Sets a switch to on or off and checks the result. */
export const setSwitch = async (control: Locator, on: boolean) => {
	if ((await control.getAttribute("aria-checked")) !== String(on)) {
		await control.click()
	}
	await expect(control).toHaveAttribute("aria-checked", String(on))
}

/**
 * The dropdown of an antd Select. antd keeps every dropdown it has opened in the page, closed
 * ones included, so find this select's own through the list id its input controls.
 */
export const dropdownOf = async (page: Page, select: Locator) => {
	const id = await select
		.locator("input[role='combobox']")
		.first()
		.getAttribute("aria-controls")
	expect(id, "the select's input names its option list").toBeTruthy()
	return page
		.locator(".ant-select-dropdown")
		.filter({ has: page.locator(`[id="${id}"]`) })
}

/**
 * Opens select and clicks its option with exactly text.
 * - antd also renders a hidden copy of each option (role="option", for screen readers), so this
 *   matches the rendered .ant-select-item-option instead.
 * - The list is virtual: only the rows in view exist (about 8), and it scrolls on wheel events
 *   only, not on scrollTop. So scroll it to the top, then down until the option is rendered.
 */
export const pickOption = async (page: Page, select: Locator, text: string) => {
	await select.click()
	const dropdown = await dropdownOf(page, select)
	const options = dropdown.locator(".ant-select-item-option")
	await expect(options.first()).toBeVisible()
	const option = options
		.filter({ has: page.getByText(text, { exact: true }) })
		.first()
	// Wheel over the list, aiming at it again each time (a wheel the list cannot use scrolls the
	// page instead, which moves the list). Turn around at either end: an earlier pick can leave
	// the list scrolled down.
	const list = dropdown.locator(".rc-virtual-list-holder").first()
	let direction = 1
	let rendered = ""
	for (let i = 0; i < 100 && !(await option.isVisible()); i++) {
		const box = await list.boundingBox()
		if (!box) break
		await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
		await page.mouse.wheel(0, direction * 80)
		await page.waitForTimeout(120)
		const now = (await options.allInnerTexts()).join("|")
		if (now === rendered) direction = -direction
		rendered = now
	}
	await option.click()
}
