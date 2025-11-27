import { StateCreator } from "zustand"
import type { JobTask, LogEntry } from "../types"
import { jobService } from "../api"
import { LOGS_CONFIG } from "../utils/constants"

export interface TaskSlice {
	jobTasksError: string | null
	taskLogsError: string | null
	isLoadingJobTasks: boolean
	isLoadingTaskLogs: boolean
	isLoadingOlderLogs: boolean
	isLoadingNewerLogs: boolean
	jobTasks: JobTask[]
	taskLogs: LogEntry[]
	taskLogsOlderCursor: number
	taskLogsNewerCursor: number
	taskLogsHasMoreOlder: boolean
	taskLogsHasMoreNewer: boolean
	// Job task actions
	fetchJobTasks: (jobId: string) => Promise<void>
	fetchInitialTaskLogs: (
		jobId: string,
		taskId: string,
		filePath: string,
	) => Promise<void>
	fetchOlderTaskLogs: (
		jobId: string,
		taskId: string,
		filePath: string,
	) => Promise<void>
	fetchNewerTaskLogs: (
		jobId: string,
		taskId: string,
		filePath: string,
	) => Promise<void>
}
export const createTaskSlice: StateCreator<TaskSlice> = (set, get) => ({
	jobTasks: [],
	taskLogs: [],
	taskLogsOlderCursor: LOGS_CONFIG.DEFAULT_CURSOR,
	taskLogsNewerCursor: LOGS_CONFIG.DEFAULT_CURSOR,
	taskLogsHasMoreOlder: true,
	taskLogsHasMoreNewer: true,
	isLoadingJobTasks: false,
	isLoadingTaskLogs: false,
	isLoadingOlderLogs: false,
	isLoadingNewerLogs: false,
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
			taskLogsOlderCursor: LOGS_CONFIG.DEFAULT_CURSOR,
			taskLogsNewerCursor: LOGS_CONFIG.DEFAULT_CURSOR,
			taskLogsHasMoreOlder: true,
			taskLogsHasMoreNewer: true,
			isLoadingOlderLogs: false,
			isLoadingNewerLogs: false,
		})
		try {
			const response = await jobService.getTaskLogs(
				jobId,
				taskId,
				filePath,
				LOGS_CONFIG.DEFAULT_CURSOR,
				LOGS_CONFIG.INITIAL_BATCH_SIZE,
				"older",
			)
			set({
				taskLogs: response.logs,
				taskLogsOlderCursor: response.older_cursor,
				taskLogsNewerCursor: response.newer_cursor,
				taskLogsHasMoreOlder: response.has_more_older,
				taskLogsHasMoreNewer: response.has_more_newer,
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

	fetchOlderTaskLogs: async (jobId, taskId, filePath) => {
		const state = get()
		const {
			taskLogsOlderCursor,
			taskLogsHasMoreOlder,
			isLoadingOlderLogs,
			taskLogs,
		} = state

		if (isLoadingOlderLogs || !taskLogsHasMoreOlder) {
			return
		}

		set({ isLoadingOlderLogs: true, taskLogsError: null })
		try {
			const response = await jobService.getTaskLogs(
				jobId,
				taskId,
				filePath,
				taskLogsOlderCursor,
				LOGS_CONFIG.SUBSEQUENT_BATCH_SIZE,
				"older",
			)

			if (state.taskLogs.length >= LOGS_CONFIG.MAX_LOGS_IN_MEMORY) {
				set({
					taskLogs: response.logs, // Replace with ONLY the new batch
					taskLogsOlderCursor: response.older_cursor,
					taskLogsNewerCursor: response.newer_cursor,
					taskLogsHasMoreOlder: response.has_more_older,
					taskLogsHasMoreNewer: response.has_more_newer,
					isLoadingOlderLogs: false,
				})
				return
			}

			// Prepend older logs to the top
			const updatedLogs = [...response.logs, ...taskLogs]

			set({
				taskLogs: updatedLogs,
				taskLogsOlderCursor: response.older_cursor,
				taskLogsHasMoreOlder: response.has_more_older,
				isLoadingOlderLogs: false,
			})
		} catch (error) {
			set({
				isLoadingOlderLogs: false,
				taskLogsError:
					error instanceof Error
						? error.message
						: "Failed to fetch more task logs",
			})
			throw error
		}
	},

	// Fetch newer logs when scrolling towards the bottom
	fetchNewerTaskLogs: async (jobId, taskId, filePath) => {
		const state = get()
		const {
			taskLogsNewerCursor,
			taskLogsHasMoreNewer,
			isLoadingNewerLogs,
			taskLogs,
		} = state

		if (isLoadingNewerLogs || !taskLogsHasMoreNewer) {
			return
		}

		set({ isLoadingNewerLogs: true, taskLogsError: null })
		try {
			const response = await jobService.getTaskLogs(
				jobId,
				taskId,
				filePath,
				taskLogsNewerCursor,
				LOGS_CONFIG.SUBSEQUENT_BATCH_SIZE,
				"newer",
			)

			if (state.taskLogs.length >= LOGS_CONFIG.MAX_LOGS_IN_MEMORY) {
				set({
					taskLogs: response.logs, // Replace with ONLY the new batch
					taskLogsOlderCursor: response.older_cursor,
					taskLogsNewerCursor: response.newer_cursor,
					taskLogsHasMoreOlder: response.has_more_older,
					taskLogsHasMoreNewer: response.has_more_newer,
					isLoadingNewerLogs: false,
				})
				return
			}

			// append newer logs to the bottom
			const updatedLogs = [...taskLogs, ...response.logs]

			set({
				taskLogs: updatedLogs,
				taskLogsNewerCursor: response.newer_cursor,
				taskLogsHasMoreNewer: response.has_more_newer,
				isLoadingNewerLogs: false,
			})
		} catch (error) {
			set({
				isLoadingNewerLogs: false,
				taskLogsError:
					error instanceof Error
						? error.message
						: "Failed to fetch newer task logs",
			})
			throw error
		}
	},
})
