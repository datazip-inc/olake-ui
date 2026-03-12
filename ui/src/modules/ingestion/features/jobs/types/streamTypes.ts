import { CheckboxChangeEvent } from "antd"

import { StreamData } from "@/modules/ingestion/common/types"

import { AdvancedSettings } from "./jobTypes"

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

export type StreamPanelProps = {
	stream: any
	onStreamSelect?: (streamName: string, checked: boolean) => void
	isSelected: boolean
}

export type StreamHeaderProps = {
	stream: any
	toggle: (e: CheckboxChangeEvent) => void
	checked: boolean
}

export type StreamConfigurationProps = {
	destinationType?: string
	sourceType?: string
}

export interface SchemaConfigurationProps {
	stepNumber?: number
	stepTitle?: string
	sourceName: string
	sourceConnector: string
	sourceVersion: string
	sourceConfig: string
	fromJobEditFlow?: boolean
	jobId?: number
	destinationType?: string
	jobName: string
	advancedSettings?: AdvancedSettings | null
}

export interface GroupedStreamsCollapsibleListProps {
	groupedStreams: { [namespace: string]: StreamData[] }
	sourceType?: string
	destinationType?: string
}

export type CursorFieldValues = {
	primary: string
	fallback: string
}
