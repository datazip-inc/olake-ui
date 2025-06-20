import { useState, useEffect } from "react"
import { useParams, Link, useNavigate } from "react-router-dom"
import { Input, Button, Switch, message, Select } from "antd"
import { ArrowRight } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { ArrowLeft } from "@phosphor-icons/react"
import { getConnectorImage } from "../../../utils/utils"
import DeleteJobModal from "../../common/Modals/DeleteJobModal"
import ClearDataModal from "../../common/Modals/ClearDataModal"
import ClearDestinationAndSyncModal from "../../common/Modals/ClearDestinationAndSyncModal"
import { jobService } from "../../../api"

const JobSettings: React.FC = () => {
	const { jobId } = useParams<{ jobId: string }>()
	const [replicationFrequencyValue, setReplicationFrequencyValue] =
		useState("1")
	const [jobName, setJobName] = useState("")
	const navigate = useNavigate()

	const {
		jobs,
		fetchJobs,
		setShowDeleteJobModal,
		setSelectedJobId,
		// setShowClearDestinationAndSyncModal,
	} = useAppStore()

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}
	}, [fetchJobs, jobs.length])

	const job = jobs.find(j => j.id.toString() === jobId)

	const [pauseJob, setPauseJob] = useState(job ? !job.activate : true)

	const handlePauseJob = async (jobId: string, checked: boolean) => {
		try {
			await jobService.activateJob(jobId, !checked)
			message.success(
				`Successfully ${checked ? "paused" : "resumed"} job ${jobId}`,
			)
			await fetchJobs()
		} catch (error) {
			console.error("Error toggling job status:", error)
			message.error(`Failed to ${checked ? "pause" : "resume"} job ${jobId}`)
		}
	}

	const [replicationFrequency, setReplicationFrequency] = useState(
		job?.frequency ? job.frequency.split("-")[1] : "minutes",
	)

	useEffect(() => {
		if (job?.frequency) {
			const parts = job.frequency.split("-")
			if (parts.length === 2) {
				setReplicationFrequencyValue(parts[0])
				setReplicationFrequency(parts[1])
			} else {
				setReplicationFrequencyValue("1")
				setReplicationFrequency("minutes")
			}
		} else if (job) {
			setReplicationFrequencyValue("1")
			setReplicationFrequency("minutes")
		}
		if (job) {
			setPauseJob(!job.activate)
		}
		if (job?.name) {
			setJobName(job.name)
		}
	}, [job])

	// const handleClearDestinationAndSync = () => {
	// 	setShowClearDestinationAndSyncModal(true)
	// }

	const handleDeleteJob = () => {
		if (jobId) {
			setSelectedJobId(jobId)
		}
		setShowDeleteJobModal(true)
	}

	const handleSaveSettings = async () => {
		if (!jobId || !job) {
			message.error("Job details not found.")
			return
		}

		const updatedFrequency = `${replicationFrequencyValue}-${replicationFrequency}`

		try {
			const jobUpdatePayload = {
				name: jobName,
				frequency: updatedFrequency,
				activate: job.activate,
				source: {
					...job.source,
					config:
						typeof job.source.config === "string"
							? job.source.config
							: JSON.stringify(job.source.config),
				},
				destination: {
					...job.destination,
					config:
						typeof job.destination.config === "string"
							? job.destination.config
							: JSON.stringify(job.destination.config),
				},
				streams_config:
					typeof job.streams_config === "string"
						? job.streams_config
						: JSON.stringify(job.streams_config),
			}

			await jobService.updateJob(jobId, jobUpdatePayload)
			message.success("Job settings saved successfully")
			await fetchJobs()
			navigate("/jobs")
		} catch (error) {
			console.error("Error saving job settings:", error)
			message.error("Failed to save job settings")
		}
	}

	return (
		<>
			<div className="flex h-screen flex-col">
				<div className="flex-1 overflow-hidden">
					<div className="px-6 pb-4 pt-6">
						<div className="flex items-center justify-between">
							<div>
								<div className="flex items-center gap-2">
									<Link
										to="/jobs"
										className="flex items-center gap-2 p-1.5 hover:rounded-[6px] hover:bg-[#f6f6f6] hover:text-black"
									>
										<ArrowLeft className="size-5" />
									</Link>

									<div className="text-2xl font-bold">
										{job?.name || "<Job_name>"}
									</div>
								</div>
								<div className="ml-10 mt-1.5 w-fit rounded bg-blue-100 px-2 py-1 text-xs text-[#0958D9]">
									{job?.activate ? "Active" : "Inactive"}
								</div>
							</div>

							<div className="flex items-center gap-2">
								{job?.source && (
									<img
										src={getConnectorImage(job.source.type)}
										alt="Source"
										className="size-7"
									/>
								)}
								<span className="text-gray-500">{"--------------▶"}</span>
								{job?.destination && (
									<img
										src={getConnectorImage(job.destination.type)}
										alt="Destination"
										className="size-7"
									/>
								)}
							</div>
						</div>
					</div>

					<div className="flex h-full border-t px-6">
						{/* Main content */}
						<div className="mt-2 w-full pr-6 transition-all duration-300">
							<h2 className="mb-4 text-xl font-medium">Job settings</h2>

							<div className="mb-6">
								<div className="flex w-full flex-row justify-between gap-8 rounded-xl border border-[#D9D9D9] bg-white px-6 pb-2 pt-6">
									<div className="mb-6 w-1/3">
										<label className="mb-2 block text-sm text-gray-700">
											Job name:
										</label>
										<Input
											placeholder="Enter your job name"
											defaultValue={job?.name}
											value={jobName}
											onChange={e => setJobName(e.target.value)}
											className="max-w-md"
										/>
									</div>

									<div className="mb-6 w-2/3">
										<label className="mb-2 block text-sm text-gray-700">
											Replication frequency:
										</label>
										<div className="flex w-full items-center gap-2">
											<Input
												value={replicationFrequencyValue}
												defaultValue={replicationFrequencyValue}
												onChange={e =>
													setReplicationFrequencyValue(e.target.value)
												}
												className="w-2/5"
											/>
											<Select
												className="w-3/5"
												value={replicationFrequency}
												onChange={setReplicationFrequency}
											>
												<Select.Option value="minutes">Minutes</Select.Option>
												<Select.Option value="hours">Hours</Select.Option>
												<Select.Option value="days">Days</Select.Option>
												<Select.Option value="weeks">Weeks</Select.Option>
												<Select.Option value="months">Months</Select.Option>
												<Select.Option value="years">Years</Select.Option>
											</Select>
										</div>
									</div>
								</div>

								<div className="mt-6 flex items-center justify-between rounded-xl border border-[#D9D9D9] px-6 py-4">
									<span className="font-medium">Pause your job</span>
									<Switch
										checked={pauseJob}
										onChange={newlyChecked => {
											if (job?.id) {
												handlePauseJob(job.id.toString(), newlyChecked)
											}
										}}
										className={pauseJob ? "bg-blue-600" : ""}
									/>
								</div>
							</div>

							<div className="mb-6 rounded-xl border border-gray-200 bg-white px-6 pb-2">
								{/* <div className="mb-3 border-gray-200">
									<div className="flex items-center justify-between">
										<div className="flex flex-col gap-2">
											<div className="font-medium">
												Clear destination and sync:
											</div>
											<div className="text-sm text-[#8A8A8A]">
												It will delete all the data in the destination and then
												sync the data from the source
											</div>
										</div>
										<Button
											onClick={handleClearDestinationAndSync}
											className="py-4"
										>
											Clear destination and sync
										</Button>
									</div>
								</div> */}

								<div className="border-gray-200 pt-4">
									<div className="mb-2 flex items-center justify-between">
										<div className="flex flex-col gap-2">
											<div className="font-medium">Delete the job:</div>
											<div className="text-sm text-[#8A8A8A]">
												No data will be deleted in your source and destination.
											</div>
										</div>
										<button
											onClick={handleDeleteJob}
											className="rounded-[6px] border bg-[#F5222D] px-4 py-1 font-light text-white hover:bg-[#b81922]"
										>
											Delete this job
										</button>
									</div>
								</div>
								<DeleteJobModal fromJobSettings={true} />
								<ClearDataModal />
								<ClearDestinationAndSyncModal />
							</div>
						</div>
					</div>
				</div>

				{/* Footer */}
				<div className="flex justify-end border-t border-gray-200 bg-white p-4 shadow-sm">
					<Button
						type="primary"
						onClick={handleSaveSettings}
						className="flex items-center gap-1 bg-[#203FDD] hover:bg-[#132685]"
					>
						Save{" "}
						<ArrowRight
							size={16}
							className="text-white"
						/>
					</Button>
				</div>
			</div>
		</>
	)
}

export default JobSettings
