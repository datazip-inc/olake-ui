import { LogEntry } from "@/common/types"

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
	advanced_settings?: AdvancedSettings | null
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
	advanced_settings?: AdvancedSettings | null
}
export interface JobTask {
	runtime: string
	start_time: string
	status: string
	file_path: string
	job_type: JobType
}

// Sync logs come from the connector run, worker logs from the worker that ran it.
export type TaskLogSource = "all" | "sync" | "worker"

export interface TaskLogApiEntry extends LogEntry {
	source?: Exclude<TaskLogSource, "all">
	category?: string
	tip?: boolean
}

export interface TaskLogsResponse {
	logs: TaskLogApiEntry[]
	older_cursor: string
	newer_cursor: string
	has_more_older: boolean
	has_more_newer: boolean
}

export enum TaskLogsDirection {
	Older = "older",
	Newer = "newer",
}

export interface TaskLogsPaginationParams {
	source: TaskLogSource
	cursor: string
	limit: number
	direction: TaskLogsDirection
}

export interface TaskLogEntry {
	level: string
	date: string
	time: string
	message: string
	source: Exclude<TaskLogSource, "all">
	tip?: boolean
}

export type JobCreationSteps = "config" | "streams"

export type JobStatus = "active" | "inactive" | "saved" | "failed"

/* job draft persisted in localStorage under "savedJobs" */
export interface SavedJobDraft {
	id: string
	name: string
	source: {
		name: string
		type: string
		id?: number
	}
	destination: {
		name: string
		type: string
		id?: number
	}
	streams_config: string
	frequency: string
	advanced_settings: Record<string, any> | null
}

export interface JobTableProps {
	jobs: (Job | SavedJobDraft)[]
	loading: boolean
	refreshLoading?: boolean
	jobType: JobStatus
	onRefresh: () => void
	onSync: (id: string) => void
	onEdit: (id: string) => void
	onPause: (id: string, checked: boolean) => void
	onDelete: (id: string) => void
	onCancelJob: (id: string) => void
}
export interface AdvancedSettings {
	max_discover_threads?: number | null
	// Derived from the streams config, never user-set: true when any selected
	// stream upserts with positional deletes.
	index_required?: boolean
	// Engines that will read the destination tables.
	target_query_engines?: string[]
}

export interface JobConfigurationProps {
	stepNumber?: number
	stepTitle?: string
}
export enum JobType {
	Sync = "sync",
	ClearDestination = "clear",
}

export interface CronParseResult {
	frequency: string
	selectedTime?: string
	selectedAmPm?: "AM" | "PM"
	selectedDay?: string
	customCronExpression?: string
}

export type FilterButtonProps = {
	filter: string
	selectedFilters: string[]
	setSelectedFilters: (filters: string[]) => void
}
