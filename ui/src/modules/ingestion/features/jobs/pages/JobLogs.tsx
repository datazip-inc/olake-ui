import {
	ArrowLeftIcon,
	ArrowRightIcon,
	ArrowsClockwiseIcon,
	CpuIcon,
	DatabaseIcon,
	DownloadIcon,
	LightbulbIcon,
} from "@phosphor-icons/react"
import { Input, Spin, Button, Segmented, Tooltip, message } from "antd"
import clsx from "clsx"
import {
	useEffect,
	useState,
	useRef,
	useCallback,
	useMemo,
	useLayoutEffect,
} from "react"
import { useParams, useNavigate, Link, useSearchParams } from "react-router-dom"
import { Virtuoso, type VirtuosoHandle } from "react-virtuoso"

import { getLogLevelClass, getLogTextColor } from "@/common/utils"
import { getConnectorImage } from "@/modules/ingestion/common/utils"

import { LOGS_CONFIG } from "../constants"
import { useJobDetails } from "../hooks"
import { jobService } from "../services"
import { useTaskStore } from "../stores"
import { TaskLogEntry, TaskLogSource } from "../types"

const INITIAL_SCROLL_TIMEOUT = 100 // Timeout in ms for initial scroll to bottom

// Per-source styling: badge, the short accent bar and the row tint.
const LOG_SOURCE_STYLES = {
	sync: {
		label: "SYNC",
		Icon: DatabaseIcon,
		badgeClass: "border-indigo-300 bg-indigo-50 text-indigo-700",
		accentClass: "bg-indigo-400",
		rowClass: "",
	},
	worker: {
		label: "WORKER",
		Icon: CpuIcon,
		badgeClass: "border-teal-300 bg-teal-50 text-teal-700",
		accentClass: "bg-teal-600",
		rowClass: "bg-teal-50/30",
	},
} as const

const LOG_SOURCE_TABS: {
	value: TaskLogSource
	label: string
	dotClass?: string
}[] = [
	{ value: "all", label: "All logs" },
	{
		value: "sync",
		label: "Sync logs",
		dotClass: LOG_SOURCE_STYLES.sync.accentClass,
	},
	{
		value: "worker",
		label: "Worker logs",
		dotClass: LOG_SOURCE_STYLES.worker.accentClass,
	},
]

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

	const [firstItemIndex, setFirstItemIndex] = useState<number>(
		LOGS_CONFIG.VIRTUAL_LIST_START_INDEX,
	)
	const virtuosoRef = useRef<VirtuosoHandle | null>(null)

	const hasPerformedInitialScroll = useRef(false)
	const isFetchingOlderRef = useRef(false)
	const previousLogCountRef = useRef(0)

	const {
		taskLogs,
		taskLogsSource,
		isLoadingTaskLogs,
		isLoadingOlderLogs,
		isLoadingNewerLogs,
		taskLogsHasMoreOlder,
		taskLogsHasMoreNewer,
		taskLogsError,
		fetchInitialTaskLogs,
		fetchOlderTaskLogs,
		fetchNewerTaskLogs,
	} = useTaskStore()

	const { data: job } = useJobDetails(jobId || "")

	const filteredLogs = useMemo(() => {
		const search = searchText.toLowerCase()

		return taskLogs?.filter(log => {
			if (!log) {
				return false
			}

			const lowerMessage = log.message.toLowerCase()
			const lowerLevel = log.level.toLowerCase()

			const matchesSearch =
				lowerMessage.includes(search) || lowerLevel.includes(search)

			if (showOnlyErrors) {
				return (
					matchesSearch &&
					(lowerMessage.includes("error") ||
						lowerLevel.includes("error") ||
						lowerLevel.includes("fatal"))
				)
			}

			return matchesSearch
		})
	}, [taskLogs, searchText, showOnlyErrors])

	const isFiltering = useMemo(
		() => Boolean(searchText.trim().length) || showOnlyErrors,
		[searchText, showOnlyErrors],
	)

	const handleDownloadLogs = () => {
		if (!jobId || !filePath) return
		jobService.downloadTaskLogs(jobId, filePath)
		message.success("Downloading logs...")
	}

	// Every run opens on the merged view
	useEffect(() => {
		useTaskStore.setState({ taskLogsSource: "all" })
	}, [filePath])

	// Fetch initial batch of task logs (or refetch after filters are cleared),
	useEffect(() => {
		if (!jobId || !isTaskLog || !filePath || isFiltering) {
			return
		}

		setFirstItemIndex(LOGS_CONFIG.VIRTUAL_LIST_START_INDEX)
		hasPerformedInitialScroll.current = false
		isFetchingOlderRef.current = false
		previousLogCountRef.current = 0

		fetchInitialTaskLogs(jobId, historyId || "1", filePath)
	}, [jobId, isTaskLog, filePath, historyId, isFiltering, fetchInitialTaskLogs])

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

			setFirstItemIndex(LOGS_CONFIG.VIRTUAL_LIST_START_INDEX - currentCount)
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
			const timeoutId = setTimeout(() => {
				virtuosoRef.current?.scrollToIndex({
					index: filteredLogs.length - 1,
					align: "end",
				})
				hasPerformedInitialScroll.current = true
			}, INITIAL_SCROLL_TIMEOUT)

			return () => {
				clearTimeout(timeoutId)
			}
		}
	}, [isLoadingTaskLogs, filteredLogs?.length])

	const handleStartReached = useCallback(() => {
		if (
			isFiltering ||
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
		isFiltering,
		isLoadingOlderLogs,
		taskLogsHasMoreOlder,
		filteredLogs?.length,
		jobId,
		filePath,
		historyId,
		fetchOlderTaskLogs,
	])

	const handleEndReached = useCallback(() => {
		if (isFiltering || isLoadingNewerLogs || !taskLogsHasMoreNewer) return

		if (jobId && filePath) {
			fetchNewerTaskLogs(jobId, historyId || "1", filePath)
		}
	}, [
		isFiltering,
		isLoadingNewerLogs,
		taskLogsHasMoreNewer,
		jobId,
		filePath,
		historyId,
		fetchNewerTaskLogs,
	])

	const resetScrollState = () => {
		hasPerformedInitialScroll.current = false
		isFetchingOlderRef.current = false
		previousLogCountRef.current = 0
		setFirstItemIndex(LOGS_CONFIG.VIRTUAL_LIST_START_INDEX)
	}

	const handleRefresh = () => {
		if (isTaskLog && filePath && jobId) {
			resetScrollState()
			fetchInitialTaskLogs(jobId, historyId || "1", filePath)
		}
	}

	const handleSourceChange = (source: TaskLogSource) => {
		if (!isTaskLog || !filePath || !jobId || source === taskLogsSource) return
		resetScrollState()
		fetchInitialTaskLogs(jobId, historyId || "1", filePath, source)
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
		<div className="flex h-full flex-col">
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
				<h2 className="mb-1 text-xl font-bold">Logs</h2>
				<p className="mb-4 text-sm text-gray-500">
					Sync logs come from the connector run. Worker logs come from the
					machine that scheduled and ran the job.
				</p>

				<Segmented<TaskLogSource>
					className="mb-4 self-start"
					value={taskLogsSource}
					onChange={handleSourceChange}
					options={LOG_SOURCE_TABS.map(tab => ({
						value: tab.value,
						label: (
							<span className="flex items-center gap-2 px-2">
								{tab.dotClass && (
									<span className={clsx("size-2 rounded-full", tab.dotClass)} />
								)}
								{tab.label}
							</span>
						),
					}))}
				/>

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
							data-testid="refresh-logs-button"
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
							<Virtuoso<TaskLogEntry>
								ref={virtuosoRef}
								data={filteredLogs}
								startReached={handleStartReached}
								endReached={handleEndReached}
								firstItemIndex={firstItemIndex}
								overscan={LOGS_CONFIG.OVERSCAN}
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
								itemContent={(_, log) => (log ? <JobLogRow log={log} /> : null)}
							/>
						)}
					</div>
				)}
			</div>

			<div className="flex justify-end gap-x-2 border-t border-gray-200 bg-white p-4">
				<Button
					type="default"
					className="font-extralight"
					onClick={handleDownloadLogs}
					disabled={!jobId || !filePath}
				>
					<DownloadIcon size={16} />
					Download Logs
				</Button>
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

