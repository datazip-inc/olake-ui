import { CheckboxChangeEvent } from "antd/es/checkbox"
import { UnknownObject } from "./common.types"

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
