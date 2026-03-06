import { useQuery } from "@tanstack/react-query"
import { destinationService } from "../../services"
import { destinationKeys } from "../../constants/queryKeys"

export const useDestinations = () => {
	return useQuery({
		queryKey: destinationKeys.lists(),
		queryFn: () => destinationService.getDestinations(),
	})
}

export const useDestinationDetails = (id: string) => {
	return useQuery({
		queryKey: destinationKeys.detail(id),
		queryFn: () => destinationService.getDestination(id),
		enabled: !!id,
	})
}

export const useDestinationVersions = (type: string) => {
	return useQuery({
		queryKey: destinationKeys.versions(type),
		queryFn: () => destinationService.getDestinationVersions(type),
		enabled: !!type,
	})
}

/** Cached per (type, version) forever in-memory; evicted from cache after 24h of non-use */
export const useDestinationSpec = (
	type: string,
	version: string,
	sourceType: string = "",
	sourceVersion: string = "",
) => {
	return useQuery({
		queryKey: destinationKeys.spec(type, version),
		queryFn: ({ signal }) =>
			destinationService.getDestinationSpec(
				type,
				version,
				sourceType,
				sourceVersion,
				signal,
			),
		enabled: !!type && !!version,
		staleTime: Infinity,
		gcTime: 24 * 60 * 60 * 1000, // 24 hours
	})
}
