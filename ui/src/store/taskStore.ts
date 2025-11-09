import { StateCreator } from "zustand"
import type { JobTask } from "../types"
import type { TaskLog } from "../types"
import { jobService } from "../api"

export interface TaskSlice {
	jobTasksError: string | null
	taskLogsError: string | null
	isLoadingJobTasks: boolean
	isLoadingTaskLogs: boolean
	jobTasks: JobTask[]
	taskLogs: TaskLog[]
	// Job task actions
	fetchJobTasks: (jobId: string) => Promise<void>
	fetchTaskLogs: (
		jobId: string,
		taskId: string,
		filePath: string,
	) => Promise<void>
}
export const createTaskSlice: StateCreator<TaskSlice> = set => ({
	jobTasks: [],
	taskLogs: [],
	isLoadingJobTasks: false,
	isLoadingTaskLogs: false,
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

	// Task Logs actions
	fetchTaskLogs: async (jobId, taskId, filePath) => {
		set({ isLoadingTaskLogs: true, taskLogsError: null })
		try {
			const response = await jobService.getTaskLogs(jobId, taskId, filePath)
			set({
				taskLogs: response,
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
})
