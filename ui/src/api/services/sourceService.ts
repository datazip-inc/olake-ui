import api from "../axios"
import { API_CONFIG } from "../config"
import {
	Entity,
	EntityBase,
	EntityTestRequest,
	EntityTestResponse,
	StreamsDataStructure,
} from "../../types"

import { ENTITY_TYPES } from "../../utils/constants"
import { trackTestConnection } from "../utils"

export const sourceService = {
	getSources: async (): Promise<Entity[]> => {
		try {
			const response = await api.get<Entity[]>(
				API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID),
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

	createSource: async (source: EntityBase) => {
		try {
			const response = await api.post<EntityBase>(
				API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID),
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
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/${id}`,
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
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/${id}`,
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
			const response = await api.post<EntityTestResponse>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/test`,
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
			return {
				success: false,
				message:
					error instanceof Error ? error.message : "Unknown error occurred",
				data: {
					connection_result: {
						message:
							error instanceof Error ? error.message : "Unknown error occurred",
						status: "FAILED",
					},
					logs: [],
				},
			}
		}
	},

	getSourceVersions: async (type: string) => {
		try {
			const response = await api.get<{ version: string[] }>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/versions/?type=${type}`,
				{
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
			const response = await api.post<any>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/spec`,
				{
					type: type.toLowerCase(),
					version,
				},
				{ timeout: 300000, signal, disableErrorNotification: true }, //timeout is 300000 as spec takes more time as it needs to fetch the spec from olake
			)
			return response.data
		} catch (error) {
			console.error("Error getting source spec:", error)
			throw error
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
	) => {
		try {
			const response = await api.post<StreamsDataStructure>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/streams`,
				{
					name,
					type,
					job_name,
					job_id: job_id ? job_id : -1,
					version,
					config,
				},
				{ timeout: 0 },
			)
			return response.data
		} catch (error) {
			console.error("Error getting source streams:", error)
			throw error
		}
	},

	checkSourceNameUnique: async (
		sourceName: string,
	): Promise<{ unique: boolean }> => {
		try {
			const response = await api.post<{ unique: boolean }>(
				`${API_CONFIG.ENDPOINTS.PROJECT(API_CONFIG.PROJECT_ID)}/check-unique`,
				{ name: sourceName, entity_type: ENTITY_TYPES.SOURCE },
			)
			return response.data
		} catch (error) {
			console.error("Error checking source name uniqueness:", error)
			throw error
		}
	},
}
