import { create } from "zustand"
import {
	Job,
	Entity,
	EntityBase,
	JobTask,
	TaskLog,
	JobBase,
	APIResponse,
} from "../types"
import { jobService, sourceService, destinationService } from "../api"
import { authService } from "../api/services/authService"

interface AppState {
	// Data states
	jobs: Job[]
	sources: Entity[]
	destinations: Entity[]
	jobTasks: JobTask[]
	taskLogs: TaskLog[]

	// Selected items
	selectedJobId: string | null
	selectedHistoryId: string | null
	selectedSource: Entity
	selectedDestination: Entity

	// Auth state
	isAuthenticated: boolean
	isAuthLoading: boolean

	// Loading states
	isLoadingJobs: boolean
	isLoadingSources: boolean
	isLoadingDestinations: boolean
	isLoadingJobTasks: boolean
	isLoadingTaskLogs: boolean

	// Error states
	jobsError: string | null
	sourcesError: string | null
	destinationsError: string | null
	jobTasksError: string | null
	taskLogsError: string | null

	// Modal states
	showTestingModal: boolean
	showSuccessModal: boolean
	showEntitySavedModal: boolean
	showSourceCancelModal: boolean
	showDeleteModal: boolean
	showDeleteJobModal: boolean
	showClearDataModal: boolean
	showClearDestinationAndSyncModal: boolean
	showEditSourceModal: boolean
	showEditDestinationModal: boolean

	// Auth actions
	initAuth: () => Promise<void>
	login: (username: string, password: string) => Promise<void>
	logout: () => void

	// Selection actions
	setSelectedJobId: (id: string | null) => void
	setSelectedHistoryId: (id: string | null) => void
	setSelectedSource: (source: Entity) => void
	setSelectedDestination: (destination: Entity) => void

	// Job actions
	fetchJobs: () => Promise<Job[]>
	addJob: (job: JobBase) => Promise<Job>
	updateJob: (id: string, job: Partial<Job>) => Promise<Job>
	deleteJob: (id: string) => Promise<void>

	// Job task actions
	fetchJobTasks: (jobId: string) => Promise<void>
	fetchTaskLogs: (
		jobId: string,
		taskId: string,
		filePath: string,
	) => Promise<void>

	// Source actions
	fetchSources: () => Promise<Entity[]>
	addSource: (source: EntityBase) => Promise<APIResponse<EntityBase>>
	updateSource: (id: string, source: EntityBase) => Promise<APIResponse<Entity>>
	deleteSource: (id: string) => Promise<void>

	// Destination actions
	fetchDestinations: () => Promise<Entity[]>
	addDestination: (destination: EntityBase) => Promise<EntityBase>
	updateDestination: (
		id: string,
		destination: Partial<Entity>,
	) => Promise<APIResponse<EntityBase>>
	deleteDestination: (id: string) => Promise<void>

	// Modal actions
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
	// Data states
	jobs: [],
	sources: [],
	destinations: [],
	jobTasks: [],
	taskLogs: [],

	// Selected items
	selectedJobId: null,
	selectedHistoryId: null,
	selectedSource: {} as Entity,
	selectedDestination: {} as Entity,

	// Auth state
	isAuthenticated: false,
	isAuthLoading: true,

	// Loading states
	isLoadingJobs: false,
	isLoadingSources: false,
	isLoadingDestinations: false,
	isLoadingJobTasks: false,
	isLoadingTaskLogs: false,

	// Error states
	jobsError: null,
	sourcesError: null,
	destinationsError: null,
	jobTasksError: null,
	taskLogsError: null,

	// Modal states
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
				set({ isAuthenticated: false, isAuthLoading: false })
				return
			}
			set({ isAuthenticated: true, isAuthLoading: false })
		} catch (error) {
			set({
				isAuthLoading: false,
				isAuthenticated: false,
				jobsError:
					error instanceof Error ? error.message : "Failed to initialize auth",
			})
		}
	},

	login: async (username: string, password: string) => {
		set({ isAuthLoading: true })
		try {
			await authService.login({ username, password })
			set({ isAuthenticated: true, isAuthLoading: false })
		} catch (error) {
			set({ isAuthLoading: false })
			throw error
		}
	},

	logout: () => {
		authService.logout()
		set({ isAuthenticated: false })
	},

	// Selection actions
	setSelectedJobId: id => set({ selectedJobId: id }),
	setSelectedHistoryId: id => set({ selectedHistoryId: id }),
	setSelectedSource: source => set({ selectedSource: source }),
	setSelectedDestination: destination =>
		set({ selectedDestination: destination }),

	// Job actions
	fetchJobs: async () => {
		set({ isLoadingJobs: true, jobsError: null })
		try {
			const jobs = await jobService.getJobs()
			set({ jobs, isLoadingJobs: false })
			return jobs
		} catch (error) {
			set({
				isLoadingJobs: false,
				jobsError:
					error instanceof Error ? error.message : "Failed to fetch jobs",
			})
			throw error
		}
	},

	addJob: async jobData => {
		try {
			const newJob = await jobService.createJob(jobData)
			set(state => ({ jobs: [...state.jobs, newJob] }))
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
				jobs: state.jobs.map(job =>
					job.id.toString() === id ? updatedJob : job,
				),
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

	// Source actions
	fetchSources: async () => {
		set({ isLoadingSources: true, sourcesError: null })
		try {
			const sources = await sourceService.getSources()
			set({ sources, isLoadingSources: false })
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
			set(state => ({ sources: [...state.sources, newSource.data as Entity] }))
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
			const updatedSourceData = updatedSource.data as Entity

			set(state => ({
				sources: state.sources.map(source =>
					source.id.toString() === id ? updatedSourceData : source,
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
			await sourceService.deleteSource(numericId.toString())
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
			set({ destinations, isLoadingDestinations: false })
			return destinations
		} catch (error) {
			set({
				isLoadingDestinations: false,
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
			set(state => ({
				destinations: [
					...state.destinations,
					newDestination as unknown as Entity,
				],
			}))
			return newDestination
		} catch (error) {
			set({
				destinationsError:
					error instanceof Error ? error.message : "Failed to add destination",
			})
			throw error
		}
	},

	updateDestination: async (id, destinationData) => {
		try {
			const updatedDestination = await destinationService.updateDestination(
				id,
				destinationData as EntityBase,
			)
			const updatedDestData = updatedDestination.data as Entity

			set(state => ({
				destinations: state.destinations.map(destination =>
					destination.id.toString() === id ? updatedDestData : destination,
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

	// Modal actions
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
