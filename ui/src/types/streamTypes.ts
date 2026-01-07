import type { CheckboxChangeEvent } from "antd/es/checkbox"
import type { IngestionMode, UnknownObject } from "./index"

export enum SyncMode {
	FULL_REFRESH = "full_refresh",
	CDC = "cdc",
	INCREMENTAL = "incremental",
	STRICT_CDC = "strict_cdc",
}

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
	sync_mode:
		| SyncMode.FULL_REFRESH
		| SyncMode.CDC
		| SyncMode.INCREMENTAL
		| SyncMode.STRICT_CDC
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
					config?: {
						destination_column_name?: string
					}
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
		destination_database?: string
		destination_table?: string
		source_defined_primary_key?: string[]
		default_stream_properties: DefaultStreamProperties
		[key: string]: unknown
	}
}

export interface DefaultStreamProperties {
	normalization: boolean
	append_mode: boolean
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
		mode: SyncMode,
		cursorField?: string,
	) => void
	useDirectForms?: boolean
}

export interface SelectedStream {
	stream_name: string
	partition_regex: string
	normalization: boolean
	filter?: string
	disabled?: boolean
	append_mode?: boolean
}

export interface SelectedStreamsByNamespace {
	[namespace: string]: SelectedStream[]
}
export interface StreamsDataStructure {
	selected_streams: SelectedStreamsByNamespace
	streams: StreamData[]
}

export interface SchemaConfigurationProps {
	selectedStreams: string[] | StreamsDataStructure
	setSelectedStreams: React.Dispatch<
		React.SetStateAction<string[] | StreamsDataStructure>
	>
	stepNumber?: number
	stepTitle?: string
	useDirectForms?: boolean
	sourceName: string
	sourceConnector: string
	sourceVersion: string
	sourceConfig: string
	initialStreamsData?: StreamsDataStructure
	fromJobEditFlow?: boolean
	jobId?: number
	destinationType?: string
	jobName: string
	onLoadingChange?: (isLoading: boolean) => void
}

export interface ExtendedStreamConfigurationProps extends StreamConfigurationProps {
	onUpdate?: (stream: any) => void
	isSelected: boolean
	initialNormalization: boolean
	initialPartitionRegex: string
	initialFullLoadFilter?: string
	fromJobEditFlow?: boolean
	initialSelectedStreams?: StreamsDataStructure
	destinationType?: string
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
	onIngestionModeChange: (
		streamName: string,
		namespace: string,
		appendMode: boolean,
	) => void
	sourceType?: string
}

export interface GroupedStreamsCollapsibleListProps {
	groupedStreams: { [namespace: string]: StreamData[] }
	selectedStreams: SelectedStreamsByNamespace
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
			| SelectedStreamsByNamespace
			| {
					selected_streams: SelectedStreamsByNamespace
					streams: StreamData[]
			  }
		>
	>
	onIngestionModeChange: (ingestionMode: IngestionMode) => void
	sourceType?: string
	destinationType?: string
}

export interface StreamSchemaProps {
	initialData: StreamData
	onColumnsChange?: (columns: string[]) => void
	onSyncModeChange?: (mode: SyncMode.FULL_REFRESH | SyncMode.CDC) => void
}

export type CursorFieldValues = {
	primary: string
	fallback: string
}
