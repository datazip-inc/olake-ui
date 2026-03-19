import { useQuery } from "@tanstack/react-query"

import { tableKeys } from "../../constants"
import { tableService } from "../../services"
import {
	mapGetTablesResponseToTables,
	mapGetTableRunsResponseToTableRuns,
	mapTableCronApiModelToTableCronFormModel,
	mapProcessLogsResponse,
} from "../../utils"

export const useTables = (
	catalog: string,
	database: string,
	enabled = true,
) => {
	return useQuery({
		queryKey: tableKeys.list(catalog, database),
		queryFn: () => tableService.getTables(catalog, database),
		select: data => mapGetTablesResponseToTables(data.tables),
		enabled: enabled && !!catalog && !!database,
		refetchOnWindowFocus: false,
	})
}

export const useTableMetrics = (
	catalog: string,
	database: string,
	tableName: string,
	enabled = true,
) => {
	return useQuery({
		queryKey: tableKeys.metrics(catalog, database, tableName),
		queryFn: () => tableService.getTableMetrics(catalog, database, tableName),
		enabled: enabled && !!catalog && !!database && !!tableName,
		refetchOnWindowFocus: false,
	})
}

export const useTableRuns = (
	catalog: string,
	database: string,
	tableName: string,
) => {
	return useQuery({
		queryKey: tableKeys.runs(catalog, database, tableName),
		queryFn: () => tableService.getTableRuns(catalog, database, tableName),
		select: data => mapGetTableRunsResponseToTableRuns(data.runs),
		enabled: !!catalog && !!database && !!tableName,
		refetchOnWindowFocus: false,
	})
}

export const useTableCronFormConfig = (
	catalog: string,
	database: string,
	tableName: string,
) => {
	return useQuery({
		queryKey: tableKeys.cron(catalog, database, tableName),
		queryFn: () =>
			tableService.getTableCronConfig(catalog, database, tableName),
		select: mapTableCronApiModelToTableCronFormModel,
		enabled: !!catalog && !!database && !!tableName,
		refetchOnWindowFocus: false,
	})
}

export const useProcessLogs = (runId: string) => {
	return useQuery({
		queryKey: tableKeys.processLogs(runId),
		queryFn: () => tableService.getProcessLogs(runId),
		select: mapProcessLogsResponse,
		enabled: !!runId,
		// always fetch fresh data
		staleTime: 0,
		gcTime: 0,
		refetchOnWindowFocus: false,
	})
}
