import type { CleanupConfigFormModel } from "../types"

/** Fusion table properties behind the Cleanup tab. */
export const SNAPSHOT_EXPIRE_ENABLED_PROPERTY_KEY = "snapshot-expire.enabled"
export const SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY = "snapshot-expire.trigger.cron"
export const SNAPSHOT_KEEP_DURATION_PROPERTY_KEY = "snapshot.keep.duration"
export const SNAPSHOT_KEEP_MIN_COUNT_PROPERTY_KEY = "snapshot.keep.min-count"

// Fusion defaults: every 12 hours, keep 7 days, keep at least 30 snapshots.
export const CLEANUP_DEFAULT_TRIGGER_INTERVAL = "0 */12 * * *"
export const DEFAULT_CLEANUP_OLDER_THAN_DAYS = 7
export const DEFAULT_MIN_SNAPSHOTS_TO_KEEP = 30

/** Form defaults when a table has no saved cleanup config: the feature starts switched on. */
export const DEFAULT_CLEANUP_CONFIG: CleanupConfigFormModel = {
	expireSnapshots: true,
	olderThanDays: DEFAULT_CLEANUP_OLDER_THAN_DAYS,
	minSnapshotsToKeep: DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	cron: { frequency: CLEANUP_DEFAULT_TRIGGER_INTERVAL, customCron: "" },
}

export const MAINTENANCE_TAB = {
	OPTIMIZE: "optimize",
	CLEANUP: "cleanup",
} as const

export const CLEANUP_DESCRIPTION =
	"Removes old snapshots from table history. Data files are not deleted in this version."
