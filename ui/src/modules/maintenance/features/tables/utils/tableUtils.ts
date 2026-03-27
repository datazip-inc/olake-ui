import {
	CheckCircleIcon,
	SpinnerIcon,
	WarningCircleIcon,
} from "@phosphor-icons/react"

import {
	bytesToMb,
	formatTimestampToUtcTime,
	toStartCase,
} from "@/common/utils"

import { KNOWN_CRON_TRIGGER_INTERVALS } from "../constants"
import type {
	CronConfigOption,
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
		case "SUCCESS":
			return {
				Icon: CheckCircleIcon,
				bgClass: "bg-olake-success-bg",
				textClass: "text-olake-success",
				label: "Success",
				iconClass: undefined,
			}
		case "RUNNING":
			return {
				Icon: SpinnerIcon,
				bgClass: "bg-olake-warning-bg",
				textClass: "text-olake-warning",
				label: "Running",
				iconClass: undefined,
			}
		case "FAILED":
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
				label: toStartCase(status || "Unknown"),
			}
	}
}

export const getRunLogsStatusConfig = (status?: string) => {
	switch (status) {
		case "SUCCESS":
			return {
				Icon: CheckCircleIcon,
				bgClass: "bg-olake-success-bg",
				textClass: "text-olake-success",
				label: "Success",
				iconClass: undefined,
			}
		case "RUNNING":
			return {
				Icon: SpinnerIcon,
				bgClass: "bg-olake-warning-bg",
				textClass: "text-olake-warning",
				label: "Running",
				iconClass: undefined,
			}
		case "FAILED":
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
				label: toStartCase(status || "Unknown"),
			}
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
	byOLake: table.byOLake,
	minor: table.minor,
	major: table.major,
	full: table.full,
	enabled: table.enabled,
})

// Applies mapFusionTableToTable across the full tables array returned by the API.
export const mapGetTablesResponseToTables = (tables: FusionTable[]): Table[] =>
	tables.map(mapFusionTableToTable)

// Returns the run id only when one of full/minor/major is currently RUNNING.
export const getCancelRunID = (table: Table): string | null => {
	const runs: Array<FusionTable["full"]> = [
		table.full,
		table.minor,
		table.major,
	]
	const running = runs.find(run => run?.status === "RUNNING")
	const runId = running?.runID
	return typeof runId === "string" && runId.trim() ? runId : null
}

const formatRunDuration = (durationMs: number): string => {
	if (durationMs <= 0) return "--"
	const totalSeconds = Math.floor(durationMs / 1000)
	if (totalSeconds < 1) return "<1s"
	const days = Math.floor(totalSeconds / 86_400)
	const hours = Math.floor((totalSeconds % 86_400) / 3_600)
	const minutes = Math.floor((totalSeconds % 3_600) / 60)
	const seconds = totalSeconds % 60

	if (days > 0) return `${days}d ${hours}h ${minutes}m ${seconds}s`
	if (hours > 0) return `${hours}h ${minutes}m ${seconds}s`
	if (minutes > 0) return `${minutes}m ${seconds}s`
	return `${seconds}s`
}

const runTypeToLabel: Record<RunType, TableRun["type"]> = {
	MINOR: "Lite",
	MAJOR: "Medium",
	FULL: "Full",
}

export const mapGetTableRunsResponseToTableRuns = (
	response: GetTableRunsApiResponse,
): TableRun[] => {
	const runs: FusionRun[] = response.result?.list ?? []
	return runs.map(run => ({
		id: run.processId,
		runId: run.processId,
		status: run.status,
		type: runTypeToLabel[run.optimizingType],
		startTime: formatTimestampToUtcTime(run.startTime),
		duration: formatRunDuration(run.duration),
		metrics: mapRunMetricsPayloadToRows(run.summary ?? {}),
	}))
}

const isRecord = (value: unknown): value is Record<string, unknown> =>
	typeof value === "object" && value !== null && !Array.isArray(value)

const normalizeMetricKey = (key: string): string =>
	key
		.split(/[_-]+/)
		.filter(Boolean)
		.map(part => part.charAt(0).toUpperCase() + part.slice(1))
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

// Converts metrics payload from API/run rows into display-ready label/value rows for the sidebar.
export const mapRunMetricsPayloadToRows = (
	payload: unknown,
): RunMetricRow[] => {
	const metricsObject = extractRunMetricsObject(payload)
	return Object.entries(metricsObject)
		.filter(([key]) => !HIDDEN_RUN_METRIC_KEYS.has(key))
		.map(([key, value]) => ({
			label: normalizeMetricKey(key),
			value: formatMetricValue(value),
		}))
}

// Maps a raw cron string to dropdown state — preset strings become a frequency value, anything else becomes a custom cron entry.
export const mapTriggerIntervalToCronConfigOption = (
	triggerInterval: string,
): CronConfigOption => {
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

const LITE_DEFAULT_TRIGGER_INTERVAL = "0 * * * *"
const MEDIUM_DEFAULT_TRIGGER_INTERVAL = "0 */8 * * *"
const FULL_DEFAULT_TRIGGER_INTERVAL = "never"

// Expands a TableCronApiModel into TableCronFormModel by resolving each trigger interval through mapTriggerIntervalToCronConfigOption.
export const mapTableCronApiModelToTableCronFormModel = (
	config: TableCronApiModel,
): TableCronFormModel => ({
	minorCron: mapTriggerIntervalToCronConfigOption(
		config.minorTriggerInterval.trim() || LITE_DEFAULT_TRIGGER_INTERVAL,
	),
	majorCron: mapTriggerIntervalToCronConfigOption(
		config.majorTriggerInterval.trim() || MEDIUM_DEFAULT_TRIGGER_INTERVAL,
	),
	fullCron: mapTriggerIntervalToCronConfigOption(
		config.fullTriggerInterval.trim() || FULL_DEFAULT_TRIGGER_INTERVAL,
	),
	targetFileSize: config.targetFileSize,
})

// Extracts cron/config fields from /details payload into existing TableCronApiModel shape.
export const mapTableDetailsResponseToTableDetailsApiModel = (
	data: TableDetailsApiResponse,
): TableDetailsApiModel => {
	const baseMetrics = data.result?.baseMetrics ?? {}
	const properties = data.result?.properties ?? {}
	const targetFileSizeRaw = properties["write.target-file-size-bytes"]
	const targFileSizeInMB = bytesToMb(
		Number.parseInt(targetFileSizeRaw ?? "", 10),
	)

	return {
		enabledForOptimization:
			(properties["self-optimizing.enabled"] ?? "").toLowerCase() === "true",
		minorTriggerInterval:
			properties["self-optimizing.minor.trigger.interval"] ?? "",
		majorTriggerInterval:
			properties["self-optimizing.major.trigger.interval"] ?? "",
		fullTriggerInterval:
			properties["self-optimizing.full.trigger.interval"] ?? "",
		targetFileSize: targFileSizeInMB,
		averageFileSize: baseMetrics.averageFileSize ?? "--",
		fileCount: baseMetrics.fileCount ?? 0,
		lastCommitTime: baseMetrics.lastCommitTime ?? 0,
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
	details?: Pick<
		TableDetailsApiModel,
		"fileCount" | "averageFileSize" | "lastCommitTime"
	>,
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
		dataFiles,
		deleteFiles,
	}
}
