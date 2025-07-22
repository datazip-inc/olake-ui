import type { CheckboxChangeEvent } from "antd/es/checkbox"
import type { UnknownObject } from "./index"

export type FilterOperator = "=" | "!=" | ">" | "<" | ">=" | "<="
export type LogicalOperator = "and" | "or"

export type FilterCondition = {
	columnName: string
	operator: FilterOperator
	value: string
}

export type MultiFilterCondition = {
	conditions: FilterCondition[]
	logicalOperator: LogicalOperator
}

export type StreamData = {
	sync_mode: "full_refresh" | "cdc" | "incremental"
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
		supported_sync_modes?: string[]
		source_defined_cursor?: boolean
		default_cursor_field?: string[]
		available_cursor_fields?: string[]
		cursor_field?: string
		source_defined_primary_key?: string[]
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
		mode: "full_refresh" | "cdc" | "incremental",
	) => void
	useDirectForms?: boolean
}

export interface SelectedStream {
	stream_name: string
	partition_regex: string
	normalization: boolean
	filter?: string
}

export interface StreamsDataStructure {
	selected_streams: {
		[namespace: string]: SelectedStream[]
	}
	streams: StreamData[]
}

export interface CombinedStreamsData {
	selected_streams: {
		[namespace: string]: SelectedStream[]
	}
	streams: StreamData[]
}

export interface SchemaConfigurationProps {
	selectedStreams:
		| string[]
		| {
				[namespace: string]: SelectedStream[]
		  }
		| CombinedStreamsData
	setSelectedStreams: React.Dispatch<
		React.SetStateAction<
			| string[]
			| {
					[namespace: string]: SelectedStream[]
			  }
			| CombinedStreamsData
		>
	>
	stepNumber?: number | string
	stepTitle?: string
	useDirectForms?: boolean
	sourceName: string
	sourceConnector: string
	sourceVersion: string
	sourceConfig: string
	initialStreamsData?: CombinedStreamsData
	fromJobEditFlow?: boolean
	jobId?: number
}

export interface ExtendedStreamConfigurationProps
	extends StreamConfigurationProps {
	onUpdate?: (stream: any) => void
	isSelected: boolean
	initialNormalization: boolean
	initialPartitionRegex: string
	initialFullLoadFilter?: string
	fromJobEditFlow?: boolean
	onNormalizationChange: (
		streamName: string,
		namespace: string,
		normalization: boolean,
	) => void
	onPartitionRegexChange: (
		streamName: string,
		namespace: string,
		partitionRegex: string,
	) => void
	onFullLoadFilterChange?: (
		streamName: string,
		namespace: string,
		filterValue: string,
	) => void
}

export interface GroupedStreamsCollapsibleListProps {
	groupedStreams: { [namespace: string]: StreamData[] }
	selectedStreams: {
		[namespace: string]: SelectedStream[]
	}
	setActiveStreamData: (stream: StreamData) => void
	activeStreamData: StreamData | null
	onStreamSelect: (
		streamName: string,
		checked: boolean,
		namespace: string,
	) => void
	setSelectedStreams: React.Dispatch<
		React.SetStateAction<
			| string[]
			| {
					[namespace: string]: SelectedStream[]
			  }
			| {
					selected_streams: {
						[namespace: string]: SelectedStream[]
					}
					streams: StreamData[]
			  }
		>
	>
}

export interface StreamSchemaProps {
	initialData: StreamData
	onColumnsChange?: (columns: string[]) => void
	onSyncModeChange?: (mode: "full_refresh" | "cdc") => void
}