const JobLogRow: React.FC<{ log: TaskLogEntry }> = ({ log }) => {
	const { label, Icon, badgeClass, accentClass, rowClass } =
		LOG_SOURCE_STYLES[log.source]
	return (
		<div
			className={clsx(
				"relative grid grid-cols-[8rem_6rem_7rem_6rem_minmax(0,1fr)] border-b border-gray-100",
				rowClass,
			)}
		>
			{/* short source accent, vertically centred in the row */}
			<span
				className={clsx(
					"absolute left-1 top-1/2 h-8 w-1 -translate-y-1/2 rounded-full",
					accentClass,
				)}
			/>
			<div className="px-4 py-3 font-mono text-xs font-medium text-gray-500">
				{log.date}
			</div>
			<div className="px-4 py-3 font-mono text-xs font-medium text-gray-500">
				{log.time}
			</div>
			<div className="min-w-0 px-2 py-3 leading-5">
				<span
					className={clsx(
						"inline-flex items-center gap-1 rounded-md border px-2 py-[3px] text-xs font-semibold uppercase",
						badgeClass,
					)}
				>
					<Icon
						size={12}
						className="shrink-0"
					/>
					{label}
				</span>
			</div>
			<div className="px-4 py-3 font-sans text-xs font-medium leading-5">
				<span
					className={clsx(
						"rounded-md px-2 py-[5px] text-xs capitalize",
						getLogLevelClass(log.level),
					)}
				>
					{log.level}
				</span>
			</div>
			{log.tip ? (
				<div className="m-1 flex items-start gap-2 rounded-md bg-emerald-50 px-3 py-2 font-mono text-xs font-medium text-emerald-800">
					<LightbulbIcon
						size={14}
						weight="fill"
						className="mt-px shrink-0"
					/>
					{log.message}
				</div>
			) : (
				<div
					className={clsx(
						"px-4 py-3 font-mono text-xs font-medium",
						getLogTextColor(log.level),
					)}
				>
					{log.message}
				</div>
			)}
		</div>
	)
}

export default JobLogs
