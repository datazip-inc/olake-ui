import api from "../axios"
import { API_CONFIG } from "../config"
import {
	Entity,
	APIResponse,
	EntityBase,
	EntityTestRequest,
	EntityTestResponse,
} from "../../types"

export const sourceService = {
	getSources: async (): Promise<Entity[]> => {
		try {
			const response = await api.get<APIResponse<Entity[]>>(
				API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID),
			)

			return response.data.data.map(item => ({
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
			const response = await api.post<APIResponse<EntityBase>>(
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
			const response = await api.put<APIResponse<Entity>>(
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
			)
		} catch (error) {
			console.error("Error deleting source:", error)
			throw error
		}
	},

	testSourceConnection: async (source: EntityTestRequest) => {
		try {
			const response = await api.post<APIResponse<EntityTestResponse>>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/test`,
				{
					type: source.type.toLowerCase(),
					version: source.version,
					config: source.config,
				},
				{ timeout: 0 },
			)
			return {
				success: response.data.success,
				message: response.data.message,
				data: response.data.data,
			}
		} catch (error) {
			console.error("Error testing source connection:", error)
			return {
				success: false,
				message:
					error instanceof Error ? error.message : "Unknown error occurred",
			}
		}
	},

	getSourceVersions: async (type: string) => {
		try {
			const response = await api.get<APIResponse<{ version: string[] }>>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/versions/?type=${type}`,
			)
			return response.data
		} catch (error) {
			console.error("Error getting source versions:", error)
			throw error
		}
	},

	getSourceSpec: async (type: string, version: string) => {
		try {
			const response = await api.post<APIResponse<Record<string, unknown>>>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/spec`,
				{
					type: type.toLowerCase(),
					version,
				},
			)
			return response.data
		} catch (error) {
			console.error("Error getting source spec:", error)
			throw error
		}
	},

	getSourceStreams: async (
		name: string,
		type: string,
		version: string,
		config: string,
	) => {
		try {
			const response = await api.post<APIResponse<Record<string, unknown>>>(
				`${API_CONFIG.ENDPOINTS.SOURCES(API_CONFIG.PROJECT_ID)}/streams`,
				{
					name,
					type,
					version: version === "" ? "latest" : version,
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
}
