import { useQuery } from "@tanstack/react-query"

import { platformService } from "../services"

export const useOptimizationStatus = () => {
	return useQuery({
		queryKey: ["platform", "optimization-status"],
		queryFn: () => platformService.getOptimizationStatus(),
		staleTime: Infinity,
		refetchOnWindowFocus: false,
		retry: 1,
	})
}
