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
} as const

export const RUN_TYPE_LABEL = {
	LITE: "Lite",
	MEDIUM: "Medium",
	FULL: "Full",
} as const

export const MINOR_CRON_PROPERTY_KEY = "self-optimizing.minor.trigger.cron"
export const MAJOR_CRON_PROPERTY_KEY = "self-optimizing.major.trigger.cron"
export const FULL_CRON_PROPERTY_KEY = "self-optimizing.full.trigger.cron"

export const compactionSlots: Array<{
	key: "minor" | "major" | "full"
	tag: "L" | "M" | "F"
	name: string
}> = [
	{ key: "minor", tag: "L", name: RUN_TYPE_LABEL.LITE },
	{ key: "major", tag: "M", name: RUN_TYPE_LABEL.MEDIUM },
	{ key: "full", tag: "F", name: RUN_TYPE_LABEL.FULL },
]
