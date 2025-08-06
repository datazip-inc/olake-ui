import { useEffect, useState } from "react"
import { useParams, useNavigate, Link, useSearchParams } from "react-router-dom"
import { Input, Spin, message, Button, Tooltip } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, ArrowRight, ArrowsClockwise } from "@phosphor-icons/react"
import {
	getConnectorImage,
	getLogLevelClass,
	getLogTextColor,
} from "../../../utils/utils"

const JobLogs: React.FC = () => {
	const { jobId, historyId } = useParams<{
		jobId: string
		historyId: string
		taskId?: string
	}>()
	const [searchParams] = useSearchParams()
	const filePath = searchParams.get("file")
	const isTaskLog = Boolean(filePath)

	const navigate = useNavigate()
	const [searchText, setSearchText] = useState("")
	const [showOnlyErrors, setShowOnlyErrors] = useState(false)

	const { Search } = Input

	const {
		jobs,
		taskLogs,
		isLoadingTaskLogs,
		taskLogsError,
		fetchTaskLogs,
		fetchJobs,
	} = useAppStore()

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}

		if (jobId) {
			if (isTaskLog && filePath) {
				fetchTaskLogs(jobId, historyId || "1", filePath).catch(error => {
					message.error("Failed to fetch task logs")
					console.error(error)
				})
			}
		}
	}, [
		jobId,
		historyId,
		filePath,
		isTaskLog,
		fetchTaskLogs,
		jobs.length,
		fetchJobs,
	])

	const job = jobs.find(j => j.id === Number(jobId))

	const filteredLogs = taskLogs.filter(function (log) {
		if (typeof log !== "object" || log === null) {
			return false
		}

		const message = (log as any).message || ""
		const level = (log as any).level || ""

		const searchLowerCase = searchText.toLowerCase()
		const messageLowerCase = message.toString().toLowerCase()
		const levelLowerCase = level.toString().toLowerCase()

		const matchesSearch =
			messageLowerCase.includes(searchLowerCase) ||
			levelLowerCase.includes(searchLowerCase)

		if (showOnlyErrors) {
			return (
				matchesSearch &&
				(messageLowerCase.includes("error") ||
					levelLowerCase.includes("error") ||
					levelLowerCase.includes("fatal"))
			)
		}

		return matchesSearch
	})

	if (taskLogsError) {
		return (
			<div className="p-6">
				<Button
					onClick={() => {
						if (isTaskLog && filePath) {
							if (jobId) {
								fetchTaskLogs(jobId, historyId || "1", filePath)
							}
						} else {
							if (jobId && historyId) {
							}
						}
					}}
					className="mt-4"
				>
					Retry
				</Button>
			</div>
		)
	}

	return (
		<div className="flex h-screen flex-col">
			<div className="mb-3 flex items-center justify-between px-6 pt-3">
				<div>
					<div className="mb-2 flex items-center">
						<div className="flex items-center gap-2">
							<div>
								<Link
									to={`/jobs/${jobId}/history`}
									className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-[#f6f6f6] hover:text-black"
								>
									<ArrowLeft className="size-5" />
								</Link>
							</div>
							<div className="flex flex-col items-start">
								<div className="text-2xl font-bold">
									{job?.name || "Jobname"}{" "}
								</div>
							</div>
						</div>
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

			<div className="flex flex-1 flex-col overflow-hidden border-t border-gray-200 p-6">
				<h2 className="mb-4 text-xl font-bold">Logs</h2>

				<div className="mb-4 flex items-center gap-3">
					<Search
						placeholder="Search Logs"
						allowClear
						className="w-1/4"
						value={searchText}
						onChange={e => setSearchText(e.target.value)}
					/>
					<Tooltip title="Click to refetch the logs">
						<Button
							icon={<ArrowsClockwise size={16} />}
							onClick={() => {
								if (isTaskLog && filePath) {
									fetchTaskLogs(jobId!, historyId || "1", filePath)
										.then(() => {
											message.success("Logs refetched successfully")
										})
										.catch(error => {
											message.error("Failed to refetch task logs")
											console.error(error)
										})
								}
							}}
							className="flex items-center"
						></Button>
					</Tooltip>
					<Button
						type={showOnlyErrors ? "primary" : "default"}
						onClick={() => setShowOnlyErrors(!showOnlyErrors)}
						className="flex items-center"
					>
						Errors
					</Button>
				</div>

				{isLoadingTaskLogs ? (
					<div className="flex items-center justify-center p-12">
						<Spin size="large" />
					</div>
				) : (
					<div
						className={`overflow-scroll rounded-xl ${filteredLogs.length > 0 ? "border" : ""} bg-white`}
					>
						<table className="min-w-full">
							<tbody>
								{filteredLogs.map((log, index) => {
									// Handle rendering differently based on the log type
									if (isTaskLog) {
										// TaskLog structure
										const taskLog = log as any
										return (
											<tr key={index}>
												<td className="w-32 px-4 py-3 text-sm text-gray-500">
													{/* Extract date from ISO timestamp if possible */}
													{taskLog.time
														? new Date(taskLog.time).toLocaleDateString()
														: ""}
												</td>
												<td className="w-24 px-4 py-3 text-sm text-gray-500">
													{/* Extract time from ISO timestamp if possible */}
													{taskLog.time
														? new Date(taskLog.time).toLocaleTimeString(
																"en-US",
																{ timeZone: "UTC", hour12: false },
															)
														: ""}
												</td>
												<td className="w-24 px-4 py-3 text-sm">
													<span
														className={`rounded-md px-2 py-[5px] text-xs capitalize ${getLogLevelClass(
															taskLog.level,
														)}`}
													>
														{taskLog.level}
													</span>
												</td>
												<td
													className={`px-4 py-3 text-sm ${getLogTextColor(taskLog.level)}`}
												>
													{taskLog.message}
												</td>
											</tr>
										)
									} else {
										const jobLog = log as any
										return (
											<tr key={index}>
												<td className="w-32 px-4 py-3 text-sm text-gray-500">
													{jobLog.date}
												</td>
												<td className="w-24 px-4 py-3 text-sm text-gray-500">
													{jobLog.time}
												</td>
												<td className="w-24 px-4 py-3 text-sm">
													<span
														className={`rounded-xl px-2 py-[5px] text-xs capitalize ${getLogLevelClass(
															jobLog.level,
														)}`}
													>
														{jobLog.level}
													</span>
												</td>
												<td
													className={`px-4 py-3 text-sm text-gray-700 ${getLogTextColor(jobLog.level)}`}
												>
													{jobLog.message}
												</td>
											</tr>
										)
									}
								})}
							</tbody>
						</table>
					</div>
				)}
			</div>

			<div className="flex justify-end border-t border-gray-200 bg-white p-4">
				<Button
					type="primary"
					className="bg-primary font-extralight text-white"
					onClick={() => navigate(`/jobs/${jobId}/settings`)}
				>
					View job configurations
					<ArrowRight size={16} />
				</Button>
			</div>
		</div>
	)
}

export default JobLogs
