import { useQuery } from "@tanstack/react-query"
import { jobsKeys } from "../../constants"
import { jobService } from "../../services"

export const useJobs = () => {
	return useQuery({
		queryKey: jobsKeys.lists(),
		queryFn: () => jobService.getJobs(),
	})
}

export const useJobDetails = (id: string, options?: { staleTime?: number }) => {
	return useQuery({
		queryKey: jobsKeys.detail(id),
		queryFn: () => jobService.getJob(id),
		enabled: !!id,
		staleTime: options?.staleTime,
	})
}

export const useJobTasks = (
	jobId: string,
	options?: {
		refetchInterval?:
			| number
			| false
			| ((query: any) => number | false | undefined)
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
	options?: { staleTime?: number },
) => {
	return useQuery({
		queryKey: jobsKeys.clearDestination(jobId),
		queryFn: () => jobService.getClearDestinationStatus(jobId),
		enabled: !!jobId,
		staleTime: options?.staleTime,
	})
}
