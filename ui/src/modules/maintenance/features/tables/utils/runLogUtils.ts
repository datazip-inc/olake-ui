import { DRIVER_SOURCE_KEY } from "../constants"
import type {
	GetProcessLogsApiResponse,
	ProcessLogEntry,
	RunLogEntry,
	RunLogSource,
} from "../types"

// Converts a single ProcessLogEntry into a RunLogEntry, parsing the ISO timestamp into separate date and time display strings.
const mapProcessEntry = (
	entry: ProcessLogEntry,
	index: number,
	sourceKey: string,
): RunLogEntry => {
	let date = ""
	let time = ""

	if (entry.time) {
		const d = new Date(entry.time)
		date = d.toLocaleDateString()
		time = d.toLocaleTimeString("en-US", {
			timeZone: "UTC",
			hour12: false,
		})
	}

	return {
		id: `${sourceKey}-${index}`,
		date,
		time,
		level: entry.level.toLowerCase() as RunLogEntry["level"],
		message: entry.message,
	}
}

// Transforms the raw process logs API response into three domain objects: driver logs, task log sources for the sidebar, and a keyed map of log entries per source.
export const mapProcessLogsResponse = (response: GetProcessLogsApiResponse) => {
	const { driverLog, taskLogs } = response

	const driverLogs: RunLogEntry[] = (driverLog.content ?? []).map(
		(entry, index) => mapProcessEntry(entry, index, DRIVER_SOURCE_KEY),
	)

	// Each task log maps to a sidebar source entry keyed by task-{taskId}.
	const taskSources: RunLogSource[] = (taskLogs ?? []).map(task => ({
		key: `task-${task.taskId}`,
		label: `Subtask ${task.taskId}`,
		hasError: (task.content ?? []).some(
			entry => entry.level.toLowerCase() === "error",
		),
	}))

	const logsBySource: Record<string, RunLogEntry[]> = {
		[DRIVER_SOURCE_KEY]: driverLogs,
		...Object.fromEntries(
			(taskLogs ?? []).map(task => [
				`task-${task.taskId}`,
				(task.content ?? []).map((entry, index) =>
					mapProcessEntry(entry, index, `task-${task.taskId}`),
				),
			]),
		),
	}

	return { driverLogs, taskSources, logsBySource }
}
