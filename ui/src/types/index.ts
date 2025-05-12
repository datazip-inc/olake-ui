import { CheckboxChangeEvent } from "antd/es/checkbox"

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

export type UnknownObject = {
	[key: string]: unknown | UnknownObject
}

export type StreamData = {
	sync_mode: "full_refresh" | "cdc"
	skip_nested_flattening?: boolean
	cursor_field?: string[]
	destination_sync_mode: string
	selected_columns: string[] | null
	sort_key: string[] | null
	stream: {
		name: string
		namespace?: string
		json_schema: UnknownObject
		type_schema?: {
			properties: Record<
				string,
				{
					type: string | string[]
					format?: string
					properties?: Record<string, any>
				}
			>
		}
		supported_sync_modes?: ["full_refresh"] | ["full_refresh", "incremental"]
		source_defined_cursor?: boolean
		default_cursor_field?: string[]
		[key: string]: unknown
	}
}

export type StreamPanelProps = {
	stream: any
	activeStreamData: any | null
	setActiveStreamData: (stream: any) => void
	onStreamSelect?: (streamName: string, checked: boolean) => void
	isSelected: boolean
}

export type StreamHeaderProps = {
	stream: any
	toggle: (e: CheckboxChangeEvent) => void
	checked: boolean
	activeStreamData: any | null
	setActiveStreamData: (stream: any) => void
}

export type StreamConfigurationProps = {
	stream: StreamData
	onSyncModeChange?: (
		streamName: string,
		namespace: string,
		mode: "full_refresh" | "cdc",
	) => void
	useDirectForms?: boolean
}

export type JobCreationSteps = "source" | "destination" | "schema" | "config"

export interface Entity {
	id: number
	name: string
	type: string
	version: string
	config: string
	created_at: string
	updated_at: string
	created_by: string
	updated_by: string
	jobs: EntityJob[]
}

export interface EntityJob {
	activate: boolean
	dest_name?: string
	source_name?: string
	dest_type?: string
	source_type?: string
	id: number
	job_name: string
	last_run_state: string
	last_runtime: string
	name: string
}

export interface EntityBase {
	name: string
	type: string
	version: string
	config: string
}

export interface APIResponse<T> {
	success: boolean
	message: string
	data: T
}

export interface EntityTestResponse {
	type: string
	version: string
	config: string
}

export interface LoginArgs {
	username: string
	password: string
}
