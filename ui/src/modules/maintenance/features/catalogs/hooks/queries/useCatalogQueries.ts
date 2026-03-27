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

export const useCatalogDatabases = (catalogName: string) => {
	return useQuery({
		queryKey: catalogKeys.databases(catalogName),
		queryFn: () => catalogService.getCatalogDatabases(catalogName),
		enabled: !!catalogName,
		refetchOnWindowFocus: false,
		select: data => data.result ?? [],
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

export const useCatalogVersions = (enabled = true) => {
	return useQuery({
		queryKey: catalogKeys.versions(CATALOG_TYPE),
		queryFn: () => catalogService.getCatalogVersions(),
		enabled,
		refetchOnWindowFocus: false,
	})
}

/** Cached per (type, version) forever in-memory; evicted from cache after 24h of non-use */
export const useCatalogSpec = (
	version: string,
	isEditMode: boolean,
	enabled = true,
) => {
	return useQuery({
		queryKey: catalogKeys.spec(CATALOG_TYPE, version),
		queryFn: ({ signal }) => catalogService.getCatalogSpec(version, signal),
		select: data => mapCatalogSpecResponse(data, isEditMode),
		enabled: enabled && !!version,
		staleTime: Infinity,
		gcTime: 24 * 60 * 60 * 1000, // 24 hours
	})
}
