import {
	OperationAccepted,
	runOperation,
} from "@/common/services/operationsService"
import { SpecResponse, TestConnectionResponse } from "@/common/types"
import { API_CONFIG } from "@/config"
import { trackTestConnection } from "@/core/analytics/analyticsUtils"
import { api } from "@/core/api"
import {
	Entity,
	EntityBase,
	EntityTestRequest,
} from "@/modules/ingestion/common/types"
import { describeConnectionFailure } from "@/modules/ingestion/common/utils"
import {
	getConnectorInLowerCase,
	normalizeConnectorType,
} from "@/modules/ingestion/common/utils"

export const destinationService = {
	getDestinations: async () => {
		try {
			const response = await api.get<Entity[]>(
				API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID),
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

	getDestination: async (id: string) => {
		try {
			const response = await api.get<Entity>(
				`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/${id}`,
			)
			const config = response.data.config
				? JSON.parse(response.data.config)
				: null
			return {
				...response.data,
				type: normalizeConnectorType(response.data.type),
				config,
			}
		} catch (error) {
			console.error("Error fetching destination from API:", error)
			throw error
		}
	},

	createDestination: async (
		destination: Omit<EntityBase, "id" | "createdAt">,
	) => {
		const response = await api.post<EntityBase>(
			API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID),
			destination,
		)
		return response.data
	},

	updateDestination: async (id: string, destination: EntityBase) => {
		try {
			const response = await api.put<EntityBase>(
				`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/${id}`,
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
			`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/${id}`,
			{ showNotification: true },
		)
		return
	},

	testDestinationConnection: async (
		destination: EntityTestRequest,
		existing: boolean = false,
		source_type: string = "",
		source_version: string = "",
	) => {
		try {
			// The POST only starts the check; the verdict arrives by polling, so the
			// request is not held open for the connector container.
			const data = await runOperation<TestConnectionResponse>(async () => {
				const response = await api.post<OperationAccepted>(
					`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/test`,
					{
						type: getConnectorInLowerCase(destination.type),
						version: destination.version,
						config: destination.config,
						source_type: source_type,
						source_version: source_version,
					},
					// Untimed: resolves the driver image from the registry server-side
					// before starting the workflow, and that lookup has no deadline.
					{ timeout: 0, disableErrorNotification: true },
				)
				return response.data
			})
			trackTestConnection(false, destination, data, existing)

			return {
				success: true,
				message: "success",
				data,
			}
		} catch (error) {
			console.error("Error testing destination connection:", error)
			const errorMessage = describeConnectionFailure(error)
			return {
				success: false,
				message: errorMessage,
				data: {
					connection_result: {
						message: errorMessage,
						status: "FAILED",
					},
					logs: [],
				},
			}
		}
	},

	getDestinationVersions: async (type: string) => {
		const response = await api.get<{ version: string[] }>(
			`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/versions`,
			{
				params: { type },
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
		try {
			const normalizedType = normalizeConnectorType(type)
			// Resumable: leaving the form mid-fetch and returning rejoins the same
			// operation rather than starting a second container.
			return await runOperation<SpecResponse>(
				async () => {
					const response = await api.post<OperationAccepted>(
						`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/spec`,
						{
							type: normalizedType,
							version: version,
							source_type: source_type,
							source_version: source_version,
						},
						{
							// Untimed: resolves the driver image from the registry
							// server-side before starting the workflow.
							timeout: 0,
							signal,
							disableErrorNotification: true,
						},
					)
					return response.data
				},
				{ signal },
			)
		} catch (error: any) {
			console.error("Error getting destination spec:", error)
			const serverMessage = error?.response?.data?.message
			throw new Error(
				serverMessage ?? error?.message ?? "Failed to fetch destination spec",
			)
		}
	},
}
