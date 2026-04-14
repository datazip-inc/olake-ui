import { LogEntry } from "@/common/types"

import { TaskLogEntry } from "../types"

export const mapLogEntriesToTaskLogEntries = (
	logs: LogEntry[],
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
		}
	})
}
