import { create } from "zustand"
import { Job, Source, Destination, JobHistory, JobLog } from "../types"
import { jobService, sourceService, destinationService } from "../api"
import { mockJobs, mockSources, mockDestinations } from "../api/mockData"

interface AppState {
	// Data
	jobs: Job[]
	sources: Source[]
	destinations: Destination[]
	jobHistory: JobHistory[]
	jobLogs: JobLog[]

	// Loading states
	isLoadingJobs: boolean
	isLoadingSources: boolean
	isLoadingDestinations: boolean
	isLoadingJobHistory: boolean
	isLoadingJobLogs: boolean

	// Error states
	jobsError: string | null
	sourcesError: string | null
	destinationsError: string | null
	jobHistoryError: string | null
	jobLogsError: string | null

	// Selected job
	selectedJobId: string | null
	selectedHistoryId: string | null

	//Modals
	showTestingModal: boolean
	showSuccessModal: boolean
	showEntitySavedModal: boolean

	// Actions - Jobs
	fetchJobs: () => Promise<void>
	addJob: (job: Omit<Job, "id" | "createdAt">) => Promise<Job>
	updateJob: (id: string, job: Partial<Job>) => Promise<Job>
	deleteJob: (id: string) => Promise<void>
	runJob: (id: string) => Promise<void>
	setSelectedJobId: (id: string | null) => void

	// Actions - Job History
	fetchJobHistory: (jobId: string) => Promise<void>
	setSelectedHistoryId: (id: string | null) => void

	// Actions - Job Logs
	fetchJobLogs: (jobId: string, historyId: string) => Promise<void>

	// Actions - Sources
	fetchSources: () => Promise<void>
	addSource: (source: Omit<Source, "id" | "createdAt">) => Promise<Source>
	updateSource: (id: string, source: Partial<Source>) => Promise<Source>
	deleteSource: (id: string) => Promise<void>

	// Actions - Destinations
	fetchDestinations: () => Promise<void>
	addDestination: (
		destination: Omit<Destination, "id" | "createdAt">,
	) => Promise<Destination>
	updateDestination: (
		id: string,
		destination: Partial<Destination>,
	) => Promise<Destination>
	deleteDestination: (id: string) => Promise<void>

	setShowTestingModal: (show: boolean) => void
	setShowSuccessModal: (show: boolean) => void
	setShowEntitySavedModal: (show: boolean) => void
}

export const useAppStore = create<AppState>(set => ({
	// Initial data
	jobs: [],
	sources: [],
	destinations: [],
	jobHistory: [],
	jobLogs: [],

	// Initial loading states
	isLoadingJobs: false,
	isLoadingSources: false,
	isLoadingDestinations: false,
	isLoadingJobHistory: false,
	isLoadingJobLogs: false,

	// Initial error states
	jobsError: null,
	sourcesError: null,
	destinationsError: null,
	jobHistoryError: null,
	jobLogsError: null,

	// Selected job
	selectedJobId: null,
	selectedHistoryId: null,

	// Modals
	showTestingModal: false,
	showSuccessModal: false,
	showEntitySavedModal: false,

	// Jobs actions
	fetchJobs: async () => {
		set({ isLoadingJobs: true, jobsError: null })
		try {
			// Use mock data for development
			set({ jobs: mockJobs, isLoadingJobs: false })

			// Uncomment for real API call
			// const jobs = await jobService.getJobs();
			// set({ jobs, isLoadingJobs: false });
		} catch (error) {
			set({
				isLoadingJobs: false,
				jobsError:
					error instanceof Error ? error.message : "Failed to fetch jobs",
			})
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
			await jobService.deleteJob(id)
			set(state => ({
				jobs: state.jobs.filter(job => job.id !== id),
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

	// Job Logs actions
	fetchJobLogs: async () => {
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

	// Sources actions
	fetchSources: async () => {
		set({ isLoadingSources: true, sourcesError: null })
		try {
			// Use mock data for development
			set({ sources: mockSources, isLoadingSources: false })

			// Uncomment for real API call
			// const sources = await sourceService.getSources();
			// set({ sources, isLoadingSources: false });
		} catch (error) {
			set({
				isLoadingSources: false,
				sourcesError:
					error instanceof Error ? error.message : "Failed to fetch sources",
			})
		}
	},

	addSource: async sourceData => {
		try {
			const newSource = await sourceService.createSource(sourceData)
			set(state => ({ sources: [...state.sources, newSource] }))
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
			await sourceService.deleteSource(id)
			set(state => ({
				sources: state.sources.filter(source => source.id !== id),
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
			// Use mock data for development
			set({ destinations: mockDestinations, isLoadingDestinations: false })

			// Uncomment for real API call
			// const destinations = await destinationService.getDestinations();
			// set({ destinations, isLoadingDestinations: false });
		} catch (error) {
			set({
				isLoadingDestinations: false,
				destinationsError:
					error instanceof Error
						? error.message
						: "Failed to fetch destinations",
			})
		}
	},

	addDestination: async destinationData => {
		try {
			const newDestination =
				await destinationService.createDestination(destinationData)
			set(state => ({
				destinations: [...state.destinations, newDestination],
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
			await destinationService.deleteDestination(id)
			set(state => ({
				destinations: state.destinations.filter(
					destination => destination.id !== id,
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
}))
