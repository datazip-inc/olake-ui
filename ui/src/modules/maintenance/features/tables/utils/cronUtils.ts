import { Cron } from "croner"

import type { CronConfigOption } from "../types"

export const getCronFromConfig = (config: CronConfigOption) => {
	if (config.frequency === "never") return ""
	return config.frequency === "custom"
		? config.customCron.trim()
		: config.frequency
}

export const isValidCronExpression = (cron: string): boolean => {
	const parts = cron.trim().split(" ")
	// Optimization supports 5 field cron expressions only
	if (parts.length !== 5) return false

	try {
		new Cron(cron)
		return true
	} catch {
		return false
	}
}

const getParsedDate = (value: Date) => value.toUTCString()

export const getEarliestNextRun = (
	configs: CronConfigOption[],
): string | undefined => {
	let earliest: Date | null = null

	for (const config of configs) {
		const cron = getCronFromConfig(config).trim()
		if (!cron || !isValidCronExpression(cron)) continue

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
	if (!normalizedCron || !isValidCronExpression(normalizedCron)) return []

	try {
		return new Cron(normalizedCron, { timezone: "UTC" })
			.nextRuns(3)
			.map(run => getParsedDate(run).replace(/:\d{2} GMT$/, " GMT"))
	} catch {
		return []
	}
}
