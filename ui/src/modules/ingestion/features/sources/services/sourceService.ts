import { AxiosError } from "axios"

import { SpecResponse, TestConnectionResponse } from "@/common/types"
import { API_CONFIG } from "@/config"
import { trackTestConnection } from "@/core/analytics/analyticsUtils"
import { api } from "@/core/api"
import {
	Entity,
	EntityBase,
	EntityTestRequest,
	StreamsDataStructure,
} from "@/modules/ingestion/common/types"

export const sourceService = {
	getSources: async (): Promise<Entity[]> => {
		try {
			const response = await api.get<Entity[]>(
				API_CONFIG.ENDPOINTS.ETL.SOURCES(API_CONFIG.PROJECT_ID),
				{ timeout: 0 }, // Disable timeout for this request since it can take longer
			)

			return response.data.map(item => ({
				...item,
				config: JSON.parse(item.config),
			}))
		} catch (error) {
			console.error("Error fetching sources from API:", error)
			throw error
		}
	},

	getSource: async (id: string): Promise<Entity> => {
		try {
			const response = await api.get<Entity>(
				`${API_CONFIG.ENDPOINTS.ETL.SOURCES(API_CONFIG.PROJECT_ID)}/${id}`,
			)

			const source = {
				...response.data,
				config: JSON.parse(response.data.config),
			}

			return source
		} catch (error) {
			console.error("Error fetching source from API:", error)
			throw error
		}
	},

	createSource: async (source: EntityBase) => {
		try {
			const response = await api.post<EntityBase>(
				API_CONFIG.ENDPOINTS.ETL.SOURCES(API_CONFIG.PROJECT_ID),
				source,
			)
			return response.data
		} catch (error) {
			console.error("Error creating source:", error)
			throw error
		}
	},

	updateSource: async (id: string, source: EntityBase) => {
		try {
			const response = await api.put<Entity>(
				`${API_CONFIG.ENDPOINTS.ETL.SOURCES(API_CONFIG.PROJECT_ID)}/${id}`,
				{
					name: source.name,
					type: source.type.toLowerCase(),
					version: source.version,
					config:
						typeof source.config === "string"
							? source.config
							: JSON.stringify(source.config),
				},
				{ showNotification: true },
			)
			return response.data
		} catch (error) {
			console.error("Error updating source:", error)
			throw error
		}
	},

	deleteSource: async (id: string) => {
		try {
			await api.delete(
				`${API_CONFIG.ENDPOINTS.ETL.SOURCES(API_CONFIG.PROJECT_ID)}/${id}`,
				{ showNotification: true },
			)
		} catch (error) {
			console.error("Error deleting source:", error)
			throw error
		}
	},

	testSourceConnection: async (
		source: EntityTestRequest,
		existing: boolean = false,
	) => {
		try {
			const response = await api.post<TestConnectionResponse>(
				`${API_CONFIG.ENDPOINTS.ETL.SOURCES(API_CONFIG.PROJECT_ID)}/test`,
				{
					type: source.type.toLowerCase(),
					version: source.version,
					config: source.config,
				},
				{ timeout: 0, disableErrorNotification: true }, // Disable timeout for this request since it can take longer
			)

			trackTestConnection(true, source, response.data, existing)
			return {
				success: true,
				message: "success",
				data: response.data,
			}
		} catch (error) {
			console.error("Error testing source connection:", error)
			const errorMessage =
				error instanceof AxiosError
					? (error.response?.data?.message ??
						"Network error - please check your connection")
					: "Unknown error occurred"
			return {
				success: false,
				message: errorMessage,
				data: {
					connection_result: {
						message: errorMessage,
						status: "FAILED",
					},
					logs: [
						{
							level: "error",
							time: new Date().toISOString(),
							message: errorMessage,
						},
					],
				},
			}
		}
	},

	getSourceVersions: async (type: string) => {
		try {
			const response = await api.get<{ version: string[] }>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/versions`,
				{
					params: { type },
					timeout: 0, // Disable timeout for this request since it can take longer
				},
			)
			return response.data
		} catch (error) {
			console.error("Error getting source versions:", error)
			throw error
		}
	},

	getSourceSpec: async (
		type: string,
		version: string,
		signal?: AbortSignal,
	) => {
		try {
			const response = await api.post<SpecResponse>(
				`${API_CONFIG.ENDPOINTS.ETL.SOURCES(API_CONFIG.PROJECT_ID)}/spec`,
				{
					type: type.toLowerCase(),
					version,
				},
				{ timeout: 300000, signal, disableErrorNotification: true }, //timeout is 300000 as spec takes more time as it needs to fetch the spec from olake
			)
			return response.data
		} catch (error: any) {
			console.error("Error getting source spec:", error)
			const serverMessage = error?.response?.data?.message
			throw new Error(
				serverMessage ?? error?.message ?? "Failed to fetch source spec",
			)
		}
	},

	//fetches source specific streams
	getSourceStreams: async (
		name: string,
		type: string,
		version: string,
		config: string,
		job_name: string,
		job_id?: number,
		max_discover_threads?: number | null,
		signal?: AbortSignal,
	) => {
		try {
			const response = await api.post<StreamsDataStructure>(
				`${API_CONFIG.ENDPOINTS.ETL.SOURCES(API_CONFIG.PROJECT_ID)}/streams`,
				{
					name,
					type,
					job_name,
					job_id: job_id ? job_id : -1,
					version,
					config,
					max_discover_threads,
				},
				{ timeout: 0, signal, disableErrorNotification: true },
			)
			return response.data
		} catch (error: any) {
			console.error("Error getting source streams:", error)
			const serverMessage = error?.response?.data?.message
			throw new Error(
				serverMessage ?? error?.message ?? "Failed to get source streams",
			)
		}
	},
}
