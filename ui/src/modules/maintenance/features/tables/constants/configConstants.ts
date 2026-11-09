import type {
	CleanupConfigFormModel,
	KeepDurationUnit,
	OptimizeConfigFormModel,
	TableConfigFormModel,
} from "../types"
import {
	CLEANUP_DEFAULT_TRIGGER_INTERVAL,
	DEFAULT_CRON_CONFIG,
	DEFAULT_TARGET_FILE_SIZE,
} from "./cronConstants"

// Fusion defaults: keep 7 days, keep at least 30 snapshots.
export const DEFAULT_KEEP_DURATION_VALUE = 7
export const DEFAULT_KEEP_DURATION_UNIT: KeepDurationUnit = "d"
export const DEFAULT_MIN_SNAPSHOTS_TO_KEEP = 30
// Fusion rejects a smaller snapshot.keep.min-count (TableConfigurations.SNAPSHOT_MIN_COUNT_LOWEST):
// OLake's pebble index can be invalidated when too few snapshots remain.
export const MIN_SNAPSHOTS_TO_KEEP_LOWEST = 10

// Units Fusion accepts for snapshot.keep.duration, in the order the picker lists them.
export const KEEP_DURATION_UNIT_OPTIONS: Array<{
	value: KeepDurationUnit
	label: string
}> = [
	{ value: "d", label: "Days" },
	{ value: "h", label: "Hours" },
	{ value: "min", label: "Minutes" },
]

// Form defaults when a table has no saved config: both features start switched on.
export const DEFAULT_OPTIMIZE_CONFIG: OptimizeConfigFormModel = {
	optimizationEnabled: true,
	minorCron: DEFAULT_CRON_CONFIG,
	majorCron: DEFAULT_CRON_CONFIG,
	fullCron: DEFAULT_CRON_CONFIG,
	targetFileSize: DEFAULT_TARGET_FILE_SIZE,
}

export const DEFAULT_CLEANUP_CONFIG: CleanupConfigFormModel = {
	expireSnapshots: true,
	keepDuration: {
		value: DEFAULT_KEEP_DURATION_VALUE,
		unit: DEFAULT_KEEP_DURATION_UNIT,
	},
	minSnapshotsToKeep: DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	cron: { frequency: CLEANUP_DEFAULT_TRIGGER_INTERVAL, customCron: "" },
}

export const DEFAULT_TABLE_CONFIG_FORM: TableConfigFormModel = {
	optimize: DEFAULT_OPTIMIZE_CONFIG,
	cleanup: DEFAULT_CLEANUP_CONFIG,
}
