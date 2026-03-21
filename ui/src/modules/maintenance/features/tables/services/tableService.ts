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
			{ showNotification: true },
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
			undefined,
			{ showNotification: true },
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
			API_CONFIG.ENDPOINTS.FUSION_PROCESS(runId),
			{ showNotification: true },
		)
		return response.data
	},

	downloadProcessLogFile: async (
		processId: string,
		fileId: string,
	): Promise<void> => {
		const path = `${API_CONFIG.ENDPOINTS.FUSION_PROCESS(processId)}/file/${encodeURIComponent(fileId)}`
		const url = `${API_CONFIG.BASE_URL}${path}`

		try {
			// Pre-flight check to verify endpoint is accessible
			// Check endpoint with minimal data transfer
			await api.get(url, {
				headers: { Range: "bytes=0-0" },
				responseType: "blob",
			})

			// if successful, trigger download
			const link = document.createElement("a")
			link.href = url
			link.style.display = "none"
			document.body.appendChild(link)
			link.click()
			document.body.removeChild(link)
		} catch (error) {
			throw error
		}
	},

	downloadProcessLogsArchive: async (processId: string): Promise<void> => {
		const path = `${API_CONFIG.ENDPOINTS.FUSION_PROCESS(processId)}/download`
		const url = `${API_CONFIG.BASE_URL}${path}`

		try {
			// Pre-flight check to verify endpoint is accessible
			// Check endpoint with minimal data transfer
			await api.get(url, {
				headers: { Range: "bytes=0-0" },
				responseType: "blob",
			})

			// if successful, trigger download
			const link = document.createElement("a")
			link.href = url
			link.style.display = "none"
			document.body.appendChild(link)
			link.click()
			document.body.removeChild(link)
		} catch (error) {
			throw error
		}
	},
}
