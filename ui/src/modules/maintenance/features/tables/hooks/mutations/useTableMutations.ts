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
			// off disables both. On the way up, a feature whose schedules are all unset or
			// Never gets the default schedules, since an empty cron never runs.
			let config: UpdateTableConfigApiRequest = {
				enabled_for_optimization: enabled.toString(),
				snapshot_expiry_enabled: enabled.toString(),
			}

			if (enabled) {
				const details = mapTableDetailsResponseToTableDetailsApiModel(
					await tableService.getTableDetails(catalog, database, tableName),
				)

				// Only the schedules are replaced; a saved target file size survives the toggle.
				if (!details.optimizationConfigured) {
					const { minor_cron, major_cron, full_cron, target_file_size } =
						mapOptimizeConfigToApiRequest(DEFAULT_OPTIMIZE_CONFIG)
					config = {
						minor_cron,
						major_cron,
						full_cron,
						...(details.targetFileSize === undefined
							? { target_file_size }
							: {}),
						...config,
					}
				}

				// Cleanup defaults are seeded per property: the cron where it is unset or Never, the
				// rest only where the table has none, so every other saved value survives the toggle.
				const cleanupDefaults = mapCleanupConfigToApiRequest(
					DEFAULT_CLEANUP_CONFIG,
				)
				config = {
					...(!details.snapshotExpiryCron?.trim()
						? { snapshot_expiry_cron: cleanupDefaults.snapshot_expiry_cron }
						: {}),
					...(details.snapshotKeepDuration === undefined
						? { snapshot_keep_duration: cleanupDefaults.snapshot_keep_duration }
						: {}),
					...(details.snapshotKeepMinCount === undefined
						? {
								snapshot_keep_min_count:
									cleanupDefaults.snapshot_keep_min_count,
							}
						: {}),
					...config,
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
