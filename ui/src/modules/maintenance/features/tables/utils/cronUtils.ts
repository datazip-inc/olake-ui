import { Cron } from "croner"

import type { CronConfigOption } from "../types"

export const getCronFromConfig = (config: CronConfigOption) => {
	if (config.frequency === "never") return ""
	return config.frequency === "custom"
		? config.customCron.trim()
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

const getParsedDate = (value: Date) => value.toUTCString()

export const getEarliestNextRun = (
	configs: CronConfigOption[],
): string | undefined => {
	let earliest: Date | null = null

	for (const config of configs) {
		const cron = getCronFromConfig(config).trim()
		const cronError = isValidCronExpression(cron)
		if (!cron || cronError) continue

		try {
			const next = new Cron(cron, { timezone: "UTC" }).nextRun()
			if (!next) continue
			if (!earliest || next < earliest) {
				earliest = next
			}
		} catch {
			// skip invalid
		}
	}

	if (!earliest) return undefined
	return getParsedDate(earliest).replace(/:\d{2} GMT$/, " GMT")
}

export const getNextRuns = (cron: string): string[] => {
	const normalizedCron = cron.trim()
	const cronError = isValidCronExpression(normalizedCron)
	if (!normalizedCron || cronError) return []

	try {
		return new Cron(normalizedCron, { timezone: "UTC" })
			.nextRuns(3)
			.map(run => getParsedDate(run).replace(/:\d{2} GMT$/, " GMT"))
	} catch {
		return []
	}
}
