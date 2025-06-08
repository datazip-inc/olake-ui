import api from "../axios"
import { API_CONFIG } from "../config"
import {
	APIResponse,
	Entity,
	EntityBase,
	EntityTestRequest,
	EntityTestResponse,
} from "../../types"

const normalizeDestinationType = (type: string): string => {
	const typeMap: Record<string, string> = {
		"amazon s3": "s3",
		"apache iceberg": "iceberg",
	}
	return typeMap[type.toLowerCase()] || type.toLowerCase()
}

const normalizeCatalogType = (catalog: string | null): string => {
	if (!catalog) return "none"

	const catalogMap: Record<string, string> = {
		"aws glue": "glue",
		"rest catalog": "rest",
		"jdbc catalog": "jdbc",
		"hive catalog": "hive",
	}
	return catalogMap[catalog.toLowerCase()] || catalog.toLowerCase()
}

export const destinationService = {
	getDestinations: async () => {
		try {
			const response = await api.get<APIResponse<Entity[]>>(
				API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID),
			)
			const destinations: Entity[] = response.data.data.map(item => {
				const config = JSON.parse(item.config)
				return {
					...item,
					config,
					status: "active",
				}
			})

			return destinations
		} catch (error) {
			console.error("Error fetching sources from API:", error)
			throw error
		}
	},

	createDestination: async (
		destination: Omit<EntityBase, "id" | "createdAt">,
	) => {
		const response = await api.post<EntityBase>(
			API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID),
			destination,
		)
		return response.data
	},

	updateDestination: async (id: string, destination: EntityBase) => {
		try {
			const response = await api.put<APIResponse<EntityBase>>(
				`${API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID)}/${id}`,
				{
					name: destination.name,
					type: destination.type,
					version: destination.version,
					config:
						typeof destination.config === "string"
							? destination.config
							: JSON.stringify(destination.config),
				},
			)
			return response.data
		} catch (error) {
			console.error("Error updating destination:", error)
			throw error
		}
	},

	deleteDestination: async (id: number) => {
		await api.delete(
			`${API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID)}/${id}`,
		)
		return
	},

	testDestinationConnection: async (destination: EntityTestRequest) => {
		try {
			const response = await api.post<APIResponse<EntityTestResponse>>(
				`${API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID)}/test`,
				{
					type: destination.type.toLowerCase(),
					version: destination.version,
					config: destination.config,
				},
				{ timeout: 0 },
			)
			return {
				success: response.data.success,
				message: response.data.message,
				data: response.data.data,
			}
		} catch (error) {
			console.error("Error testing destination connection:", error)
			return {
				success: false,
				message:
					error instanceof Error ? error.message : "Unknown error occurred",
			}
		}
	},

	getDestinationVersions: async (type: string) => {
		const response = await api.get<APIResponse<{ version: string[] }>>(
			`${API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID)}/versions/?type=${type}`,
		)
		return response.data
	},

	getDestinationSpec: async (
		type: string,
		catalog: string | null,
		version: string,
	) => {
		const normalizedType = normalizeDestinationType(type)
		let normalizedCatalog = normalizeCatalogType(catalog)

		if (normalizedType === "iceberg" && normalizedCatalog === "none") {
			normalizedCatalog = "glue"
		}

		const response = await api.post<APIResponse<any>>(
			`${API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID)}/spec`,
			{
				type: normalizedType,
				version: version === "" ? "latest" : version,
				catalog: normalizedCatalog,
			},
		)
		return response.data
	},
}
