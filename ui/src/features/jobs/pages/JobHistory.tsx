import { useState, useRef, useEffect } from "react"
import clsx from "clsx"
import { useParams, useNavigate, useLocation, Link } from "react-router-dom"
import { Table, Button, Input, Spin, Pagination, Tooltip } from "antd"
import {
	ArrowLeftIcon,
	ArrowRightIcon,
	ArrowsClockwiseIcon,
	EyeIcon,
} from "@phosphor-icons/react"

import { useJobDetails, useJobTasks } from "../hooks"
import {
	getConnectorImage,
	getStatusClass,
	getStatusLabel,
} from "@/common/utils"
import { parseDateToTimestamp } from "@/utils"
import { getJobTypeClass, getJobTypeLabel } from "../utils"
import { getStatusIcon } from "@/common/components/statusIcons"
import { JobType } from "../types"

interface JobHistoryNavState {
	waitForNewSync?: boolean
	syncStartTime?: number
}

const POLL_INTERVAL = 500
const MAX_RETRIES = 4
const TOLERANCE_MS = 2000

const JobHistory: React.FC = () => {
	const { jobId } = useParams<{ jobId: string }>()
	const navigate = useNavigate()
	const location = useLocation()
	const [searchText, setSearchText] = useState("")
	const [currentPage, setCurrentPage] = useState(1)
	const pageSize = 8

	// Get navigation state to check if we should wait for new sync
	const { waitForNewSync, syncStartTime } =
		(location.state as JobHistoryNavState) || {}

	// Polling state: starts true if navigated with waitForNewSync
	const [isPolling, setIsPolling] = useState(
		!!waitForNewSync && !!syncStartTime,
	)
	const retryCountRef = useRef(0)

	const {
		data: jobTasks = [],
		isLoading: isLoadingJobTasks,
		error: jobTasksErrorObj,
		refetch: refetchJobTasks,
	} = useJobTasks(jobId || "", {
		refetchInterval: isPolling ? POLL_INTERVAL : false,
	})

	// Stop polling when sync is found or retries exhausted
	useEffect(() => {
		if (!isPolling) return

		const foundRecentSync = jobTasks.some(task => {
			const ts = parseDateToTimestamp(task.start_time)
			return (
				task.job_type === JobType.Sync &&
				ts !== null &&
				Math.abs(ts - syncStartTime!) <= TOLERANCE_MS
			)
		})

		if (foundRecentSync) {
			setIsPolling(false)
			return
		}

		retryCountRef.current += 1
		if (retryCountRef.current >= MAX_RETRIES) {
			setIsPolling(false)
		}
	}, [jobTasks, isPolling, syncStartTime])

	const jobTasksError = jobTasksErrorObj
		? jobTasksErrorObj instanceof Error
			? jobTasksErrorObj.message
			: "Failed to fetch job tasks"
		: null

	const { data: job } = useJobDetails(jobId || "")

	const handleViewLogs = (filePath: string) => {
		if (jobId) {
			navigate(
				`/jobs/${jobId}/history/1/logs?file=${encodeURIComponent(filePath)}`,
			)
		}
	}

	const { Search } = Input

	const columns = [
		{
			title: "Start time",
			dataIndex: "start_time",
			key: "start_time",
			render: (text: string) => {
				return text.replace("_", " ").replace(/-(\d+)-(\d+)$/, ":$1:$2")
			},
		},
		{
			title: "Runtime",
			dataIndex: "runtime",
			key: "runtime",
		},
		{
			title: "Status",
			dataIndex: "status",
			key: "status",
			render: (status: string) => (
				<div
					className={clsx(
						"flex w-fit items-center justify-center gap-1 rounded-md px-4 py-1",
						getStatusClass(status),
					)}
				>
					{getStatusIcon(status.toLowerCase())}
					<span>{getStatusLabel(status.toLowerCase())}</span>
				</div>
			),
		},
		{
			title: "Job Type",
			dataIndex: "job_type",
			key: "job_type",
			render: (job_type: JobType) => (
				<div
					className={clsx(
						"flex w-fit items-center justify-center gap-1 rounded-md px-4 py-1",
						getJobTypeClass(job_type),
					)}
				>
					<span>{getJobTypeLabel(job_type)}</span>
				</div>
			),
		},
		{
			title: "Actions",
			key: "actions",
			render: (_: any, record: any) => (
				<Button
					type="default"
					icon={<EyeIcon size={16} />}
					onClick={() => handleViewLogs(record.file_path)}
				>
					View logs
				</Button>
			),
		},
	]

	const filteredTasks = jobTasks?.filter(
		item =>
			item.start_time.toLowerCase().includes(searchText.toLowerCase()) ||
			item.status.toLowerCase().includes(searchText.toLowerCase()),
	)

	const startIndex = (currentPage - 1) * pageSize
	const endIndex = Math.min(startIndex + pageSize, filteredTasks?.length)
	const currentPageData = filteredTasks?.slice(startIndex, endIndex)

	if (jobTasksError && !isLoadingJobTasks && !isPolling) {
		return (
			<div className="p-6">
				<div className="text-red-500">
					Error loading job tasks: {jobTasksError}
				</div>
			</div>
		)
	}

	return (
		<div className="flex h-screen flex-col">
			<div className="mb-3 flex items-center justify-between px-6 pt-3">
				<div>
					<div className="flex items-start gap-2">
						<Link
							to="/jobs"
							className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
						>
							<ArrowLeftIcon className="size-5" />
						</Link>

						<div className="flex flex-col items-start">
							<div className="text-2xl font-bold">
								{job?.name || "<Job_name>"}
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
				<h2 className="mb-4 text-xl font-bold">Job Logs & History</h2>

				<div className="mb-4 flex gap-2">
					<Search
						placeholder="Search Tasks"
						allowClear
						className="w-1/4"
						value={searchText}
						onChange={e => setSearchText(e.target.value)}
					/>
					<Tooltip title="Click to refetch">
						<Button
							data-testid="refresh-tasks-button"
							icon={<ArrowsClockwiseIcon size={16} />}
							onClick={() => refetchJobTasks()}
							className="flex items-center"
						></Button>
					</Tooltip>
				</div>

				{isPolling || isLoadingJobTasks ? (
					<div className="flex items-center justify-center p-12">
						<Spin size="large" />
					</div>
				) : (
					<>
						<Table
							dataSource={currentPageData}
							columns={columns}
							rowKey="file_path"
							pagination={false}
							className="overflow-scroll rounded-lg border"
						/>
					</>
				)}
			</div>

			<div
				style={{
					position: "fixed",
					bottom: 80,
					right: 40,
					display: "flex",
					justifyContent: "flex-end",
					padding: "8px 0",
					backgroundColor: "#fff",
					zIndex: 100,
				}}
			>
				<Pagination
					current={currentPage}
					onChange={setCurrentPage}
					total={filteredTasks?.length}
					pageSize={pageSize}
					showSizeChanger={false}
				/>
			</div>

			<div style={{ height: "80px" }}></div>

			<div className="flex justify-end border-t border-gray-200 bg-white p-4">
				<Button
					type="primary"
					className="font-extralight text-white"
					onClick={() => navigate(`/jobs/${jobId}/settings`)}
				>
					View job configurations
					<ArrowRightIcon size={16} />
				</Button>
			</div>
		</div>
	)
}

export default JobHistory
