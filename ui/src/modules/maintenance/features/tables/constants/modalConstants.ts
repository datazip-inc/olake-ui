import type { CompactionCronKey } from "../types"
import { RUN_TYPE_LABEL } from "./tableConstants"

/** Shared size of the configure and success modals so their inner layouts can agree. */
export const TABLE_MODAL_WIDTH = 696

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

/** Single edits one table's saved config; bulk writes the same form to many tables. */
export const CONFIGURE_MODE = {
	SINGLE: "single",
	BULK: "bulk",
} as const

export const MAINTENANCE_TAB = {
	OPTIMIZE: "optimize",
	CLEANUP: "cleanup",
} as const

export const MAINTENANCE_TAB_ITEMS = [
	{ key: MAINTENANCE_TAB.OPTIMIZE, label: "Optimize" },
	{ key: MAINTENANCE_TAB.CLEANUP, label: "Cleanup" },
]

/** All user-facing copy of the Optimize tab. */
export const OPTIMIZE_TEXT = {
	switchLabel: "Optimization",
	targetFileSizeLabel: "Target file size",
	targetFileSizeTooltip:
		"Desired file size after compaction; full optimization creates files of this size.",
} as const

/** The three compaction schedules of the Optimize tab, in display order. */
export const OPTIMIZE_SCHEDULES: ReadonlyArray<{
	key: CompactionCronKey
	title: string
	tooltip: string
}> = [
	{
		key: "minorCron",
		title: RUN_TYPE_LABEL.LITE,
		tooltip:
			"Converts equality deletes to position deletes and merge small files",
	},
	{
		key: "majorCron",
		title: RUN_TYPE_LABEL.MEDIUM,
		tooltip: "Applies deletes and merges data files",
	},
	{
		key: "fullCron",
		title: RUN_TYPE_LABEL.FULL,
		tooltip:
			"Rewrites all data and applies deletes, creating files of the target size.",
	},
]

/** All user-facing copy of the Cleanup tab. */
export const CLEANUP_TEXT = {
	description:
		"Removes old snapshots from table history. Data files are not deleted in this version.",
	switchLabel: "Expire Snapshots",
	keepDurationLabel: "Remove snapshots older than",
	keepDurationTooltip:
		"Snapshots older than this are expired on each cleanup run.",
	minKeepLabel: "Minimum snapshots to keep",
	minKeepTooltip:
		"Cleanup never drops below this many snapshots, even if they are older than the limit.",
	cronTitle: "Cleanup Cron",
	cronTooltip: "How often expired snapshots are removed.",
} as const
