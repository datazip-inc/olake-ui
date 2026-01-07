import api from "../axios"
import { API_CONFIG } from "../config"
import {
	Job,
	JobBase,
	JobTask,
	StreamsDataStructure,
	TaskLogsDirection,
	TaskLogsPaginationParams,
	TaskLogsResponse,
} from "../../types"
import { AxiosError } from "axios"
import { normalizeConnectorType } from "../../utils/utils"
import { ENTITY_TYPES } from "../../utils/constants"

export const jobService = {
	getJobs: async (): Promise<Job[]> => {
		try {
			const response = await api.get<Job[]>(
				API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID),
				{ timeout: 0 },
			)

			const jobs = response.data.map(job => ({
				...job,
				destination: {
					...job.destination,
					type: normalizeConnectorType(job.destination.type),
				},
			}))

			return jobs
		} catch (error) {
			console.error("Error fetching jobs from API:", error)
			throw error
		}
	},

	getJob: async (id: string): Promise<Job> => {
		try {
			const response = await api.get<Job>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${id}`,
			)

			const job = {
				...response.data,
				destination: {
					...response.data.destination,
					type: normalizeConnectorType(response.data.destination.type),
				},
			}

			return job
		} catch (error) {
			console.error("Error fetching job from API:", error)
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
		params: TaskLogsPaginationParams = {
			cursor: -1,
			limit: 1000,
			direction: TaskLogsDirection.Older,
		},
	): Promise<TaskLogsResponse> => {
		try {
			const { cursor, limit, direction } = params
			const query = new URLSearchParams({
				cursor: String(cursor),
				limit: String(limit),
				direction,
			})

			const response = await api.post<TaskLogsResponse>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/tasks/${taskId}/logs?${query.toString()}`,
				{ file_path: filePath },
				{ showNotification: false }, // no toast for logs
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
				`${API_CONFIG.ENDPOINTS.PROJECT(API_CONFIG.PROJECT_ID)}/check-unique`,
				{ name: jobName, entity_type: ENTITY_TYPES.JOB },
			)
			return response.data
		} catch (error) {
			console.error("Error checking job name uniqueness:", error)
			throw error
		}
	},

	clearDestination: async (jobId: string): Promise<{ message: string }> => {
		try {
			const response = await api.post<{ message: string }>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/clear-destination`,
			)

			return response.data
		} catch (error) {
			console.error("Error clearing destination:", error)
			throw error
		}
	},
	getClearDestinationStatus: async (
		jobId: string,
	): Promise<{ running: boolean }> => {
		try {
			const response = await api.get<{ running: boolean }>(
				`${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/clear-destination`,
			)

			return response.data
		} catch (error) {
			console.error("Error fetching clear destination status:", error)
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

	downloadTaskLogs: async (jobId: string, filePath: string): Promise<void> => {
		const params = new URLSearchParams({
			file_path: filePath,
		})

		const url = `${API_CONFIG.BASE_URL}${API_CONFIG.ENDPOINTS.JOBS(API_CONFIG.PROJECT_ID)}/${jobId}/logs/download?${params.toString()}`

		try {
			// Pre-flight check to verify endpoint is accessible
			// Check endpoint with minimal data transfer
			await api.get(url, {
				headers: { Range: "bytes=0-0" },
				responseType: "blob",
			})

			// if successful, trigger download
			const link = document.createElement("a")
			link.href = url
			link.style.display = "none"
			document.body.appendChild(link)
			link.click()
			document.body.removeChild(link)
		} catch (error) {
			throw error
		}
	},
}
