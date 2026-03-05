import { useMutation } from "@tanstack/react-query"
import { jobService } from "../../services"
import { JobBase } from "../../types"
import { jobsKeys } from "../../constants"

// mutationKey: jobsKeys.all() tells the global MutationCache (in App.tsx) to
// invalidate all job-related queries after this mutation succeeds.
// Use it when the mutation changes job data (create, update, delete, status changes).
// Omit it for mutations that don't affect the job query cache.

export const useCreateJob = () => {
	return useMutation({
		mutationKey: jobsKeys.all(),
		mutationFn: (job: JobBase) => jobService.createJob(job),
	})
}

export const useUpdateJob = () => {
	return useMutation({
		mutationKey: jobsKeys.all(),
		mutationFn: ({ jobId, job }: { jobId: string; job: JobBase }) =>
			jobService.updateJob(jobId, job),
	})
}

export const useDeleteJob = () => {
	return useMutation({
		mutationKey: jobsKeys.all(),
		mutationFn: (jobId: number) => jobService.deleteJob(jobId),
	})
}

export const useCancelJob = () => {
	return useMutation({
		mutationKey: jobsKeys.all(),
		mutationFn: (jobId: string) => jobService.cancelJob(jobId),
	})
}

export const useSyncJob = () => {
	return useMutation({
		mutationKey: jobsKeys.all(),
		mutationFn: (jobId: string) => jobService.syncJob(jobId),
	})
}

export const useActivateJob = () => {
	return useMutation({
		mutationKey: jobsKeys.all(),
		mutationFn: ({ jobId, activate }: { jobId: string; activate: boolean }) =>
			jobService.activateJob(jobId, activate),
	})
}

export const useClearDestination = () => {
	return useMutation({
		mutationKey: jobsKeys.all(),
		mutationFn: (jobId: string) => jobService.clearDestination(jobId),
	})
}

export const useDownloadTaskLogs = () => {
	return useMutation({
		mutationFn: ({ jobId, filePath }: { jobId: string; filePath: string }) =>
			jobService.downloadTaskLogs(jobId, filePath),
	})
}
