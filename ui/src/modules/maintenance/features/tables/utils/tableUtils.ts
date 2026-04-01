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

import {
	KNOWN_CRON_TRIGGER_INTERVALS,
	LITE_DEFAULT_TRIGGER_INTERVAL,
	MEDIUM_DEFAULT_TRIGGER_INTERVAL,
	FULL_DEFAULT_TRIGGER_INTERVAL,
	RUN_STATUS,
	RUN_TYPE,
	RUN_TYPE_LABEL,
} from "../constants"
import type {
	CronConfigOption,
	FusionCompactionRun,
	CompactionRun,
	FusionRun,
	FusionTable,
	RunType,
	RunMetricRow,
	Table,
	TableDetailsApiResponse,
	TableCronApiModel,
	TableDetailsApiModel,
	TableCronFormModel,
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

const mapCompactionRun = (run: FusionCompactionRun | null): CompactionRun => {
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
	minor: mapCompactionRun(table.minor),
	major: mapCompactionRun(table.major),
	full: mapCompactionRun(table.full),
	enabled: table.enabled,
})

// Applies mapFusionTableToTable across the full tables array returned by the API.
export const mapGetTablesResponseToTables = (tables: FusionTable[]): Table[] =>
	tables.map(mapFusionTableToTable)

// Returns the run id only when one of full/minor/major is currently RUNNING.
export const getCancelRunID = (table: Table): string | null => {
	const runs: Array<CompactionRun> = [table.full, table.minor, table.major]
	const running = runs.find(run => run?.status === RUN_STATUS.RUNNING)
	const runId = running?.runID
	return typeof runId === "string" && runId.trim() ? runId : null
}

const runTypeToLabel: Record<RunType, TableRun["type"]> = {
	[RUN_TYPE.MINOR]: RUN_TYPE_LABEL.LITE,
	[RUN_TYPE.MAJOR]: RUN_TYPE_LABEL.MEDIUM,
	[RUN_TYPE.FULL]: RUN_TYPE_LABEL.FULL,
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
			type: runTypeToLabel[run.optimizingType],
			startTime: formatTimestampToUtcTime(run.startTime),
			duration: formatDuration(run.duration),
			metrics: mapRunMetricsPayloadToRows(run.summary ?? {}),
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
// Maps a raw cron string to dropdown state — preset strings become a frequency value, anything else becomes a custom cron entry.
export const mapTriggerIntervalToCronConfigOption = (
	triggerInterval: string,
): CronConfigOption => {
	if (triggerInterval === "") {
		return { frequency: "never", customCron: "" }
	}

	if (KNOWN_CRON_TRIGGER_INTERVALS.has(triggerInterval)) {
		return {
			frequency: triggerInterval,
			customCron: "",
		}
	}

	return {
		frequency: "custom",
		customCron: triggerInterval,
	}
}

// Expands a TableCronApiModel into TableCronFormModel by resolving each trigger interval through mapTriggerIntervalToCronConfigOption.
export const mapTableCronApiModelToTableCronFormModel = (
	config: TableCronApiModel,
): TableCronFormModel => ({
	minorCron: mapTriggerIntervalToCronConfigOption(
		config.minorTriggerInterval?.trim() ?? LITE_DEFAULT_TRIGGER_INTERVAL,
	),
	majorCron: mapTriggerIntervalToCronConfigOption(
		config.majorTriggerInterval?.trim() ?? MEDIUM_DEFAULT_TRIGGER_INTERVAL,
	),
	fullCron: mapTriggerIntervalToCronConfigOption(
		config.fullTriggerInterval?.trim() ?? FULL_DEFAULT_TRIGGER_INTERVAL,
	),
	targetFileSize: config.targetFileSize,
})

// Extracts cron/config fields from /details payload into existing TableCronApiModel shape.
export const mapTableDetailsResponseToTableDetailsApiModel = (
	data: TableDetailsApiResponse,
): TableDetailsApiModel => {
	const baseMetrics = data.result?.baseMetrics ?? {}
	const properties = data.result?.properties ?? {}
	const targetFileSizeRaw = properties["self-optimizing.target-size"]
	const targFileSizeInMB = bytesToMb(
		Number.parseInt(targetFileSizeRaw ?? "", 10),
	)

	return {
		enabledForOptimization:
			(properties["self-optimizing.enabled"] ?? "").toLowerCase() === "true",
		minorTriggerInterval: properties["self-optimizing.minor.trigger.cron"],
		majorTriggerInterval: properties["self-optimizing.major.trigger.cron"],
		fullTriggerInterval: properties["self-optimizing.full.trigger.cron"],
		targetFileSize: targFileSizeInMB,
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
		...mapTableCronApiModelToTableCronFormModel(tableDetails),
	}
}

// Maps the snapshots/metrics payload to only data-files and delete-files.
export const mapTableMetricsResponseToFileSummary = (
	data: TableMetricsApiResponse,
): TableMetricsFileSummary => {
	const filesSummary = data.result?.list?.[0]?.filesSummaryForChart
	return {
		"data-files": Number(filesSummary?.["data-files"] ?? 0),
		"delete-files": Number(filesSummary?.["delete-files"] ?? 0),
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
		lastCommitTime: details?.lastCommitTime,
		totalSize: details?.totalSize,
		dataFiles,
		deleteFiles,
	}
}
