import { SpecResponse, TestConnectionResponse } from "@/common/types"
import { API_CONFIG } from "@/config"
import { trackTestConnection } from "@/core/analytics/analyticsUtils"
import { api } from "@/core/api"

import type {
	CatalogFormData,
	CatalogPayload,
	CatalogTestRequest,
	DestinationEntity,
	GetCatalogDatabasesResponse,
	GetCatalogsResponse,
} from "../types"

const DESTINATION_TYPE = "iceberg"

export const catalogService = {
	getIcebergDestinations: async () => {
		const response = await api.get<DestinationEntity[]>(
			API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID),
			{ timeout: 0 },
		)

		return response.data.filter(
			item => item.type.toLowerCase() === DESTINATION_TYPE,
		)
	},

	getCatalogs: async () => {
		const response = await api.get<GetCatalogsResponse>(
			API_CONFIG.ENDPOINTS.OPT.CATALOGS(),
		)
		return response.data
	},

	getCatalogDatabases: async (
		catalogName: string,
	): Promise<GetCatalogDatabasesResponse> => {
		const response = await api.get<GetCatalogDatabasesResponse>(
			`${API_CONFIG.ENDPOINTS.OPT.CATALOGS(catalogName)}/databases`,
		)
		return response.data
	},

	getCatalog: async (catalogName: string) => {
		const response = await api.get<CatalogFormData>(
			API_CONFIG.ENDPOINTS.OPT.CATALOG(catalogName),
		)
		return response.data
	},

	createCatalog: async (config: CatalogFormData) => {
		await api.post<void>(API_CONFIG.ENDPOINTS.OPT.CATALOG(), config, {
			disableErrorNotification: true,
		})
		return
	},

	updateCatalog: async (catalogName: string, config: CatalogFormData) => {
		const response = await api.put<CatalogPayload>(
			API_CONFIG.ENDPOINTS.OPT.CATALOG(catalogName),
			config,
		)
		return response.data
	},

	deleteCatalog: async (catalogName: string) => {
		await api.delete(API_CONFIG.ENDPOINTS.OPT.CATALOG(catalogName), {
			showNotification: true,
		})
		return
	},

	testCatalogConnection: async (
		catalog: CatalogTestRequest,
		existing: boolean = false,
	) => {
		try {
			const response = await api.post<TestConnectionResponse>(
				`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/test`,
				{
					type: DESTINATION_TYPE,
					version: catalog.version,
					config: catalog.config,
				},
				{ timeout: 0, disableErrorNotification: true },
			)
			trackTestConnection(
				false,
				{ ...catalog, type: DESTINATION_TYPE },
				response.data,
				existing,
			)

			return {
				success: true,
				message: "success",
				data: response.data,
			}
		} catch (error) {
			console.error("Error testing catalog connection:", error)
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

	getCatalogVersions: async () => {
		const response = await api.get<{ version: string[] }>(
			`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/versions/?type=${DESTINATION_TYPE}`,
			{
				timeout: 0,
			},
		)
		return response.data
	},

	getCatalogSpec: async (version: string, signal?: AbortSignal) => {
		try {
			const response = await api.post<SpecResponse>(
				`${API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID)}/spec`,
				{
					type: DESTINATION_TYPE,
					version: version,
				},
				{ timeout: 300000, signal, disableErrorNotification: true },
			)
			return response.data
		} catch (error: any) {
			console.error("Error getting catalog spec:", error)
			const serverMessage = error?.response?.data?.message
			throw new Error(
				serverMessage ?? error?.message ?? "Failed to fetch catalog spec",
			)
		}
	},
}
