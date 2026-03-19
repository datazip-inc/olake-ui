import type { CronConfigOption } from "../types"

export const CRON_FREQUENCY_OPTIONS = [
	{ label: "Never", value: "never" },
	{ label: "Every 30 min", value: "*/30 * * * *" },
	{ label: "Every hour", value: "0 * * * *" },
	{ label: "Every 8 hours", value: "0 */8 * * *" },
	{ label: "Every 12 hours", value: "0 */12 * * *" },
	{ label: "Every 24 hours", value: "0 0 * * *" },
	{ label: "Custom", value: "custom" },
]

export const KNOWN_CRON_TRIGGER_INTERVALS: Set<string> = new Set(
	CRON_FREQUENCY_OPTIONS.map(option => option.value).filter(
		value => value !== "custom",
	),
)

export const DEFAULT_CRON_CONFIG: CronConfigOption = {
	frequency: "0 0 * * *",
	customCron: "",
}
