/** The job wizard's streams step: the streams list, a stream's configuration, and bulk changes. */

import { Page } from "@playwright/test"

import { expect } from "../../fixtures"
import {
	pickOption,
	dropdownOf,
	setSwitch,
	switchNear,
} from "../common/controls"

/** The row of a stream in the streams list; clicking it opens the stream's configuration. */
export const streamRow = (page: Page, stream: string) =>
	page.getByRole("button", { name: stream, exact: true })

/** Unticks "Sync all", so only the streams selectStream ticks are synced. */
export const deselectAllStreams = async (page: Page) => {
	await page.getByRole("checkbox", { name: "Sync all" }).uncheck()
}

/** Ticks a stream and opens its configuration. */
export const selectStream = async (page: Page, stream: string) => {
	const row = streamRow(page, stream)
	await row.getByRole("checkbox").check()
	await row.click()
	await expect(
		page.getByText(new RegExp(`/${stream}$`)).first(),
		`${stream} configuration is open`,
	).toBeVisible()
}

/** Opens a stream's configuration without changing its selection. */
export const openStream = async (page: Page, stream: string) => {
	await streamRow(page, stream).click()
	await expect(page.getByText(new RegExp(`/${stream}$`)).first()).toBeVisible()
}

/** A tab of the open stream's configuration. */
export const streamTab = async (
	page: Page,
	tab: "Config" | "Schema" | "Partitioning",
) => {
	await page.getByRole("button", { name: tab, exact: true }).click()
}

export type SyncModeLabel =
	| "Full Refresh"
	| "Full Refresh + Incremental"
	| "Full Refresh + CDC"
	| "CDC Only"

export const syncModeRadio = (page: Page, mode: SyncModeLabel) =>
	page.getByRole("radio", { name: mode, exact: true })

export const setSyncMode = async (page: Page, mode: SyncModeLabel) => {
	await syncModeRadio(page, mode).check()
	await expect(syncModeRadio(page, mode)).toBeChecked()
}

export type IngestionModeLabel = "Upsert" | "Append"

export const setIngestionMode = async (
	page: Page,
	mode: IngestionModeLabel,
) => {
	const radio = page.getByRole("radio", { name: mode, exact: true })
	await radio.check()
	await expect(radio).toBeChecked()
}

/** Sets the open stream's Normalization. */
export const setNormalization = async (page: Page, on: boolean) => {
	await setSwitch(switchNear(page, "Normalization"), on)
}

/** Sets the open stream's "Sync new columns automatically", on its Schema tab. */
export const setSyncNewColumns = async (page: Page, on: boolean) => {
	await streamTab(page, "Schema")
	await setSwitch(switchNear(page, "Sync new columns automatically"), on)
	await streamTab(page, "Config")
}

export type UpsertTypeLabel =
	| "Equality Deletes"
	| "Positional Deletes"
	| "Deletion Vector"

/** The "Upsert Type:" select of the open stream. */
const upsertTypeSelect = (page: Page) =>
	page
		.getByText("Upsert Type:", { exact: true })
		.locator("xpath=ancestor::div[.//div[contains(@class,'ant-select')]][1]")
		.locator(".ant-select")
		.first()

/** The upsert types the open stream offers (at most 3, so all of them are rendered). */
export const offeredUpsertTypes = async (page: Page): Promise<string[]> => {
	const select = upsertTypeSelect(page)
	await select.click()
	const options = (await dropdownOf(page, select)).locator(
		".ant-select-item-option",
	)
	await expect(options.first()).toBeVisible()
	const labels = (await options.allInnerTexts()).map(text => text.trim())
	await page.keyboard.press("Escape")
	return labels
}

export const setUpsertType = async (
	page: Page,
	type: UpsertTypeLabel,
	formatVersion: "V2" | "V3",
) => {
	await pickOption(page, upsertTypeSelect(page), type)
	await expect(upsertTypeSelect(page)).toContainText(type)
	await expect(page.getByText("Iceberg format version:")).toContainText(
		formatVersion,
	)
}

export type FilterCondition = {
	column: string
	operator: ">" | "<" | "=" | "!=" | ">=" | "<="
	value: string
}

