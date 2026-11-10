import { API_CONFIG } from "@/config/apiConfig"

import type { TableRunsQuery } from "../types"

export const tableKeys = {
	all: () => ["projects", API_CONFIG.PROJECT_ID, "tables"] as const,

	list: (catalog: string, database: string) =>
		[...tableKeys.all(), "list", catalog, database] as const,

	table: (catalog: string, database: string, tableName: string) =>
		[...tableKeys.all(), "table", catalog, database, tableName] as const,
	// page/pageSize are left out when every run is fetched at once and paged in the browser.
	runs: ({
		catalog,
		database,
		tableName,
		page,
		pageSize,
		status,
		runType,
		excludeRunType,
	}: Omit<TableRunsQuery, "page" | "pageSize"> &
		Partial<Pick<TableRunsQuery, "page" | "pageSize">>) =>
		[
			...tableKeys.table(catalog, database, tableName),
			"runs",
			{ page, pageSize, status, runType, excludeRunType },
		] as const,
	cron: (catalog: string, database: string, tableName: string) =>
		[...tableKeys.table(catalog, database, tableName), "cron"] as const,
	details: (catalog: string, database: string, tableName: string) =>
		[...tableKeys.table(catalog, database, tableName), "details"] as const,
	metrics: (catalog: string, database: string, tableName: string) =>
		[...tableKeys.table(catalog, database, tableName), "metrics"] as const,
	processLogs: (runId: string) =>
		[...tableKeys.all(), "processLogs", runId] as const,
}
