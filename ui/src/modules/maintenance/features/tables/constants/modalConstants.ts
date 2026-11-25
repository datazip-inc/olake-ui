import type { CompactionCronKey, MaintenanceTabKey } from "../types"
import {
	FULL_CUSTOM_DEFAULT_CRON,
	LITE_CUSTOM_DEFAULT_CRON,
	MEDIUM_CUSTOM_DEFAULT_CRON,
} from "./cronConstants"
import { RUN_TYPE_LABEL } from "./tableConstants"

export const DEFAULT_TABLE_MODAL_STYLES = {
	content: {
		padding: 0,
		overflow: "hidden",
		borderRadius: 20,
	},
	body: {
		padding: 0,
	},
} as const

// Single edits one table's saved config; bulk writes the same form to many tables.
export const CONFIGURE_MODE = {
	SINGLE: "single",
	BULK: "bulk",
} as const

export const MAINTENANCE_TAB = {
	OPTIMIZE: "optimize",
	CLEANUP: "cleanup",
} as const

export const MAINTENANCE_TAB_ITEMS: Array<{
	key: MaintenanceTabKey
	label: string
}> = [
	{ key: MAINTENANCE_TAB.OPTIMIZE, label: "Optimize" },
	{ key: MAINTENANCE_TAB.CLEANUP, label: "Cleanup" },
]

// The three compaction schedules of the Optimize tab, in display order.
export const OPTIMIZE_SCHEDULES: Array<{
	key: CompactionCronKey
	title: string
	tooltip: string
	// Cron used when the Custom box is left empty.
	customDefault: string
}> = [
	{
		key: "minorCron",
		title: RUN_TYPE_LABEL.LITE,
		tooltip:
			"Converts equality deletes to position deletes and merge small files",
		customDefault: LITE_CUSTOM_DEFAULT_CRON,
	},
	{
		key: "majorCron",
		title: RUN_TYPE_LABEL.MEDIUM,
		tooltip: "Applies deletes and merges data files",
		customDefault: MEDIUM_CUSTOM_DEFAULT_CRON,
	},
	{
		key: "fullCron",
		title: RUN_TYPE_LABEL.FULL,
		tooltip:
			"Rewrites all data and applies deletes, creating files of the target size.",
		customDefault: FULL_CUSTOM_DEFAULT_CRON,
	},
]
