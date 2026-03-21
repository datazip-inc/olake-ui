export type RunLogLevel =
	| "DEBUG"
	| "INFO"
	| "WARNING"
	| "WARN"
	| "ERROR"
	| "FATAL"

export interface RunLogEntry {
	id: string
	date: string
	time: string
	level: RunLogLevel
	message: string
}

export interface RunLogSource {
	key: string
	label: string
	hasError: boolean
}

// API Response Types
export interface ProcessLogEntry {
	level: RunLogLevel
	time: string
	processId: string
	taskId: string
	logger: string
	message: string
}

export interface ProcessDriverLog {
	exists: boolean
	content: ProcessLogEntry[]
}

export interface ProcessTaskLog {
	taskId: string
	exists: boolean
	content: ProcessLogEntry[]
}

export interface GetProcessLogsApiResponse {
	processId: string
	exists: boolean
	driverLog: ProcessDriverLog
	taskLogs: ProcessTaskLog[]
}
