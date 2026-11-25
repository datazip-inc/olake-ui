import { Cron } from "croner"

import { KNOWN_CRON_TRIGGER_INTERVALS } from "../constants"
import type { CronConfigOption } from "../types"

// Cron string a schedule saves; an empty Custom box falls back to customDefault (its placeholder).
export const getCronFromConfig = (
	config: CronConfigOption,
	customDefault: string,
) => {
	if (config.frequency === "never") return ""
	return config.frequency === "custom"
		? config.customCron.trim() || customDefault
		: config.frequency
}

// Returns null if the cron expression is valid, or an error string if error.
export const isValidCronExpression = (cron: string): string | null => {
	const parts = cron.trim().split(" ")
	// Optimization supports 5 field cron expressions only
	if (parts.length !== 5)
		return `Cron expression must have 5 fields, but got ${parts.length}`

	try {
		new Cron(cron)
		return null
	} catch (error) {
		return error instanceof Error ? error.message : "Invalid cron expression"
	}
}

// Validation message for a schedule's cron, or null when it is "" (never) or a valid expression.
export const getCronError = (cron: string): string | null =>
	cron ? isValidCronExpression(cron) : null

// Maps a raw cron string to dropdown state: preset strings become a frequency value, anything else a custom entry.
export const mapTriggerIntervalToCronConfigOption = (
	triggerInterval: string,
): CronConfigOption => {
	if (triggerInterval === "") {
		return { frequency: "never", customCron: "" }
	}

	if (KNOWN_CRON_TRIGGER_INTERVALS.has(triggerInterval)) {
		return {
			frequency: triggerInterval,
			customCron: "",
		}
	}

	return {
		frequency: "custom",
		customCron: triggerInterval,
	}
}

// "Fri, 06 Nov 2026 10:42 GMT": UTC, seconds dropped.
const formatRunTime = (run: Date) =>
	run.toUTCString().replace(/:\d{2} GMT$/, " GMT")

// Next `count` fire times of one expression; empty for "never" or an invalid expression.
const getUpcomingRunDates = (cron: string, count: number): Date[] => {
	const normalized = cron.trim()
	if (!normalized || isValidCronExpression(normalized)) return []
	try {
		return new Cron(normalized, { timezone: "UTC" }).nextRuns(count)
	} catch {
		return []
	}
}

// Earliest `count` upcoming runs across several schedules' crons, merged and sorted.
// Two schedules firing in the same minute show as one upcoming time.
export const getNextRunsAcross = (crons: string[], count = 3): string[] => {
	const byTime = new Map<number, Date>()
	for (const cron of crons) {
		for (const run of getUpcomingRunDates(cron, count)) {
			byTime.set(run.getTime(), run)
		}
	}
	return [...byTime.values()]
		.sort((a, b) => a.getTime() - b.getTime())
		.slice(0, count)
		.map(formatRunTime)
}

export const getNextRuns = (cron: string): string[] =>
	getUpcomingRunDates(cron, 3).map(formatRunTime)
