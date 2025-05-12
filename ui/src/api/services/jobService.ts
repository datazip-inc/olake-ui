import api from "../axios"
import {
	APIResponse,
	Job,
	JobBase,
	JobHistory,
	JobLog,
	JobTask,
	TaskLog,
} from "../../types"

export const jobService = {
	// Get all jobs
	getJobs: async () => {
		try {
			const response = await api.get<APIResponse<Job[]>>(
				"/api/v1/project/123/jobs",
			)

			const jobs: Job[] = response.data.data.map(item => {
				return {
					...item,
				}
			})

			return jobs
		} catch (error) {
			console.error("Error fetching jobs from API:", error)
			throw error
		}
	},

	// Get job by id
	getJobById: async (id: string) => {
		const response = await api.get<Job>(`/jobs/${id}`)
		return response.data
	},

	// Create new job
	createJob: async (job: JobBase) => {
		const response = await api.post<Job>("/api/v1/project/123/jobs", job)
		return response.data
	},

	// Update job
	updateJob: async (id: string, job: Partial<Job>) => {
		const response = await api.put<Job>(`/api/v1/project/123/jobs/${id}`, job)
		return response.data
	},

	// Delete job
	deleteJob: async (id: number) => {
		await api.delete(`/api/v1/project/123/jobs/${id}`)
		return
	},

	// Run job
	runJob: async (id: string) => {
		const response = await api.post(`/jobs/${id}/run`)
		return response.data
	},

	// Stop job
	stopJob: async (id: string) => {
		const response = await api.post(`/jobs/${id}/stop`)
		return response.data
	},

	// Get job history
	getJobHistory: async (id: string) => {
		const response = await api.get<JobHistory[]>(`/jobs/${id}/history`)
		return response.data
	},

	// Get job logs
	getJobLogs: async (jobId: string, historyId: string) => {
		const response = await api.get<JobLog[]>(
			`/jobs/${jobId}/history/${historyId}/logs`,
		)
		return response.data
	},

	// Sync job
	syncJob: async (id: string) => {
		const response = await api.post<APIResponse<any>>(
			`/api/v1/project/123/jobs/${id}/sync`,
			{},
			{ timeout: 0 }, // Disable timeout for this request since it can take longer
		)
		return response.data
	},

	// Get job tasks
	getJobTasks: async (id: string) => {
		try {
			const response = await api.get<APIResponse<JobTask[]>>(
				`/api/v1/project/123/jobs/${id}/tasks`,
				{ timeout: 0 }, // Disable timeout for this request
			)

			return response.data
		} catch (error) {
			console.error("Error fetching job tasks:", error)
			throw error
		}
	},

	// Get task logs
	getTaskLogs: async (jobId: string, taskId: string, filePath: string) => {
		try {
			const response = await api.post<APIResponse<TaskLog[]>>(
				`/api/v1/project/123/jobs/${jobId}/tasks/${taskId}/logs`,
				{ file_path: filePath },
			)

			return response.data
		} catch (error) {
			console.error("Error fetching task logs:", error)
			throw error
		}
	},

	// Activate/Deactivate job
	activateJob: async (jobId: string, activate: boolean) => {
		try {
			const response = await api.post<APIResponse<any>>(
				`/api/v1/project/123/jobs/${jobId}/activate`,
				{ activate },
			)
			return response.data
		} catch (error) {
			console.error("Error toggling job activation:", error)
			throw error
		}
	},
}
