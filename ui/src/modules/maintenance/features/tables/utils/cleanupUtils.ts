import {
	CLEANUP_DEFAULT_TRIGGER_INTERVAL,
	DEFAULT_CLEANUP_CONFIG,
	DEFAULT_CLEANUP_OLDER_THAN_DAYS,
	DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
} from "../constants"
import type {
	CleanupConfigFormModel,
	TableCronApiModel,
	UpdateTableCronApiRequest,
} from "../types"
import { getCronFromConfig } from "./cronUtils"
import { mapTriggerIntervalToCronConfigOption } from "./tableUtils"

const MINUTES_PER_DAY = 24 * 60
const UNIT_TO_MINUTES: Record<string, number> = {
	d: MINUTES_PER_DAY,
	h: 60,
	min: 1,
	s: 1 / 60,
}

/**
 * Days in a Fusion duration ("7d", "36h", "90min", "7"; a bare number is days),
 * rounded up to a whole day. Unparsable input falls back to the default.
 */
export const parseKeepDurationToDays = (raw?: string): number => {
	const match = (raw ?? "").trim().match(/^(\d+)\s*(d|h|min|s)?$/)
	if (!match) return DEFAULT_CLEANUP_OLDER_THAN_DAYS
	const minutes = Number(match[1]) * UNIT_TO_MINUTES[match[2] ?? "d"]
	return Math.max(1, Math.ceil(minutes / MINUTES_PER_DAY))
}

/** Fusion duration string for the form's day count. */
export const formatKeepDuration = (days: number) => `${days}d`

export const mapTableCronApiModelToCleanupConfig = (
	config: TableCronApiModel,
): CleanupConfigFormModel => ({
	// Never configured: start switched on, like the Optimization switch.
	expireSnapshots: config.snapshotExpiryConfigured
		? config.snapshotExpiryEnabled
		: true,
	olderThanDays: parseKeepDurationToDays(config.snapshotKeepDuration),
	minSnapshotsToKeep:
		config.snapshotKeepMinCount && config.snapshotKeepMinCount > 0
			? config.snapshotKeepMinCount
			: DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	cron: mapTriggerIntervalToCronConfigOption(
		config.snapshotExpiryCron?.trim() ?? CLEANUP_DEFAULT_TRIGGER_INTERVAL,
	),
})

/** Request fields that write the cleanup form to the table properties. */
export const mapCleanupConfigToApiRequest = (
	config: CleanupConfigFormModel,
): Pick<
	UpdateTableCronApiRequest,
	| "snapshot_expiry_enabled"
	| "snapshot_expiry_cron"
	| "snapshot_keep_duration"
	| "snapshot_keep_min_count"
> => ({
	snapshot_expiry_enabled: String(config.expireSnapshots),
	snapshot_expiry_cron: getCronFromConfig(config.cron),
	snapshot_keep_duration: formatKeepDuration(config.olderThanDays),
	snapshot_keep_min_count: config.minSnapshotsToKeep,
})

/** Request fields that switch cleanup on with defaults for a never-configured table. */
export const defaultCleanupApiRequest = () =>
	mapCleanupConfigToApiRequest(DEFAULT_CLEANUP_CONFIG)
