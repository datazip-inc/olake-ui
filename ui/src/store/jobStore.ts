import { StateCreator } from "zustand"
import { Job, JobBase } from "../types"
import { jobService } from "../api"

export interface JobSlice {
	jobs: Job[]
	jobsError: string | null
	isLoadingJobs: boolean
	selectedJob: Job | null
	isLoadingSelectedJob: boolean
	selectedJobError: string | null
	fetchJobs: () => Promise<Job[]>
	fetchSelectedJob: (id: string) => Promise<Job>
	addJob: (job: JobBase) => Promise<Job>
	updateJob: (id: string, job: Partial<Job>) => Promise<Job>
	deleteJob: (id: string) => Promise<void>
}

export const createJobSlice: StateCreator<JobSlice> = set => ({
	jobs: [],
	jobsError: null,
	isLoadingJobs: false,
	selectedJob: null,
	isLoadingSelectedJob: false,
	selectedJobError: null,

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

	fetchSelectedJob: async (id: string) => {
		set({ isLoadingSelectedJob: true, selectedJobError: null })
		try {
			const job = await jobService.getJob(id)
			set({
				selectedJob: job,
				isLoadingSelectedJob: false,
			})
			return job
		} catch (error) {
			set({
				isLoadingSelectedJob: false,
				selectedJobError:
					error instanceof Error ? error.message : "Failed to fetch job",
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
})
