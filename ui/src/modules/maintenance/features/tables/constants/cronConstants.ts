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

// Defaults the UI writes when a feature is switched on while every one of its schedules is Never
// (or unset), from the configure modal or the Status toggle.
export const LITE_DEFAULT_TRIGGER_INTERVAL = "0 * * * *"
export const MEDIUM_DEFAULT_TRIGGER_INTERVAL = "0 */8 * * *"
export const FULL_DEFAULT_TRIGGER_INTERVAL = ""
export const CLEANUP_DEFAULT_TRIGGER_INTERVAL = "0 */12 * * *"
export const DEFAULT_TARGET_FILE_SIZE = 512

// Shown as the Custom box placeholder, and saved when the box is left empty.
export const LITE_CUSTOM_DEFAULT_CRON = "*/5 * * * *"
export const MEDIUM_CUSTOM_DEFAULT_CRON = "*/30 * * * *"
export const FULL_CUSTOM_DEFAULT_CRON = "0 * * * *"
export const CLEANUP_CUSTOM_DEFAULT_CRON = "0 * * * *"
