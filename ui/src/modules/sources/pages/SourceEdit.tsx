import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Select, Switch, message, Table, Spin } from "antd"
import { GenderNeuter, Notebook } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { ArrowLeft } from "@phosphor-icons/react"
import type { ColumnsType } from "antd/es/table"
import { SourceJob } from "../../../types"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import FixedSchemaForm from "../../../utils/FormFix"
import StepTitle from "../../common/components/StepTitle"
import DeleteModal from "../../common/Modals/DeleteModal"
import { getConnectorImage } from "../../../utils/utils"
import EditSourceModal from "../../common/Modals/EditSourceModal"
import { sourceService } from "../../../api"
import { formatDistanceToNow } from "date-fns"

interface SourceEditProps {
	fromJobFlow?: boolean
	stepNumber?: string | number
	stepTitle?: string
	initialData?: any
}

const SourceEdit: React.FC<SourceEditProps> = ({
	fromJobFlow = false,
	stepNumber,
	stepTitle,
	initialData,
}) => {
	const { sourceId } = useParams<{ sourceId: string }>()
	const navigate = useNavigate()
	const [activeTab, setActiveTab] = useState("config")
	const [connector, setConnector] = useState<string | null>(null)
	const [selectedVersion, setSelectedVersion] = useState("")
	const [sourceName, setSourceName] = useState("")
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [showAllJobs, setShowAllJobs] = useState(false)
	const [formData, setFormData] = useState<any>({})
	const { setShowDeleteModal, setSelectedSource } = useAppStore()
	const [mockAssociatedJobs] = useState<any[]>([])
	const [source, setSource] = useState<any>(null)
	const [loading, setLoading] = useState(false)
	const [schema, setSchema] = useState<any>(null)

	// Add a state for tracking paused job IDs
	const [pausedJobIds, setPausedJobIds] = useState<string[]>([])

	const {
		sources,
		jobs,
		fetchSources,
		fetchJobs,
		updateSource,
		setShowEditSourceModal,
	} = useAppStore()

	useEffect(() => {
		if (!sources.length) {
			fetchSources()
		}

		if (sourceId) {
			const source = sources.find(s => s.id.toString() === sourceId)
			if (source) {
				setSource(source)
				setSourceName(source.name)
				let normalizedType = source.type
				if (source.type.toLowerCase() === "mongodb") normalizedType = "MongoDB"
				if (source.type.toLowerCase() === "postgres")
					normalizedType = "Postgres"
				if (source.type.toLowerCase() === "mysql") normalizedType = "MySQL"
				setConnector(normalizedType)
				setSelectedVersion(source.version)
				setFormData(source.config)
			} else {
				message.error("Source not found")
				navigate("/sources")
			}
		}
	}, [sourceId, sources, fetchSources, jobs.length, fetchJobs, navigate])

	useEffect(() => {
		if (initialData) {
			setSourceName(initialData.name || "")
			let normalizedType = initialData.type
			if (initialData.type.toLowerCase() === "mongodb")
				normalizedType = "MongoDB"
			if (initialData.type.toLowerCase() === "postgres")
				normalizedType = "Postgres"
			if (initialData.type.toLowerCase() === "mysql") normalizedType = "MySQL"
			setConnector(normalizedType)
			setSelectedVersion(initialData.version || "latest")

			// Set form data from initialData
			if (initialData.config) {
				if (typeof initialData.config === "string") {
					try {
						setFormData(JSON.parse(initialData.config))
					} catch (error) {
						console.error("Error parsing source config:", error)
						setFormData(initialData.config)
					}
				} else {
					setFormData(initialData.config)
				}
			}
		}
	}, [initialData])

	useEffect(() => {
		const fetchSourceSpec = async () => {
			try {
				setLoading(true)
				const response = await sourceService.getSourceSpec(
					connector as string,
					selectedVersion,
				)
				if (response.success && response.data?.spec) {
					setSchema(response.data.spec)
				} else {
					console.error("Failed to get source spec:", response.message)
				}
			} catch (error) {
				console.error("Error fetching source spec:", error)
			} finally {
				setLoading(false)
			}
		}

		if (connector) {
			fetchSourceSpec()
		}

		return () => {
			// Cleanup function to prevent memory leaks
			setLoading(false)
		}
	}, [connector, selectedVersion])

	// Mock associated jobs for the source
	const associatedJobs = jobs.slice(0, 5).map(job => ({
		...job,
		state: "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		destination: {
			name: "AWS S3 Data Lake",
			type: "Amazon S3",
			paused: false,
			config: {
				s3_bucket: "prod-data-lake",
				s3_region: "us-west-2",
				writer: "parquet",
			},
		},
		paused: false,
	}))

	// Additional jobs that will be shown when "View all" is clicked
	const additionalJobs = jobs.slice(5, 10).map(job => ({
		...job,
		state: "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		destination: {
			name: "AWS S3 Warehouse",
			type: "AWS Glue Catalog",
			paused: false,
			config: {
				database: "analytics_db",
				region: "us-west-2",
			},
		},
		paused: false,
	}))

	const displayedJobs = showAllJobs
		? [...associatedJobs, ...additionalJobs]
		: associatedJobs

	const handleSave = () => {
		if (mockAssociatedJobs.length > 0) {
			setShowEditSourceModal(true)
			return
		}

		saveSource()
	}

	const saveSource = () => {
		let configToSave = { ...formData }

		const sourceData = {
			name: sourceName,
			type: connector || "MongoDB",
			status: "active" as const,
			config: configToSave,
		}

		if (sourceId) {
			updateSource(sourceId, sourceData)
				.then(() => {
					message.success("Source updated successfully")
					navigate("/sources")
				})
				.catch(error => {
					message.error("Failed to update source")
					console.error(error)
				})
		}
	}

	const handleDelete = () => {
		const sourceToDelete = {
			id: sourceId || "",
			name: sourceName || "",
			type: connector,
			...formData,
			associatedJobs: mockAssociatedJobs,
		}
		setSelectedSource(sourceToDelete as any)
		setShowDeleteModal(true)
	}

	const handleTestConnection = () => {
		message.success("Connection test successful")
	}

	const handleViewAllJobs = () => {
		setShowAllJobs(true)
	}

	const handlePauseAllJobs = (checked: boolean) => {
		if (checked) {
			const allJobIds = displayedJobs.map(job => job.id.toString())
			setPausedJobIds(allJobIds)
		} else {
			setPausedJobIds([])
		}
		message.info(`${checked ? "Pausing" : "Resuming"} all jobs for this source`)
	}

	const handlePauseJob = (jobId: string, checked: boolean) => {
		// Update the pausedJobIds state to track which jobs are paused
		if (checked) {
			setPausedJobIds(prev => [...prev, jobId])
		} else {
			setPausedJobIds(prev => prev.filter(id => id !== jobId))
		}

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
			dataIndex: "activate",
			key: "activate",
			render: (activate: boolean) => (
				<span
					className={`rounded px-2 py-1 text-xs ${
						!activate
							? "bg-[#FFF1F0] text-[#F5222D]"
							: "bg-[#E6F4FF] text-[#0958D9]"
					}`}
				>
					{activate ? "Active" : "Inactive"}
				</span>
			),
		},
		{
			title: "Last runtime",
			dataIndex: "last_run_time",
			render: (text: string) =>
				formatDistanceToNow(new Date(text), { addSuffix: true }),
		},
		{
			title: "Last runtime status",
			dataIndex: "last_run_state",
			key: "last_run_state",
			render: (last_run_state: string) => (
				<div
					className={`flex w-fit items-center justify-center gap-1 rounded-[6px] px-4 py-1 ${
						last_run_state === "success"
							? "bg-[#f6ffed] text-[#389E0D]"
							: last_run_state === "failed"
								? "bg-[#fff1f0] text-[#cf1322]"
								: ""
					}`}
				>
					{last_run_state}
				</div>
			),
		},
		{
			title: "Destination",
			dataIndex: "dest_name",
			key: "dest_name",
			render: (dest_name: string, record: any) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(record.dest_type || "")}
						alt={record.dest_type || ""}
						className="mr-2 size-6"
					/>
					{dest_name}
				</div>
			),
		},
		{
			title: "Running status",
			dataIndex: "id",
			key: "pause",
			render: (activate: boolean, record: any) => (
				<Switch
					checked={activate}
					onChange={checked => handlePauseJob(record.id.toString(), !checked)}
					className={
						!pausedJobIds.includes(record.id.toString())
							? "bg-blue-600"
							: "bg-gray-200"
					}
				/>
			),
		},
	]

	return (
		<div className={`flex h-screen flex-col ${fromJobFlow ? "pb-32" : ""}`}>
			{/* Header */}
			{!fromJobFlow && (
				<div className="flex items-center gap-2 px-6 pb-0 pt-6">
					<Link
						to="/sources"
						className="flex items-center gap-2 p-1.5 hover:rounded-[6px] hover:bg-[#f6f6f6] hover:text-black"
					>
						<ArrowLeft className="size-5" />
					</Link>

					<div className="flex items-center">
						<h1 className="text-2xl font-bold">{sourceName}</h1>
					</div>
				</div>
			)}

			{/* Main content */}
			<div className="mt-2 flex flex-1 overflow-hidden border border-t border-[#D9D9D9]">
				{/* Left content */}
				<div
					className={`${
						docsMinimized ? "w-full" : "w-3/4"
					} overflow-auto p-6 pt-4 transition-all duration-300`}
				>
					{fromJobFlow && stepNumber && stepTitle && (
						<div>
							<StepTitle
								stepNumber={stepNumber}
								stepTitle={stepTitle}
							/>
						</div>
					)}

					{!fromJobFlow && (
						<div className="mb-4">
							<div className="flex w-fit rounded-[6px] bg-[#f5f5f5] p-1">
								<button
									className={`w-56 rounded-[6px] px-3 py-1.5 text-sm font-normal ${
										activeTab === "config"
											? "mr-1 bg-[#203fdd] text-center text-[#F0F0F0]"
											: "mr-1 bg-[#F5F5F5] text-center text-[#0A0A0A]"
									}`}
									onClick={() => setActiveTab("config")}
								>
									Config
								</button>

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
							</div>
						</div>
					)}

					{activeTab === "config" ? (
						<div className="bg-white">
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
											<Select
												value={connector}
												onChange={value => {
													setConnector(value)
												}}
												className="h-8 w-full"
												options={[
													{
														value: "MongoDB",
														label: (
															<div className="flex items-center">
																<img
																	src={getConnectorImage("MongoDB")}
																	alt="MongoDB"
																	className="mr-2 size-5"
																/>
																<span>MongoDB</span>
															</div>
														),
													},
													{
														value: "Postgres",
														label: (
															<div className="flex items-center">
																<img
																	src={getConnectorImage("Postgres")}
																	alt="Postgres"
																	className="mr-2 size-5"
																/>
																<span>Postgres</span>
															</div>
														),
													},
													{
														value: "MySQL",
														label: (
															<div className="flex items-center">
																<img
																	src={getConnectorImage("MySQL")}
																	alt="MySQL"
																	className="mr-2 size-5"
																/>
																<span>MySQL</span>
															</div>
														),
													},
												]}
											/>
										</div>
									</div>

									<div>
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Name of your source:
											<span className="text-red-500">*</span>
										</label>
										<Input
											placeholder="Enter the name of your source"
											value={sourceName}
											onChange={e => setSourceName(e.target.value)}
											className="h-8"
										/>
									</div>
								</div>
							</div>

							<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
								<div className="mb-2 flex items-center gap-1">
									<GenderNeuter className="size-6" />
									<div className="text-lg font-medium">Endpoint config</div>
								</div>
								{loading ? (
									<div className="flex h-32 items-center justify-center">
										<Spin tip="Loading schema..." />
									</div>
								) : (
									<FixedSchemaForm
										schema={schema}
										formData={formData}
										onChange={setFormData}
										hideSubmit={true}
									/>
								)}
							</div>
						</div>
					) : (
						<div className="rounded-lg p-6">
							<h3 className="mb-4 text-lg font-medium">Associated jobs</h3>

							<Table
								columns={columns}
								dataSource={source?.jobs}
								pagination={false}
								rowKey={record => record.id}
								className="min-w-full"
								rowClassName={() => "custom-row"}
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

				<DocumentationPanel
					docUrl={`https://olake.io/docs/connectors/${connector?.toLowerCase()}/config`}
					isMinimized={docsMinimized}
					onToggle={toggleDocsPanel}
					showResizer={true}
				/>
			</div>

			<EditSourceModal />

			{/* Delete Modal */}
			<DeleteModal fromSource={true} />

			{/* Footer */}
			{!fromJobFlow && (
				<div className="flex justify-between border-t border-gray-200 bg-white p-4">
					<div>
						<button
							className="rounded-[6px] border border-[#F5222D] px-4 py-1 text-[#F5222D] hover:bg-[#F5222D] hover:text-white"
							onClick={handleDelete}
						>
							Delete
						</button>
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
							onClick={handleSave}
						>
							Save changes
						</button>
					</div>
				</div>
			)}
		</div>
	)
}

export default SourceEdit
