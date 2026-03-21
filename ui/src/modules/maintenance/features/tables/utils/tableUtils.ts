import { formatTimestampToUtcTime } from "@/common/utils"

import { KNOWN_CRON_TRIGGER_INTERVALS } from "../constants"
import type {
	CronConfigOption,
	FusionRun,
	FusionTable,
	RunMetricRow,
	Table,
	TableCronApiModel,
	TableCronFormModel,
	TableRun,
} from "../types"

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

// Converts a FusionRun into a TableRun, formatting timestamps and duration into display strings.
export const mapFusionRunToTableRun = (run: FusionRun): TableRun => ({
	id: run["run-id"],
	runId: run["run-id"],
	status: run.status,
	type: run.type,
	startTime: formatTimestampToUtcTime(run["start-time"]),
	duration: formatRunDuration(run.duration),
	metrics: mapRunMetricsPayloadToRows(run.metrics),
})

// Applies mapFusionRunToTableRun across the full runs array returned by the API.
export const mapGetTableRunsResponseToTableRuns = (
	runs: FusionRun[],
): TableRun[] => runs.map(mapFusionRunToTableRun)

const isRecord = (value: unknown): value is Record<string, unknown> =>
	typeof value === "object" && value !== null && !Array.isArray(value)

const normalizeMetricKey = (key: string): string =>
	key
		.split(/[_-]+/)
		.filter(Boolean)
		.map(part => part.charAt(0).toUpperCase() + part.slice(1))
		.join(" ")

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
	return Object.entries(metricsObject).map(([key, value]) => ({
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

// Expands a TableCronApiModel into TableCronFormModel by resolving each trigger interval through mapTriggerIntervalToCronConfigOption.
export const mapTableCronApiModelToTableCronFormModel = (
	config: TableCronApiModel,
): TableCronFormModel => ({
	minorCron: mapTriggerIntervalToCronConfigOption(config.minorTriggerInterval),
	majorCron: mapTriggerIntervalToCronConfigOption(config.majorTriggerInterval),
	fullCron: mapTriggerIntervalToCronConfigOption(config.fullTriggerInterval),
	targetFileSize: config.targetFileSize,
})
