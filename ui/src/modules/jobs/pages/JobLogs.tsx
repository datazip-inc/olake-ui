import {
	useEffect,
	useState,
	useRef,
	useCallback,
	useMemo,
	useLayoutEffect,
} from "react"
import clsx from "clsx"
import { useParams, useNavigate, Link, useSearchParams } from "react-router-dom"
import { Input, Spin, Button, Tooltip } from "antd"
import {
	ArrowLeftIcon,
	ArrowRightIcon,
	ArrowsClockwiseIcon,
} from "@phosphor-icons/react"
import { Virtuoso, type VirtuosoHandle } from "react-virtuoso"

import { useAppStore } from "../../../store"
import {
	getConnectorImage,
	getLogLevelClass,
	getLogTextColor,
} from "../../../utils/utils"

const JobLogs: React.FC = () => {
	const { jobId, historyId } = useParams<{
		jobId: string
		historyId: string
	}>()
	const [searchParams] = useSearchParams()
	const filePath = searchParams.get("file")
	const isTaskLog = Boolean(filePath)

	const navigate = useNavigate()
	const [searchText, setSearchText] = useState("")
	const [showOnlyErrors, setShowOnlyErrors] = useState(false)
	const { Search } = Input

	const START_INDEX = 1000000 // High starting index to allow space for prepending older log entries without disrupting Virtuoso's list virtualization
	const [firstItemIndex, setFirstItemIndex] = useState(START_INDEX)
	const virtuosoRef = useRef<VirtuosoHandle | null>(null)

	const hasPerformedInitialScroll = useRef(false)
	const isFetchingOlderRef = useRef(false)
	const previousLogCountRef = useRef(0)

	const {
		jobs,
		taskLogs,
		isLoadingTaskLogs,
		isLoadingOlderLogs,
		isLoadingNewerLogs,
		taskLogsHasMoreOlder,
		taskLogsHasMoreNewer,
		taskLogsError,
		fetchInitialTaskLogs,
		fetchOlderTaskLogs,
		fetchNewerTaskLogs,
		fetchJobs,
	} = useAppStore()

	const filteredLogs = useMemo(() => {
		return taskLogs?.filter(log => {
			if (typeof log !== "object" || log === null) return false

			const message = (log as any).message || ""
			const level = (log as any).level || ""
			const search = searchText.toLowerCase()

			const matchesSearch =
				message.toLowerCase().includes(search) ||
				level.toLowerCase().includes(search)

			if (showOnlyErrors) {
				return (
					matchesSearch &&
					(message.toLowerCase().includes("error") ||
						level.toLowerCase().includes("error") ||
						level.toLowerCase().includes("fatal"))
				)
			}
			return matchesSearch
		})
	}, [taskLogs, searchText, showOnlyErrors])

	useEffect(() => {
		if (!jobs.length) fetchJobs()

		if (jobId && isTaskLog && filePath) {
			setFirstItemIndex(START_INDEX)
			hasPerformedInitialScroll.current = false
			isFetchingOlderRef.current = false
			previousLogCountRef.current = 0

			fetchInitialTaskLogs(jobId, historyId || "1", filePath)
		}
	}, [
		jobs.length,
		fetchJobs,
		jobId,
		isTaskLog,
		filePath,
		historyId,
		fetchInitialTaskLogs,
	])

	const job = jobs.find(j => j.id === Number(jobId))

	// Handle Scroll Position & Index Shifting synchronously to prevent visual jumping
	useLayoutEffect(() => {
		const currentCount = filteredLogs?.length || 0
		const prevCount = previousLogCountRef.current

		if (currentCount === 0 || prevCount === 0) {
			previousLogCountRef.current = currentCount
			return
		}

		const diff = currentCount - prevCount

		// CASE 1: PREPEND (Standard Scroll Up)
		// We added items to the top. Shift the virtual index backward
		// to anchor the user's view to the same specific log line.
		if (diff > 0 && isFetchingOlderRef.current) {
			setFirstItemIndex(prev => prev - diff)
			isFetchingOlderRef.current = false
		}
		// CASE 2: MEMORY RESET (Limit Reached)
		// The list shrank because we dropped newer logs to save memory.
		// Jump to the last item of this new batch.
		else if (diff < 0 && isFetchingOlderRef.current) {
			requestAnimationFrame(() => {
				virtuosoRef.current?.scrollToIndex({
					index: currentCount - 1,
				})
			})

			setFirstItemIndex(START_INDEX - currentCount)
			isFetchingOlderRef.current = false
		}

		previousLogCountRef.current = currentCount
	}, [filteredLogs?.length])

	// Initial Scroll to Bottom
	useEffect(() => {
		if (
			!isLoadingTaskLogs &&
			filteredLogs?.length > 0 &&
			!hasPerformedInitialScroll.current
		) {
			setTimeout(() => {
				virtuosoRef.current?.scrollToIndex({
					index: filteredLogs.length - 1,
					align: "end",
				})
				hasPerformedInitialScroll.current = true
			}, 100)
		}
	}, [isLoadingTaskLogs, filteredLogs?.length])

	const handleStartReached = useCallback(() => {
		if (
			isLoadingOlderLogs ||
			!taskLogsHasMoreOlder ||
			!filteredLogs ||
			filteredLogs.length === 0
		)
			return

		isFetchingOlderRef.current = true

		if (jobId && filePath) {
			fetchOlderTaskLogs(jobId, historyId || "1", filePath)
		}
	}, [
		isLoadingOlderLogs,
		taskLogsHasMoreOlder,
		filteredLogs?.length,
		jobId,
		filePath,
		historyId,
		fetchOlderTaskLogs,
	])

	const handleEndReached = useCallback(() => {
		if (isLoadingNewerLogs || !taskLogsHasMoreNewer) return

		if (jobId && filePath) {
			fetchNewerTaskLogs(jobId, historyId || "1", filePath)
		}
	}, [
		isLoadingNewerLogs,
		taskLogsHasMoreNewer,
		jobId,
		filePath,
		historyId,
		fetchNewerTaskLogs,
	])

	const handleRefresh = () => {
		if (isTaskLog && filePath && jobId) {
			hasPerformedInitialScroll.current = false
			setFirstItemIndex(START_INDEX)
			fetchInitialTaskLogs(jobId, historyId || "1", filePath)
		}
	}

	if (taskLogsError) {
		return (
			<div className="p-6">
				<Button
					onClick={handleRefresh}
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
						className={clsx(
							"h-full rounded-xl bg-white",
							filteredLogs?.length && "border",
						)}
					>
						{!filteredLogs || filteredLogs.length === 0 ? (
							<div className="flex h-full items-center justify-center p-4 text-sm text-gray-500">
								No logs found
							</div>
						) : (
							<Virtuoso
								ref={virtuosoRef}
								data={filteredLogs}
								startReached={handleStartReached}
								endReached={handleEndReached}
								firstItemIndex={firstItemIndex}
								overscan={1000}
								followOutput={false}
								components={{
									Header: () =>
										isLoadingOlderLogs || isLoadingNewerLogs ? (
											<div className="flex justify-center bg-white/90 p-2 text-xs text-gray-500">
												<Spin size="small" />
												<span className="ml-2">Loading logs...</span>
											</div>
										) : null,
								}}
								itemContent={(_, log) => {
									if (!log) return null
									const item = log as any
									const hasTimeField = Boolean(item.time)
									return (
										<div className="grid grid-cols-[8rem_6rem_6rem_minmax(0,1fr)] border-b border-gray-100">
											<div className="px-4 py-3 text-sm text-gray-500">
												{hasTimeField
													? new Date(item.time).toLocaleDateString()
													: item.date}
											</div>
											<div className="px-4 py-3 text-sm text-gray-500">
												{hasTimeField
													? new Date(item.time).toLocaleTimeString("en-US", {
															timeZone: "UTC",
															hour12: false,
														})
													: item.time}
											</div>
											<div className="px-4 py-3 text-sm">
												<span
													className={clsx(
														"rounded-md px-2 py-[5px] text-xs capitalize",
														getLogLevelClass(item.level),
													)}
												>
													{item.level}
												</span>
											</div>
											<div
												className={clsx(
													"px-4 py-3 text-sm",
													getLogTextColor(item.level),
												)}
											>
												{item.message}
											</div>
										</div>
									)
								}}
							/>
						)}
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
