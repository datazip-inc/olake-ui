import { useQuery } from "@tanstack/react-query"

import { jobsKeys } from "../../constants"
import { jobService } from "../../services"

export const useJobs = (options?: {
	refetchInterval?: number | false | undefined
}) => {
	return useQuery({
		queryKey: jobsKeys.lists(),
		queryFn: () => jobService.getJobs(),
		refetchInterval: options?.refetchInterval,
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
