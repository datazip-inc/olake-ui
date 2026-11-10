import {
	CONFIGURE_MODE,
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

// Page of Fusion optimizing processes (compaction and snapshot expiry together).
export interface GetTableRunsApiResponse {
	result?: {
		list?: FusionRun[]
		total?: number
	}
}

// Table properties written by the configure modal and the Status toggle; omitted fields are left untouched.
export interface UpdateTableConfigApiRequest {
	minor_cron?: string
	major_cron?: string
	full_cron?: string
	enabled_for_optimization?: string
	target_file_size?: number
	snapshot_expiry_enabled?: string
	snapshot_expiry_cron?: string
	// Fusion duration, e.g. "7d"
	snapshot_keep_duration?: string
	snapshot_keep_min_count?: number
}

export interface UpdateTablesConfigApiRequest {
	tables: string[]
	sql_input: UpdateTableConfigApiRequest
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
	// Status column: on when optimization or snapshot expiry is enabled.
	maintenanceEnabled: boolean
}

export type LatestRunsByKind = Pick<
	Table,
	"minor" | "major" | "full" | "expiry"
>

export type LatestRunSlotKey = keyof LatestRunsByKind

export interface TableRun {
	id: string
	runId: string
	status: RunStatus
	runType: RunType
	type: RunTypeLabel
	startTime: string
	duration: string
	metrics: RunMetricRow[]
}

export interface TableConfigApiModel {
	minorTriggerInterval?: string
	majorTriggerInterval?: string
	fullTriggerInterval?: string
	enabledForOptimization: boolean
	// False when the table has no compaction cron set at all.
	optimizationConfigured: boolean
	targetFileSize?: number
	// False when the table has neither snapshot-expire.enabled nor a cron set.
	snapshotExpiryConfigured: boolean
	// Fusion's rule: snapshot-expire.enabled when set, else whether a cron is set.
	snapshotExpiryEnabled: boolean
	snapshotExpiryCron?: string
	// Raw Fusion duration string, e.g. "7d"
	snapshotKeepDuration?: string
	snapshotKeepMinCount?: number
}

export interface TableDetailsApiModel extends TableConfigApiModel {
	averageFileSize: string
	fileCount: number
	lastCommitTime: number
	totalSize: string
}

export type TableDetailsViewModel = TableDetailsApiModel & TableConfigFormModel

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

export interface OptimizeConfigFormModel {
	// Saved switch state; a never-configured table starts switched on.
	optimizationEnabled: boolean
	minorCron: CronConfigOption
	majorCron: CronConfigOption
	fullCron: CronConfigOption
	targetFileSize: number
}

export type CompactionCronKey = keyof Pick<
	OptimizeConfigFormModel,
	"minorCron" | "majorCron" | "fullCron"
>

// Units Fusion accepts for snapshot.keep.duration (seconds are folded into minutes).
export type KeepDurationUnit = "d" | "h" | "min"

export interface KeepDuration {
	value: number
	unit: KeepDurationUnit
}

// Cleanup tab of the configure modal: snapshot-expiry settings stored as table properties in Fusion.
export interface CleanupConfigFormModel {
	expireSnapshots: boolean
	// Snapshots older than this are removed.
	keepDuration: KeepDuration
	// Lower bound on snapshots kept regardless of age.
	minSnapshotsToKeep: number
	cron: CronConfigOption
}

export interface TableConfigFormModel {
	optimize: OptimizeConfigFormModel
	cleanup: CleanupConfigFormModel
}

export interface TableRunsQuery {
	catalog: string
	database: string
	tableName: string
	page: number
	pageSize: number
	status?: string
	// Fusion `type` filter: one process type; undefined means all.
	runType?: RunType
	// Dropped after fetching; Fusion cannot exclude a process type itself.
	excludeRunType?: RunType
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

export type MaintenanceTabKey = ValueOf<typeof MAINTENANCE_TAB>

export type ConfigureMode = ValueOf<typeof CONFIGURE_MODE>
