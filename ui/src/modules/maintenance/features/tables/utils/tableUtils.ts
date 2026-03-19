import { formatUtcTime } from "@/common/utils"

import { KNOWN_CRON_TRIGGER_INTERVALS } from "../constants"
import type {
	CronConfigOption,
	FusionRun,
	FusionTable,
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

const formatRunDuration = (durationSeconds: number): string => {
	if (!Number.isFinite(durationSeconds) || durationSeconds <= 0) return "--"
	const totalSeconds = Math.floor(durationSeconds)
	const minutes = Math.floor(totalSeconds / 60)
	const seconds = totalSeconds % 60

	if (minutes === 0) return `${seconds}s`
	if (seconds === 0) return `${minutes}m`
	return `${minutes}m ${seconds}s`
}

// Converts a FusionRun into a TableRun, formatting timestamps and duration into display strings.
export const mapFusionRunToTableRun = (run: FusionRun): TableRun => ({
	id: run["run-id"],
	runId: run["run-id"],
	status: run.status,
	type: run.type,
	startTime: formatUtcTime(run["start-time"]),
	duration: formatRunDuration(run.duration),
	metrics: run.metrics,
})

// Applies mapFusionRunToTableRun across the full runs array returned by the API.
export const mapGetTableRunsResponseToTableRuns = (
	runs: FusionRun[],
): TableRun[] => runs.map(mapFusionRunToTableRun)

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
