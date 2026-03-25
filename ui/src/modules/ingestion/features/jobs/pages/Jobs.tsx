import { GitCommitIcon, PlusIcon } from "@phosphor-icons/react"
import { Button, Tabs, Empty, message, Spin } from "antd"
import { useState, useEffect } from "react"
import { useNavigate } from "react-router-dom"

import { trackEvent, AnalyticsEvent } from "@/core/analytics"

import { JobTable, JobEmptyState, DeleteJobModal } from "../components"
import { JOB_STATUS } from "../constants"
import {
	useJobs,
	useJobsWithNotification,
	useSyncJob,
	useActivateJob,
	useCancelJob,
} from "../hooks"
import { savedJobsService } from "../services"
import { useJobStore } from "../stores"
import { Job, JobStatus, SavedJobDraft } from "../types"

const Jobs: React.FC = () => {
	const [activeTab, setActiveTab] = useState<JobStatus>(
		JOB_STATUS.ACTIVE as JobStatus,
	)
	const navigate = useNavigate()
	const { setShowDeleteJobModal, setSelectedJobId } = useJobStore()
	const {
		data: jobs = [],
		isLoading: isLoadingJobs,
		error: jobsError,
		refetch: refetchJobs,
	} = useJobs({ refetchInterval: 5000 })
	const { refetch: refetchJobsWithNotification, isFetching: isRefreshingJobs } =
		useJobsWithNotification()
	const { mutateAsync: syncJob } = useSyncJob()
	const { mutateAsync: activateJob } = useActivateJob()
	const { mutateAsync: cancelJob } = useCancelJob()

	const handleCreateJob = () => {
		trackEvent(AnalyticsEvent.CreateJobClicked)
		navigate("/jobs/new")
	}

	const handleRefreshJobs = async () => {
		await refetchJobsWithNotification()
	}

	const handleSyncJob = async (id: string) => {
		try {
			await syncJob(id)
			navigate(`/jobs/${id}/history`, {
				state: {
					waitForNewSync: true,
					syncStartTime: Date.now(),
				},
			}) // navigate to job history so that user can see the tasks running
		} catch (error) {
			console.error("Error syncing job:", error)
		}
	}

	const handleEditJob = (id: string) => {
		if (activeTab === JOB_STATUS.SAVED) {
			const savedJob = savedJobs.find(job => job.id.toString() === id)
			if (savedJob) {
				const initialData = {
					sourceId: savedJob.source?.id ?? null,
					destinationId: savedJob.destination?.id ?? null,
					jobName: savedJob.name,
					cronExpression: savedJob.frequency,
					advanced_settings: savedJob.advanced_settings ?? null,
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
		await activateJob({ jobId: id, activate: !checked })
	}

	// cancels the running job
	const handleCancelJob = async (id: string) => {
		await cancelJob(id)
	}

	const handleDeleteJob = (id: string) => {
		if (activeTab === JOB_STATUS.SAVED) {
			setSavedJobs(savedJobsService.remove(id))
			message.success("Saved job deleted successfully")
		} else {
			setShowDeleteJobModal(true)
			setSelectedJobId(id)
		}
	}
	const [savedJobs, setSavedJobs] = useState<SavedJobDraft[]>([])

	useEffect(() => {
		setSavedJobs(savedJobsService.getAll())
	}, [])

	const filteredJobs: (Job | SavedJobDraft)[] = (() => {
		switch (activeTab) {
			case JOB_STATUS.ACTIVE:
				return jobs.filter(job => job?.activate === true)
			case JOB_STATUS.INACTIVE:
				return jobs.filter(job => job?.activate === false)
			case JOB_STATUS.SAVED:
				return savedJobs
			case JOB_STATUS.FAILED:
				return jobs.filter(
					job => (job?.last_run_state ?? "").toLowerCase() === "failed",
				)
			default:
				return []
		}
	})()

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
				<div className="text-red-500">
					Error loading jobs: {jobsError.message}
				</div>
				<Button
					onClick={() => refetchJobs()}
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
								tip="Loading jobs..."
							/>
						</div>
					) : tab.key === JOB_STATUS.ACTIVE && showEmpty ? (
						<JobEmptyState />
					) : filteredJobs.length === 0 ? (
						<Empty
							image={Empty.PRESENTED_IMAGE_SIMPLE}
							description="No jobs configured"
							className="flex flex-col items-center"
						/>
					) : (
						<JobTable
							jobs={filteredJobs}
							loading={isLoadingJobs || isRefreshingJobs}
							jobType={activeTab}
							onRefresh={handleRefreshJobs}
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
