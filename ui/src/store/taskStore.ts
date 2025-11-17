import { StateCreator } from "zustand"
import type { JobTask, LogEntry } from "../types"
import { jobService } from "../api"
import { LOGS_CONFIG } from "../utils/constants"

export interface TaskSlice {
	jobTasksError: string | null
	taskLogsError: string | null
	isLoadingJobTasks: boolean
	isLoadingTaskLogs: boolean
	isLoadingMoreLogs: boolean
	jobTasks: JobTask[]
	taskLogs: LogEntry[]
	taskLogsCursor: number
	taskLogsHasMore: boolean
	// Job task actions
	fetchJobTasks: (jobId: string) => Promise<void>
	fetchInitialTaskLogs: (
		jobId: string,
		taskId: string,
		filePath: string,
	) => Promise<void>
	fetchMoreTaskLogs: (
		jobId: string,
		taskId: string,
		filePath: string,
	) => Promise<void>
	resetTaskLogs: () => void
}
export const createTaskSlice: StateCreator<TaskSlice> = (set, get) => ({
	jobTasks: [],
	taskLogs: [],
	taskLogsCursor: LOGS_CONFIG.DEFAULT_CURSOR,
	taskLogsHasMore: true,
	isLoadingJobTasks: false,
	isLoadingTaskLogs: false,
	isLoadingMoreLogs: false,
	jobTasksError: null,
	taskLogsError: null,

	fetchJobTasks: async jobId => {
		set({ isLoadingJobTasks: true, jobTasksError: null })
		try {
			const response = await jobService.getJobTasks(jobId)
			set({
				jobTasks: response,
				isLoadingJobTasks: false,
			})
		} catch (error) {
			set({
				isLoadingJobTasks: false,
				jobTasksError:
					error instanceof Error ? error.message : "Failed to fetch job tasks",
			})
			throw error
		}
	},

	// Fetch initial batch of logs (first load)
	fetchInitialTaskLogs: async (jobId, taskId, filePath) => {
		set({
			isLoadingTaskLogs: true,
			taskLogsError: null,
			taskLogs: [],
			taskLogsCursor: LOGS_CONFIG.DEFAULT_CURSOR,
			taskLogsHasMore: true,
		})
		try {
			const response = await jobService.getTaskLogs(
				jobId,
				taskId,
				filePath,
				LOGS_CONFIG.DEFAULT_CURSOR,
				LOGS_CONFIG.INITIAL_BATCH_SIZE,
			)
			set({
				taskLogs: response.logs,
				taskLogsCursor: response.cursor,
				taskLogsHasMore: response.hasMore,
				isLoadingTaskLogs: false,
			})
		} catch (error) {
			set({
				isLoadingTaskLogs: false,
				taskLogsError:
					error instanceof Error ? error.message : "Failed to fetch task logs",
			})
			throw error
		}
	},

	// Fetch more logs (infinite scroll)
	fetchMoreTaskLogs: async (jobId, taskId, filePath) => {
		const state = get()
		const { taskLogsCursor, taskLogsHasMore, isLoadingMoreLogs, taskLogs } =
			state

		if (isLoadingMoreLogs || !taskLogsHasMore) {
			return
		}

		set({ isLoadingMoreLogs: true, taskLogsError: null })
		try {
			const response = await jobService.getTaskLogs(
				jobId,
				taskId,
				filePath,
				taskLogsCursor,
				LOGS_CONFIG.SUBSEQUENT_BATCH_SIZE,
			)

			// Prepend older logs to the top
			const updatedLogs = [...response.logs, ...taskLogs]

			// trim logs if exceeding max memory
			const trimmedLogs =
				updatedLogs.length > LOGS_CONFIG.MAX_LOGS_IN_MEMORY
					? updatedLogs.slice(0, LOGS_CONFIG.MAX_LOGS_IN_MEMORY)
					: updatedLogs

			set({
				taskLogs: trimmedLogs,
				taskLogsCursor: response.cursor,
				taskLogsHasMore: response.hasMore,
				isLoadingMoreLogs: false,
			})
		} catch (error) {
			set({
				isLoadingMoreLogs: false,
				taskLogsError:
					error instanceof Error
						? error.message
						: "Failed to fetch more task logs",
			})
			throw error
		}
	},

	// Reset logs state
	resetTaskLogs: () => {
		set({
			taskLogs: [],
			taskLogsCursor: LOGS_CONFIG.DEFAULT_CURSOR,
			taskLogsHasMore: true,
			isLoadingTaskLogs: false,
			isLoadingMoreLogs: false,
			taskLogsError: null,
		})
	},
})
