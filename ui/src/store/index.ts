import { create } from "zustand"
import {
	Job,
	Entity,
	JobHistory,
	JobLog,
	EntityBase,
	JobTask,
	TaskLog,
} from "../types"
import { jobService, sourceService, destinationService } from "../api"
import { authService } from "../api/services/authService"

interface AppState {
	// Data
	jobs: Job[]
	sources: Entity[]
	destinations: Entity[]
	jobHistory: JobHistory[]
	jobLogs: JobLog[]
	jobTasks: JobTask[]
	taskLogs: TaskLog[]

	// Auth state
	isAuthenticated: boolean
	isAuthLoading: boolean

	// Loading states
	isLoadingJobs: boolean
	isLoadingSources: boolean
	isLoadingDestinations: boolean
	isLoadingJobHistory: boolean
	isLoadingJobLogs: boolean
	isLoadingJobTasks: boolean
	isLoadingTaskLogs: boolean

	// Error states
	jobsError: string | null
	sourcesError: string | null
	destinationsError: string | null
	jobHistoryError: string | null
	jobLogsError: string | null
	jobTasksError: string | null
	taskLogsError: string | null

	selectedJobId: string | null
	selectedHistoryId: string | null
	selectedSource: Entity
	selectedDestination: Entity

	//Modals
	showTestingModal: boolean
	showSuccessModal: boolean
	showEntitySavedModal: boolean
	showSourceCancelModal: boolean
	showDeleteModal: boolean

	// Auth actions
	initAuth: () => Promise<void>

	showDeleteJobModal: boolean
	showClearDataModal: boolean
	showClearDestinationAndSyncModal: boolean
	showEditSourceModal: boolean
	showEditDestinationModal: boolean
	// Actions - Jobs
	fetchJobs: () => Promise<Job[]>
	addJob: (job: JobBase) => Promise<Job>
	updateJob: (id: string, job: Partial<Job>) => Promise<Job>
	deleteJob: (id: string) => Promise<void>
	runJob: (id: string) => Promise<void>
	setSelectedJobId: (id: string | null) => void
	setSelectedSource: (source: Entity) => void
	setSelectedDestination: (destination: Entity) => void

	// Actions - Job History
	fetchJobHistory: (jobId: string) => Promise<void>
	setSelectedHistoryId: (id: string | null) => void

	// Actions - Job Tasks
	fetchJobTasks: (jobId: string) => Promise<void>

	// Actions - Task Logs
	fetchTaskLogs: (
		jobId: string,
		taskId: string,
		filePath: string,
	) => Promise<void>

	// Actions - Job Logs
	fetchJobLogs: (jobId: string, historyId: string) => Promise<void>

	// Actions - Sources
	fetchSources: () => Promise<Entity[]>
	addSource: (source: EntityBase) => Promise<EntityBase>
	updateSource: (id: string, source: Partial<Entity>) => Promise<Entity>
	deleteSource: (id: string) => Promise<void>

	// Actions - Destinations
	fetchDestinations: () => Promise<Entity[]>
	addDestination: (destination: EntityBase) => Promise<EntityBase>
	updateDestination: (
		id: string,
		destination: Partial<Entity>,
	) => Promise<Entity>
	deleteDestination: (id: string) => Promise<void>

	setShowTestingModal: (show: boolean) => void
	setShowSuccessModal: (show: boolean) => void
	setShowEntitySavedModal: (show: boolean) => void
	setShowSourceCancelModal: (show: boolean) => void
	setShowDeleteModal: (show: boolean) => void
	setShowDeleteJobModal: (show: boolean) => void
	setShowClearDataModal: (show: boolean) => void
	setShowClearDestinationAndSyncModal: (show: boolean) => void
	setShowEditSourceModal: (show: boolean) => void
	setShowEditDestinationModal: (show: boolean) => void
}

