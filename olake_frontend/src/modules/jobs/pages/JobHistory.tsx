import { useEffect, useState } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Table, Button, Input, Spin, message } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, Eye } from "@phosphor-icons/react"

const JobHistory: React.FC = () => {
	const { jobId } = useParams<{ jobId: string }>()
	const navigate = useNavigate()
	const [searchText, setSearchText] = useState("")

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
				return "text-green-500"
			case "failed":
				return "text-red-500"
			case "running":
				return "text-blue-500"
			default:
				return "text-gray-500"
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
				<span className={getStatusClass(status)}>{status}</span>
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
		<div className="p-6">
			<div className="mb-6">
				<Link
					to="/jobs"
					className="mb-4 flex items-center text-blue-600"
				>
					<ArrowLeft
						size={16}
						className="mr-1"
					/>{" "}
					Back to Jobs
				</Link>

				<div className="mb-2 flex items-center">
					<h1 className="text-2xl font-bold">{job?.name || "<Job_name>"}</h1>
					<span className="ml-2 rounded bg-blue-100 px-2 py-1 text-xs text-blue-600">
						{job?.status || "Active"}
					</span>
				</div>

				<div className="mb-6 flex items-center justify-between">
					<div className="flex items-center">
						<div className="mr-8 flex items-center">
							<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-green-500 text-white">
								S
							</div>
							<span>Source</span>
						</div>
						<div className="w-16 border-t-2 border-dashed border-gray-300"></div>
						<div className="ml-8 flex items-center">
							<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-red-500 text-white">
								D
							</div>
							<span>Destination</span>
						</div>
					</div>

					<div className="flex items-center">
						<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-blue-500 text-white">
							D
						</div>
						<div className="ml-1 flex h-8 w-8 items-center justify-center rounded-full bg-amber-500 text-white">
							A
						</div>
					</div>
				</div>
			</div>

			<h2 className="mb-4 text-xl font-bold">Job history</h2>

			<div className="mb-4">
				<Search
					placeholder="Search Jobs"
					allowClear
					className="w-72"
					value={searchText}
					onChange={e => setSearchText(e.target.value)}
				/>
			</div>

			{isLoadingJobHistory ? (
				<div className="flex items-center justify-center p-12">
					<Spin size="large" />
				</div>
			) : (
				<Table
					dataSource={filteredHistory}
					columns={columns}
					rowKey="id"
					pagination={{
						pageSize: 10,
						showSizeChanger: false,
					}}
					className="overflow-hidden rounded-lg border"
				/>
			)}

			<div className="mt-6 flex justify-end">
				<Button
					type="primary"
					className="bg-blue-600"
					onClick={() => navigate(`/jobs/${jobId}/settings`)}
				>
					View job configurations →
				</Button>
			</div>
		</div>
	)
}

export default JobHistory
