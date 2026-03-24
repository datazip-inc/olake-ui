import { useQuery } from "@tanstack/react-query"

import { jobsKeys } from "../../constants"
import { jobService } from "../../services"

export const useJobs = () => {
	return useQuery({
		queryKey: jobsKeys.lists(),
		queryFn: () => jobService.getJobs(),
	})
}

export const useJobsWithNotification = () => {
	return useQuery({
		queryKey: jobsKeys.lists(),
		queryFn: () => jobService.getJobs(true),
		enabled: false,
	})
}

export const useJobDetails = (
	id: string,
	options?: {
		staleTime?: number
		refetchOnMount?: boolean | "always"
		gcTime?: number
	},
) => {
	return useQuery({
		queryKey: jobsKeys.detail(id),
		queryFn: () => jobService.getJob(id),
		enabled: !!id,
		staleTime: options?.staleTime,
		refetchOnMount: options?.refetchOnMount,
		gcTime: options?.gcTime,
		refetchOnWindowFocus: false,
	})
}

export const useJobTasks = (
	jobId: string,
	options?: {
		refetchInterval?: number | false | undefined
	},
) => {
	return useQuery({
		queryKey: jobsKeys.tasks(jobId),
		queryFn: () => jobService.getJobTasks(jobId),
		enabled: !!jobId,
		refetchInterval: options?.refetchInterval,
	})
}

export const useClearDestinationStatus = (
	jobId: string,
	options?: { refetchOnWindowFocus?: boolean },
) => {
	const query = useQuery({
		queryKey: jobsKeys.clearDestination(jobId),
		queryFn: () => jobService.getClearDestinationStatus(jobId),
		enabled: !!jobId,
		refetchOnWindowFocus: options?.refetchOnWindowFocus,
		refetchOnMount: "always",
		staleTime: 0,
		gcTime: 0,
	})

	return {
		...query,
		// TanStack Query's SWR behavior serves cached data while refetching in the background.
		// `isFetchedAfterMount` suppresses stale values until the first fresh response is confirmed.
		data: query.isFetchedAfterMount ? query.data : undefined,
	}
}
