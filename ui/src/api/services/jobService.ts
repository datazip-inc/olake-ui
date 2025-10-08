import api from "../axios"
import { API_CONFIG } from "../config"
import { APIResponse, Job, JobBase, JobTask, TaskLog } from "../../types"

export const jobService = {
	getJobs: async (): Promise<Job[]> => {
		try {
			const response = await api.get<APIResponse<Job[]>>(
				API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID),
			)

			return response.data.data
		} catch (error) {
			console.error("Error fetching jobs from API:", error)
			throw error
		}
	},

	createJob: async (job: JobBase): Promise<Job> => {
		try {
			const response = await api.post<Job>(
				API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID),
				job,
			)
			return response.data
		} catch (error) {
			console.error("Error creating job:", error)
			throw error
		}
	},

	updateJob: async (id: string, job: Partial<Job>): Promise<Job> => {
		try {
			const response = await api.put<Job>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}`,
				job,
			)
			return response.data
		} catch (error) {
			console.error("Error updating job:", error)
			throw error
		}
	},

	deleteJob: async (id: number): Promise<void> => {
		try {
			await api.delete(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}`,
			)
		} catch (error) {
			console.error("Error deleting job:", error)
			throw error
		}
	},

	cancelJob: async (id: string): Promise<string> => {
		try {
			const response = await api.get<APIResponse<any>>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}/cancel`,
			)
			return response.data.data.message
		} catch (error) {
			console.error("Error canceling job:", error)
			throw error
		}
	},

	syncJob: async (id: string): Promise<any> => {
		try {
			const response = await api.post<APIResponse<any>>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}/sync`,
				{},
				{ timeout: 0 }, // Disable timeout for this request since it can take longer
			)
			return response.data
		} catch (error) {
			console.error("Error syncing job:", error)
			throw error
		}
	},

	getJobTasks: async (id: string): Promise<APIResponse<JobTask[]>> => {
		try {
			const response = await api.get<APIResponse<JobTask[]>>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}/tasks`,
				{ timeout: 0 }, // Disable timeout for this request
			)
			return response.data
		} catch (error) {
			console.error("Error fetching job tasks:", error)
			throw error
		}
	},

	getTaskLogs: async (
		jobId: string,
		taskId: string,
		filePath: string,
	): Promise<APIResponse<TaskLog[]>> => {
		try {
			const response = await api.post<APIResponse<TaskLog[]>>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/tasks/${taskId}/logs`,
				{ file_path: filePath },
				{ timeout: 0 }, // Disable timeout for this request since it can take longer
			)
			return response.data
		} catch (error) {
			console.error("Error fetching task logs:", error)
			throw error
		}
	},

	//This either pauses or resumes the job
	activateJob: async (
		jobId: string,
		activate: boolean,
	): Promise<APIResponse<any>> => {
		try {
			const response = await api.post<APIResponse<any>>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/activate`,
				{ activate },
			)
			return response.data
		} catch (error) {
			console.error("Error toggling job activation:", error)
			throw error
		}
	},

	checkJobNameUnique: async (jobName: string): Promise<{ unique: boolean }> => {
		try {
			const response = await api.post<APIResponse<{ unique: boolean }>>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/check-unique`,
				{ job_name: jobName },
			)
			return response.data.data
		} catch (error) {
			console.error("Error checking job name uniqueness:", error)
			throw error
		}
	},
}
