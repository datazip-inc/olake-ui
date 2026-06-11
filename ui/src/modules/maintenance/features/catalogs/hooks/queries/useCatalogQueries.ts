import { useQuery } from "@tanstack/react-query"

import { catalogKeys } from "../../constants"
import { catalogService } from "../../services"
import {
	mapGetCatalogResponseToFormData,
	mapGetCatalogsResponseToCatalogs,
	mapCatalogSpecResponse,
} from "../../utils"

const CATALOG_TYPE = "iceberg"

export const useCatalogs = (enabled = true) => {
	return useQuery({
		queryKey: catalogKeys.list(),
		queryFn: () => catalogService.getCatalogs(),
		select: data => mapGetCatalogsResponseToCatalogs(data),
		enabled,
		refetchOnWindowFocus: false,
	})
}

export const useIcebergDestinations = (enabled = true) => {
	return useQuery({
		queryKey: catalogKeys.icebergDestinations(),
		queryFn: () => catalogService.getIcebergDestinations(),
		enabled,
		refetchOnWindowFocus: false,
	})
}

export const useCatalogDatabases = (catalogName: string) => {
	return useQuery({
		queryKey: catalogKeys.databases(catalogName),
		queryFn: () => catalogService.getCatalogDatabases(catalogName),
		enabled: !!catalogName,
		refetchOnWindowFocus: false,
		select: data =>
			(data.result ?? []).slice().sort((a, b) => a.localeCompare(b)),
	})
}

export const useCatalogDetails = (catalogName: string) => {
	return useQuery({
		queryKey: catalogKeys.detail(catalogName),
		queryFn: () => catalogService.getCatalog(catalogName),
		select: mapGetCatalogResponseToFormData,
		enabled: !!catalogName,
		refetchOnWindowFocus: false,
	})
}

/** Cached per (type, version) forever in-memory; evicted from cache after 24h of non-use */
export const useCatalogSpec = (isEditMode: boolean, enabled = true) => {
	return useQuery({
		queryKey: catalogKeys.spec(CATALOG_TYPE),
		queryFn: ({ signal }) => catalogService.getCatalogSpec(signal),
		select: data => mapCatalogSpecResponse(data, isEditMode),
		enabled,
		staleTime: Infinity,
		gcTime: 24 * 60 * 60 * 1000, // 24 hours
	})
}
