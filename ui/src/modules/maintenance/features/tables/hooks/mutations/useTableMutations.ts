import { useMutation } from "@tanstack/react-query"

import {
	DEFAULT_TARGET_FILE_SIZE,
	FULL_DEFAULT_TRIGGER_INTERVAL,
	LITE_DEFAULT_TRIGGER_INTERVAL,
	MEDIUM_DEFAULT_TRIGGER_INTERVAL,
	tableKeys,
} from "../../constants"
import { tableService } from "../../services"
import type {
	CancelRunRequest,
	ToggleTableOptimizingRequest,
	UpdateTableCronApiRequest,
} from "../../types"

export const useToggleTableOptimizing = () => {
	return useMutation({
		mutationKey: tableKeys.all(),
		mutationFn: async ({
			catalog,
			database,
			tableName,
			enabled,
		}: ToggleTableOptimizingRequest) => {
			// When disabling optimization, no configuration checks are needed.
			if (!enabled) {
				return tableService.setTableOptimizing(
					catalog,
					database,
					tableName,
					enabled,
				)
			}

			// When enabling, fetch the raw API response to see if cron properties already exist.
			const details = await tableService.getTableDetails(
				catalog,
				database,
				tableName,
			)
			const properties = details.result?.properties ?? {}

			const cronPropertyKeys = [
				"self-optimizing.minor.trigger.cron",
				"self-optimizing.major.trigger.cron",
				"self-optimizing.full.trigger.cron",
			]
			const isConfigured = cronPropertyKeys.some(key => key in properties)

			if (isConfigured) {
				// Toggle optimization ON without overriding the user's existing config.
				return tableService.setTableOptimizing(
					catalog,
					database,
					tableName,
					enabled,
				)
			}

			// If unconfigured, initialize the table with defaults while enabling it.
			return tableService.updateTableCronConfig(catalog, database, tableName, {
				minor_cron: LITE_DEFAULT_TRIGGER_INTERVAL,
				major_cron: MEDIUM_DEFAULT_TRIGGER_INTERVAL,
				full_cron: FULL_DEFAULT_TRIGGER_INTERVAL,
				target_file_size: DEFAULT_TARGET_FILE_SIZE,
				enabled_for_optimization: "true",
			})
		},
	})
}

/** Scoped to the specific table — only its cron/metrics/runs queries are invalidated on success. */
export const useUpdateTableCronConfig = (
	catalog: string,
	database: string,
	tableName: string,
) => {
	return useMutation({
		mutationKey: tableKeys.table(catalog, database, tableName),
		mutationFn: (payload: UpdateTableCronApiRequest) =>
			tableService.updateTableCronConfig(catalog, database, tableName, payload),
	})
}

export const useCancelTableRun = () => {
	return useMutation({
		mutationKey: tableKeys.all(),
		mutationFn: ({ catalog, database, tableName, runId }: CancelRunRequest) =>
			tableService.cancelTableRun(catalog, database, tableName, runId),
	})
}

export const useDownloadProcessLogFile = () => {
	return useMutation({
		mutationFn: ({
			processId,
			fileId,
		}: {
			processId: string
			fileId: string
		}) => tableService.downloadProcessLogFile(processId, fileId),
	})
}

export const useDownloadProcessLogsArchive = () => {
	return useMutation({
		mutationFn: (processId: string) =>
			tableService.downloadProcessLogsArchive(processId),
	})
}
