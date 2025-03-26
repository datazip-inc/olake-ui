import { CheckboxChangeEvent } from "antd/es/checkbox"

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
		supported_sync_modes?: ["full_refresh"] | ["full_refresh", "incremental"]
		source_defined_cursor?: boolean
		default_cursor_field?: string[]
		[key: string]: unknown
	}
}

export type StreamsCollapsibleListProps = {
	streamsToDisplay: StreamData[]
	allChecked: boolean
	handleToggleAllStreams: (e: CheckboxChangeEvent) => void
	activeStreamData: StreamData | null
	setActiveStreamData: (stream: StreamData) => void
	selectedStreams: string[]
	onStreamSelect: (streamName: string, checked: boolean) => void
}

export type StreamPanelProps = {
	stream: StreamData
	activeStreamData: StreamData | null
	setActiveStreamData: (stream: StreamData) => void
	onStreamSelect?: (streamName: string, checked: boolean) => void
	isSelected: boolean
}

export type StreamHeaderProps = {
	stream: StreamData
	toggle: (e: CheckboxChangeEvent) => void
	checked: boolean
	activeStreamData: StreamData | null
	setActiveStreamData: (stream: StreamData) => void
}

export type StreamConfigurationProps = {
	stream: StreamData
}