/**
 * Turns the open stream's Data Filter on with up to two conditions joined by AND or OR. Data
 * Filter needs Normalization on.
 */
export const setDataFilter = async (
	page: Page,
	conditions: FilterCondition[],
	join: "AND" | "OR",
) => {
	await setSwitch(switchNear(page, "Data Filter"), true)
	for (const [index, condition] of conditions.entries()) {
		if (index > 0) {
			await page.getByRole("button", { name: "New Column filter" }).click()
		}
		const block = page
			.getByText(index === 0 ? "Column I" : "Column II", { exact: true })
			.locator("xpath=ancestor::div[.//label[normalize-space()='Value']][1]")
		const selects = block.locator(".ant-select")
		await pickOption(page, selects.nth(0), condition.column)
		await pickOption(page, selects.nth(1), condition.operator)
		await block.getByRole("textbox").fill(condition.value)
		await expect(selects.nth(0)).toContainText(condition.column)
		await expect(selects.nth(1)).toContainText(condition.operator)
		await expect(block.getByRole("textbox")).toHaveValue(condition.value)
	}
	if (conditions.length > 1) {
		await page.getByRole("button", { name: join, exact: true }).click()
	}
}

/**
 * "All Append" or "All Upsert": confirms the modal, then checks that each of streams took mode.
 */
export const switchAllIngestion = async (
	page: Page,
	mode: IngestionModeLabel,
	streams: string[],
) => {
	await page.getByText(`All ${mode}`, { exact: true }).click()
	const dialog = page.getByRole("dialog").filter({
		hasText: `Switch to ${mode} for all tables ?`,
	})
	await expect(dialog).toBeVisible()
	await expect(dialog).toContainText(
		"You can change mode for individual tables",
	)
	await dialog.getByRole("button", { name: "Confirm" }).click()
	await expect(dialog).toBeHidden()
	for (const stream of streams) {
		await openStream(page, stream)
		await streamTab(page, "Config")
		await expect(
			page.getByRole("radio", { name: mode, exact: true }),
		).toBeChecked()
	}
}

/**
 * Bulk Configure: sets mode on streams through the Select Streams, Apply Configurations and
 * Summary of Changes dialogs, then checks each stream took it.
 */
export const bulkConfigureSyncMode = async (
	page: Page,
	streams: string[],
	mode: SyncModeLabel,
) => {
	await page.getByRole("button", { name: "Bulk Configure" }).click()
	const bulk = page.getByRole("dialog").filter({ hasText: "Select Streams" })
	await expect(bulk).toBeVisible()
	await expect(bulk).toContainText("Select streams you wish to bulk configure")
	for (const stream of streams) {
		await bulk.getByRole("checkbox", { name: stream, exact: true }).check()
	}
	await bulk.getByRole("button", { name: "Configure Streams" }).click()
	const apply = page.getByRole("dialog").filter({
		hasText: "Apply Configurations",
	})
	await expect(apply).toBeVisible()
	// Nothing is marked as changed yet, and the sync mode sits in a collapsed section.
	const review = apply.getByRole("button", { name: "Review Changes" })
	await expect(review).toBeDisabled()
	const ingestion = apply.getByRole("button", {
		name: "Toggle sync and ingestion settings",
	})
	await expect(ingestion).toHaveAttribute("aria-expanded", "false")
	await ingestion.click()
	await expect(ingestion).toHaveAttribute("aria-expanded", "true")
	await apply.getByRole("radio", { name: mode, exact: true }).check()
	await expect(review).toBeEnabled()
	await review.click()
	const summary = page.getByRole("dialog").filter({
		hasText: "Summary of Changes",
	})
	await expect(summary).toBeVisible()
	await summary.getByRole("button", { name: "Apply Changes" }).click()
	const configured = page.getByText(
		`${streams.length} streams configured successfully`,
	)
	await expect(configured).toBeVisible()
	await expect(configured).toBeHidden({ timeout: 30_000 })
	for (const stream of streams) {
		await openStream(page, stream)
		await streamTab(page, "Config")
		await expect(syncModeRadio(page, mode)).toBeChecked()
	}
}
