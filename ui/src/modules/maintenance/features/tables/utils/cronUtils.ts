import parser from "cron-parser"

import type { CronConfigOption } from "../types"

export const getCronFromConfig = (config: CronConfigOption) => {
	if (config.frequency === "never") return ""
	return config.frequency === "custom"
		? config.customCron.trim()
		: config.frequency
}

export const isValidCronExpression = (cron: string): boolean => {
	const parts = cron.trim().split(" ")
	if (parts.length !== 5) return false

	try {
		parser.parse(cron)
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
			const interval = parser.parse(cron, {
				currentDate: new Date(),
				tz: "UTC",
			})
			const next = interval.next().toDate()
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
		const interval = parser.parse(normalizedCron, {
			currentDate: new Date(),
			tz: "UTC",
		})

		const runs: string[] = []
		for (let index = 0; index < 3; index += 1) {
			const run = getParsedDate(interval.next().toDate()).replace(
				/:\d{2} GMT$/,
				" GMT",
			)
			runs.push(run)
		}
		return runs
	} catch {
		return []
	}
}
