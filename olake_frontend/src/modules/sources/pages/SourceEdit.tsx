import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Radio, Select, Switch, message } from "antd"
import { CornersIn, CornersOut } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { ArrowLeft, CaretDown } from "@phosphor-icons/react"

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

	return (
		<div className="flex h-screen flex-col">
			{/* Header */}
			<div className="p-6 pb-0">
				<Link
					to="/sources"
					className="mb-4 flex items-center text-blue-600"
				>
					<ArrowLeft
						size={16}
						className="mr-1"
					/>{" "}
					Back to Sources
				</Link>

				<div className="mb-4 flex items-center">
					<h1 className="text-2xl font-bold">
						{isNewSource
							? "Create New Source"
							: sourceName || "MongoDB_Source_1"}
					</h1>
					{!isNewSource && (
						<span className="ml-2 rounded bg-blue-100 px-2 py-1 text-xs text-blue-600">
							Active
						</span>
					)}
				</div>
			</div>

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden">
				{/* Left content */}
				<div
					className={`${
						docsMinimized ? "w-full" : "w-3/4"
					} overflow-auto p-6 pt-0 transition-all duration-300`}
				>
					<div className="mb-4">
						<div className="flex border-b border-gray-200">
							<button
								className={`px-4 py-3 text-sm font-medium ${
									activeTab === "config"
										? "mr-1 w-48 rounded-lg bg-[#203fdd] text-center text-white"
										: "mr-1 w-48 bg-gray-200 text-center"
								}`}
								onClick={() => setActiveTab("config")}
							>
								Config
							</button>
							{!isNewSource && (
								<button
									className={`px-4 py-3 text-sm font-medium ${
										activeTab === "jobs"
											? "mr-1 w-48 rounded-lg bg-[#203fdd] text-center text-white"
											: "mr-1 w-48 bg-gray-200 text-center"
									}`}
									onClick={() => setActiveTab("jobs")}
								>
									Associated jobs
								</button>
							)}
						</div>
					</div>

					{activeTab === "config" ? (
						<div className="rounded-lg border border-gray-200 bg-white p-6">
							<div className="mb-6">
								<h3 className="mb-4 text-lg font-medium">
									Capture information
								</h3>

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

							<div className="mb-6">
								<h3 className="mb-4 text-lg font-medium">Endpoint config</h3>
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
										onChange={e => setConnectionUri(e.target.value)}
									/>
								</div>
							</div>

							<div className="mb-6 rounded-lg border border-gray-200">
								<div
									className="flex cursor-pointer items-center justify-between p-4"
									onClick={() => setShowAdvanced(!showAdvanced)}
								>
									<span className="font-medium">Advanced configurations</span>
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
						<div className="rounded-lg border border-gray-200 bg-white p-6">
							<h3 className="mb-4 text-lg font-medium">Associated jobs</h3>

							<table className="min-w-full">
								<thead>
									<tr className="border-b border-gray-200">
										<th className="px-4 py-3 text-left font-medium text-gray-700">
											Name
										</th>
										<th className="px-4 py-3 text-left font-medium text-gray-700">
											State
										</th>
										<th className="px-4 py-3 text-left font-medium text-gray-700">
											Last runtime
										</th>
										<th className="px-4 py-3 text-left font-medium text-gray-700">
											Last runtime status
										</th>
										<th className="px-4 py-3 text-left font-medium text-gray-700">
											Destination
										</th>
										<th className="px-4 py-3 text-left font-medium text-gray-700">
											Pause job
										</th>
									</tr>
								</thead>
								<tbody>
									{displayedJobs.map((job, index) => (
										<tr
											key={index}
											className="border-b border-gray-100"
										>
											<td className="px-4 py-3">{job.name}</td>
											<td className="px-4 py-3">
												<span
													className={`rounded px-2 py-1 text-xs ${
														job.state === "Inactive"
															? "bg-red-100 text-red-600"
															: "bg-blue-100 text-blue-600"
													}`}
												>
													{job.state}
												</span>
											</td>
											<td className="px-4 py-3">{job.lastRuntime}</td>
											<td className="px-4 py-3">
												<span className="flex items-center text-green-500">
													<span className="mr-2 h-2 w-2 rounded-full bg-green-500"></span>
													{job.lastRuntimeStatus}
												</span>
											</td>
											<td className="px-4 py-3">
												<div className="flex items-center">
													<div className="mr-2 flex h-6 w-6 items-center justify-center rounded-full bg-red-500 text-white">
														<span>D</span>
													</div>
													{job.destination}
												</div>
											</td>
											<td className="px-4 py-3">
												<Switch
													checked={job.paused}
													onChange={checked => handlePauseJob(job.id, checked)}
													className={job.paused ? "bg-blue-600" : "bg-gray-200"}
												/>
											</td>
										</tr>
									))}
								</tbody>
							</table>

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
					)}
				</div>

				{/* Documentation panel with iframe */}
				{!docsMinimized && (
					<div className="h-[calc(100vh-120px)] w-1/4 overflow-hidden border-l border-gray-200 bg-white">
						<div className="flex items-center justify-between border-b border-gray-200 p-4">
							<div className="flex items-center">
								<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-white">
									<span className="font-bold">M</span>
								</div>
								<span className="text-lg font-bold">MongoDB</span>
							</div>
							<Button
								type="text"
								icon={<CornersIn size={16} />}
								onClick={toggleDocsPanel}
								className="hover:bg-gray-100"
							/>
						</div>

						<div className="h-[calc(100%-60px)] w-full">
							<iframe
								src="https://olake.io/docs/category/mongodb"
								className="h-full w-full border-0"
								title="MongoDB Documentation"
								sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
							/>
						</div>
					</div>
				)}

				{/* Minimized docs panel button */}
				{docsMinimized && (
					<div className="fixed bottom-6 right-6">
						<Button
							type="primary"
							className="flex items-center bg-blue-600"
							onClick={toggleDocsPanel}
							icon={
								<CornersOut
									size={16}
									className="mr-2"
								/>
							}
						>
							Show Documentation
						</Button>
					</div>
				)}
			</div>

			{/* Footer with buttons */}
			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<div>
					{!isNewSource && (
						<Button
							danger
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

export default SourceEdit
