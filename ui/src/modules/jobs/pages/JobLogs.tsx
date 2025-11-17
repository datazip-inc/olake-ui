import { useEffect, useState, useRef, useCallback } from "react"
import clsx from "clsx"
import { useParams, useNavigate, Link, useSearchParams } from "react-router-dom"
import { Input, Spin, Button, Tooltip } from "antd"
import {
	ArrowLeftIcon,
	ArrowRightIcon,
	ArrowsClockwiseIcon,
} from "@phosphor-icons/react"

import { useAppStore } from "../../../store"
import {
	getConnectorImage,
	getLogLevelClass,
	getLogTextColor,
} from "../../../utils/utils"
import { LOGS_CONFIG } from "../../../utils/constants"

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

	const scrollContainerRef = useRef<HTMLDivElement>(null)
	const previousScrollSnapshot = useRef<{ height: number; top: number } | null>(
		null,
	)
	const { Search } = Input

	const {
		jobs,
		taskLogs,
		isLoadingTaskLogs,
		isLoadingMoreLogs,
		taskLogsHasMore,
		taskLogsError,
		fetchInitialTaskLogs,
		fetchMoreTaskLogs,
		fetchJobs,
	} = useAppStore()

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}

		if (jobId) {
			if (isTaskLog && filePath) {
				fetchInitialTaskLogs(jobId, historyId || "1", filePath)
			}
		}
	}, [jobId, historyId, filePath, isTaskLog, fetchInitialTaskLogs, fetchJobs])

	// Scroll to bottom on initial load
	useEffect(() => {
		if (!isLoadingTaskLogs && scrollContainerRef.current) {
			const container = scrollContainerRef.current
			container.scrollTo({
				top: container.scrollHeight,
			})
		}
	}, [isLoadingTaskLogs])

	const handleScroll = useCallback(() => {
		if (!scrollContainerRef.current) return

		if (!isTaskLog || !jobId || !filePath || !historyId) {
			return
		}

		const { scrollTop, scrollHeight, clientHeight } = scrollContainerRef.current

		// Calculate how much we've scrolled as a percentage of total scrollable content
		const scrollableHeight = scrollHeight - clientHeight
		if (scrollableHeight <= 0) {
			return
		}

		const scrolledPercentage = scrollTop / scrollableHeight

		// Trigger when we've scrolled up to 50% of total content (or less)
		const isNearTop =
			scrolledPercentage <= LOGS_CONFIG.SCROLL_THRESHOLD_PERCENTAGE

		if (isNearTop && !isLoadingMoreLogs && taskLogsHasMore) {
			previousScrollSnapshot.current = {
				height: scrollHeight,
				top: scrollTop,
			}
			fetchMoreTaskLogs(jobId, historyId, filePath)
		}
	}, [
		isTaskLog,
		jobId,
		historyId,
		filePath,
		isLoadingMoreLogs,
		taskLogsHasMore,
		fetchMoreTaskLogs,
	])

	// Scroll to the previous position when loading more logs
	useEffect(() => {
		if (isLoadingMoreLogs) {
			return
		}

		if (!previousScrollSnapshot.current) {
			return
		}

		const container = scrollContainerRef.current
		if (!container) {
			previousScrollSnapshot.current = null
			return
		}

		const { height, top } = previousScrollSnapshot.current
		const heightDiff = container.scrollHeight - height

		container.scrollTop = top + heightDiff
		previousScrollSnapshot.current = null
	}, [isLoadingMoreLogs, taskLogs.length])

	// Add event listener to the scroll container
	useEffect(() => {
		const container = scrollContainerRef.current
		if (container) {
			container.addEventListener("scroll", handleScroll)
			return () => container.removeEventListener("scroll", handleScroll)
		}

		return () => {
			const cleanupContainer = scrollContainerRef.current
			if (cleanupContainer) {
				cleanupContainer.removeEventListener("scroll", handleScroll)
			}
		}
	}, [handleScroll])

	const job = jobs.find(j => j.id === Number(jobId))

	const filteredLogs = taskLogs?.filter(function (log) {
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
								fetchInitialTaskLogs(jobId, historyId || "1", filePath)
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

	const handleRefresh = () => {
		if (isTaskLog && filePath && jobId) {
			fetchInitialTaskLogs(jobId, historyId || "1", filePath)
		}
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
									className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
								>
									<ArrowLeftIcon className="size-5" />
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
							icon={<ArrowsClockwiseIcon size={16} />}
							onClick={handleRefresh}
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

				{isLoadingTaskLogs && !taskLogs.length ? (
					<div className="flex items-center justify-center p-12">
						<Spin size="large" />
					</div>
				) : (
					<div
						ref={scrollContainerRef}
						className={clsx(
							"overflow-scroll rounded-xl bg-white",
							filteredLogs?.length && filteredLogs.length > 0 && "border",
						)}
					>
						{isLoadingMoreLogs && (
							<div className="sticky top-0 z-10 flex items-center justify-center gap-2 border-b border-gray-100 bg-white/90 p-2 text-xs text-gray-500">
								<Spin size="default" />
								<span>Loading older logs...</span>
							</div>
						)}
						<table className="min-w-full">
							<tbody>
								{filteredLogs?.map((log, index) => {
									if (isTaskLog) {
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
														className={clsx(
															"rounded-md px-2 py-[5px] text-xs capitalize",
															getLogLevelClass(taskLog.level),
														)}
													>
														{taskLog.level}
													</span>
												</td>
												<td
													className={clsx(
														"px-4 py-3 text-sm",
														getLogTextColor(taskLog.level),
													)}
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
														className={clsx(
															"rounded-xl px-2 py-[5px] text-xs capitalize",
															getLogLevelClass(jobLog.level),
														)}
													>
														{jobLog.level}
													</span>
												</td>
												<td
													className={clsx(
														"px-4 py-3 text-sm text-gray-700",
														getLogTextColor(jobLog.level),
													)}
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
					<ArrowRightIcon size={16} />
				</Button>
			</div>
		</div>
	)
}
export default JobLogs
