import { useMutation } from "@tanstack/react-query"

import { catalogKeys } from "../../constants"
import { catalogService } from "../../services"
import type { CatalogFormData } from "../../types"

export const useCreateCatalog = () => {
	return useMutation({
		mutationKey: catalogKeys.all(),
		mutationFn: (config: CatalogFormData) =>
			catalogService.createCatalog(config),
	})
}

export const useUpdateCatalog = () => {
	return useMutation({
		mutationKey: catalogKeys.all(),
		mutationFn: ({
			catalogName,
			config,
		}: {
			catalogName: string
			config: CatalogFormData
		}) => catalogService.updateCatalog(catalogName, config),
	})
}

export const useDeleteCatalog = () => {
	return useMutation({
		mutationKey: catalogKeys.all(),
		mutationFn: (catalogName: string) =>
			catalogService.deleteCatalog(catalogName),
	})
}

