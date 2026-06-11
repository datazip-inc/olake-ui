import { SpecResponse } from "@/common/types"
import { API_CONFIG } from "@/config"
import { api } from "@/core/api"

import type {
	CatalogFormData,
	CatalogPayload,
	DestinationEntity,
	GetCatalogDatabasesResponse,
	GetCatalogsResponse,
} from "../types"

const DESTINATION_TYPE = "iceberg"

export const catalogService = {
	getIcebergDestinations: async () => {
		const response = await api.get<DestinationEntity[]>(
			API_CONFIG.ENDPOINTS.ETL.DESTINATIONS(API_CONFIG.PROJECT_ID),
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
			{ disableErrorNotification: true },
		)
		return response.data
	},

	deleteCatalog: async (catalogName: string) => {
		await api.delete(API_CONFIG.ENDPOINTS.OPT.CATALOG(catalogName), {
			showNotification: true,
		})
		return
	},

	getCatalogSpec: async (signal?: AbortSignal): Promise<SpecResponse> => {
		try {
			const response = await api.get<SpecResponse>(
				API_CONFIG.ENDPOINTS.OPT.CATALOG_SPEC,
				{
					timeout: 300000,
					signal,
					disableErrorNotification: true,
				},
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
