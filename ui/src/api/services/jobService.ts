import api from "../axios"
import { Job, JobHistory, JobLog } from "../../types"
import { mockJobs } from "../mockData"

const useMockData = true

export const jobService = {
	// Get all jobs
	getJobs: async () => {
		const response = await api.get<Job[]>("/jobs")
		return response.data
	},

	// Get job by id
	getJobById: async (id: string) => {
		const response = await api.get<Job>(`/jobs/${id}`)
		return response.data
	},

	// Create new job
	createJob: async (job: Omit<Job, "id" | "createdAt">) => {
		const response = await api.post<Job>("/jobs", job)
		return response.data
	},

	// Update job
	updateJob: async (id: string, job: Partial<Job>) => {
		const response = await api.put<Job>(`/jobs/${id}`, job)
		return response.data
	},

	// Delete job
	deleteJob: async (id: string) => {
		if (useMockData) {
			const index = mockJobs.findIndex(j => j.id === id)
			if (index === -1) throw new Error("Job not found")

			mockJobs.splice(index, 1)
			return new Promise<void>(resolve => {
				setTimeout(() => resolve(), 300)
			})
		}
		// const response = await api.delete(`/jobs/${id}`)
		// return response.data
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
}
