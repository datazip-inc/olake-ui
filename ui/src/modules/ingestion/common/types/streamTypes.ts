import type { UnknownObject } from "@/modules/ingestion/common/types"

export interface StreamIdentifier {
	streamName: string
	namespace: string
}

export type FilterOperator = "=" | "!=" | ">" | "<" | ">=" | "<="
export type LogicalOperator = "and" | "or"

export enum SyncMode {
	FULL_REFRESH = "full_refresh",
	CDC = "cdc",
	INCREMENTAL = "incremental",
	STRICT_CDC = "strict_cdc",
}

export interface FilterConfigCondition {
	column: string
	operator: FilterOperator
	value: any
}

export type MultiFilterCondition = {
	conditions: FilterConfigCondition[]
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
	sort_key: string[] | null
	stream: {
		name: string
		namespace?: string
		json_schema: UnknownObject
		type_schema?: {
			properties: Record<
				string,
				{
					destination_column_name?: string
					olake_column?: boolean
					type: string[]
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

export interface SelectedColumns {
	columns: string[]
	sync_new_columns: boolean
}

export interface SelectedStream {
	stream_name: string
	partition_regex: string
	normalization: boolean
	filter?: string
	disabled?: boolean
	append_mode?: boolean
	selected_columns?: SelectedColumns
	filter_config?: FilterConfig
}

export interface FilterConfig {
	logical_operator: LogicalOperator
	conditions: FilterConfigCondition[]
}

export interface SelectedStreamsByNamespace {
	[namespace: string]: SelectedStream[]
}
export interface StreamsDataStructure {
	selected_streams: SelectedStreamsByNamespace
	streams: StreamData[]
}
