import api from "../axios"
import { API_CONFIG } from "../config"
import {
	APIResponse,
	Job,
	JobBase,
	JobTask,
	StreamsDataStructure,
	TaskLog,
} from "../../types"
import { AxiosError } from "axios"

export const jobService = {
	getJobs: async (): Promise<Job[]> => {
		try {
			const response = await api.get<Job[]>(
				API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID),
			)

			return response.data
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
				{ timeout: 30000, showNotification: true },
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
				{ showNotification: true },
			)
		} catch (error) {
			console.error("Error deleting job:", error)
			throw error
		}
	},

	cancelJob: async (id: string): Promise<string> => {
		try {
			const response = await api.get<any>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}/cancel`,
				{ showNotification: true },
			)
			return response.data.message
		} catch (error) {
			console.error("Error canceling job:", error)
			throw error
		}
	},

	syncJob: async (id: string): Promise<any> => {
		try {
			const response = await api.post<any>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}/sync`,
				{},
				{ timeout: 0, showNotification: true }, // Disable timeout for this request since it can take longer
			)
			return response.data
		} catch (error) {
			console.error("Error syncing job:", error)
			throw error instanceof AxiosError && error.response?.data.message
				? error.response?.data.message
				: "Failed to sync job"
		}
	},

	getJobTasks: async (id: string): Promise<JobTask[]> => {
		try {
			const response = await api.get<JobTask[]>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}/tasks`,
				{ timeout: 0, showNotification: true }, // Disable timeout for this request, no toast for fetching tasks
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
	): Promise<TaskLog[]> => {
		try {
			const response = await api.post<TaskLog[]>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/tasks/${taskId}/logs`,
				{ file_path: filePath },
				{ timeout: 0, showNotification: true }, // Disable timeout for this request since it can take longer, no toast for logs
			)
			return response.data
		} catch (error) {
			console.error("Error fetching task logs:", error)
			throw error
		}
	},

	//This either pauses or resumes the job
	activateJob: async (jobId: string, activate: boolean): Promise<any> => {
		try {
			const response = await api.post<any>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/activate`,
				{ activate },
				{ showNotification: true },
			)
			return response.data
		} catch (error) {
			console.error("Error toggling job activation:", error)
			throw error
		}
	},

	checkJobNameUnique: async (jobName: string): Promise<{ unique: boolean }> => {
		try {
			const response = await api.post<{ unique: boolean }>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/check-unique`,
				{ job_name: jobName },
			)
			return response.data
		} catch (error) {
			console.error("Error checking job name uniqueness:", error)
			throw error
		}
	},

	clearDestination: async (
		jobId: string,
	): Promise<APIResponse<{ message: string }>> => {
		try {
			const response = await api.post<APIResponse<{ message: string }>>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/clear-destination`,
			)

			return response.data
		} catch (error) {
			console.error("Error clearing destination:", error)
			throw error
		}
	},
	getStreamDifference: async (
		jobId: string,
		streamsConfig: string,
	): Promise<{ difference_streams: StreamsDataStructure }> => {
		try {
			const response = await api.post<{
				difference_streams: StreamsDataStructure
			}>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/stream-difference`,
				{ updated_streams_config: streamsConfig },
				{ timeout: 30000 },
			)
			return response.data
		} catch (error) {
			console.error("Error getting stream difference:", error)
			throw error
		}
	},
}
