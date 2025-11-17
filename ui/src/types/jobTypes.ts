import { EntityJob, LogEntry } from "./entityTypes"

export interface Job {
	id: number
	name: string
	source: {
		id?: number
		name: string
		type: string
		version: string
		config: string
	}
	destination: {
		id?: number
		name: string
		type: string
		version: string
		config: string
	}
	streams_config: string
	frequency: string
	last_run_type: JobType
	last_run_state: string
	last_run_time: string
	created_at: string
	updated_at: string
	created_by: string
	updated_by: string
	activate: boolean
}
export interface JobBase {
	name: string
	source: {
		id?: number
		name: string
		type: string
		version: string
		config: string
	}
	destination: {
		id?: number
		name: string
		type: string
		version: string
		config: string
	}
	frequency: string
	streams_config: string
	difference_streams?: string
	activate?: boolean
}
export interface JobTask {
	runtime: string
	start_time: string
	status: string
	file_path: string
	job_type: JobType
}
export interface TaskLog {
	logs: LogEntry[]
	cursor: number
	hasMore: boolean
}

export interface TaskLogResponse {
	logs: LogEntry[]
	cursor: number
	has_more: boolean
}
export type JobCreationSteps = "config" | "source" | "destination" | "streams"

export type JobStatus = "active" | "inactive" | "saved" | "failed"

export interface JobTableProps {
	jobs: Job[]
	loading: boolean
	jobType: JobStatus
	onSync: (id: string) => void
	onEdit: (id: string) => void
	onPause: (id: string, checked: boolean) => void
	onDelete: (id: string) => void
	onCancelJob: (id: string) => void
}

export interface JobConfigurationProps {
	jobName: string
	setJobName: React.Dispatch<React.SetStateAction<string>>
	cronExpression?: string
	setCronExpression: React.Dispatch<React.SetStateAction<string>>
	stepNumber?: number
	stepTitle?: string
	jobNameFilled?: boolean
}

export interface JobConnectionProps {
	sourceType: string
	destinationType: string
	jobName: string
	remainingJobs?: number
	jobs: EntityJob[]
}

export enum JobType {
	Sync = "sync",
	ClearDestination = "clear",
}
