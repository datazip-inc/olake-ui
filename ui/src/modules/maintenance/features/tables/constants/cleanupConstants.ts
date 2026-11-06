import type { CleanupConfigFormModel } from "../types"
import { DEFAULT_CRON_CONFIG } from "./cronConstants"

export const DEFAULT_CLEANUP_OLDER_THAN_DAYS = 7
export const DEFAULT_MIN_SNAPSHOTS_TO_KEEP = 100

export const DEFAULT_CLEANUP_CONFIG: CleanupConfigFormModel = {
	expireSnapshots: false,
	olderThanDays: DEFAULT_CLEANUP_OLDER_THAN_DAYS,
	minSnapshotsToKeep: DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	cron: DEFAULT_CRON_CONFIG,
}

export const MAINTENANCE_TAB = {
	OPTIMIZE: "optimize",
	CLEANUP: "cleanup",
} as const

export const CLEANUP_DESCRIPTION =
	"Removes old snapshots from table history. Data files are not deleted in this version."
