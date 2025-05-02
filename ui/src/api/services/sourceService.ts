import api from "../axios"
import {
	Entity,
	APIResponse,
	EntityBase,
	EntityTestResponse,
} from "../../types"

export const sourceService = {
	getSources: async () => {
		try {
			const response = await api.get<APIResponse<Entity[]>>(
				"/api/v1/project/123/sources",
			)

			const sources: Entity[] = response.data.data.map(item => {
				const config = JSON.parse(item.config)

				return {
					...item,
					config,
				}
			})

			return sources
		} catch (error) {
			console.error("Error fetching sources from API:", error)
			throw error
		}
	},

	// Create new source
	createSource: async (source: EntityBase) => {
		const response = await api.post<EntityBase>(
			"/api/v1/project/123/sources",
			source,
		)
		return response.data
	},

	// Update source
	updateSource: async (id: string, source: any) => {
		try {
			const response = await api.put<APIResponse<any>>(
				`/api/v1/project/123/sources/${id}`,
				{
					name: source.name,
					type: source.type.toLowerCase(),
					version: source.version,
					config: JSON.stringify(source.config),
				},
			)
			return response.data
		} catch (error) {
			console.error("Error updating source:", error)
			throw error
		}
	},

	// Delete source
	deleteSource: async (id: number) => {
		await api.delete(`/api/v1/project/123/sources/${id}`)
		return
	},

	// Test source connection
	testSourceConnection: async (source: EntityTestResponse) => {
		try {
			const response = await api.post<APIResponse<EntityTestResponse>>(
				"/api/v1/project/123/sources/test",
				{
					type: source.type.toLowerCase(),
					version: "1.0.0",
					config: source.config,
				},
			)
			return {
				success: response.data.success,
				message: response.data.message,
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
		const response = await api.get<APIResponse<{ version: string[] }>>(
			`/api/v1/project/123/sources/versions/?type=${type}`,
		)
		return response.data
	},

	getSourceSpec: async (type: string, version: string) => {
		const response = await api.post<APIResponse<any>>(
			`/api/v1/project/123/sources/spec`,
			{
				type: type.toLowerCase(),
				version: version,
			},
		)
		return response.data
	},

	getSourceStreams: async (
		name: string,
		type: string,
		version: string,
		config: string,
	) => {
		try {
			const response = await api.post<APIResponse<any>>(
				"/api/v1/project/123/sources/streams",
				{
					name: name,
					type: type,
					version: version,
					config: config,
				},
			)
			return response.data
		} catch (error) {
			console.error("Error getting source streams:", error)
			throw error
		}
	},
}
