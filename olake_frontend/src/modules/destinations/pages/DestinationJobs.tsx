import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Button, Table, Switch, message } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft } from "@phosphor-icons/react"

const DestinationJobs: React.FC = () => {
	const { destinationId } = useParams<{ destinationId: string }>()
	const navigate = useNavigate()
	const [showAllJobs, setShowAllJobs] = useState(false)

	const { destinations, jobs, fetchDestinations, fetchJobs } = useAppStore()

	useEffect(() => {
		if (!destinations.length) {
			fetchDestinations()
		}

		if (!jobs.length) {
			fetchJobs()
		}
	}, [fetchDestinations, fetchJobs, destinations.length, jobs.length])

	const destination = destinations.find(d => d.id === destinationId)

	// Mock associated jobs for the destination
	const associatedJobs = jobs.slice(0, 5).map(job => ({
		...job,
		state: Math.random() > 0.7 ? "Inactive" : "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		source: "Mongo DB Athena",
		paused: false,
	}))

	// Additional jobs that will be shown when "View all" is clicked
	const additionalJobs = jobs.slice(5, 10).map(job => ({
		...job,
		state: Math.random() > 0.7 ? "Inactive" : "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		source: "Mongo DB Athena",
		paused: false,
	}))

	const displayedJobs = showAllJobs
		? [...associatedJobs, ...additionalJobs]
		: associatedJobs

	const handleViewAllJobs = () => {
		setShowAllJobs(true)
	}

	const handlePauseAllJobs = (checked: boolean) => {
		message.info(
			`${checked ? "Pausing" : "Resuming"} all jobs for this destination`,
		)
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
			title: "Source",
			dataIndex: "source",
			key: "source",
			render: (text: string) => (
				<div className="flex items-center">
					<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-green-500 text-white">
						<span>S</span>
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
					to={`/destinations/${destinationId}`}
					className="mb-4 flex items-center text-blue-600"
				>
					<ArrowLeft
						size={16}
						className="mr-1"
					/>{" "}
					Back to Destination
				</Link>

				<div className="mb-2 flex items-center">
					<h1 className="text-2xl font-bold">
						{destination?.name || "Destination"}
					</h1>
					<span className="ml-2 rounded bg-blue-100 px-2 py-1 text-xs text-blue-600">
						{destination?.status || "Active"}
					</span>
				</div>
			</div>

			<div className="rounded-lg border border-gray-200 bg-white p-6">
				<div className="mb-4 flex">
					<button
						className={`border-b border-gray-200 px-4 py-3 text-sm font-medium text-gray-500 hover:text-gray-700`}
						onClick={() => navigate(`/destinations/${destinationId}`)}
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

export default DestinationJobs
