import { useQuery } from "@tanstack/react-query"

import { platformService } from "../services"

export const useCompactionStatus = () => {
	return useQuery({
		queryKey: ["platform", "compaction-status"],
		queryFn: () => platformService.getCompactionStatus(),
		staleTime: Infinity,
		refetchOnWindowFocus: false,
		retry: 1,
	})
}
