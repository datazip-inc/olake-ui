import { useMutation, useQueryClient } from "@tanstack/react-query"

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
	GetTablesApiResponse,
	ToggleTableOptimizingRequest,
	UpdateTableCronApiRequest,
	UpdateTablesConfigApiRequest,
} from "../../types"

// Refreshes only the toggled table's row, from what fusion stored, rather than the whole list.
export const useToggleTableOptimizing = () => {
	const queryClient = useQueryClient()

	return useMutation({
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
				const { result } = await tableService.getTableOptimizing(
					catalog,
					database,
					tableName,
				)

				const isConfigured = [
					result.minorTriggerCron,
					result.majorTriggerCron,
					result.fullTriggerCron,
				].some(cron => cron != null)

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

			const response = await tableService.updateTableConfig(
				catalog,
				database,
				tableName,
				config,
			)
			if (!response.success) return response

			const listKey = tableKeys.list(catalog, database)
			try {
				const { result } = await tableService.getTableOptimizing(
					catalog,
					database,
					tableName,
				)
				queryClient.setQueryData<GetTablesApiResponse>(
					listKey,
					data =>
						data && {
							result: {
								...data.result,
								tables: data.result.tables.map(table =>
									table.name === tableName
										? { ...table, enabled: result.enabled }
										: table,
								),
							},
						},
				)
			} catch {
				// saved already, fall back to refreshing the whole list
				await queryClient.invalidateQueries({ queryKey: listKey })
			}
			return response
		},
	})
}

export const useBulkUpdateTableCronConfig = (
	catalog: string,
	database: string,
) => {
	return useMutation({
		mutationKey: tableKeys.list(catalog, database),
		mutationFn: (payload: UpdateTablesConfigApiRequest) =>
			tableService.bulkUpdateTableConfig(catalog, database, payload),
	})
}

/** Scoped to the specific table — only its cron/metrics/runs queries are invalidated on success. */
export const useUpdateTableCronConfig = (
	catalog: string,
	database: string,
	tableName: string,
) => {
	return useMutation({
		mutationKey: tableKeys.list(catalog, database),
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
