import api from "../axios"
import { API_CONFIG } from "../config"
import {
	Entity,
	EntityBase,
	EntityTestRequest,
	EntityTestResponse,
} from "../../types"
import {
	getConnectorInLowerCase,
	normalizeConnectorType,
} from "../../utils/utils"

export const destinationService = {
	getDestinations: async () => {
		try {
			const response = await api.get<Entity[]>(
				API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID),
				{ timeout: 0 }, // Disable timeout for this request since it can take longer
			)
			const destinations: Entity[] = response.data.map(item => {
				const config = JSON.parse(item.config)
				return {
					...item,
					type: normalizeConnectorType(item.type),
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
			const response = await api.put<EntityBase>(
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
				{ showNotification: true },
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
			{ showNotification: true },
		)
		return
	},

	testDestinationConnection: async (
		destination: EntityTestRequest,
		source_type: string = "",
		source_version: string = "",
	) => {
		try {
			const response = await api.post<EntityTestResponse>(
				`${API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID)}/test`,
				{
					type: getConnectorInLowerCase(destination.type),
					version: destination.version,
					config: destination.config,
					source_type: source_type,
					source_version: source_version,
				},
				//timeout is 0 as test connection takes more time as it needs to connect to the destination
				{ timeout: 0, disableErrorNotification: true },
			)
			return {
				success: true,
				message: "success",
				data: response.data,
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
		const response = await api.get<{ version: string[] }>(
			`${API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID)}/versions/?type=${type}`,
			{
				timeout: 0,
			},
		)
		return response.data
	},

	getDestinationSpec: async (
		type: string,
		version: string,
		source_type: string = "",
		source_version: string = "",
		signal?: AbortSignal,
	) => {
		const normalizedType = normalizeConnectorType(type)
		const response = await api.post<any>(
			`${API_CONFIG.ENDPOINTS.DESTINATIONS(API_CONFIG.PROJECT_ID)}/spec`,
			{
				type: normalizedType,
				version: version,
				source_type: source_type,
				source_version: source_version,
			},
			//timeout is 300000 as spec takes more time as it needs to fetch the spec from the destination
			{ timeout: 300000, signal, disableErrorNotification: true },
		)
		return response.data
	},

	checkDestinationNameUnique: async (
		destinationName: string,
	): Promise<{ unique: boolean }> => {
		try {
			const response = await api.post<{ unique: boolean }>(
				`${API_CONFIG.ENDPOINTS.PROJECT(API_CONFIG.PROJECT_ID)}/check-unique`,
				{ name: destinationName, entity_type: "destination" },
			)
			return response.data
		} catch (error) {
			console.error("Error checking destination name uniqueness:", error)
			throw error
		}
	},
}
