import type {
	CleanupConfigFormModel,
	CronConfigOption,
	KeepDurationUnit,
	OptimizeConfigFormModel,
	TableConfigFormModel,
} from "../types"
import {
	CLEANUP_DEFAULT_TRIGGER_INTERVAL,
	DEFAULT_TARGET_FILE_SIZE,
	FULL_DEFAULT_TRIGGER_INTERVAL,
	LITE_DEFAULT_TRIGGER_INTERVAL,
	MEDIUM_DEFAULT_TRIGGER_INTERVAL,
} from "./cronConstants"

// Preset dropdown state for a default trigger interval; an empty interval is "never".
const presetCron = (interval: string): CronConfigOption => ({
	frequency: interval || "never",
	customCron: "",
})

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
	minorCron: presetCron(LITE_DEFAULT_TRIGGER_INTERVAL),
	majorCron: presetCron(MEDIUM_DEFAULT_TRIGGER_INTERVAL),
	fullCron: presetCron(FULL_DEFAULT_TRIGGER_INTERVAL),
	targetFileSize: DEFAULT_TARGET_FILE_SIZE,
}

export const DEFAULT_CLEANUP_CONFIG: CleanupConfigFormModel = {
	expireSnapshots: true,
	keepDuration: {
		value: DEFAULT_KEEP_DURATION_VALUE,
		unit: DEFAULT_KEEP_DURATION_UNIT,
	},
	minSnapshotsToKeep: DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	cron: presetCron(CLEANUP_DEFAULT_TRIGGER_INTERVAL),
}

export const DEFAULT_TABLE_CONFIG_FORM: TableConfigFormModel = {
	optimize: DEFAULT_OPTIMIZE_CONFIG,
	cleanup: DEFAULT_CLEANUP_CONFIG,
}
