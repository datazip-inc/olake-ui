import type { CleanupConfigFormModel, KeepDurationUnit } from "../types"

/** Fusion table properties behind the Cleanup tab. */
export const SNAPSHOT_EXPIRE_ENABLED_PROPERTY_KEY = "snapshot-expire.enabled"
export const SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY = "snapshot-expire.trigger.cron"
export const SNAPSHOT_KEEP_DURATION_PROPERTY_KEY = "snapshot.keep.duration"
export const SNAPSHOT_KEEP_MIN_COUNT_PROPERTY_KEY = "snapshot.keep.min-count"

// Fusion defaults: every 12 hours, keep 7 days, keep at least 30 snapshots.
export const CLEANUP_DEFAULT_TRIGGER_INTERVAL = "0 */12 * * *"
export const DEFAULT_KEEP_DURATION_VALUE = 7
export const DEFAULT_KEEP_DURATION_UNIT: KeepDurationUnit = "d"
export const DEFAULT_MIN_SNAPSHOTS_TO_KEEP = 30

/** Units Fusion accepts for snapshot.keep.duration, in the order the picker lists them. */
export const KEEP_DURATION_UNIT_OPTIONS: Array<{
	value: KeepDurationUnit
	label: string
}> = [
	{ value: "d", label: "Days" },
	{ value: "h", label: "Hours" },
	{ value: "min", label: "Minutes" },
]

/** Form defaults when a table has no saved cleanup config: the feature starts switched on. */
export const DEFAULT_CLEANUP_CONFIG: CleanupConfigFormModel = {
	expireSnapshots: true,
	keepDuration: {
		value: DEFAULT_KEEP_DURATION_VALUE,
		unit: DEFAULT_KEEP_DURATION_UNIT,
	},
	minSnapshotsToKeep: DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	cron: { frequency: CLEANUP_DEFAULT_TRIGGER_INTERVAL, customCron: "" },
}

export const MAINTENANCE_TAB = {
	OPTIMIZE: "optimize",
	CLEANUP: "cleanup",
} as const

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
