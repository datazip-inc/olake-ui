import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { Button, Tabs, Empty, message, Spin } from "antd"
import { GitCommitIcon, PlusIcon } from "@phosphor-icons/react"

import { useAppStore } from "@store"
import { jobService } from "@api"
import analyticsService from "@api/services/analyticsService"
import { JobType } from "@app-types/jobTypes"
import { JOB_TYPES } from "@utils/constants"
import JobTable from "../components/JobTable"
import JobEmptyState from "../components/JobEmptyState"
import DeleteJobModal from "@modules/common/Modals/DeleteJobModal"

const Jobs: React.FC = () => {
	const [activeTab, setActiveTab] = useState<JobType>(
		JOB_TYPES.ACTIVE as JobType,
	)
	const navigate = useNavigate()
	const {
		jobs,
		isLoadingJobs,
		jobsError,
		fetchJobs,
		setShowDeleteJobModal,
		setSelectedJobId,
	} = useAppStore()

	useEffect(() => {
		fetchJobs().catch(error => {
			message.error("Failed to fetch jobs")
			console.error(error)
		})
	}, [fetchJobs])

	const handleCreateJob = () => {
		analyticsService.trackEvent("create_job_clicked")
		navigate("/jobs/new")
	}

	const handleSyncJob = async (id: string) => {
		try {
			navigate(`/jobs/${id}/history`) // navigate to job history so that user can see the tasks running
			await jobService.syncJob(id)
			message.success("Job sync started successfully")
			await fetchJobs()
		} catch (error) {
			message.error("Failed to sync job")
			console.error(error)
		}
	}

	const handleEditJob = (id: string) => {
		if (activeTab === JOB_TYPES.SAVED) {
			const savedJob = savedJobs.find(job => job.id.toString() === id)
			if (savedJob) {
				const initialData = {
					sourceName: savedJob.source.name,
					sourceConnector: savedJob.source.type,
					sourceVersion: savedJob.source.version,
					sourceFormData: JSON.parse(savedJob.source.config),
					destinationName: savedJob.destination.name,
					destinationConnector: savedJob.destination.type,
					destinationVersion: savedJob.destination.version,
					destinationFormData: JSON.parse(savedJob.destination.config),
					selectedStreams: JSON.parse(savedJob.streams_config),
					jobName: savedJob.name,
					cronExpression: savedJob.frequency,
					isJobNameFilled: true,
				}
				navigate("/jobs/new", {
					state: {
						initialData,
						savedJobId: savedJob.id,
					},
				})
				return
			}
		}
		navigate(`/jobs/${id}/edit`)
	}

	const handlePauseJob = async (id: string, checked: boolean) => {
		const job = jobs.find(j => j.id.toString() === id)
		await jobService.activateJob(id, !checked)
		message.success(
			`Successfully ${checked ? "paused" : "resumed"} ${job?.name || id}`,
		)
		await fetchJobs()
	}

	// cancels the running job
	const handleCancelJob = async (id: string) => {
		try {
			const response = await jobService.cancelJob(id)
			message.success(response)
		} catch (error) {
			console.error("Error canceling job:", error)
			message.error("Failed to cancel run")
		}
	}

	const handleDeleteJob = (id: string) => {
		if (activeTab === JOB_TYPES.SAVED) {
			const savedJobsFromStorage = JSON.parse(
				localStorage.getItem("savedJobs") || "[]",
			)
			const updatedSavedJobs = savedJobsFromStorage.filter(
				(job: any) => job.id !== id,
			)
			localStorage.setItem("savedJobs", JSON.stringify(updatedSavedJobs))
			setSavedJobs(updatedSavedJobs)
			message.success("Saved job deleted successfully")
		} else {
			setShowDeleteJobModal(true)
			setSelectedJobId(id)
		}
	}
	const [filteredJobs, setFilteredJobs] = useState<typeof jobs>([])
	const [savedJobs, setSavedJobs] = useState<typeof jobs>([])

	useEffect(() => {
		const savedJobsFromStorage = JSON.parse(
			localStorage.getItem("savedJobs") || "[]",
		)
		setSavedJobs(savedJobsFromStorage)
	}, [])

	useEffect(() => {
		updateJobsList()
	}, [activeTab, jobs, savedJobs])

	const updateJobsList = () => {
		switch (activeTab) {
			case JOB_TYPES.ACTIVE:
				setFilteredJobs(jobs.filter(job => job.activate === true))
				break
			case JOB_TYPES.INACTIVE:
				setFilteredJobs(jobs.filter(job => job.activate === false))
				break
			case JOB_TYPES.SAVED:
				setFilteredJobs(savedJobs)
				break
			case JOB_TYPES.FAILED:
				setFilteredJobs(
					jobs.filter(job => job.last_run_state?.toLowerCase() === "failed"),
				)
				break
			default:
				// Handle unexpected activeTab values gracefully
				setFilteredJobs([])
		}
	}

	const showEmpty = !isLoadingJobs && jobs.length === 0

	const tabItems = [
		{ key: JOB_TYPES.ACTIVE, label: "Active jobs" },
		{ key: JOB_TYPES.INACTIVE, label: "Inactive jobs" },
		{ key: JOB_TYPES.SAVED, label: "Saved jobs" },
		{ key: JOB_TYPES.FAILED, label: "Failed jobs" },
	]

	if (jobsError) {
		return (
			<div className="p-6">
				<div className="text-red-500">Error loading jobs: {jobsError}</div>
				<Button
					onClick={() => fetchJobs()}
					className="mt-4"
				>
					Retry
				</Button>
			</div>
		)
	}

	return (
		<div className="p-6">
			<div className="mb-4 flex items-center justify-between">
				<div className="flex items-center gap-2">
					<GitCommitIcon className="mr-2 size-6" />
					<h1 className="text-2xl font-bold">Jobs</h1>
				</div>
				<button
					className="flex items-center justify-center gap-1 rounded-md bg-primary px-4 py-2 font-light text-white hover:bg-primary-600"
					onClick={handleCreateJob}
				>
					<PlusIcon className="size-4 text-white" />
					Create Job
				</button>
			</div>

			<p className="mb-6 text-gray-600">
				A list of all your jobs stacked at one place for you to see
			</p>

			<Tabs
				activeKey={activeTab}
				onChange={key => setActiveTab(key as JobType)}
				items={tabItems.map(tab => ({
					key: tab.key,
					label: tab.label,
					children: isLoadingJobs ? (
						<div className="flex items-center justify-center py-16">
							<Spin
								size="large"
								tip="Loading sources..."
							/>
						</div>
					) : tab.key === JOB_TYPES.ACTIVE && showEmpty ? (
						<JobEmptyState handleCreateJob={handleCreateJob} />
					) : filteredJobs.length === 0 ? (
						<Empty
							image={Empty.PRESENTED_IMAGE_SIMPLE}
							description="No jobs configured"
							className="flex flex-col items-center"
						/>
					) : (
						<JobTable
							jobs={filteredJobs}
							loading={isLoadingJobs}
							jobType={activeTab}
							onSync={handleSyncJob}
							onEdit={handleEditJob}
							onPause={handlePauseJob}
							onDelete={handleDeleteJob}
							onCancelJob={handleCancelJob}
						/>
					),
				}))}
			/>
			<DeleteJobModal />
		</div>
	)
}

export default Jobs
