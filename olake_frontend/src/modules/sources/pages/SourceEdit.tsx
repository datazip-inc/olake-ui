import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Radio, Select, Switch, message, Table } from "antd"
import { Check, GearFine, GenderNeuter, Notebook } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { ArrowLeft, CaretDown } from "@phosphor-icons/react"
import type { ColumnsType } from "antd/es/table"
import { SourceJob } from "../../../types"
import DocumentationPanel from "../../common/components/DocumentationPanel"

const SourceEdit: React.FC = () => {
	const { sourceId } = useParams<{ sourceId: string }>()
	const navigate = useNavigate()
	const isNewSource = sourceId === "new"
	const [activeTab, setActiveTab] = useState("config")
	const [connector, setConnector] = useState("MongoDB")
	const [connectionType, setConnectionType] = useState("uri")
	const [connectionUri, setConnectionUri] = useState("")
	const [sourceName, setSourceName] = useState("")
	const [srvEnabled, setSrvEnabled] = useState(false)
	const [showAdvanced, setShowAdvanced] = useState(false)
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [showAllJobs, setShowAllJobs] = useState(false)

	const {
		sources,
		jobs,
		fetchSources,
		fetchJobs,
		//  addSource, updateSource
	} = useAppStore()

	useEffect(() => {
		if (!sources.length) {
			fetchSources()
		}

		if (!jobs.length) {
			fetchJobs()
		}

		if (!isNewSource && sourceId) {
			const source = sources.find(s => s.id === sourceId)
			if (source) {
				setSourceName(source.name)
				// Set other fields based on source data
			}
		}
	}, [sourceId, isNewSource, sources, fetchSources, jobs.length, fetchJobs])

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

	// const handleSave = () => {
	//   const sourceData = {
	//     name: sourceName || `MongoDB_source_${Math.floor(Math.random() * 1000)}`,
	//     type: connector,
	//     status: "active" as const,
	//   };

	//   if (isNewSource) {
	//     addSource(sourceData)
	//       .then(() => {
	//         message.success("Source created successfully");
	//         navigate("/sources");
	//       })
	//       .catch((error) => {
	//         message.error("Failed to create source");
	//         console.error(error);
	//       });
	//   } else if (sourceId) {
	//     updateSource(sourceId, sourceData)
	//       .then(() => {
	//         message.success("Source updated successfully");
	//         navigate("/sources");
	//       })
	//       .catch((error) => {
	//         message.error("Failed to update source");
	//         console.error(error);
	//       });
	//   }
	// };

	const handleDelete = () => {
		message.success("Source deleted successfully")
		navigate("/sources")
	}

	const handleTestConnection = () => {
		message.success("Connection test successful")
	}

	const handleCreateJob = () => {
		message.info("Creating job from this source")
		navigate("/jobs/new")
	}

	const handleViewAllJobs = () => {
		setShowAllJobs(true)
	}

	const handlePauseAllJobs = (checked: boolean) => {
		message.info(`${checked ? "Pausing" : "Resuming"} all jobs for this source`)
	}

	const handlePauseJob = (jobId: string, checked: boolean) => {
		message.info(`${checked ? "Pausing" : "Resuming"} job ${jobId}`)
	}

	const toggleDocsPanel = () => {
		setDocsMinimized(!docsMinimized)
	}

	const columns: ColumnsType<SourceJob> = [
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
			title: "Destination",
			dataIndex: "destination",
			key: "destination",
			render: (destination: string) => (
				<div className="flex items-center">
					<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-white">
						<span>D</span>
					</div>
					{destination}
				</div>
			),
		},
		{
			title: "Pause job",
			dataIndex: "id",
			key: "pause",
			render: (_: string, record: SourceJob) => (
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
					to="/sources"
					className="mb-4 flex items-center"
				>
					<ArrowLeft className="size-5" />
				</Link>

				<div className="mb-4 flex items-center">
					<h1 className="text-2xl font-bold">
						{isNewSource
							? "Create New Source"
							: sourceName || "MongoDB_Source_1"}
					</h1>
				</div>
			</div>

			{/* Main content */}
			<div className="mt-2 flex flex-1 overflow-hidden border border-t border-[#D9D9D9]">
				{/* Left content */}
				<div
					className={`${
						docsMinimized ? "w-full" : "w-3/4"
					} overflow-auto p-6 pt-4 transition-all duration-300`}
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
							{!isNewSource && (
								<button
									className={`w-56 rounded-[6px] px-3 py-1.5 text-sm font-normal ${
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
						<div className="bg-white p-6">
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
											<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-green-500 text-white">
												<span>M</span>
											</div>
											<Select
												value={connector}
												onChange={setConnector}
												className="w-full"
												options={[
													{ value: "MongoDB", label: "MongoDB" },
													{ value: "PostgreSQL", label: "PostgreSQL" },
													{ value: "MySQL", label: "MySQL" },
													{ value: "Kafka", label: "Kafka" },
												]}
											/>
										</div>
									</div>

									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Name of your source:
										</label>
										<Input
											placeholder="Enter the name of your source"
											value={sourceName}
											onChange={e => setSourceName(e.target.value)}
										/>
									</div>
								</div>
							</div>

							<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
								<div className="mb-2 flex items-center gap-1">
									<GenderNeuter className="size-6" />
									<div className="text-lg font-medium">Endpoint config</div>
								</div>
								<div className="mb-4 flex">
									<Radio.Group
										value={connectionType}
										onChange={e => setConnectionType(e.target.value)}
										className="flex"
									>
										<Radio
											value="uri"
											className="mr-8"
										>
											Connection URI
										</Radio>
										<Radio value="hosts">Hosts</Radio>
									</Radio.Group>
								</div>

								<div className="mb-4">
									<label className="mb-2 block text-sm font-medium text-gray-700">
										Connection URI:
									</label>
									<Input
										placeholder="Enter your connection URI"
										value={connectionUri}
										className="w-2/5"
										onChange={e => setConnectionUri(e.target.value)}
									/>
								</div>
							</div>

							<div className="mb-6 rounded-xl border border-[#D9D9D9]">
								<div
									className="flex cursor-pointer items-center justify-between p-4"
									onClick={() => setShowAdvanced(!showAdvanced)}
								>
									<div className="flex items-center gap-1">
										<GearFine className="size-5" />
										<div className="font-medium">Advanced configurations</div>
									</div>
									<CaretDown
										className={`transform transition-transform ${
											showAdvanced ? "rotate-180" : ""
										}`}
										size={16}
									/>
								</div>

								{showAdvanced && (
									<div className="border-t border-gray-200 p-4">
										<div className="grid grid-cols-2 gap-6">
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Replica set:
												</label>
												<Input placeholder="Input" />
											</div>
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Auth DB:
												</label>
												<Input placeholder="Input" />
											</div>
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Read preference:
												</label>
												<Input placeholder="Input" />
											</div>
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Server RAM:
												</label>
												<Input placeholder="Input" />
											</div>
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Max threads:
												</label>
												<Input placeholder="Input" />
											</div>
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Default mode:
												</label>
												<Input placeholder="Input" />
											</div>
											<div className="col-span-2">
												<div className="flex items-center justify-between">
													<span className="font-medium">SRV</span>
													<Switch
														checked={srvEnabled}
														onChange={setSrvEnabled}
														className={srvEnabled ? "bg-blue-600" : ""}
													/>
												</div>
											</div>
										</div>
									</div>
								)}
							</div>
						</div>
					) : (
						<div className="rounded-lg p-6">
							<h3 className="mb-4 text-lg font-medium">Associated jobs</h3>

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
					docUrl="https://olake.io/docs/category/mongodb"
					isMinimized={docsMinimized}
					onToggle={toggleDocsPanel}
					showResizer={true}
				/>
			</div>

			{/* Footer with buttons */}
			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<div>
					{!isNewSource && (
						<button
							className="rounded-[6px] border border-[#F5222D] px-4 py-1 text-[#F5222D] hover:bg-[#F5222D] hover:text-white"
							onClick={handleDelete}
						>
							Delete
						</button>
					)}
				</div>
				<div className="flex space-x-4">
					<button
						onClick={handleTestConnection}
						className="flex items-center justify-center gap-2 rounded-[6px] border border-[#D9D9D9] px-4 py-1 font-light hover:bg-[#EBEBEB]"
					>
						Test connection
					</button>
					<button
						className="flex items-center justify-center gap-1 rounded-[6px] bg-[#203FDD] px-4 py-1 font-light text-white hover:bg-[#132685]"
						onClick={handleCreateJob}
					>
						Use source
					</button>
				</div>
			</div>
		</div>
	)
}

export default SourceEdit
