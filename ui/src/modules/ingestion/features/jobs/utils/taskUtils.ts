import { TaskLogApiEntry, TaskLogEntry } from "../types"

export const mapLogEntriesToTaskLogEntries = (
	logs: TaskLogApiEntry[],
): TaskLogEntry[] => {
	return logs.map(log => {
		const level = log.level ?? ""
		const message = log.message ?? ""
		const timeRaw = log.time ?? ""

		let date = ""
		let time = ""

		if (timeRaw) {
			const dateObj = new Date(timeRaw)
			date = dateObj.toLocaleDateString()
			time = dateObj.toLocaleTimeString("en-US", {
				timeZone: "UTC",
				hour12: false,
			})
		}

		return {
			level,
			message,
			time,
			date,
			source: log.source ?? "sync",
			tip: log.tip,
		}
	})
}
