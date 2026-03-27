import { useMutation } from "@tanstack/react-query"

import { tableKeys } from "../../constants"
import { tableService } from "../../services"
import type {
	CancelRunRequest,
	ToggleTableOptimizingRequest,
	UpdateTableCronApiRequest,
} from "../../types"

export const useToggleTableOptimizing = () => {
	return useMutation({
		mutationKey: tableKeys.all(),
		mutationFn: ({
			catalog,
			database,
			tableName,
			enabled,
		}: ToggleTableOptimizingRequest) =>
			tableService.setTableOptimizing(catalog, database, tableName, enabled),
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
