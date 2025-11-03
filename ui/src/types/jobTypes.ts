import { EntityJob } from "./entityTypes"

export interface Job {
	id: number
	name: string
	source: {
		name: string
		type: string
		version: string
		config: string
	}
	destination: {
		name: string
		type: string
		version: string
		config: string
	}
	streams_config: string
	frequency: string
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
	activate?: boolean
}
export interface JobTask {
	runtime: string
	start_time: string
	status: string
	file_path: string
}
export interface TaskLog {
	level: string
	message: string
	time: string
}
export type JobCreationSteps = "config" | "source" | "destination" | "streams"

export type JobType = "active" | "inactive" | "saved" | "failed"

export interface JobTableProps {
	jobs: Job[]
	loading: boolean
	jobType: JobType
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
