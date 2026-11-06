import { useMutation } from "@tanstack/react-query"

import {
	DEFAULT_TARGET_FILE_SIZE,
	FULL_CRON_PROPERTY_KEY,
	FULL_DEFAULT_TRIGGER_INTERVAL,
	LITE_DEFAULT_TRIGGER_INTERVAL,
	MAJOR_CRON_PROPERTY_KEY,
	MEDIUM_DEFAULT_TRIGGER_INTERVAL,
	MINOR_CRON_PROPERTY_KEY,
	SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY,
	tableKeys,
} from "../../constants"
import { tableService } from "../../services"
import type {
	CancelRunRequest,
	ToggleTableOptimizingRequest,
	UpdateTableConfigApiRequest,
	UpdateTablesConfigApiRequest,
} from "../../types"
import { DEFAULT_CLEANUP_API_REQUEST } from "../../utils"

export const useToggleTableOptimizing = () => {
	return useMutation({
		mutationKey: tableKeys.all(),
		mutationFn: async ({
			catalog,
			database,
			tableName,
			enabled,
		}: ToggleTableOptimizingRequest) => {
			// Status covers both features: on enables optimization and snapshot expiry,
			// off disables both. Never-configured tables get the defaults on the way up.
			let config: UpdateTableConfigApiRequest = {
				enabled_for_optimization: enabled.toString(),
				snapshot_expiry_enabled: enabled.toString(),
			}

			if (enabled) {
				const details = await tableService.getTableDetails(
					catalog,
					database,
					tableName,
				)
				const properties = details.result?.properties ?? {}

				const isOptimizationConfigured = [
					MINOR_CRON_PROPERTY_KEY,
					MAJOR_CRON_PROPERTY_KEY,
					FULL_CRON_PROPERTY_KEY,
				].some(key => key in properties)

				if (!isOptimizationConfigured) {
					config = {
						...config,
						minor_cron: LITE_DEFAULT_TRIGGER_INTERVAL,
						major_cron: MEDIUM_DEFAULT_TRIGGER_INTERVAL,
						full_cron: FULL_DEFAULT_TRIGGER_INTERVAL,
						target_file_size: DEFAULT_TARGET_FILE_SIZE,
					}
				}

				if (!(SNAPSHOT_EXPIRE_CRON_PROPERTY_KEY in properties)) {
					config = { ...config, ...DEFAULT_CLEANUP_API_REQUEST }
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

export const useBulkUpdateTableConfig = (catalog: string, database: string) => {
	return useMutation({
		mutationKey: tableKeys.list(catalog, database),
		mutationFn: (payload: UpdateTablesConfigApiRequest) =>
			tableService.bulkUpdateTableConfig(catalog, database, payload),
	})
}

/** Writes one table's optimize + cleanup properties. */
export const useUpdateTableConfig = (
	catalog: string,
	database: string,
	tableName: string,
) => {
	return useMutation({
		mutationKey: tableKeys.list(catalog, database),
		mutationFn: (payload: UpdateTableConfigApiRequest) =>
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
