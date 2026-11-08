import {
	CheckCircleIcon,
	SpinnerIcon,
	WarningCircleIcon,
} from "@phosphor-icons/react"

import {
	bytesToMb,
	formatDuration,
	formatTimestampToUtcTime,
	getRelativeTimeString,
	toSentenceCase,
} from "@/common/utils"

import { mapTableConfigApiModelToCleanupConfig } from "./cleanupUtils"
import { mapTriggerIntervalToCronConfigOption } from "./cronUtils"
import {
	DEFAULT_TARGET_FILE_SIZE,
	FULL_CRON_PROPERTY_KEY,
	FULL_DEFAULT_TRIGGER_INTERVAL,
	LITE_DEFAULT_TRIGGER_INTERVAL,
	MAJOR_CRON_PROPERTY_KEY,
	MEDIUM_DEFAULT_TRIGGER_INTERVAL,
	MINOR_CRON_PROPERTY_KEY,
	OPTIMIZATION_ENABLED_PROPERTY_KEY,
	RUN_STATUS,
	RUN_TYPE,
	RUN_TYPE_LABEL,
	SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY,
	SNAPSHOT_EXPIRE_ENABLED_PROPERTY_KEY,
	SNAPSHOT_KEEP_DURATION_PROPERTY_KEY,
	SNAPSHOT_KEEP_MIN_COUNT_PROPERTY_KEY,
	TARGET_FILE_SIZE_PROPERTY_KEY,
} from "../constants"
import type {
	FusionLatestRun,
	LatestRun,
	FusionRun,
	FusionTable,
	RunType,
	RunMetricRow,
	Table,
	TableDetailsApiResponse,
	TableConfigApiModel,
	TableDetailsApiModel,
	TableConfigFormModel,
	TableDetailsViewModel,
	GetTableRunsApiResponse,
	TableMetricsApiResponse,
	TableMetricsFileSummary,
	TableMetricsModalData,
	TableRun,
} from "../types"

const DEFAULT_RUN_STATUS_CONFIG = {
	Icon: WarningCircleIcon,
	bgClass: "bg-olake-surface-muted",
	textClass: "text-olake-text-tertiary",
	label: "Unknown",
	iconClass: undefined,
}

export const getRunStatusConfig = (status?: string) => {
	switch (status) {
		case RUN_STATUS.SUCCESS:
			return {
				Icon: CheckCircleIcon,
				bgClass: "bg-olake-success-bg",
				textClass: "text-olake-success",
				label: "Success",
				iconClass: undefined,
			}
		case RUN_STATUS.SKIPPED:
			return {
				Icon: WarningCircleIcon,
				bgClass: "bg-olake-surface-muted",
				textClass: "text-olake-text-tertiary",
				label: "Skipped",
				iconClass: undefined,
			}
		case RUN_STATUS.CLOSED:
			return {
				Icon: WarningCircleIcon,
				bgClass: "bg-olake-surface-muted",
				textClass: "text-olake-text-tertiary",
				label: "Closed",
				iconClass: undefined,
			}
		case RUN_STATUS.RUNNING:
			return {
				Icon: SpinnerIcon,
				bgClass: "bg-olake-warning-bg",
				textClass: "text-olake-warning",
				label: "Running",
				iconClass: undefined,
			}
		case RUN_STATUS.FAILED:
			return {
				Icon: WarningCircleIcon,
				bgClass: "bg-olake-error-bg",
				textClass: "text-olake-error",
				label: "Failed",
				iconClass: undefined,
			}
		default:
			return {
				...DEFAULT_RUN_STATUS_CONFIG,
				label: toSentenceCase(status || "Unknown"),
			}
	}
}

