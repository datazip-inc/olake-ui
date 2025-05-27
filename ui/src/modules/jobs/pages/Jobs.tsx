import { useState, useEffect } from "react"
import { Button, Tabs, Empty, message, Spin } from "antd"
import { useNavigate } from "react-router-dom"
import { useAppStore } from "../../../store"
import JobTable from "../components/JobTable"
import FirstJob from "../../../assets/FirstJob.svg"
import JobsTutorial from "../../../assets/JobsTutorial.svg"
import { GitCommit, PlayCircle, Plus } from "@phosphor-icons/react"
import DeleteJobModal from "../../common/Modals/DeleteJobModal"
import { jobService } from "../../../api"

const Jobs: React.FC = () => {
	const [activeTab, setActiveTab] = useState("active")
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
		navigate("/jobs/new")
	}

	const handleSyncJob = async (id: string) => {
		try {
			navigate(`/jobs/${id}/history`)
			await jobService.syncJob(id)
			message.success("Job sync started successfully")
			await fetchJobs()
		} catch (error) {
			message.error("Failed to sync job")
			console.error(error)
		}
	}

	const handleEditJob = (id: string) => {
		message.info(`Editing job ${id}`)
		navigate(`/jobs/${id}/edit`)
	}

	const handlePauseJob = async (id: string, checked: boolean) => {
		await jobService.activateJob(id, !checked)
		message.success(`Successfully ${checked ? "paused" : "resumed"} job ${id}`)
		await fetchJobs()
	}

	const handleDeleteJob = (id: string) => {
		if (activeTab === "saved") {
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
		if (activeTab === "active") {
			setFilteredJobs(jobs.filter(job => job.activate === true))
		} else if (activeTab === "inactive") {
			setFilteredJobs(jobs.filter(job => job.activate === false))
		} else if (activeTab === "saved") {
			setFilteredJobs(savedJobs)
		} else if (activeTab === "failed") {
			setFilteredJobs(
				jobs.filter(job => job.last_run_state?.toLowerCase() == "failed"),
			)
		}
	}

	const showEmpty = !isLoadingJobs && jobs.length === 0

	const tabItems = [
		{ key: "active", label: "Active jobs" },
		{ key: "inactive", label: "Inactive jobs" },
		{ key: "saved", label: "Saved jobs" },
		{ key: "failed", label: "Failed jobs" },
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
					<GitCommit className="mr-2 size-6" />
					<h1 className="text-2xl font-bold">Jobs</h1>
				</div>
				<button
					className="flex items-center justify-center gap-1 rounded-[6px] bg-[#203FDD] px-4 py-2 font-light text-white hover:bg-[#132685]"
					onClick={handleCreateJob}
				>
					<Plus className="size-4 text-white" />
					Create Job
				</button>
			</div>

			<p className="mb-6 text-gray-600">
				A list of all your jobs stacked at one place for you to see
			</p>

			<Tabs
				activeKey={activeTab}
				onChange={setActiveTab}
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
					) : tab.key === "active" && showEmpty ? (
						<div className="flex flex-col items-center justify-center py-16">
							<img
								src={FirstJob}
								alt="Empty state"
								className="mb-8 h-64 w-96"
							/>
							<div className="mb-2 text-[#193AE6]">Welcome User !</div>
							<h2 className="mb-2 text-2xl font-bold">
								Ready to run your first Job
							</h2>
							<p className="mb-8 text-[#0A0A0A]">
								Get started and experience the speed of O<b>Lake</b> by running
								jobs
							</p>
							<Button
								type="primary"
								className="mb-12 bg-[#193AE6] text-sm"
								onClick={handleCreateJob}
							>
								<GitCommit />
								Create your first Job
							</Button>
							<div className="w-[412px] rounded-xl border-[1px] border-[#D9D9D9] bg-white p-6 shadow-sm">
								<div className="flex items-center gap-4">
									<img
										src={JobsTutorial}
										alt="Job Tutorial"
										className="rounded-lg"
									/>
									<div className="flex-1">
										<div className="mb-1 flex items-center gap-1 text-xs">
											<PlayCircle color="#9f9f9f" />
											<span className="text-[#9F9F9F]">OLake/ Tutorial</span>
										</div>
										<div className="text-xs">
											Checkout this tutorial, to know more about running jobs
										</div>
									</div>
								</div>
							</div>
						</div>
					) : filteredJobs.length === 0 ? (
						<Empty
							image={Empty.PRESENTED_IMAGE_SIMPLE}
							description="No data"
							className="flex flex-col items-start"
						/>
					) : (
						<JobTable
							jobs={filteredJobs}
							loading={isLoadingJobs}
							jobType={activeTab as "active" | "inactive" | "saved" | "failed"}
							onSync={handleSyncJob}
							onEdit={handleEditJob}
							onPause={handlePauseJob}
							onDelete={handleDeleteJob}
						/>
					),
				}))}
			/>
			<DeleteJobModal />
		</div>
	)
}

export default Jobs