export const useAppStore = create<AppState>(set => ({
	// Initial data
	jobs: [],
	sources: [],
	destinations: [],
	jobHistory: [],
	jobLogs: [],
	jobTasks: [],
	taskLogs: [],

	// Initial auth state
	isAuthenticated: false,
	isAuthLoading: true,

	// Initial loading states
	isLoadingJobs: false,
	isLoadingSources: false,
	isLoadingDestinations: false,
	isLoadingJobHistory: false,
	isLoadingJobLogs: false,
	isLoadingJobTasks: false,
	isLoadingTaskLogs: false,

	// Initial error states
	jobsError: null,
	sourcesError: null,
	destinationsError: null,
	jobHistoryError: null,
	jobLogsError: null,
	jobTasksError: null,
	taskLogsError: null,

	// Selected job
	selectedJobId: null,
	selectedHistoryId: null,
	selectedSource: {} as Entity,
	selectedDestination: {} as Entity,

	// Modals
	showTestingModal: false,
	showSuccessModal: false,
	showEntitySavedModal: false,
	showSourceCancelModal: false,
	showDeleteModal: false,
	showDeleteJobModal: false,
	showClearDataModal: false,
	showClearDestinationAndSyncModal: false,
	showEditSourceModal: false,
	showEditDestinationModal: false,
	// Auth actions
	initAuth: async () => {
		set({ isAuthLoading: true })
		try {
			if (!authService.isLoggedIn()) {
				// Login with default credentials
				await authService.login("olake", "password")
			}
			set({ isAuthenticated: true, isAuthLoading: false })
		} catch (error) {
			set({
				isAuthLoading: false,
				jobsError:
					error instanceof Error ? error.message : "Failed to initialize auth",
			})
		}
	},

	fetchJobs: async () => {
		set({ isLoadingJobs: true, jobsError: null })
		try {
			const jobs = await jobService.getJobs()
			set({ jobs: jobs, isLoadingJobs: false })
			return jobs
		} catch (error) {
			set({
				isLoadingSources: false,
				jobsError:
					error instanceof Error ? error.message : "Failed to fetch jobs",
			})
			throw error
		}
	},

	addJob: async jobData => {
		try {
			const newJob = await jobService.createJob(jobData)
			return newJob
		} catch (error) {
			set({
				jobsError: error instanceof Error ? error.message : "Failed to add job",
			})
			throw error
		}
	},

	updateJob: async (id, jobData) => {
		try {
			const updatedJob = await jobService.updateJob(id, jobData)
			set(state => ({
				jobs: state.jobs.map(job => (job.id === id ? updatedJob : job)),
			}))
			return updatedJob
		} catch (error) {
			set({
				jobsError:
					error instanceof Error ? error.message : "Failed to update job",
			})
			throw error
		}
	},

	deleteJob: async id => {
		try {
			const numericId = typeof id === "string" ? parseInt(id, 10) : id

			await jobService.deleteJob(numericId)
			set(state => ({
				jobs: state.jobs.filter(job => job.id !== numericId),
			}))
		} catch (error) {
			set({
				jobsError:
					error instanceof Error ? error.message : "Failed to delete job",
			})
			throw error
		}
	},

	runJob: async id => {
		try {
			await jobService.runJob(id)
			// Optionally update job status in state
			set(state => ({
				jobs: state.jobs.map(job =>
					job.id === id ? { ...job, status: "active" } : job,
				),
			}))
		} catch (error) {
			set({
				jobsError: error instanceof Error ? error.message : "Failed to run job",
			})
			throw error
		}
	},

	setSelectedJobId: id => {
		set({ selectedJobId: id })
	},

	// Job History actions
	fetchJobHistory: async jobId => {
		set({ isLoadingJobHistory: true, jobHistoryError: null })
		try {
			// Mock data for development
			const mockHistory: JobHistory[] = [
				{
					id: "hist-1",
					jobId,
					startTime: "2025-02-25 07:05 PM",
					runtime: "30 seconds",
					status: "success",
				},
				{
					id: "hist-2",
					jobId,
					startTime: "2025-02-25 06:40 PM",
					runtime: "45 seconds",
					status: "success",
				},
				{
					id: "hist-3",
					jobId,
					startTime: "2025-02-25 06:25 PM",
					runtime: "35 seconds",
					status: "failed",
				},
				{
					id: "hist-4",
					jobId,
					startTime: "2025-02-25 05:35 PM",
					runtime: "54 seconds",
					status: "running",
				},
				{
					id: "hist-5",
					jobId,
					startTime: "2025-02-25 03:27 PM",
					runtime: "42 seconds",
					status: "success",
				},
				{
					id: "hist-6",
					jobId,
					startTime: "2025-02-25 10:25 AM",
					runtime: "45 seconds",
					status: "success",
				},
				{
					id: "hist-7",
					jobId,
					startTime: "2025-02-25 07:25 AM",
					runtime: "53 seconds",
					status: "scheduled",
				},
				{
					id: "hist-8",
					jobId,
					startTime: "2025-02-25 07:25 AM",
					runtime: "1 minute 03 seconds",
					status: "scheduled",
				},
				{
					id: "hist-9",
					jobId,
					startTime: "2025-02-25 07:25 AM",
					runtime: "1 minute 29 seconds",
					status: "success",
				},
				{
					id: "hist-10",
					jobId,
					startTime: "2025-02-25 07:25 AM",
					runtime: "1 minute 45 seconds",
					status: "success",
				},
			]

			set({ jobHistory: mockHistory, isLoadingJobHistory: false })

			// Uncomment for real API call
			// const history = await jobService.getJobHistory(jobId);
			// set({ jobHistory: history, isLoadingJobHistory: false });
		} catch (error) {
			set({
				isLoadingJobHistory: false,
				jobHistoryError:
					error instanceof Error
						? error.message
						: "Failed to fetch job history",
			})
		}
	},

	setSelectedHistoryId: id => {
		set({ selectedHistoryId: id })
	},

	setSelectedSource: source => {
		set({ selectedSource: source })
	},

	setSelectedDestination: destination => {
		set({ selectedDestination: destination })
	},

	// Job Tasks actions
	fetchJobTasks: async jobId => {
		set({ isLoadingJobTasks: true, jobTasksError: null })
		try {
			const response = await jobService.getJobTasks(jobId)
			set({
				jobTasks: response.data,
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
				taskLogs: response.data,
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

	// Job Logs actions
	fetchJobLogs: async (jobId, historyId) => {
		set({ isLoadingJobLogs: true, jobLogsError: null })
		try {
			// Mock data for development
			const logMessage =
				"Lorem ipsum dolor sit amet consectetur. Urna neque imperdiet nisl libero praesent diam hendrerit urna tortor."
			const mockLogs: JobLog[] = Array.from({ length: 20 }, (_, i) => {
				let level: "debug" | "info" | "warning" | "error" = "debug"
				if (i % 7 === 0) level = "info"
				if (i % 11 === 0) level = "warning"
				if (i % 13 === 0) level = "error"

				return {
					date: "26/02/2025",
					time: "00:07:23",
					level,
					message: logMessage,
				}
			})

			set({ jobLogs: mockLogs, isLoadingJobLogs: false })

			// Uncomment for real API call
			// const logs = await jobService.getJobLogs(jobId, historyId);
			// set({ jobLogs: logs, isLoadingJobLogs: false });
		} catch (error) {
			set({
				isLoadingJobLogs: false,
				jobLogsError:
					error instanceof Error ? error.message : "Failed to fetch job logs",
			})
		}
	},

	fetchSources: async () => {
		set({ isLoadingSources: true, sourcesError: null })
		try {
			const sources = await sourceService.getSources()
			set({ sources: sources, isLoadingSources: false })
			return sources
		} catch (error) {
			set({
				isLoadingSources: false,
				sourcesError:
					error instanceof Error ? error.message : "Failed to fetch sources",
			})
			throw error
		}
	},

	addSource: async sourceData => {
		try {
			const newSource = await sourceService.createSource(sourceData)
			return newSource
		} catch (error) {
			set({
				sourcesError:
					error instanceof Error ? error.message : "Failed to add source",
			})
			throw error
		}
	},

	updateSource: async (id, sourceData) => {
		try {
			const updatedSource = await sourceService.updateSource(id, sourceData)
			set(state => ({
				sources: state.sources.map(source =>
					source.id === id ? updatedSource : source,
				),
			}))
			return updatedSource
		} catch (error) {
			set({
				sourcesError:
					error instanceof Error ? error.message : "Failed to update source",
			})
			throw error
		}
	},

	deleteSource: async id => {
		try {
			const numericId = typeof id === "string" ? parseInt(id, 10) : id

			await sourceService.deleteSource(numericId)
			set(state => ({
				sources: state.sources.filter(source => source.id !== numericId),
			}))
		} catch (error) {
			set({
				sourcesError:
					error instanceof Error ? error.message : "Failed to delete source",
			})
			throw error
		}
	},

	// Destinations actions
	fetchDestinations: async () => {
		set({ isLoadingDestinations: true, destinationsError: null })
		try {
			const destinations = await destinationService.getDestinations()
			set({ destinations: destinations, isLoadingDestinations: false })
			return destinations
		} catch (error) {
			set({
				isLoadingSources: false,
				destinationsError:
					error instanceof Error
						? error.message
						: "Failed to fetch destinations",
			})
			throw error
		}
	},

	addDestination: async destinationData => {
		try {
			const newDestination =
				await destinationService.createDestination(destinationData)
			return newDestination
		} catch (error) {
			set({
				sourcesError:
					error instanceof Error ? error.message : "Failed to add source",
			})
			throw error
		}
	},

	updateDestination: async (id, destinationData) => {
		try {
			const updatedDestination = await destinationService.updateDestination(
				id,
				destinationData,
			)
			set(state => ({
				destinations: state.destinations.map(destination =>
					destination.id === id ? updatedDestination : destination,
				),
			}))
			return updatedDestination
		} catch (error) {
			set({
				destinationsError:
					error instanceof Error
						? error.message
						: "Failed to update destination",
			})
			throw error
		}
	},

	deleteDestination: async id => {
		try {
			const numericId = typeof id === "string" ? parseInt(id, 10) : id

			await destinationService.deleteDestination(numericId)
			set(state => ({
				destinations: state.destinations.filter(
					destination => destination.id !== numericId,
				),
			}))
		} catch (error) {
			set({
				destinationsError:
					error instanceof Error
						? error.message
						: "Failed to delete destination",
			})
			throw error
		}
	},

	setShowTestingModal: show => set({ showTestingModal: show }),
	setShowSuccessModal: show => set({ showSuccessModal: show }),
	setShowEntitySavedModal: show => set({ showEntitySavedModal: show }),
	setShowSourceCancelModal: show => set({ showSourceCancelModal: show }),
	setShowDeleteModal: show => set({ showDeleteModal: show }),
	setShowDeleteJobModal: show => set({ showDeleteJobModal: show }),
	setShowClearDataModal: show => set({ showClearDataModal: show }),
	setShowClearDestinationAndSyncModal: show =>
		set({ showClearDestinationAndSyncModal: show }),
	setShowEditSourceModal: show => set({ showEditSourceModal: show }),
	setShowEditDestinationModal: show => set({ showEditDestinationModal: show }),
}))
