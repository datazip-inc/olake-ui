import { useMutation } from "@tanstack/react-query"
import { useRef } from "react"

import { sourceService } from "../../services"
import { DiscoverSourceStreamsParams } from "../../types"

export const useDiscoverSourceStreams = () => {
	const abortControllerRef = useRef<AbortController | null>(null)

	const cancel = () => {
		abortControllerRef.current?.abort()
		abortControllerRef.current = null
	}

	const mutation = useMutation({
		mutationFn: (params: DiscoverSourceStreamsParams) => {
			cancel()
			const abortController = new AbortController()
			abortControllerRef.current = abortController
			return sourceService.getSourceStreams(
				params.name,
				params.type,
				params.version,
				params.config,
				params.job_name,
				params.job_id,
				params.max_discover_threads,
				abortController.signal,
			)
		},
	})

	return {
		...mutation,
		cancel,
	}
}
