import { useState, useEffect } from "react"
import { useParams, Link, useNavigate } from "react-router-dom"
import { Input, Button, Switch, message, Select, Radio, Tooltip } from "antd"
import { ArrowRight, Info } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { ArrowLeft } from "@phosphor-icons/react"
import {
	getConnectorImage,
	generateCronExpression,
	parseCronExpression,
	validateCronExpression,
	isValidCronExpression,
} from "../../../utils/utils"
import DeleteJobModal from "../../common/Modals/DeleteJobModal"
import ClearDataModal from "../../common/Modals/ClearDataModal"
import ClearDestinationAndSyncModal from "../../common/Modals/ClearDestinationAndSyncModal"
import { jobService } from "../../../api"
import { DAYS, FREQUENCY_OPTIONS } from "../../../utils/constants"
import parser from "cron-parser"

const JobSettings: React.FC = () => {
	const { jobId } = useParams<{ jobId: string }>()
	const [jobName, setJobName] = useState("")
	const navigate = useNavigate()

	// Cron-related states
	const [selectedTime, setSelectedTime] = useState("1")
	const [selectedAmPm, setSelectedAmPm] = useState<"AM" | "PM">("AM")
	const [selectedDay, setSelectedDay] = useState("Sunday")
	const [frequency, setFrequency] = useState("minutes")
	const [customCronExpression, setCustomCronExpression] = useState("")
	const [cronExpression, setCronExpression] = useState("* * * * *")
	const [nextRuns, setNextRuns] = useState<string[]>([])

	// Configuration object for all select options
	const selectConfig = {
		frequency: FREQUENCY_OPTIONS,
		time: Array.from({ length: 12 }, (_, i) => ({
			value: (i + 1).toString(),
			label: (i + 1).toString(),
		})),
		days: DAYS.map(day => ({ value: day, label: day })),
	}

	const { jobs, fetchJobs, setShowDeleteJobModal, setSelectedJobId } =
		useAppStore()

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}
	}, [fetchJobs, jobs.length])

	const job = jobs.find(j => j.id.toString() === jobId)
	const [pauseJob, setPauseJob] = useState(job ? !job.activate : true)

	const getParsedDate = (value: Date) => value.toUTCString()

	const updateNextRuns = (cronValue: string) => {
		if (!cronValue || !isValidCronExpression(cronValue)) {
			setNextRuns([])
			return
		}

		try {
			const interval = parser.parse(cronValue, {
				currentDate: new Date(),
				tz: "UTC",
			})
			const data = []
			for (let i = 0; i < 3; i++) {
				data.push(getParsedDate(interval.next().toDate()))
			}
			setNextRuns(data)
		} catch (error) {
			console.error(
				"Invalid cron expression:",
				error instanceof Error ? error.message : String(error),
			)
			setNextRuns([])
		}
	}

	// Parse initial cron expression and set states
	useEffect(() => {
		if (job?.frequency) {
			const result = parseCronExpression(job.frequency, DAYS)

			setFrequency(result.frequency)
			if (result.customCronExpression) {
				setCustomCronExpression(result.customCronExpression)
			}
			if (result.selectedTime) {
				setSelectedTime(result.selectedTime)
			}
			if (result.selectedAmPm) {
				setSelectedAmPm(result.selectedAmPm)
			}
			if (result.selectedDay) {
				setSelectedDay(result.selectedDay)
			}

			setCronExpression(job.frequency)
			updateNextRuns(job.frequency)
		}
		if (job) {
			setPauseJob(!job.activate)
			setJobName(job.name)
		}
	}, [job])

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

	// Unified handler for all cron expression updates
	const updateCronExpression = (
		freq?: string,
		time?: string,
		amPm?: "AM" | "PM",
		day?: string,
	) => {
		const f = freq || frequency
		const t = time || selectedTime
		const ap = amPm || selectedAmPm
		const d = day || selectedDay

		if (f === "custom") {
			setCronExpression(customCronExpression)
			updateNextRuns(customCronExpression)
		} else {
			const newCronExpression = generateCronExpression(f, t, ap, d)
			setCronExpression(newCronExpression)
			updateNextRuns(newCronExpression)
		}
	}

	const handleFrequencyChange = (selectedUnit: string) => {
		setFrequency(selectedUnit)
		if (selectedUnit === "custom") {
			setCronExpression(customCronExpression)
			updateNextRuns(customCronExpression)
		} else {
			updateCronExpression(selectedUnit)
		}
	}

	const handleTimeChange = (value: string) => {
		setSelectedTime(value)
		if (frequency !== "custom") {
			updateCronExpression(undefined, value)
		}
	}

	const handleAmPmChange = (value: "AM" | "PM") => {
		setSelectedAmPm(value)
		if (frequency !== "custom") {
			updateCronExpression(undefined, undefined, value)
		}
	}

	const handleDayChange = (value: string) => {
		setSelectedDay(value)
		if (frequency !== "custom") {
			updateCronExpression(undefined, undefined, undefined, value)
		}
	}

	const handleCustomCronChange = (value: string) => {
		setCustomCronExpression(value)
		setCronExpression(value)
		updateNextRuns(value)
	}

	const handleDeleteJob = () => {
		if (jobId) {
			setSelectedJobId(jobId)
		}
		setShowDeleteJobModal(true)
	}

	// Helper to determine if time selection should be shown
	const isTimeSelectionFrequency = (freq: string): boolean => {
		return freq === "days" || freq === "weeks"
	}

	const shouldShowTimeSelection =
		isTimeSelectionFrequency(frequency) && frequency !== "custom"

	const handleSaveSettings = async () => {
		if (!jobId || !job) {
			message.error("Job details not found.")
			return
		}

		if (!jobName.trim()) {
			message.error("Job name is required")
			return
		}
		if (!validateCronExpression(cronExpression)) {
			return
		}

		try {
			const jobUpdatePayload = {
				name: jobName,
				frequency: cronExpression,
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
										className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
									>
										<ArrowLeft className="size-5" />
									</Link>

									<div className="text-2xl font-bold">{job?.name}</div>
								</div>
								<div className="ml-10 mt-1.5 w-fit rounded bg-primary-200 px-2 py-1 text-xs text-primary-700">
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
											value={jobName}
											onChange={e => setJobName(e.target.value)}
											className="max-w-md"
										/>
									</div>

									<div className="mb-6 w-2/3">
										<div className="flex gap-4">
											<div>
												<label className="mb-2 block text-sm">Frequency</label>
												<Select
													className="w-40"
													value={frequency}
													onChange={handleFrequencyChange}
													options={selectConfig.frequency}
												/>
											</div>

											{frequency === "custom" && (
												<div>
													<div className="mb-2 flex items-center gap-1">
														<label className="block text-sm">
															Cron Expression
														</label>
														<Tooltip title="Cron format: minute hour day month weekday. Example: 0 0 * * * runs every day at midnight.">
															<Info
																size={16}
																className="cursor-help text-slate-900"
															/>
														</Tooltip>
													</div>
													<Input
														className="w-64"
														placeholder="Enter cron expression (Eg : * * * * *)"
														value={customCronExpression}
														onChange={e =>
															handleCustomCronChange(e.target.value)
														}
													/>
												</div>
											)}

											{frequency === "weeks" && (
												<div>
													<label className="mb-2 block text-sm">
														Select Day
													</label>
													<Select
														className="w-36"
														value={selectedDay}
														onChange={handleDayChange}
														options={selectConfig.days}
														placeholder="Select Day"
													/>
												</div>
											)}

											{shouldShowTimeSelection && (
												<div className={frequency === "weeks" ? "" : "ml-4"}>
													<label className="mb-2 block text-sm">
														Job Start Time{" "}
														<span className="text-gray-500">
															(12H Format UTC)
														</span>
													</label>
													<div className="flex items-center gap-1">
														<Select
															className="w-24"
															value={selectedTime}
															onChange={handleTimeChange}
															options={selectConfig.time}
														/>
														<Radio.Group
															value={selectedAmPm}
															onChange={e => handleAmPmChange(e.target.value)}
														>
															<Radio.Button value="AM">AM</Radio.Button>
															<Radio.Button value="PM">PM</Radio.Button>
														</Radio.Group>
													</div>
												</div>
											)}
										</div>
										{nextRuns.length > 0 && (
											<div className="mt-4 flex gap-2">
												<span className="text-sm font-medium">
													Next 3 Runs (UTC):
												</span>
												<div className="flex gap-4">
													{nextRuns.map((run, index) => (
														<span
															key={index}
															className="text-sm text-gray-600"
														>
															{run}
														</span>
													))}
												</div>
											</div>
										)}
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
								<div className="border-gray-200 pt-4">
									<div className="mb-2 flex items-center justify-between">
										<div className="flex flex-col gap-2">
											<div className="font-medium">Delete the job:</div>
											<div className="text-sm text-text-tertiary">
												No data will be deleted in your source and destination.
											</div>
										</div>
										<button
											onClick={handleDeleteJob}
											className="rounded-md border bg-danger px-4 py-1 font-light text-white hover:bg-danger-dark"
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

				<div className="flex justify-end border-t border-gray-200 bg-white p-4 shadow-sm">
					<Button
						type="primary"
						onClick={handleSaveSettings}
						className="flex items-center gap-1 bg-primary hover:bg-primary-600"
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
