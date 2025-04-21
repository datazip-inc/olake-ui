import { useEffect, useState } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Table, Button, Input, Spin, message, Pagination } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, ArrowRight, Eye } from "@phosphor-icons/react"
import { getConnectorImage } from "../../../utils/utils"

const JobHistory: React.FC = () => {
	const { jobId } = useParams<{ jobId: string }>()
	const navigate = useNavigate()
	const [searchText, setSearchText] = useState("")
	const [currentPage, setCurrentPage] = useState(1)
	const pageSize = 8

	const {
		jobs,
		jobHistory,
		isLoadingJobHistory,
		jobHistoryError,
		fetchJobHistory,
		fetchJobs,
	} = useAppStore()

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}

		if (jobId) {
			fetchJobHistory(jobId).catch(error => {
				message.error("Failed to fetch job history")
				console.error(error)
			})
		}
	}, [jobId, fetchJobHistory, jobs.length, fetchJobs])

	const job = jobs.find(j => j.id === jobId)
	const handleViewLogs = (historyId: string) => {
		if (jobId) {
			navigate(`/jobs/${jobId}/history/${historyId}/logs`)
		}
	}

	const { Search } = Input

	const getStatusClass = (status: string) => {
		switch (status) {
			case "success":
				return "text-[#52C41A] bg-[#F6FFED]"
			case "failed":
				return "text-[#F5222D] bg-[#FFF1F0]"
			case "running":
				return "text-[#0958D9] bg-[#E6F4FF]"
			case "scheduled":
				return "text-[rgba(0,0,0,88)] bg-[#f0f0f0]"
			default:
				return "text-[rgba(0,0,0,88)] bg-[#f0f0f0]"
		}
	}

	const columns = [
		{
			title: "Start time (UTC)",
			dataIndex: "startTime",
			key: "startTime",
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
				<span className={`${getStatusClass(status)} rounded-xl px-2 py-2`}>
					{status}
				</span>
			),
		},
		{
			title: "Actions",
			key: "actions",

			render: (_: any, record: any) => (
				<Button
					type="default"
					icon={<Eye size={16} />}
					onClick={() => handleViewLogs(record.id)}
				>
					View logs
				</Button>
			),
		},
	]

	const filteredHistory = jobHistory.filter(
		item =>
			item.startTime.toLowerCase().includes(searchText.toLowerCase()) ||
			item.status.toLowerCase().includes(searchText.toLowerCase()),
	)

	// Calculate current page data for display
	const startIndex = (currentPage - 1) * pageSize
	const endIndex = Math.min(startIndex + pageSize, filteredHistory.length)
	const currentPageData = filteredHistory.slice(startIndex, endIndex)

	if (jobHistoryError) {
		return (
			<div className="p-6">
				<div className="text-red-500">
					Error loading job history: {jobHistoryError}
				</div>
				<Button
					onClick={() => jobId && fetchJobHistory(jobId)}
					className="mt-4"
				>
					Retry
				</Button>
			</div>
		)
	}

	return (
		<div className="flex h-screen flex-col">
			<div className="mb-6 flex items-center justify-between px-6 pt-3">
				<div>
					<div className="flex items-center gap-2">
						<Link
							to="/jobs"
							className="items-cente mt-[2px] flex"
						>
							<ArrowLeft size={20} />
						</Link>

						<div className="text-2xl font-bold">
							{job?.name || "<Job_name>"}
						</div>
					</div>
					<div className="ml-6 mt-1.5 w-fit rounded bg-[#E6F4FF] px-2 py-1 text-xs capitalize text-[#0958D9]">
						{job?.status || "Active"}
					</div>
				</div>

				<div className="flex items-center gap-2">
					{job?.source && (
						<img
							src={getConnectorImage(job.source)}
							alt="Source"
							className="size-7"
						/>
					)}
					<span className="text-gray-500">{"--------------▶"}</span>
					{job?.destination && (
						<img
							src={getConnectorImage(job.destination)}
							alt="Destination"
							className="size-7"
						/>
					)}
				</div>
			</div>

			<div className="flex flex-1 flex-col overflow-hidden border-t border-gray-200 p-6">
				<h2 className="mb-4 text-xl font-bold">Job history</h2>

				<div className="mb-4">
					<Search
						placeholder="Search Jobs"
						allowClear
						className="w-1/4"
						value={searchText}
						onChange={e => setSearchText(e.target.value)}
					/>
				</div>

				{isLoadingJobHistory ? (
					<div className="flex items-center justify-center p-12">
						<Spin size="large" />
					</div>
				) : (
					<>
						<Table
							dataSource={currentPageData}
							columns={columns}
							rowKey="id"
							pagination={false}
							className="overflow-scroll rounded-lg border"
						/>
					</>
				)}
			</div>

			{/* Fixed pagination at bottom right */}
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
					total={filteredHistory.length}
					pageSize={pageSize}
					showSizeChanger={false}
				/>
			</div>

			{/* Add padding at bottom to prevent content from being hidden behind fixed pagination */}
			<div style={{ height: "80px" }}></div>

			<div className="flex justify-end border-t border-gray-200 bg-white p-4">
				<Button
					type="primary"
					className="font-extralight text-white"
					onClick={() => navigate(`/jobs/${jobId}/settings`)}
				>
					View job configurations
					<ArrowRight size={16} />
				</Button>
			</div>
		</div>
	)
}

export default JobHistory
