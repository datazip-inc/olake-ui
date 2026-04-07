import { keepPreviousData, useQuery } from "@tanstack/react-query"

import { tableKeys } from "../../constants"
import { tableService } from "../../services"
import {
	mapGetTablesResponseToTables,
	mapGetTableRunsResponseToTableRuns,
	mapTableDetailsResponseToTableDetailsViewModel,
	mapTableMetricsResponseToFileSummary,
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

export const useTableDetails = (
	catalog: string,
	database: string,
	tableName: string,
	enabled = true,
) => {
	return useQuery({
		queryKey: tableKeys.details(catalog, database, tableName),
		queryFn: () => tableService.getTableDetails(catalog, database, tableName),
		select: mapTableDetailsResponseToTableDetailsViewModel,
		enabled: enabled && !!catalog && !!database && !!tableName,
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
		select: mapTableMetricsResponseToFileSummary,
		enabled: enabled && !!catalog && !!database && !!tableName,
		refetchOnWindowFocus: false,
	})
}

export const useTableRuns = (
	catalog: string,
	database: string,
	tableName: string,
	page: number,
	pageSize: number,
	status?: string,
) => {
	return useQuery({
		queryKey: tableKeys.runs(
			catalog,
			database,
			tableName,
			page,
			pageSize,
			status,
		),
		queryFn: () =>
			tableService.getTableRuns(
				catalog,
				database,
				tableName,
				page,
				pageSize,
				status,
			),
		select: mapGetTableRunsResponseToTableRuns,
		enabled: !!catalog && !!database && !!tableName,
		placeholderData: keepPreviousData,
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
