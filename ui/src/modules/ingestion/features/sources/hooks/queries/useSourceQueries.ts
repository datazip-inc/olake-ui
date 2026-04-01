import { useQuery } from "@tanstack/react-query"

import { sourceKeys } from "../../constants/queryKeys"
import { sourceService } from "../../services"

export const useSources = () => {
	return useQuery({
		queryKey: sourceKeys.lists(),
		queryFn: () => sourceService.getSources(),
	})
}

export const useSourceDetails = (id: string) => {
	return useQuery({
		queryKey: sourceKeys.detail(id),
		queryFn: () => sourceService.getSource(id),
		enabled: !!id,
		refetchOnWindowFocus: false,
	})
}

export const useSourceVersions = (type: string) => {
	return useQuery({
		queryKey: sourceKeys.versions(type),
		queryFn: () => sourceService.getSourceVersions(type),
		enabled: !!type,
		refetchOnWindowFocus: false,
	})
}

/** Cached per (type, version) forever in-memory; evicted from cache after 24h of non-use */
export const useSourceSpec = (type: string, version: string) => {
	return useQuery({
		queryKey: sourceKeys.spec(type, version),
		queryFn: ({ signal }) => sourceService.getSourceSpec(type, version, signal),
		enabled: !!type && !!version,
		staleTime: Infinity,
		gcTime: 24 * 60 * 60 * 1000, // 24 hours
	})
}
