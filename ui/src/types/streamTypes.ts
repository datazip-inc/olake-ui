import type { CheckboxChangeEvent } from "antd/es/checkbox"
import type { UnknownObject } from "./index"

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