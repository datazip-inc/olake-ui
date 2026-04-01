import { API_CONFIG } from "@/config"
import { api } from "@/core/api"

import type {
	GetProcessLogsApiResponse,
	GetTableRunsApiResponse,
	TableDetailsApiResponse,
	GetTablesApiResponse,
	TableMetricsApiResponse,
	UpdateTableCronApiRequest,
} from "../types"

export const tableService = {
	getTables: async (catalog: string, database: string) => {
		const response = await api.get<GetTablesApiResponse>(
			API_CONFIG.ENDPOINTS.OPT.TABLES(catalog, database),
		)
		return response.data
	},

	getTableDetails: async (
		catalog: string,
		database: string,
		tableName: string,
	) => {
		const response = await api.get<TableDetailsApiResponse>(
			`${API_CONFIG.ENDPOINTS.OPT.TABLE(catalog, database, tableName)}/details`,
		)
		return response.data
	},

	getTableMetrics: async (
		catalog: string,
		database: string,
		tableName: string,
	) => {
		const response = await api.get<TableMetricsApiResponse>(
			`${API_CONFIG.ENDPOINTS.OPT.TABLE(catalog, database, tableName)}/snapshots?page=1&pageSize=1`,
		)
		return response.data
	},

	getTableRuns: async (
		catalog: string,
		database: string,
		tableName: string,
		page: number,
		pageSize: number,
		status?: string,
	) => {
		const searchParams = new URLSearchParams({
			page: String(page),
			pageSize: String(pageSize),
		})
		if (status) {
			searchParams.set("status", status)
		}

		const response = await api.get<GetTableRunsApiResponse>(
			`${API_CONFIG.ENDPOINTS.OPT.TABLE(catalog, database, tableName)}/optimizing-processes?${searchParams.toString()}`,
		)
		return response.data
	},

	cancelTableRun: async (
		catalog: string,
		database: string,
		tableName: string,
		runId: string,
	) => {
		const response = await api.post(
			`${API_CONFIG.ENDPOINTS.OPT.TABLE(catalog, database, tableName)}/optimizing-processes/${encodeURIComponent(runId)}/cancel`,
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
		const response = await api.put(
			`${API_CONFIG.ENDPOINTS.OPT.TABLE_CONFIG(catalog, database, tableName)}/config`,
			{ enabled_for_optimization: enabled.toString() },
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
			`${API_CONFIG.ENDPOINTS.OPT.TABLE_CONFIG(catalog, database, tableName)}/config`,
			payload,
		)
		return response.data
	},
	getProcessLogs: async (runId: string) => {
		const response = await api.get<GetProcessLogsApiResponse>(
			API_CONFIG.ENDPOINTS.OPT.PROCESS(runId),
			{ showNotification: true },
		)
		return response.data
	},

	downloadProcessLogFile: async (
		processId: string,
		fileId: string,
	): Promise<void> => {
		const path = `${API_CONFIG.ENDPOINTS.OPT.PROCESS(processId)}/file/${encodeURIComponent(fileId)}`
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
		const path = `${API_CONFIG.ENDPOINTS.OPT.PROCESS(processId)}/download`
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
