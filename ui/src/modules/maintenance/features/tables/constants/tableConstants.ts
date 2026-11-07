import type { FilterKey, LatestRunSlotKey } from "../types"

export const PAGE_SIZE = 6

export const RUN_STATUS = {
	SUCCESS: "SUCCESS",
	RUNNING: "RUNNING",
	FAILED: "FAILED",
	SKIPPED: "SKIPPED",
	CLOSED: "CLOSED",
} as const

export const RUN_TYPE = {
	MINOR: "MINOR",
	MAJOR: "MAJOR",
	FULL: "FULL",
	/** Snapshot expiry (cleanup) run */
	EXPIRY: "EXPIRY",
} as const

export const RUN_TYPE_LABEL = {
	LITE: "Lite",
	MEDIUM: "Medium",
	FULL: "Full",
	EXPIRY: "Expire",
} as const

/** Run History type filter; values double as the `type` query param of the runs endpoint. */
export const RUN_TYPE_FILTER = {
	ALL: "all",
	OPTIMIZATION: "optimization",
	EXPIRY: "expiry",
} as const

export const RUN_TYPE_FILTER_OPTIONS = [
	{ key: RUN_TYPE_FILTER.ALL, label: "All Types" },
	{ key: RUN_TYPE_FILTER.OPTIMIZATION, label: "Optimization" },
	{ key: RUN_TYPE_FILTER.EXPIRY, label: "Expire" },
] as const

export const RUN_STATUS_FILTER_OPTIONS = [
	{ key: "all", label: "All Runs" },
	{ key: "failed", label: "Failed Runs" },
] as const

export const TABLE_FILTER_OPTIONS: ReadonlyArray<{
	key: FilterKey
	label: string
}> = [
	{ key: "all", label: "All tables" },
	{ key: "olake", label: "OLake Ingested" },
	{ key: "external", label: "Imported Tables" },
]

/** Fusion table properties behind the configure modal and the Status toggle. */
export const MINOR_CRON_PROPERTY_KEY = "self-optimizing.minor.trigger.cron"
export const MAJOR_CRON_PROPERTY_KEY = "self-optimizing.major.trigger.cron"
export const FULL_CRON_PROPERTY_KEY = "self-optimizing.full.trigger.cron"
export const TARGET_FILE_SIZE_PROPERTY_KEY = "self-optimizing.target-size"
export const OPTIMIZATION_ENABLED_PROPERTY_KEY = "self-optimizing.enabled"
export const SNAPSHOT_EXPIRE_ENABLED_PROPERTY_KEY = "snapshot-expire.enabled"
export const SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY = "snapshot-expire.trigger.cron"
export const SNAPSHOT_KEEP_DURATION_PROPERTY_KEY = "snapshot.keep.duration"
export const SNAPSHOT_KEEP_MIN_COUNT_PROPERTY_KEY = "snapshot.keep.min-count"

/** Last Run Status slots: the three compaction levels plus snapshot expiry. */
export const LATEST_RUN_SLOTS: ReadonlyArray<{
	key: LatestRunSlotKey
	tag: "L" | "M" | "F" | "E"
	name: string
}> = [
	{ key: "minor", tag: "L", name: RUN_TYPE_LABEL.LITE },
	{ key: "major", tag: "M", name: RUN_TYPE_LABEL.MEDIUM },
	{ key: "full", tag: "F", name: RUN_TYPE_LABEL.FULL },
	{ key: "expiry", tag: "E", name: RUN_TYPE_LABEL.EXPIRY },
]
