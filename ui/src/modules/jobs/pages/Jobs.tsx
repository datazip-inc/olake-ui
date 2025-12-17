import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"
import { Button, Tabs, Empty, message, Spin } from "antd"
import { GitCommitIcon, PlusIcon } from "@phosphor-icons/react"

import { useAppStore } from "../../../store"
import { jobService } from "../../../api"
import analyticsService from "../../../api/services/analyticsService"
import { JobStatus } from "../../../types/jobTypes"
import { JOB_STATUS } from "../../../utils/constants"
import JobTable from "../components/JobTable"
import JobEmptyState from "../components/JobEmptyState"
import DeleteJobModal from "../../common/Modals/DeleteJobModal"
import { AnalyticsEvent } from "../../../api/enums"

const Jobs: React.FC = () => {
	const [activeTab, setActiveTab] = useState<JobStatus>(
		JOB_STATUS.ACTIVE as JobStatus,
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
		fetchJobs()
	}, [])

	const handleCreateJob = () => {
		analyticsService.trackEvent(AnalyticsEvent.CreateJobClicked)
		navigate("/jobs/new")
	}

	const handleSyncJob = async (id: string) => {
		try {
			await jobService.syncJob(id)
			navigate(`/jobs/${id}/history`, {
				state: {
					waitForNewSync: true,
					syncStartTime: Date.now(),
				},
			}) // navigate to job history so that user can see the tasks running
		} catch (error) {
			message.error(error as string)
			console.error("Error syncing job:", error)
		}
	}

	const handleEditJob = (id: string) => {
		if (activeTab === JOB_STATUS.SAVED) {
			const savedJob = savedJobs.find(job => job.id.toString() === id)
			if (savedJob) {
				const initialData = {
					sourceName: savedJob.source.name,
					sourceConnector: savedJob.source.type,
					sourceVersion: savedJob.source.version,
					sourceFormData: JSON.parse(savedJob.source.config),
					sourceId: savedJob.source.id,
					destinationName: savedJob.destination.name,
					destinationConnector: savedJob.destination.type,
					destinationVersion: savedJob.destination.version,
					destinationFormData: JSON.parse(savedJob.destination.config),
					destinationId: savedJob.destination.id,
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
		await jobService.activateJob(id, !checked)
		await fetchJobs()
	}

	// cancels the running job
	const handleCancelJob = async (id: string) => {
		try {
			await jobService.cancelJob(id)
		} catch (error) {
			console.error("Error canceling job:", error)
		}
	}

	const handleDeleteJob = (id: string) => {
		if (activeTab === JOB_STATUS.SAVED) {
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
			case JOB_STATUS.ACTIVE:
				setFilteredJobs(jobs.filter(job => job?.activate === true))
				break
			case JOB_STATUS.INACTIVE:
				setFilteredJobs(jobs.filter(job => job?.activate === false))
				break
			case JOB_STATUS.SAVED:
				setFilteredJobs(savedJobs)
				break
			case JOB_STATUS.FAILED:
				setFilteredJobs(
					jobs.filter(
						job => (job?.last_run_state ?? "").toLowerCase() === "failed",
					),
				)
				break
			default:
				// Handle unexpected activeTab values gracefully
				setFilteredJobs([])
		}
	}

	const showEmpty = !isLoadingJobs && jobs.length === 0

	const tabItems = [
		{ key: JOB_STATUS.ACTIVE, label: "Active jobs" },
		{ key: JOB_STATUS.INACTIVE, label: "Inactive jobs" },
		{ key: JOB_STATUS.SAVED, label: "Saved jobs" },
		{ key: JOB_STATUS.FAILED, label: "Failed jobs" },
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
				onChange={key => setActiveTab(key as JobStatus)}
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
					) : tab.key === JOB_STATUS.ACTIVE && showEmpty ? (
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
