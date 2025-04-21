import { useEffect, useState } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Spin, message, Button } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, ArrowRight } from "@phosphor-icons/react"
import { getConnectorImage } from "../../../utils/utils"

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
				return "text-blue-600 bg-[#F0F5FF]"
			case "info":
				return "text-[#531DAB] bg-[#F9F0FF]"
			case "warning":
				return "text-[#FAAD14] bg-[#FFFBE6]"
			case "error":
				return "text-red-500 bg-[#FFF1F0]"
			default:
				return "text-gray-600"
		}
	}

	const getLogTextColor = (level: string) => {
		switch (level) {
			case "warning":
				return "text-[#FAAD14]"
			case "error":
				return "text-[#F5222D]"
			default:
				return "text-[#000000"
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
		<div className="flex h-screen flex-col">
			<div className="mb-6 flex items-center justify-between px-6 pt-3">
				<div>
					<div className="mb-2 flex items-center">
						<div className="flex items-center gap-2">
							<div>
								<Link
									to={`/jobs/${jobId}/history`}
									className="mt-[2px] flex items-center"
								>
									<ArrowLeft size={20} />
								</Link>
							</div>
							<div className="text-2xl font-bold">
								{job?.name || "Jobname"} [Timestamp]
							</div>
						</div>
					</div>
					<span className="ml-6 rounded bg-[#E6F4FF] px-2 py-1 text-xs capitalize text-[#203FDD]">
						{job?.status || "Active"}
					</span>
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
				<h2 className="mb-4 text-xl font-bold">Logs</h2>

				<div className="mb-4">
					<Search
						placeholder="Search Logs"
						allowClear
						className="w-1/4"
						value={searchText}
						onChange={e => setSearchText(e.target.value)}
					/>
				</div>

				{isLoadingJobLogs ? (
					<div className="flex items-center justify-center p-12">
						<Spin size="large" />
					</div>
				) : (
					<div className="overflow-scroll rounded-xl border bg-white">
						<table className="min-w-full">
							<tbody>
								{filteredLogs.map((log, index) => (
									<tr key={index}>
										<td className="w-32 px-4 py-3 text-sm text-gray-500">
											{log.date}
										</td>
										<td className="w-24 px-4 py-3 text-sm text-gray-500">
											{log.time}
										</td>
										<td className="w-24 px-4 py-3 text-sm">
											<span
												className={`rounded-xl px-2 py-[5px] text-xs capitalize ${getLogLevelClass(
													log.level,
												)}`}
											>
												{log.level}
											</span>
										</td>
										<td
											className={`px-4 py-3 text-sm text-gray-700 ${getLogTextColor(log.level)}`}
										>
											{log.message}
										</td>
									</tr>
								))}
							</tbody>
						</table>
					</div>
				)}
			</div>

			<div className="flex justify-end border-t border-gray-200 bg-white p-4">
				<Button
					type="primary"
					className="bg-[#203FDD] font-extralight text-white"
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
