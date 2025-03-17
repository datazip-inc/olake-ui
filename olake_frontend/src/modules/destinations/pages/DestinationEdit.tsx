import React, { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Radio, Select, Switch, message } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, Check, Notebook } from "@phosphor-icons/react"
import { DestinationJob } from "../../../types"
import Table, { ColumnsType } from "antd/es/table"
import DocumentationPanel from "../../common/components/DocumentationPanel"

const DestinationEdit: React.FC = () => {
	const { destinationId } = useParams<{ destinationId: string }>()
	const navigate = useNavigate()
	const isNewDestination = destinationId === "new"
	const [activeTab, setActiveTab] = useState("config")
	const [connector, setConnector] = useState("Amazon S3")
	const [authType, setAuthType] = useState("iam")
	const [iamInfo, setIamInfo] = useState("")
	const [connectionUrl, setConnectionUrl] = useState("")
	const [destinationName, setDestinationName] = useState("")
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [showAllJobs, setShowAllJobs] = useState(false)

	const {
		destinations,
		jobs,
		fetchDestinations,
		fetchJobs,
		// addDestination,
		// updateDestination,
	} = useAppStore()

	useEffect(() => {
		if (!destinations.length) {
			fetchDestinations()
		}

		if (!jobs.length) {
			fetchJobs()
		}

		if (!isNewDestination && destinationId) {
			const destination = destinations.find(d => d.id === destinationId)
			if (destination) {
				setDestinationName(destination.name)
				setConnector(destination.type)
				// Set other fields based on destination data
			}
		}
	}, [
		destinationId,
		isNewDestination,
		destinations,
		fetchDestinations,
		jobs.length,
		fetchJobs,
	])

	// Mock associated jobs for the destination
	const associatedJobs = jobs.slice(0, 5).map(job => ({
		...job,
		state: Math.random() > 0.7 ? "Inactive" : "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		source: "MongoDB Source",
		paused: false,
	}))

	// Additional jobs that will be shown when "View all" is clicked
	const additionalJobs = jobs.slice(5, 10).map(job => ({
		...job,
		state: Math.random() > 0.7 ? "Inactive" : "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		source: "MongoDB Source",
		paused: false,
	}))

	const displayedJobs = showAllJobs
		? [...associatedJobs, ...additionalJobs]
		: associatedJobs

	// const handleSave = () => {
	// 	const destinationData = {
	// 		name:
	// 			destinationName ||
	// 			`${connector}_destination_${Math.floor(Math.random() * 1000)}`,
	// 		type: connector,
	// 		status: "active" as const,
	// 	}

	// 	if (isNewDestination) {
	// 		addDestination(destinationData)
	// 			.then(() => {
	// 				message.success("Destination created successfully")
	// 				navigate("/destinations")
	// 			})
	// 			.catch(error => {
	// 				message.error("Failed to create destination")
	// 				console.error(error)
	// 			})
	// 	} else if (destinationId) {
	// 		updateDestination(destinationId, destinationData)
	// 			.then(() => {
	// 				message.success("Destination updated successfully")
	// 				navigate("/destinations")
	// 			})
	// 			.catch(error => {
	// 				message.error("Failed to update destination")
	// 				console.error(error)
	// 			})
	// 	}
	// }

	const handleDelete = () => {
		message.success("Destination deleted successfully")
		navigate("/destinations")
	}

	const handleTestConnection = () => {
		message.success("Connection test successful")
	}

	const handleCreateJob = () => {
		message.info("Creating job from this destination")
		navigate("/jobs/new")
	}

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

	const toggleDocsPanel = () => {
		setDocsMinimized(!docsMinimized)
	}

	const columns: ColumnsType<DestinationJob> = [
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
							? "bg-[#FFF1F0] text-[#F5222D]"
							: "bg-[#E6F4FF] text-[#0958D9]"
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
				<button className="flex items-center gap-2 rounded bg-[#F6FFED] px-2 text-[#389E0D]">
					<Check className="size-4" />
					{status}
				</button>
			),
		},
		{
			title: "Source",
			dataIndex: "source",
			key: "source",
			render: (source: string) => (
				<div className="flex items-center">
					<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-white">
						<span>D</span>
					</div>
					{source}
				</div>
			),
		},
		{
			title: "Pause job",
			dataIndex: "id",
			key: "pause",
			render: (_: string, record: DestinationJob) => (
				<Switch
					checked={record.paused}
					onChange={checked => handlePauseJob(record.id, checked)}
					className={record.paused ? "bg-blue-600" : "bg-gray-200"}
				/>
			),
		},
	]

	return (
		<div className="flex h-screen flex-col">
			{/* Header */}
			<div className="flex gap-2 px-6 pb-0 pt-6">
				<Link
					to="/destinations"
					className="mb-4 flex items-center"
				>
					<ArrowLeft className="size-5" />
				</Link>

				<div className="mb-4 flex items-center">
					<h1 className="text-2xl font-bold">
						{isNewDestination
							? "Create New Destination"
							: destinationName || "<Destination_name>"}
					</h1>
				</div>
			</div>

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden border border-t border-[#D9D9D9]">
				{/* Left content */}
				<div
					className={`${
						docsMinimized ? "w-full" : "w-3/4"
					} mt-4 overflow-auto p-6 pt-0 transition-all duration-300`}
				>
					<div className="mb-4">
						<div className="flex">
							<button
								className={`w-56 rounded-xl px-3 py-1.5 text-sm font-normal ${
									activeTab === "config"
										? "mr-1 bg-[#203fdd] text-center text-[#F0F0F0]"
										: "mr-1 bg-[#F5F5F5] text-center text-[#0A0A0A]"
								}`}
								onClick={() => setActiveTab("config")}
							>
								Config
							</button>
							{!isNewDestination && (
								<button
									className={`w-56 rounded-xl px-3 py-1.5 text-sm font-normal ${
										activeTab === "jobs"
											? "mr-1 bg-[#203fdd] text-center text-[#F0F0F0]"
											: "mr-1 bg-[#F5F5F5] text-center text-[#0A0A0A]"
									}`}
									onClick={() => setActiveTab("jobs")}
								>
									Associated jobs
								</button>
							)}
						</div>
					</div>

					{activeTab === "config" ? (
						<div className="rounded-lg">
							<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
								<div className="mb-4 flex items-center gap-1 text-lg font-medium">
									<Notebook className="size-5" />
									Capture information
								</div>

								<div className="grid grid-cols-2 gap-6">
									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Connector:
										</label>
										<div className="flex items-center">
											<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-white">
												<span>
													{connector === "Amazon S3"
														? "S"
														: connector.charAt(0)}
												</span>
											</div>
											<Select
												value={connector}
												onChange={setConnector}
												className="w-full"
												options={[
													{ value: "Amazon S3", label: "Amazon S3" },
													{ value: "Snowflake", label: "Snowflake" },
													{ value: "BigQuery", label: "BigQuery" },
													{ value: "Redshift", label: "Redshift" },
												]}
											/>
										</div>
									</div>

									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Name of your destination:
										</label>
										<Input
											placeholder="Enter the name of your destination"
											value={destinationName}
											onChange={e => setDestinationName(e.target.value)}
										/>
									</div>
								</div>
							</div>

							<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
								<h3 className="mb-4 text-lg font-medium">Endpoint config</h3>
								<div className="mb-4 flex">
									<Radio.Group
										value={authType}
										onChange={e => setAuthType(e.target.value)}
										className="flex"
									>
										<Radio
											value="iam"
											className="mr-8"
										>
											IAM
										</Radio>
										<Radio value="keys">Access keys</Radio>
									</Radio.Group>
								</div>

								<div className="mb-4">
									<label className="mb-2 block text-sm font-medium text-gray-700">
										IAM info:
									</label>
									<Input
										placeholder="Enter your IAM info"
										value={iamInfo}
										onChange={e => setIamInfo(e.target.value)}
									/>
								</div>

								<div className="mb-4">
									<label className="mb-2 block text-sm font-medium text-gray-700">
										Connection URL:
									</label>
									<Input
										placeholder="Enter your connection URL"
										value={connectionUrl}
										onChange={e => setConnectionUrl(e.target.value)}
									/>
								</div>
							</div>
						</div>
					) : (
						<div className="">
							<h3 className="mb-4 text-base font-medium">Associated jobs</h3>

							<Table
								columns={columns}
								dataSource={displayedJobs}
								pagination={false}
								rowKey={record => record.id}
								className="min-w-full"
							/>

							{!showAllJobs && additionalJobs.length > 0 && (
								<div className="mt-6 flex justify-center">
									<Button
										type="default"
										onClick={handleViewAllJobs}
										className="w-full border-none bg-[#E9EBFC] font-medium text-[#203FDD]"
									>
										View all associated jobs
									</Button>
								</div>
							)}

							<div className="mt-6 flex items-center justify-between rounded-xl border border-[#D9D9D9] p-4">
								<span className="font-medium">Pause all associated jobs</span>
								<Switch
									onChange={handlePauseAllJobs}
									className="bg-gray-200"
								/>
							</div>
						</div>
					)}
				</div>

				{/* Documentation panel with iframe */}
				<DocumentationPanel
					title="Documentation"
					icon="D"
					docUrl="https://olake.io/docs/category/mongodb"
					isMinimized={docsMinimized}
					onToggle={toggleDocsPanel}
					showResizer={true}
				/>
			</div>

			{/* Footer with buttons */}
			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<div>
					{!isNewDestination && (
						<Button
							className="border border-[#F5222D] text-[#F5222D]"
							onClick={handleDelete}
						>
							Delete
						</Button>
					)}
				</div>
				<div className="flex space-x-4">
					<Button onClick={handleTestConnection}>Test connection</Button>
					<Button
						type="primary"
						className="bg-blue-600"
						onClick={handleCreateJob}
					>
						Create job
					</Button>
				</div>
			</div>
		</div>
	)
}

export default DestinationEdit
