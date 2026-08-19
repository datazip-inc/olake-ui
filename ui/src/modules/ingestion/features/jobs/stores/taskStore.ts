import { create } from "zustand"

import { LOGS_CONFIG } from "../constants"
import { jobService } from "../services"
import {
	TaskLogsDirection,
	type TaskLogEntry,
	type TaskLogsPaginationParams,
} from "../types"
import { mapLogEntriesToTaskLogEntries } from "../utils"

export interface TaskState {
	taskLogsError: string | null
	isLoadingTaskLogs: boolean
	isLoadingOlderLogs: boolean
	isLoadingNewerLogs: boolean
	taskLogs: TaskLogEntry[]
	taskLogsOlderCursor: number
	taskLogsNewerCursor: number
	taskLogsHasMoreOlder: boolean
	taskLogsHasMoreNewer: boolean
	// Exclusive upper chunk index already loaded from the initial tail fetch.
	taskLogsLoadedNewerBound: number | null
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

export const useTaskStore = create<TaskState>()((set, get) => ({
	taskLogs: [],
	taskLogsOlderCursor: LOGS_CONFIG.DEFAULT_CURSOR,
	taskLogsNewerCursor: LOGS_CONFIG.DEFAULT_CURSOR,
	taskLogsHasMoreOlder: true,
	taskLogsHasMoreNewer: false,
	taskLogsLoadedNewerBound: null,
	isLoadingTaskLogs: false,
	isLoadingOlderLogs: false,
	isLoadingNewerLogs: false,
	taskLogsError: null,

	fetchInitialTaskLogs: async (jobId, taskId, filePath) => {
		set({
			isLoadingTaskLogs: true,
			taskLogsError: null,
			taskLogs: [],
			taskLogsOlderCursor: LOGS_CONFIG.DEFAULT_CURSOR,
			taskLogsNewerCursor: LOGS_CONFIG.DEFAULT_CURSOR,
			taskLogsHasMoreOlder: true,
			taskLogsHasMoreNewer: false,
			taskLogsLoadedNewerBound: null,
			isLoadingOlderLogs: false,
			isLoadingNewerLogs: false,
		})
		try {
			const paginationParams: TaskLogsPaginationParams = {
				cursor: LOGS_CONFIG.DEFAULT_CURSOR,
				limit: LOGS_CONFIG.INITIAL_BATCH_SIZE,
				direction: TaskLogsDirection.Older,
			}

			const response = await jobService.getTaskLogs(
				jobId,
				taskId,
				filePath,
				paginationParams,
			)
			set({
				taskLogs: mapLogEntriesToTaskLogEntries(response.logs),
				taskLogsOlderCursor: response.older_cursor,
				taskLogsNewerCursor: response.newer_cursor,
				taskLogsHasMoreOlder: response.has_more_older,
				taskLogsHasMoreNewer: response.has_more_newer,
				taskLogsLoadedNewerBound: response.newer_cursor,
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
			taskLogsLoadedNewerBound,
		} = state

		if (isLoadingOlderLogs || !taskLogsHasMoreOlder) {
			return
		}

		set({ isLoadingOlderLogs: true, taskLogsError: null })
		try {
			const paginationParams: TaskLogsPaginationParams = {
				cursor: taskLogsOlderCursor,
				limit: LOGS_CONFIG.SUBSEQUENT_BATCH_SIZE,
				direction: TaskLogsDirection.Older,
			}

			const response = await jobService.getTaskLogs(
				jobId,
				taskId,
				filePath,
				paginationParams,
			)

			const hasMoreNewer =
				taskLogsLoadedNewerBound !== null &&
				response.newer_cursor < taskLogsLoadedNewerBound
					? false
					: response.has_more_newer

			if (state.taskLogs.length >= LOGS_CONFIG.MAX_LOGS_IN_MEMORY) {
				set({
					taskLogs: mapLogEntriesToTaskLogEntries(response.logs),
					taskLogsOlderCursor: response.older_cursor,
					taskLogsNewerCursor: response.newer_cursor,
					taskLogsHasMoreOlder: response.has_more_older,
					taskLogsHasMoreNewer: hasMoreNewer,
					isLoadingOlderLogs: false,
				})
				return
			}

			const normalizedLogs = mapLogEntriesToTaskLogEntries(response.logs)
			const updatedLogs = [...normalizedLogs, ...taskLogs]

			set({
				taskLogs: updatedLogs,
				taskLogsOlderCursor: response.older_cursor,
				taskLogsNewerCursor: response.newer_cursor,
				taskLogsHasMoreOlder: response.has_more_older,
				taskLogsHasMoreNewer: hasMoreNewer,
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

	fetchNewerTaskLogs: async (jobId, taskId, filePath) => {
		const state = get()
		const {
			taskLogsNewerCursor,
			taskLogsHasMoreNewer,
			isLoadingNewerLogs,
			taskLogs,
			taskLogsLoadedNewerBound,
		} = state

		if (isLoadingNewerLogs || !taskLogsHasMoreNewer) {
			return
		}

		// Tail chunks from the initial load are already in memory; skip re-fetch
		// after scrolling up (fetchOlder moves newer_cursor below loadedNewerBound).
		if (
			taskLogsLoadedNewerBound !== null &&
			taskLogsNewerCursor < taskLogsLoadedNewerBound
		) {
			set({ taskLogsHasMoreNewer: false })
			return
		}

		set({ isLoadingNewerLogs: true, taskLogsError: null })
		try {
			const paginationParams: TaskLogsPaginationParams = {
				cursor: taskLogsNewerCursor,
				limit: LOGS_CONFIG.SUBSEQUENT_BATCH_SIZE,
				direction: TaskLogsDirection.Newer,
			}

			const response = await jobService.getTaskLogs(
				jobId,
				taskId,
				filePath,
				paginationParams,
			)

			if (state.taskLogs.length >= LOGS_CONFIG.MAX_LOGS_IN_MEMORY) {
				set({
					taskLogs: mapLogEntriesToTaskLogEntries(response.logs),
					taskLogsOlderCursor: response.older_cursor,
					taskLogsNewerCursor: response.newer_cursor,
					taskLogsHasMoreOlder: response.has_more_older,
					taskLogsHasMoreNewer: response.has_more_newer,
					taskLogsLoadedNewerBound: response.newer_cursor,
					isLoadingNewerLogs: false,
				})
				return
			}

			const normalizedLogs = mapLogEntriesToTaskLogEntries(response.logs)
			const updatedLogs = [...taskLogs, ...normalizedLogs]

			set({
				taskLogs: updatedLogs,
				taskLogsOlderCursor: response.older_cursor,
				taskLogsNewerCursor: response.newer_cursor,
				taskLogsHasMoreOlder: response.has_more_older,
				taskLogsHasMoreNewer: response.has_more_newer,
				taskLogsLoadedNewerBound: response.newer_cursor,
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
}))
