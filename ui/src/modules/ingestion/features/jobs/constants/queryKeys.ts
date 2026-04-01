import { API_CONFIG } from "@/config/apiConfig"

export const jobsKeys = {
	all: () => ["projects", API_CONFIG.PROJECT_ID, "jobs"] as const,

	lists: () => [...jobsKeys.all(), "list"] as const,

	details: () => [...jobsKeys.all(), "details"] as const,
	detail: (jobId: string) => [...jobsKeys.details(), jobId] as const,

	streams: (jobId: string) => [...jobsKeys.detail(jobId), "streams"] as const,

	tasks: (jobId: string) => [...jobsKeys.detail(jobId), "tasks"] as const,

	clearDestination: (jobId: string) =>
		[...jobsKeys.detail(jobId), "clear-destination"] as const,
}
