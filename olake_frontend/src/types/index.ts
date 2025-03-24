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

export interface Source {
	id: string
	name: string
	type: string
	status: "active" | "inactive" | "saved"
	createdAt: Date
}

export interface Destination {
	id: string
	name: string
	type: string
	status: "active" | "inactive" | "saved"
	createdAt: Date
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
	destination: string
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
