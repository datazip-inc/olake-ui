export interface Job {
	id: string
	name: string
	status: "active" | "inactive" | "saved" | "failed"
	source: string
	destination: string
	createdAt: Date
	lastSync?: string
	lastSyncStatus?: "success" | "failed" | "running"
}

export interface JobBasic {
	source: string
	destination: string
	jobName: string
}

export interface JobHistory {
	id: string
	jobId: string
	startTime: string
	runtime: string
	status: "success" | "failed" | "running" | "scheduled"
}

export interface JobLog {
	date: string
	time: string
	level: "debug" | "info" | "warning" | "error"
	message: string
}
export interface SourceJob {
	id: string
	name: string
	state: string
	lastRuntime: string
	lastRuntimeStatus: string
	destination: {
		name: string
		type: string
		config: any
	}
	paused: boolean
}

export interface DestinationJob {
	id: string
	name: string
	state: string
	lastRuntime: string
	lastRuntimeStatus: string
	source: string
	paused: boolean
}
export type JobCreationSteps = "source" | "destination" | "schema" | "config"