export const getRunLogsStatusConfig = (status?: string) => {
	switch (status) {
		case RUN_STATUS.SUCCESS:
			return {
				Icon: CheckCircleIcon,
				bgClass: "bg-olake-success-bg",
				textClass: "text-olake-success",
				label: "Success",
				iconClass: undefined,
			}
		case RUN_STATUS.SKIPPED:
			return {
				Icon: WarningCircleIcon,
				bgClass: "bg-olake-surface-muted",
				textClass: "text-olake-text-tertiary",
				label: "Skipped",
				iconClass: undefined,
			}
		case RUN_STATUS.CLOSED:
			return {
				Icon: WarningCircleIcon,
				bgClass: "bg-olake-surface-muted",
				textClass: "text-olake-text-tertiary",
				label: "Closed",
				iconClass: undefined,
			}
		case RUN_STATUS.RUNNING:
			return {
				Icon: SpinnerIcon,
				bgClass: "bg-olake-warning-bg",
				textClass: "text-olake-warning",
				label: "Running",
				iconClass: undefined,
			}
		case RUN_STATUS.FAILED:
			return {
				Icon: WarningCircleIcon,
				bgClass: "bg-olake-error-bg",
				textClass: "text-olake-error-alt",
				label: "Failed",
				iconClass: undefined,
			}
		default:
			return {
				...DEFAULT_RUN_STATUS_CONFIG,
				label: toSentenceCase(status || "Unknown"),
			}
	}
}

const mapLatestRun = (run: FusionLatestRun | null): LatestRun => {
	if (!run) return null
	return {
		status: run.status,
		runID: run.runID,
		lastRun: getRelativeTimeString(run.finish_time),
	}
}

// Converts a FusionTable into a Table, attaching a stable row id derived from name + index.
export const mapFusionTableToTable = (
	table: FusionTable,
	idx: number,
): Table => ({
	id: `${table.name}-${idx}`,
	name: table.name,
	totalSize: table.totalSize,
	healthScore: table.healthScore,
	olakeCreated: table.olake_created,
	minor: mapLatestRun(table.minor),
	major: mapLatestRun(table.major),
	full: mapLatestRun(table.full),
	expiry: mapLatestRun(table.expiry),
	// The Status toggle covers both features; off switches both off.
	maintenanceEnabled: table.enabled || table.snapshot_expiry_enabled,
})

// Applies mapFusionTableToTable across the full tables array returned by the API.
export const mapGetTablesResponseToTables = (tables: FusionTable[]): Table[] =>
	tables.map(mapFusionTableToTable)

// Returns the run id only when one of full/minor/major is currently RUNNING.
// Cancel is offered for compaction runs only; expiry runs are not cancellable from the list.
export const getCancelRunID = (table: Table): string | null => {
	const runs: Array<LatestRun> = [table.full, table.minor, table.major]
	const running = runs.find(run => run?.status === RUN_STATUS.RUNNING)
	const runId = running?.runID
	return typeof runId === "string" && runId.trim() ? runId : null
}

const runTypeToLabel: Record<RunType, TableRun["type"]> = {
	[RUN_TYPE.MINOR]: RUN_TYPE_LABEL.LITE,
	[RUN_TYPE.MAJOR]: RUN_TYPE_LABEL.MEDIUM,
	[RUN_TYPE.FULL]: RUN_TYPE_LABEL.FULL,
	[RUN_TYPE.EXPIRY]: RUN_TYPE_LABEL.EXPIRY,
}

// Failed runs carry the reason outside the summary; surface it with the metrics.
const mapRunToMetricRows = (run: FusionRun): RunMetricRow[] => {
	const rows = mapRunMetricsPayloadToRows(run.summary ?? {})
	const failReason = run.failReason?.trim()
	if (failReason && run.status === RUN_STATUS.FAILED) {
		rows.push({ label: "Fail Reason", value: failReason })
	}
	return rows
}

