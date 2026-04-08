import { RUN_STATUS, RUN_TYPE, RUN_TYPE_LABEL } from "../constants"

type ValueOf<T> = T[keyof T]

// Backend API Types
export type RunStatus = ValueOf<typeof RUN_STATUS>

export type RunType = ValueOf<typeof RUN_TYPE>

export interface RunMetrics {
	[key: string]: string | number | null | undefined
}

export interface RunMetricRow {
	label: string
	value: string
}

export interface FusionCompactionRun {
	finish_time: number
	status: RunStatus
	runID?: string
}

export interface FusionTable {
	name: string
	totalSize: string
	healthScore: number
	olake_created: boolean
	major: FusionCompactionRun | null
	minor: FusionCompactionRun | null
	full: FusionCompactionRun | null
	enabled: boolean
}

export interface GetTablesApiResponse {
	catalog: string
	database: string
	tables: FusionTable[]
}

export interface TableMetrics {
	"table-metrics": {
		"file-count": {
			total: number
			"data-files": number
			"delete-files": number
		}
		"average-file-size": string
		"last-commit-time": number
	}
}

export interface TableDetailsApiResponse {
	result: {
		baseMetrics?: {
			averageFileSize?: string
			fileCount?: number
			lastCommitTime?: number
			totalSize?: string
		}
		properties?: Record<string, string>
	}
}

export interface TableMetricsApiResponse {
	result: {
		list: Array<{
			filesSummaryForChart?: {
				"data-files"?: string
				"delete-files"?: string
			}
		}>
		total: number
	}
}

export interface FusionRun {
	processId: string
	status: RunStatus
	optimizingType: RunType
	startTime: number
	duration: number
	summary?: RunMetrics
}

export interface GetTableRunsApiResponse {
	result?: {
		list?: FusionRun[]
		total?: number
	}
}

export interface UpdateTableCronApiRequest {
	minor_cron?: string
	major_cron?: string
	full_cron?: string
	enabled_for_optimization?: string
	target_file_size?: number
}

// Frontend Domain Types
export type FilterKey = "all" | "olake" | "external"

export type RunHistoryFilter = "all" | "failed"

export type CompactionRun = {
	lastRun: string
	status: RunStatus
	runID?: string
} | null
export type RunTypeLabel = ValueOf<typeof RUN_TYPE_LABEL>

export interface Table {
	id: string
	name: string
	totalSize: string
	healthScore: number
	olakeCreated: boolean
	minor: CompactionRun
	major: CompactionRun
	full: CompactionRun
	enabled: boolean
}

export interface TableRun {
	id: string
	runId: string
	status: RunStatus
	type: RunTypeLabel
	startTime: string
	duration: string
	metrics: RunMetricRow[]
}

export interface TableCronApiModel {
	minorTriggerInterval?: string
	majorTriggerInterval?: string
	fullTriggerInterval?: string
	enabledForOptimization: boolean
	targetFileSize?: number
}

export interface TableDetailsApiModel extends TableCronApiModel {
	averageFileSize: string
	fileCount: number
	lastCommitTime: number
	totalSize: string
}

export type TableDetailsViewModel = TableDetailsApiModel & TableCronFormModel

export interface TableMetricsFileSummary {
	"data-files": number
	"delete-files": number
}

export interface TableMetricsModalData {
	fileCount?: number
	averageFileSize?: string
	lastCommitTime?: number
	dataFiles?: number
	deleteFiles?: number
	totalSize?: string
}

export interface CronConfigOption {
	frequency: string
	customCron: string
}

export interface TableCronFormModel {
	minorCron: CronConfigOption
	majorCron: CronConfigOption
	fullCron: CronConfigOption
	targetFileSize?: number
}

export interface ToggleTableOptimizingRequest {
	catalog: string
	database: string
	tableName: string
	enabled: boolean
}

export interface CancelRunRequest {
	catalog: string
	database: string
	tableName: string
	runId: string
}

export interface UpdateTableConfigApiResponse {
	success: boolean
	message: string
	logs?: string[]
}

export interface ScheduleSectionProps {
	title: RunTypeLabel
	value: CronConfigOption
	onChange: (next: CronConfigOption) => void
	isFirst?: boolean
	tooltip?: string
}
