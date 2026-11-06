import {
	MAINTENANCE_TAB,
	RUN_STATUS,
	RUN_TYPE,
	RUN_TYPE_FILTER,
	RUN_TYPE_LABEL,
} from "../constants"

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

export interface FusionLatestRun {
	finish_time: number
	status: RunStatus
	runID?: string
}

export interface FusionTable {
	name: string
	totalSize: string
	healthScore: number
	olake_created: boolean
	major: FusionLatestRun | null
	minor: FusionLatestRun | null
	full: FusionLatestRun | null
	/** latest snapshot-expiry (cleanup) run */
	expiry: FusionLatestRun | null
	enabled: boolean
	snapshot_expiry_enabled: boolean
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
			commitTime?: number
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
	failReason?: string | null
	summary?: RunMetrics | null
}

/** Page of runs from the BFF runs endpoint (compaction and snapshot expiry together). */
export interface GetTableRunsApiResponse {
	list?: FusionRun[]
	total?: number
}

export interface UpdateTableCronApiRequest {
	minor_cron?: string
	major_cron?: string
	full_cron?: string
	enabled_for_optimization?: string
	target_file_size?: number
	// snapshot expiry (cleanup); omitted fields are left untouched on the table
	snapshot_expiry_enabled?: string
	snapshot_expiry_cron?: string
	/** Fusion duration, e.g. "7d" */
	snapshot_keep_duration?: string
	snapshot_keep_min_count?: number
}

export interface UpdateTablesConfigApiRequest {
	tables: string[]
	sql_input: UpdateTableCronApiRequest
}

// Frontend Domain Types
export type FilterKey = "all" | "olake" | "external"

export type RunHistoryFilter = "all" | "failed"

export type RunTypeFilter = ValueOf<typeof RUN_TYPE_FILTER>

export type LatestRun = {
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
	minor: LatestRun
	major: LatestRun
	full: LatestRun
	expiry: LatestRun
	enabled: boolean
	snapshotExpiryEnabled: boolean
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
	/** False when the table has no compaction cron set at all. */
	optimizationConfigured: boolean
	targetFileSize?: number
	/** False when the table has neither snapshot-expire.enabled nor a cron set. */
	snapshotExpiryConfigured: boolean
	snapshotExpiryCron?: string
	/** Raw Fusion duration string, e.g. "7d" */
	snapshotKeepDuration?: string
	snapshotKeepMinCount?: number
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
	commitTime?: number
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
	/** Saved switch state; a never-configured table starts switched on. */
	optimizationEnabled: boolean
	cleanup: CleanupConfigFormModel
}

/** Arguments of the run history query, shared by the service, hook and query key. */
export interface TableRunsQuery {
	catalog: string
	database: string
	tableName: string
	page: number
	pageSize: number
	status?: string
	runType?: string
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
	sessionId?: string
	success: boolean
	message?: string
	logs?: string[]
}

export interface ScheduleSectionProps {
	title: string
	value: CronConfigOption
	onChange: (next: CronConfigOption) => void
	isFirst?: boolean
	tooltip?: string
}

export type MaintenanceTabKey = ValueOf<typeof MAINTENANCE_TAB>

/** Units Fusion accepts for snapshot.keep.duration (seconds are folded into minutes). */
export type KeepDurationUnit = "d" | "h" | "min"

export interface KeepDuration {
	value: number
	unit: KeepDurationUnit
}

/** Snapshot-expiry (cleanup) settings, stored as table properties in Fusion. */
export interface CleanupConfigFormModel {
	expireSnapshots: boolean
	/** Snapshots older than this are removed. */
	keepDuration: KeepDuration
	/** Lower bound on snapshots kept regardless of age. */
	minSnapshotsToKeep: number
	cron: CronConfigOption
}
