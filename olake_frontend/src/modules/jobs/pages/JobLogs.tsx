import { useEffect, useState } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Spin, message, Button } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft } from "@phosphor-icons/react"

const JobLogs: React.FC = () => {
	const { jobId, historyId } = useParams<{
		jobId: string
		historyId: string
	}>()

	const navigate = useNavigate()
	const [searchText, setSearchText] = useState("")

	const { Search } = Input

	const {
		jobs,
		jobLogs,
		isLoadingJobLogs,
		jobLogsError,
		fetchJobLogs,
		fetchJobs,
	} = useAppStore()

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}

		if (jobId && historyId) {
			fetchJobLogs(jobId, historyId).catch(error => {
				message.error("Failed to fetch job logs")
				console.error(error)
			})
		}
	}, [jobId, historyId, fetchJobLogs, jobs.length, fetchJobs])

	const job = jobs.find(j => j.id === jobId)

	const getLogLevelClass = (level: string) => {
		switch (level) {
			case "debug":
				return "text-blue-600"
			case "info":
				return "text-blue-400"
			case "warning":
				return "text-amber-500"
			case "error":
				return "text-red-500"
			default:
				return "text-gray-600"
		}
	}

	const filteredLogs = jobLogs.filter(
		log =>
			log.message.toLowerCase().includes(searchText.toLowerCase()) ||
			log.level.toLowerCase().includes(searchText.toLowerCase()),
	)

	if (jobLogsError) {
		return (
			<div className="p-6">
				<div className="text-red-500">
					Error loading job logs: {jobLogsError}
				</div>
				<Button
					onClick={() => jobId && historyId && fetchJobLogs(jobId, historyId)}
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
					to={`/jobs/${jobId}/history`}
					className="mb-4 flex items-center text-blue-600"
				>
					<ArrowLeft
						size={16}
						className="mr-1"
					/>{" "}
					Back to Job History
				</Link>

				<div className="mb-2 flex items-center">
					<h1 className="text-2xl font-bold">
						{job?.name || "Jobname"} [Timestamp]
					</h1>
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

			<h2 className="mb-4 text-xl font-bold">Logs</h2>

			<div className="mb-4">
				<Search
					placeholder="Search Logs"
					allowClear
					className="w-72"
					value={searchText}
					onChange={e => setSearchText(e.target.value)}
				/>
			</div>

			{isLoadingJobLogs ? (
				<div className="flex items-center justify-center p-12">
					<Spin size="large" />
				</div>
			) : (
				<div className="overflow-hidden rounded-lg border bg-white">
					<table className="min-w-full">
						<tbody>
							{filteredLogs.map((log, index) => (
								<tr
									key={index}
									className="border-b border-gray-100 last:border-0"
								>
									<td className="w-32 px-4 py-3 text-sm text-gray-500">
										{log.timestamp}
									</td>
									<td
										className={`px-4 py-3 text-sm ${getLogLevelClass(
											log.level,
										)} w-24`}
									>
										{log.level}
									</td>
									<td className="px-4 py-3 text-sm text-gray-700">
										{log.message}
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
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

export default JobLogs
