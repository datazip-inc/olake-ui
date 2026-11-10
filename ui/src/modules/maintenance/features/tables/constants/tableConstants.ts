import type {
	FilterKey,
	LatestRunSlotKey,
	RunHistoryFilter,
	RunTypeFilter,
} from "../types"

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
	EXPIRY: "EXPIRY",
} as const

export const RUN_TYPE_LABEL = {
	LITE: "Lite",
	MEDIUM: "Medium",
	FULL: "Full",
	EXPIRY: "Expire",
} as const

// Run History type filter, also the `type` query param of the run history URL.
export const RUN_TYPE_FILTER = {
	ALL: "all",
	OPTIMIZATION: "optimization",
	CLEANUP: "cleanup",
} as const

// Fusion process types each filter selects (its `type` param, comma-separated); undefined means all.
export const RUN_TYPE_FILTER_PROCESS_TYPES: Record<
	RunTypeFilter,
	string | undefined
> = {
	[RUN_TYPE_FILTER.ALL]: undefined,
	[RUN_TYPE_FILTER.OPTIMIZATION]: [
		RUN_TYPE.MINOR,
		RUN_TYPE.MAJOR,
		RUN_TYPE.FULL,
	].join(","),
	[RUN_TYPE_FILTER.CLEANUP]: RUN_TYPE.EXPIRY,
}

export const RUN_TYPE_FILTER_OPTIONS: Array<{
	key: RunTypeFilter
	label: string
}> = [
	{ key: RUN_TYPE_FILTER.ALL, label: "All Types" },
	{ key: RUN_TYPE_FILTER.OPTIMIZATION, label: "Optimization" },
	{ key: RUN_TYPE_FILTER.CLEANUP, label: "Cleanup" },
]

export const RUN_STATUS_FILTER_OPTIONS: Array<{
	key: RunHistoryFilter
	label: string
}> = [
	{ key: "all", label: "All Runs" },
	{ key: "failed", label: "Failed Runs" },
]

// Fusion process status each status filter selects; undefined means no filter.
export const RUN_STATUS_FILTER_STATUSES: Record<
	RunHistoryFilter,
	string | undefined
> = {
	all: undefined,
	failed: RUN_STATUS.FAILED,
}

// Query params the run history page keeps in the URL so filters survive a reload.
export const RUN_HISTORY_QUERY_PARAM = {
	STATUS: "status",
	TYPE: "type",
} as const

export const TABLE_FILTER_OPTIONS: Array<{
	key: FilterKey
	label: string
}> = [
	{ key: "all", label: "All tables" },
	{ key: "olake", label: "OLake Ingested" },
	{ key: "external", label: "Imported Tables" },
]

// Fusion table properties behind the configure modal and the Status toggle.
export const MINOR_CRON_PROPERTY_KEY = "self-optimizing.minor.trigger.cron"
export const MAJOR_CRON_PROPERTY_KEY = "self-optimizing.major.trigger.cron"
export const FULL_CRON_PROPERTY_KEY = "self-optimizing.full.trigger.cron"
export const TARGET_FILE_SIZE_PROPERTY_KEY = "self-optimizing.target-size"
export const OPTIMIZATION_ENABLED_PROPERTY_KEY = "self-optimizing.enabled"
export const SNAPSHOT_EXPIRE_ENABLED_PROPERTY_KEY = "snapshot-expire.enabled"
export const SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY = "snapshot-expire.trigger.cron"
export const SNAPSHOT_KEEP_DURATION_PROPERTY_KEY = "snapshot.keep.duration"
export const SNAPSHOT_KEEP_MIN_COUNT_PROPERTY_KEY = "snapshot.keep.min-count"

// Last Run Status slots: the three compaction levels plus snapshot expiry.
export const LATEST_RUN_SLOTS: Array<{
	key: LatestRunSlotKey
	tag: "L" | "M" | "F" | "C"
	name: string
}> = [
	{ key: "minor", tag: "L", name: RUN_TYPE_LABEL.LITE },
	{ key: "major", tag: "M", name: RUN_TYPE_LABEL.MEDIUM },
	{ key: "full", tag: "F", name: RUN_TYPE_LABEL.FULL },
	{ key: "expiry", tag: "C", name: "Cleanup" },
]
