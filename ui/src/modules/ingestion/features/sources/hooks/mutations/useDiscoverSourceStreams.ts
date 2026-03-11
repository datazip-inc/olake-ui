import { sourceService } from "../../services"
import { useMutation } from "@tanstack/react-query"
import { DiscoverSourceStreamsParams } from "../../types"

export const useDiscoverSourceStreams = () => {
	return useMutation({
		mutationFn: (params: DiscoverSourceStreamsParams) =>
			sourceService.getSourceStreams(
				params.name,
				params.type,
				params.version,
				params.config,
				params.job_name,
				params.job_id,
				params.max_discover_threads,
			),
	})
}
