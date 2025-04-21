import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Button, Table, Switch, message } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft } from "@phosphor-icons/react"

const SourceJobs: React.FC = () => {
	const { sourceId } = useParams<{ sourceId: string }>()
	const navigate = useNavigate()
	const [showAllJobs, setShowAllJobs] = useState(false)

	const { sources, jobs, fetchSources, fetchJobs } = useAppStore()

	useEffect(() => {
		if (!sources.length) {
			fetchSources()
		}

		if (!jobs.length) {
			fetchJobs()
		}
	}, [fetchSources, fetchJobs, sources.length, jobs.length])

	const source = sources.find(s => s.id === sourceId)

	// Mock associated jobs for the source
	const associatedJobs = jobs.slice(0, 5).map(job => ({
		...job,
		state: Math.random() > 0.7 ? "Inactive" : "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		destination: "Amazon S3 destination",
		paused: false,
	}))

	// Additional jobs that will be shown when "View all" is clicked
	const additionalJobs = jobs.slice(5, 10).map(job => ({
		...job,
		state: Math.random() > 0.7 ? "Inactive" : "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		destination: "Amazon S3 destination",
		paused: false,
	}))

	const displayedJobs = showAllJobs
		? [...associatedJobs, ...additionalJobs]
		: associatedJobs

	const handleViewAllJobs = () => {
		setShowAllJobs(true)
	}

	const handlePauseAllJobs = (checked: boolean) => {
		message.info(`${checked ? "Pausing" : "Resuming"} all jobs for this source`)
	}

	const handlePauseJob = (jobId: string, checked: boolean) => {
		message.info(`${checked ? "Pausing" : "Resuming"} job ${jobId}`)
	}

	const columns = [
		{
			title: "Name",
			dataIndex: "name",
			key: "name",
		},
		{
			title: "State",
			dataIndex: "state",
			key: "state",
			render: (state: string) => (
				<span
					className={`rounded px-2 py-1 text-xs ${
						state === "Inactive"
							? "bg-red-100 text-red-600"
							: "bg-blue-100 text-blue-600"
					}`}
				>
					{state}
				</span>
			),
		},
		{
			title: "Last runtime",
			dataIndex: "lastRuntime",
			key: "lastRuntime",
		},
		{
			title: "Last runtime status",
			dataIndex: "lastRuntimeStatus",
			key: "lastRuntimeStatus",
			render: (status: string) => (
				<span className="flex items-center text-green-500">
					<span className="mr-2 h-2 w-2 rounded-full bg-green-500"></span>
					{status}
				</span>
			),
		},
		{
			title: "Destination",
			dataIndex: "destination",
			key: "destination",
			render: (text: string) => (
				<div className="flex items-center">
					<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-white">
						<span>D</span>
					</div>
					{text}
				</div>
			),
		},
		{
			title: "Pause job",
			key: "pause",

			render: (_: any, record: any) => (
				<Switch
					checked={record.paused}
					onChange={checked => handlePauseJob(record.id, checked)}
					className={record.paused ? "bg-blue-600" : "bg-gray-200"}
				/>
			),
		},
	]

	return (
		<div className="p-6">
			<div className="mb-6">
				<Link
					to={`/sources/${sourceId}`}
					className="mb-4 flex items-center text-blue-600"
				>
					<ArrowLeft
						size={16}
						className="mr-1"
					/>{" "}
					Back to Source
				</Link>

				<div className="mb-2 flex items-center">
					<h1 className="text-2xl font-bold">{source?.name || "Source"}</h1>
					<span className="ml-2 rounded bg-blue-100 px-2 py-1 text-xs text-blue-600">
						{source?.status || "Active"}
					</span>
				</div>
			</div>

			<div className="rounded-lg border border-gray-200 bg-white p-6">
				<div className="mb-4 flex">
					<button
						className={`border-b border-gray-200 px-4 py-3 text-sm font-medium text-gray-500 hover:text-gray-700`}
						onClick={() => navigate(`/sources/${sourceId}`)}
					>
						Config
					</button>
					<button
						className={`border-b-2 border-blue-600 px-4 py-3 text-sm font-medium text-blue-600`}
					>
						Associated jobs
					</button>
				</div>

				<h3 className="mb-4 text-lg font-medium">Associated jobs</h3>

				<Table
					dataSource={displayedJobs}
					columns={columns}
					rowKey="id"
					pagination={false}
					className="mb-6"
				/>

				{!showAllJobs && additionalJobs.length > 0 && (
					<div className="mt-6 flex justify-center">
						<Button
							type="default"
							onClick={handleViewAllJobs}
						>
							View all associated jobs
						</Button>
					</div>
				)}

				<div className="mt-6 flex items-center justify-between border-t border-gray-200 pt-6">
					<span className="font-medium">Pause all associated jobs</span>
					<Switch
						onChange={handlePauseAllJobs}
						className="bg-gray-200"
					/>
				</div>
			</div>
		</div>
	)
}

export default SourceJobs
