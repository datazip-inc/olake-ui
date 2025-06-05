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
export type JobCreationSteps = "source" | "destination" | "schema" | "config"

export type JobType = "active" | "inactive" | "saved" | "failed"

export interface JobTableProps {
	jobs: Job[]
	loading: boolean
	jobType: JobType
	onSync: (id: string) => void
	onEdit: (id: string) => void
	onPause: (id: string, checked: boolean) => void
	onDelete: (id: string) => void
}

export interface JobConfigurationProps {
	jobName: string
	setJobName: React.Dispatch<React.SetStateAction<string>>
	replicationFrequency: string
	setReplicationFrequency: React.Dispatch<React.SetStateAction<string>>
	replicationFrequencyValue: string
	setReplicationFrequencyValue: React.Dispatch<React.SetStateAction<string>>
	schemaChangeStrategy: string
	setSchemaChangeStrategy: React.Dispatch<React.SetStateAction<string>>
	notifyOnSchemaChanges: boolean
	setNotifyOnSchemaChanges: React.Dispatch<React.SetStateAction<boolean>>
	stepNumber?: number | string
	stepTitle?: string
}
