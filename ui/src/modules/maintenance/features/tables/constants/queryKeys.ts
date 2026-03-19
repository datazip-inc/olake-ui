import { API_CONFIG } from "@/config/apiConfig"

export const tableKeys = {
	all: () => ["projects", API_CONFIG.PROJECT_ID, "tables"] as const,

	list: (catalog: string, database: string) =>
		[...tableKeys.all(), "list", catalog, database] as const,

	table: (catalog: string, database: string, tableName: string) =>
		[...tableKeys.all(), "table", catalog, database, tableName] as const,
	runs: (catalog: string, database: string, tableName: string) =>
		[...tableKeys.table(catalog, database, tableName), "runs"] as const,
	cron: (catalog: string, database: string, tableName: string) =>
		[...tableKeys.table(catalog, database, tableName), "cron"] as const,
	metrics: (catalog: string, database: string, tableName: string) =>
		[...tableKeys.table(catalog, database, tableName), "metrics"] as const,
	processLogs: (runId: string) =>
		[...tableKeys.all(), "processLogs", runId] as const,
}
