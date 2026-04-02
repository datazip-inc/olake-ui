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
			let config: UpdateTableCronApiRequest = {
				enabled_for_optimization: enabled.toString(),
			}

			if (enabled) {
				const details = await tableService.getTableDetails(
					catalog,
					database,
					tableName,
				)
				const properties = details.result?.properties ?? {}

				const isConfigured = [
					"self-optimizing.minor.trigger.cron",
					"self-optimizing.major.trigger.cron",
					"self-optimizing.full.trigger.cron",
				].some(key => key in properties)

				if (!isConfigured) {
					config = {
						...config,
						minor_cron: LITE_DEFAULT_TRIGGER_INTERVAL,
						major_cron: MEDIUM_DEFAULT_TRIGGER_INTERVAL,
						full_cron: FULL_DEFAULT_TRIGGER_INTERVAL,
						target_file_size: DEFAULT_TARGET_FILE_SIZE,
					}
				}
			}

			return tableService.updateTableConfig(
				catalog,
				database,
				tableName,
				config,
			)
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
			tableService.updateTableConfig(catalog, database, tableName, payload),
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
