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
