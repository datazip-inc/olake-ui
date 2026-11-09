import {
	CLEANUP_DEFAULT_TRIGGER_INTERVAL,
	DEFAULT_KEEP_DURATION_UNIT,
	DEFAULT_KEEP_DURATION_VALUE,
	DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
} from "../constants"
import type {
	CleanupConfigFormModel,
	KeepDuration,
	KeepDurationUnit,
	TableConfigApiModel,
	UpdateTableConfigApiRequest,
} from "../types"
import {
	getCronFromConfig,
	mapTriggerIntervalToCronConfigOption,
} from "./cronUtils"

const DEFAULT_KEEP_DURATION: KeepDuration = {
	value: DEFAULT_KEEP_DURATION_VALUE,
	unit: DEFAULT_KEEP_DURATION_UNIT,
}

// Fusion duration ("7d", "36h", "90min", "30s"; a bare number is days) as value + unit.
// Seconds are the one unit the picker does not offer, so they are rounded up to minutes.
// Unparsable input falls back to the default.
export const parseKeepDuration = (raw?: string): KeepDuration => {
	const match = (raw ?? "").trim().match(/^(\d+)\s*(d|h|min|s)?$/)
	if (!match) return DEFAULT_KEEP_DURATION
	const value = Number(match[1])
	const unit = match[2] ?? "d"
	if (unit === "s")
		return { value: Math.max(1, Math.ceil(value / 60)), unit: "min" }
	return { value, unit: unit as KeepDurationUnit }
}

// Fusion duration string for the form value, e.g. "36h".
export const formatKeepDuration = ({ value, unit }: KeepDuration) =>
	`${value}${unit}`

// Cleanup form from the table's properties. A never-configured table starts switched on.
export const mapTableConfigApiModelToCleanupConfig = (
	config: TableConfigApiModel,
): CleanupConfigFormModel => ({
	expireSnapshots: config.snapshotExpiryConfigured
		? config.snapshotExpiryEnabled
		: true,
	keepDuration: parseKeepDuration(config.snapshotKeepDuration),
	minSnapshotsToKeep:
		config.snapshotKeepMinCount && config.snapshotKeepMinCount > 0
			? config.snapshotKeepMinCount
			: DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	cron: mapTriggerIntervalToCronConfigOption(
		config.snapshotExpiryCron?.trim() ?? CLEANUP_DEFAULT_TRIGGER_INTERVAL,
	),
})

export const mapCleanupConfigToApiRequest = (
	config: CleanupConfigFormModel,
): Pick<
	UpdateTableConfigApiRequest,
	| "snapshot_expiry_enabled"
	| "snapshot_expiry_cron"
	| "snapshot_keep_duration"
	| "snapshot_keep_min_count"
> => ({
	snapshot_expiry_enabled: String(config.expireSnapshots),
	snapshot_expiry_cron: getCronFromConfig(config.cron),
	snapshot_keep_duration: formatKeepDuration(config.keepDuration),
	snapshot_keep_min_count: config.minSnapshotsToKeep,
})
