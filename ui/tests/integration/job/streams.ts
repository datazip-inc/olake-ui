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

/** Types text into Search Streams, which keeps the streams whose name holds it. */
export const searchStreams = async (page: Page, text: string) => {
	await page.getByRole("searchbox", { name: "Search Streams" }).fill(text)
}

/**
 * Shows the streams the list's filter keeps: all, the selected ones, or the others. Selected and
 * Not Selected toggle on top of each other, and both on show every stream, so All tables, which
 * turns them off, goes first.
 */
export const filterStreams = async (
	page: Page,
	filter: "All tables" | "Selected" | "Not Selected",
) => {
	await page.getByRole("button", { name: "All tables", exact: true }).click()
	if (filter !== "All tables") {
		await page.getByRole("button", { name: filter, exact: true }).click()
	}
}

/** The checkbox that selects every stream of namespace at once. */
export const namespaceCheckbox = (page: Page, namespace: string) =>
	page
		.getByText(namespace, { exact: true })
		.locator("xpath=ancestor::div[.//input[@type='checkbox']][1]")
		.getByRole("checkbox")
		.first()

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

/**
 * On a stream of a job to S3 (parquet): there is no Ingestion Mode or Upsert Type to choose, and
 * no All Append or All Upsert, since parquet files are only ever appended to.
 */
export const expectAppendOnlyStream = async (page: Page) => {
	await expect(page.getByText("Sync mode:")).toBeVisible()
	for (const text of [
		"Ingestion Mode:",
		"Upsert Type:",
		"All Append",
		"All Upsert",
	]) {
		await expect(page.getByText(text, { exact: true })).toHaveCount(0)
	}
}

/**
 * Sets the open stream's cursor field, and its fallback cursor when given: the field an
 * Incremental sync reads instead when the cursor field is null.
 */
export const setCursor = async (
	page: Page,
	cursor: string,
	fallback?: string,
) => {
	const select = (label: string) =>
		page
			.getByText(label, { exact: true })
			.locator("xpath=ancestor::div[.//div[contains(@class,'ant-select')]][1]")
			.locator(".ant-select")
			.first()
	await pickOption(page, select("Cursor field:"), cursor)
	await expect(select("Cursor field:")).toContainText(cursor)
	if (fallback) {
		await page.getByRole("button", { name: "Add Fallback Cursor" }).click()
		await pickOption(page, select("Fallback Cursor:"), fallback)
		await expect(select("Fallback Cursor:")).toContainText(fallback)
	}
}

/** Sets the open stream's partition regex on its Partitioning tab. */
export const setPartitionRegex = async (page: Page, regex: string) => {
	await streamTab(page, "Partitioning")
	await page.getByPlaceholder("Enter your partition regex").fill(regex)
	await page.getByRole("button", { name: "Set Partition" }).click()
	await expect(page.getByText("Active partition regex:")).toBeVisible()
	await streamTab(page, "Config")
}

/** Ticks or unticks column on the open stream's Schema tab. */
export const setColumnSelected = async (
	page: Page,
	column: string,
	selected: boolean,
) => {
	await streamTab(page, "Schema")
	const checkbox = page
		.locator("div.border-b")
		.filter({
			has: page.locator("div.flex-1").getByText(column, { exact: true }),
		})
		.first()
		.getByRole("checkbox")
	await checkbox.setChecked(selected)
	await expect(checkbox).toBeChecked({ checked: selected })
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

/**
 * The open stream's Data Filter: a second condition waits for the first ("Please complete the
 * first filter before applying another."), then comes as Column II, and Remove takes it away.
 * The filter is turned off again.
 */
export const expectFilterConditionRules = async (
	page: Page,
	first: FilterCondition,
) => {
	await setSwitch(switchNear(page, "Data Filter"), true)
	const addCondition = page.getByRole("button", { name: "New Column filter" })
	const second = page.getByText("Column II", { exact: true })
	await addCondition.click()
	await expect(
		page.getByText("Please complete the first filter before applying another."),
	).toBeVisible()
	await expect(second).toHaveCount(0)

	await setDataFilter(page, [first], "AND")
	await addCondition.click()
	await expect(second).toBeVisible()
	await second
		.locator("xpath=ancestor::div[.//button[normalize-space()='Remove']][1]")
		.getByRole("button", { name: "Remove" })
		.click()
	await expect(second).toHaveCount(0)
	await setSwitch(switchNear(page, "Data Filter"), false)
}

/**
 * Bulk Configure over streams with no cursor field of one type in common: Full Refresh +
 * Incremental is disabled ("No common cursor fields across selected streams"), and the data
 * filter's second condition waits for the first. Closes the dialogs without applying anything.
 */
export const expectBulkRules = async (page: Page, streams: string[]) => {
	await page.getByRole("button", { name: "Bulk Configure" }).click()
	const bulk = page.getByRole("dialog").filter({ hasText: "Select Streams" })
	await expect(bulk).toBeVisible()
	for (const stream of streams) {
		await bulk.getByRole("checkbox", { name: stream, exact: true }).check()
	}
	await bulk.getByRole("button", { name: "Configure Streams" }).click()
	const apply = page.getByRole("dialog").filter({
		hasText: "Apply Configurations",
	})
	await expect(apply).toBeVisible()
	await apply
		.getByRole("button", { name: "Toggle sync and ingestion settings" })
		.click()
	const incremental = apply.getByRole("radio", {
		name: "Full Refresh + Incremental",
		exact: true,
	})
	await expect(incremental).toBeDisabled()
	await apply.getByText("Full Refresh + Incremental", { exact: true }).hover()
	await expect(
		page.getByRole("tooltip").filter({
			hasText: "No common cursor fields across selected streams",
		}),
	).toBeVisible()

	await setSwitch(
		apply
			.getByText("Data Filter", { exact: true })
			.first()
			.locator("xpath=ancestor::div[.//button[@role='switch']][1]")
			.getByRole("switch")
			.first(),
		true,
	)
	await apply.getByRole("button", { name: "New Column filter" }).click()
	await expect(
		page.getByText("Please complete the first filter before applying another."),
	).toBeVisible()
	await page.keyboard.press("Escape")
	await expect(apply).toBeHidden()
}
