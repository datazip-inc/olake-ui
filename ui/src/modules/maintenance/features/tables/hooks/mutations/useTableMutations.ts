import { useMutation } from "@tanstack/react-query"

import {
	DEFAULT_CLEANUP_CONFIG,
	DEFAULT_OPTIMIZE_CONFIG,
	tableKeys,
} from "../../constants"
import { tableService } from "../../services"
import type {
	CancelRunRequest,
	ToggleTableOptimizingRequest,
	UpdateTableConfigApiRequest,
	UpdateTablesConfigApiRequest,
} from "../../types"
import {
	mapCleanupConfigToApiRequest,
	mapOptimizeConfigToApiRequest,
	mapTableDetailsResponseToTableDetailsApiModel,
} from "../../utils"

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
				const details = mapTableDetailsResponseToTableDetailsApiModel(
					await tableService.getTableDetails(catalog, database, tableName),
				)

				if (!details.optimizationConfigured) {
					config = {
						...mapOptimizeConfigToApiRequest(DEFAULT_OPTIMIZE_CONFIG),
						...config,
					}
				}

				// Expiry only fires on a cron, so a table without one gets the full defaults
				// even when snapshot-expire.enabled was already set.
				if (!details.snapshotExpiryCron) {
					config = {
						...mapCleanupConfigToApiRequest(DEFAULT_CLEANUP_CONFIG),
						...config,
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

export const useBulkUpdateTableConfig = (catalog: string, database: string) => {
	return useMutation({
		mutationKey: tableKeys.list(catalog, database),
		mutationFn: (payload: UpdateTablesConfigApiRequest) =>
			tableService.bulkUpdateTableConfig(catalog, database, payload),
	})
}

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
