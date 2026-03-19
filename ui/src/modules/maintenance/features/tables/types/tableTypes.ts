// Backend API Types
export type RunStatus = "SUCCESS" | "RUNNING" | "FAILED"
export type RunType = "MINOR" | "MAJOR" | "FULL"

export interface RunMetrics {
	[key: string]: string | number | null | undefined
}

export interface FusionCompactionRun {
	"last-run": string
	status: RunStatus
}

export interface FusionTable {
	name: string
	totalSize: string
	healthScore: number
	byOLake: boolean
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

export interface FusionRun {
	"run-id": string
	status: RunStatus
	type: RunType
	"start-time": number
	duration: number
	metrics: RunMetrics
}

export interface GetTableRunsApiResponse {
	runs: FusionRun[]
	total: number
}

export interface UpdateTableCronApiRequest {
	minorTriggerInterval: string
	majorTriggerInterval: string
	fullTriggerInterval: string
	targetFileSize?: number
}

// Frontend Domain Types
export type FilterKey = "all" | "olake" | "external"
export type CompactionRun = FusionCompactionRun | null
export type CompactionScheduleTitle = "Minor" | "Major" | "Full"

export interface Table {
	id: string
	name: string
	totalSize: string
	healthScore: number
	byOLake: boolean
	minor: CompactionRun
	major: CompactionRun
	full: CompactionRun
	enabled: boolean
}

export interface TableRun {
	id: string
	runId: string
	status: RunStatus
	type: RunType
	startTime: string
	duration: string
	metrics: RunMetrics
}

export interface TableCronApiModel {
	minorTriggerInterval: string
	majorTriggerInterval: string
	fullTriggerInterval: string
	targetFileSize?: number
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
}

export interface ScheduleSectionProps {
	title: CompactionScheduleTitle
	value: CronConfigOption
	onChange: (next: CronConfigOption) => void
	isFirst?: boolean
}
