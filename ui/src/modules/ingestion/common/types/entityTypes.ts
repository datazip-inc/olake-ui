import { TestConnectionStatus } from "@/common/types"

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
	destination_name?: string
	source_name?: string
	destination_type?: string
	source_type?: string
	id: number
	last_run_state: string
	last_run_time?: string
	name: string
}
export interface EntityBase {
	name: string
	type: string
	version: string
	config: string
}
export interface EntityTestRequest {
	type: string
	version: string
	config: string
}
export interface EntityTestResponse {
	connection_result: {
		message: string
		status: TestConnectionStatus
	}
	logs: LogEntry[]
}

export type EntityType = "source" | "destination"

export interface EntityEditModalProps {
	entityType: EntityType
	open: boolean
	jobs: EntityJob[]
	onConfirm: () => Promise<void>
	onCancel: () => void
}

export type EntitySavedModalType =
	| "source"
	| "destination"
	| "config"
	| "streams"

export interface EntitySavedModalProps {
	open: boolean
	onClose: () => void
	type: EntitySavedModalType
	onComplete?: () => void
	fromJobFlow: boolean
	entityName?: string
}

export interface LogEntry {
	level: string
	time: string
	message: string
}
