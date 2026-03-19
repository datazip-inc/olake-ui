import { API_CONFIG } from "@/config"
import { api } from "@/core/api"

import type {
	GetProcessLogsApiResponse,
	GetTableRunsApiResponse,
	GetTablesApiResponse,
	TableCronApiModel,
	TableMetrics,
	UpdateTableCronApiRequest,
} from "../types"

export const tableService = {
	getTables: async (catalog: string, database: string) => {
		const response = await api.get<GetTablesApiResponse>(
			API_CONFIG.ENDPOINTS.FUSION_TABLE(catalog, database),
		)
		return response.data
	},

	getTableMetrics: async (
		catalog: string,
		database: string,
		tableName: string,
	) => {
		const response = await api.get<TableMetrics>(
			`${API_CONFIG.ENDPOINTS.FUSION_TABLE(catalog, database, tableName)}/metrics`,
		)
		return response.data
	},

	getTableRuns: async (
		catalog: string,
		database: string,
		tableName: string,
	) => {
		const response = await api.get<GetTableRunsApiResponse>(
			`${API_CONFIG.ENDPOINTS.FUSION_TABLE(catalog, database, tableName)}/runs`,
		)
		return response.data
	},

	cancelTableRun: async (
		catalog: string,
		database: string,
		tableName: string,
	) => {
		const response = await api.post(
			`${API_CONFIG.ENDPOINTS.FUSION_TABLE(catalog, database, tableName)}/runs/cancel`,
		)
		return response.data
	},

	setTableOptimizing: async (
		catalog: string,
		database: string,
		tableName: string,
		enabled: boolean,
	) => {
		const response = await api.post(
			`${API_CONFIG.ENDPOINTS.FUSION_TABLE(catalog, database, tableName)}/${enabled ? "enable-optimizing" : "disable-optimizing"}`,
		)
		return response.data
	},

	getTableCronConfig: async (
		catalog: string,
		database: string,
		tableName: string,
	) => {
		const response = await api.get<TableCronApiModel>(
			`${API_CONFIG.ENDPOINTS.FUSION_TABLE(catalog, database, tableName)}/cron`,
		)
		return response.data
	},

	updateTableCronConfig: async (
		catalog: string,
		database: string,
		tableName: string,
		payload: UpdateTableCronApiRequest,
	) => {
		const response = await api.put(
			`${API_CONFIG.ENDPOINTS.FUSION_TABLE(catalog, database, tableName)}/cron`,
			payload,
		)
		return response.data
	},
	getProcessLogs: async (runId: string) => {
		const response = await api.get<GetProcessLogsApiResponse>(
			API_CONFIG.ENDPOINTS.FUSION_PROCESS_LOGS(runId),
		)
		return response.data
	},
}
