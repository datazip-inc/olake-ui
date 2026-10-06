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

export enum UpsertType {
	EQUALITY = "eq",
	POSITIONAL = "pos",
	DELETION_VECTOR = "dv",
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
		supported_sync_modes?: SyncMode[]
		source_defined_cursor?: boolean
		default_cursor_field?: string[]
		available_cursor_fields?: string[]
		cursor_field?: string
		// Discovered default only; in streams v2 the configured sync mode lives on
		// the selected_streams entry, so the catalog may omit it.
		sync_mode?: SyncMode
		destination_database?: string
		destination_table?: string
		source_defined_primary_key?: string[]
		default_stream_properties: DefaultStreamProperties
		available_update_types?: UpsertType[]
		[key: string]: unknown
	}
}

export interface DefaultStreamProperties {
	normalization: boolean
	append_mode: boolean
	// Catalog-provided default upsert type. Stored on the selected stream as
	// update_type. Absent on older olake versions.
	update_type?: UpsertType
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
	// Only present when the stream runs in upsert mode (append_mode falsy).
	update_type?: UpsertType
	use_source_column_names?: boolean
	selected_columns?: SelectedColumns
	filter_config?: FilterConfig
	sync_mode?: SyncMode
	cursor_field?: string
	destination_database?: string
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

export enum CatalogFormat {
	LEGACY = "legacy",
	V2 = "v2",
}

export interface AvailableStreams {
	streams: StreamData[]
}

export interface SelectedStreams {
	selected_streams: SelectedStreamsByNamespace
}

export interface StreamsV2Catalog {
	available_streams: AvailableStreams
	selected_streams: SelectedStreamsByNamespace
}

// Deprecated: legacy streams.json shape, only read and written by legacyStreams.ts.
export interface LegacyStreamsConfig {
	streams: StreamData[]
	selected_streams: SelectedStreamsByNamespace
}

export type DiscoverResponse = StreamsV2Catalog | LegacyStreamsConfig

export interface StreamDifferenceResponse {
	selected_streams: SelectedStreamsByNamespace
	streams?: StreamData[]
}