export const mapGetTableRunsResponseToTableRuns = (
	response: GetTableRunsApiResponse,
): { runs: TableRun[]; total: number } => {
	const runs: FusionRun[] = response.result?.list ?? []
	return {
		runs: runs.map(run => ({
			id: run.processId,
			runId: run.processId,
			status: run.status,
			runType: run.optimizingType,
			type: runTypeToLabel[run.optimizingType] ?? run.optimizingType,
			startTime: formatTimestampToUtcTime(run.startTime),
			duration: formatDuration(run.duration) || "-",
			metrics: mapRunToMetricRows(run),
		})),
		total: response.result?.total ?? 0,
	}
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
	typeof value === "object" && value !== null && !Array.isArray(value)

const normalizeMetricKey = (key: string): string =>
	key
		.replace(/([a-z])([A-Z])/g, "$1 $2") // insert space between lower and upper case
		.split(/[_\-\s]+/)
		.filter(Boolean)
		.map(part => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
		.join(" ")

// Excludde these fields from the Run Metrics sidebar.
const HIDDEN_RUN_METRIC_KEYS = new Set([
	"input-data-records(rewrite)",
	"input-equality-delete-records",
	"input-position-delete-records",
	"output-data-records",
	"output-delete-records",
])

const formatMetricValue = (value: unknown): string => {
	if (value === null || value === undefined || value === "") return "--"
	if (typeof value === "boolean") return value ? "True" : "False"
	if (typeof value === "object") return JSON.stringify(value)
	return String(value)
}

const extractRunMetricsObject = (payload: unknown): Record<string, unknown> => {
	if (!isRecord(payload)) return {}

	const data = payload.data
	if (isRecord(data)) {
		const nestedMetrics = data.metrics
		if (isRecord(nestedMetrics)) return nestedMetrics
	}

	const rootMetrics = payload.metrics
	if (isRecord(rootMetrics)) return rootMetrics

	return payload
}

const TYPE_DISPLAY_REGEX = new RegExp(
	`\\b(${Object.keys(runTypeToLabel).join("|")})\\b`,
	"g",
)

const TYPE_REPLACEABLE_KEYS = ["optimizingType", "skipReason"]

const replaceDisplayTypes = (text: string): string =>
	text.replace(
		TYPE_DISPLAY_REGEX,
		match => runTypeToLabel[match as RunType] ?? match,
	)

// Converts metrics payload from API/run rows into display-ready label/value rows for the sidebar.
export const mapRunMetricsPayloadToRows = (
	payload: unknown,
): RunMetricRow[] => {
	const metricsObject = extractRunMetricsObject(payload)
	return Object.entries(metricsObject)
		.filter(([key]) => !HIDDEN_RUN_METRIC_KEYS.has(key))
		.map(([key, value]) => {
			const formattedValue = formatMetricValue(value)
			const shouldReplaceDisplayTypes =
				TYPE_REPLACEABLE_KEYS.includes(key) && typeof value === "string"

			return {
				label: normalizeMetricKey(key),
				value: shouldReplaceDisplayTypes
					? replaceDisplayTypes(formattedValue)
					: formattedValue,
			}
		})
}
// Expands a TableConfigApiModel into the configure modal's form, one branch per tab.
export const mapTableConfigApiModelToFormModel = (
	config: TableConfigApiModel,
): TableConfigFormModel => ({
	optimize: {
		// A never-configured table starts with the switch on.
		optimizationEnabled: config.optimizationConfigured
			? config.enabledForOptimization
			: true,
		minorCron: mapTriggerIntervalToCronConfigOption(
			config.minorTriggerInterval?.trim() ?? LITE_DEFAULT_TRIGGER_INTERVAL,
		),
		majorCron: mapTriggerIntervalToCronConfigOption(
			config.majorTriggerInterval?.trim() ?? MEDIUM_DEFAULT_TRIGGER_INTERVAL,
		),
		fullCron: mapTriggerIntervalToCronConfigOption(
			config.fullTriggerInterval?.trim() ?? FULL_DEFAULT_TRIGGER_INTERVAL,
		),
		targetFileSize: config.targetFileSize ?? DEFAULT_TARGET_FILE_SIZE,
	},
	cleanup: mapTableConfigApiModelToCleanupConfig(config),
})

// Integer property value, or undefined when the property is absent or not a number.
const parseOptionalInt = (raw?: string): number | undefined => {
	const parsed = Number.parseInt(raw ?? "", 10)
	return Number.isNaN(parsed) ? undefined : parsed
}

// Mirrors Fusion's rule (and the BFF list endpoint): snapshot-expire.enabled decides when set,
// otherwise expiry is on exactly when a trigger cron is set.
const isSnapshotExpiryEnabled = (
	properties: Record<string, string>,
): boolean => {
	const enabled = properties[SNAPSHOT_EXPIRE_ENABLED_PROPERTY_KEY]?.trim()
	if (enabled) return enabled.toLowerCase() === "true"
	return !!properties[SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY]?.trim()
}

// Extracts config fields from the /details payload into TableConfigApiModel.
export const mapTableDetailsResponseToTableDetailsApiModel = (
	data: TableDetailsApiResponse,
): TableDetailsApiModel => {
	const baseMetrics = data.result?.baseMetrics ?? {}
	const properties = data.result?.properties ?? {}
	const targetFileSizeInMB = bytesToMb(
		Number.parseInt(properties[TARGET_FILE_SIZE_PROPERTY_KEY] ?? "", 10),
	)

	return {
		enabledForOptimization:
			(properties[OPTIMIZATION_ENABLED_PROPERTY_KEY] ?? "").toLowerCase() ===
			"true",
		minorTriggerInterval: properties[MINOR_CRON_PROPERTY_KEY],
		majorTriggerInterval: properties[MAJOR_CRON_PROPERTY_KEY],
		fullTriggerInterval: properties[FULL_CRON_PROPERTY_KEY],
		targetFileSize: targetFileSizeInMB > 0 ? targetFileSizeInMB : undefined,
		optimizationConfigured: [
			MINOR_CRON_PROPERTY_KEY,
			MAJOR_CRON_PROPERTY_KEY,
			FULL_CRON_PROPERTY_KEY,
		].some(key => key in properties),
		snapshotExpiryConfigured:
			SNAPSHOT_EXPIRE_ENABLED_PROPERTY_KEY in properties ||
			SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY in properties,
		snapshotExpiryEnabled: isSnapshotExpiryEnabled(properties),
		snapshotExpiryCron: properties[SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY],
		snapshotKeepDuration: properties[SNAPSHOT_KEEP_DURATION_PROPERTY_KEY],
		snapshotKeepMinCount: parseOptionalInt(
			properties[SNAPSHOT_KEEP_MIN_COUNT_PROPERTY_KEY],
		),
		averageFileSize: baseMetrics.averageFileSize ?? "--",
		fileCount: baseMetrics.fileCount ?? 0,
		lastCommitTime: baseMetrics.lastCommitTime ?? 0,
		totalSize: baseMetrics.totalSize ?? "--",
	}
}

export const mapTableDetailsResponseToTableDetailsViewModel = (
	data: TableDetailsApiResponse,
): TableDetailsViewModel => {
	const tableDetails = mapTableDetailsResponseToTableDetailsApiModel(data)
	return {
		...tableDetails,
		...mapTableConfigApiModelToFormModel(tableDetails),
	}
}

// Maps the snapshots/metrics payload to only data-files, delete-files and commit time.
export const mapTableMetricsResponseToFileSummary = (
	data: TableMetricsApiResponse,
): TableMetricsFileSummary => {
	const latestSnapshot = data.result?.list?.[0]
	const filesSummary = latestSnapshot?.filesSummaryForChart
	return {
		"data-files": Number(filesSummary?.["data-files"] ?? 0),
		"delete-files": Number(filesSummary?.["delete-files"] ?? 0),
		commitTime: latestSnapshot?.commitTime,
	}
}

// Combines details and snapshots metrics into one modal-friendly object.
export const buildTableMetricsModalData = (
	details?: TableDetailsApiModel,
	metrics?: TableMetricsFileSummary,
): TableMetricsModalData => {
	const dataFiles = metrics?.["data-files"]
	const deleteFiles = metrics?.["delete-files"]
	const fallbackFileCount =
		dataFiles !== undefined && deleteFiles !== undefined
			? dataFiles + deleteFiles
			: undefined

	return {
		fileCount: details?.fileCount ?? fallbackFileCount,
		averageFileSize: details?.averageFileSize,
		lastCommitTime: metrics?.commitTime,
		totalSize: details?.totalSize,
		dataFiles,
		deleteFiles,
	}
}
