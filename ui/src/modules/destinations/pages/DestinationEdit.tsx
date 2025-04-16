import React, { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Select, Switch, message } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, Check, Notebook } from "@phosphor-icons/react"
import Table from "antd/es/table"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import DynamicSchemaForm from "../../common/components/DynamicSchemaForm"
import { destinationService } from "../../../api/services/destinationService"
import { RJSFSchema, UiSchema } from "@rjsf/utils"
import StepTitle from "../../common/components/StepTitle"
import DeleteModal from "../../common/Modals/DeleteModal"

interface DestinationEditProps {
	fromJobFlow?: boolean
	stepNumber?: string | number
	stepTitle?: string
	initialData?: any
}

const DestinationEdit: React.FC<DestinationEditProps> = ({
	fromJobFlow = false,
	stepNumber,
	stepTitle,
	initialData,
}) => {
	const { destinationId } = useParams<{ destinationId: string }>()
	const navigate = useNavigate()
	const isNewDestination = destinationId === "new"
	const [activeTab, setActiveTab] = useState("config")
	const [connector, setConnector] = useState("AWS S3")
	const [catalog, setCatalog] = useState<string | null>(null)
	const [destinationName, setDestinationName] = useState("")
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [showAllJobs, setShowAllJobs] = useState(false)
	const [formData, setFormData] = useState<any>({})
	const [connectorSchema, setConnectorSchema] = useState<RJSFSchema>({})
	const [connectorUiSchema] = useState<UiSchema>({})
	const [isLoading, setIsLoading] = useState(false)
	const [mockAssociatedJobs, setMockAssociatedJobs] = useState<any[]>([])

	// Mock data for each destination type
	const mockData = {
		"Amazon S3": {
			type: "PARQUET",
			writer: {
				normalization: false,
				s3_bucket: "my-test-bucket",
				s3_region: "ap-south-1",
				s3_access_key: "AKIAXXXXXXXXXXXXXXXX",
				s3_secret_key: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				s3_path: "/data/test",
			},
		},
		"Apache Iceberg": {
			type: "ICEBERG",
			writer: {
				catalog_type: "glue",
				normalization: false,
				iceberg_s3_path: "s3://my-bucket/olake_iceberg/test",
				aws_region: "ap-south-1",
				aws_access_key: "AKIAXXXXXXXXXXXXXXXX",
				aws_secret_key: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				iceberg_db: "my_database",
				grpc_port: 50051,
				server_host: "localhost",
			},
		},
		"AWS Glue Catalog": {
			type: "ICEBERG",
			writer: {
				catalog_type: "glue",
				normalization: false,
				iceberg_s3_path: "s3://my-bucket/olake_iceberg/test",
				aws_region: "ap-south-1",
				aws_access_key: "AKIAXXXXXXXXXXXXXXXX",
				aws_secret_key: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
				iceberg_db: "my_database",
				grpc_port: 50051,
				server_host: "localhost",
			},
		},
		"REST Catalog": {
			type: "ICEBERG",
			writer: {
				catalog_type: "rest",
				normalization: false,
				rest_catalog_url: "http://localhost:8181/catalog",
				iceberg_s3_path: "warehouse",
				iceberg_db: "olake_iceberg",
			},
		},
		"JDBC Catalog": {
			type: "ICEBERG",
			writer: {
				catalog_type: "jdbc",
				jdbc_url: "jdbc:postgresql://localhost:5432/iceberg",
				jdbc_username: "admin",
				jdbc_password: "password",
				normalization: false,
				iceberg_s3_path: "s3a://warehouse",
				s3_endpoint: "http://localhost:9000",
				s3_use_ssl: false,
				s3_path_style: true,
				aws_access_key: "minioadmin",
				aws_region: "ap-south-1",
				aws_secret_key: "minioadmin",
				iceberg_db: "olake_iceberg",
			},
		},
	}

	const {
		destinations,
		jobs,
		fetchDestinations,
		fetchJobs,
		setSelectedDestination,
		setShowDeleteModal,
		// addDestination,
		// updateDestination,
	} = useAppStore()

	// Fetch connector schema based on selected connector type
	const fetchConnectorSchema = async (connectorType: string) => {
		setIsLoading(true)
		try {
			let schemaData
			if (connector === "Apache Iceberg") {
				let connectorFromCatalog
				if (catalog === null) {
					connectorFromCatalog = "AWS Glue"
				} else {
					connectorFromCatalog = catalog
				}
				schemaData =
					await destinationService.getConnectorSchema(connectorFromCatalog)
				setConnectorSchema(schemaData as RJSFSchema)

				// Set mock data for the catalog after schema is loaded
				if (mockData[connectorFromCatalog as keyof typeof mockData]) {
					setFormData(mockData[connectorFromCatalog as keyof typeof mockData])
				}
			} else {
				schemaData = await destinationService.getConnectorSchema(connector)
				setConnectorSchema(schemaData as RJSFSchema)

				// Set mock data for the connector after schema is loaded
				if (mockData[connector as keyof typeof mockData]) {
					setFormData(mockData[connector as keyof typeof mockData])
				}
			}
			// const schema = await destinationService.getConnectorSchema(connectorType)
			// if (schema) {
			// 	setConnectorSchema(schema as RJSFSchema)
			// 	if (schema.uiSchema) {
			// 		setConnectorUiSchema(schema.uiSchema)
			// 	}
			// }
		} catch (error) {
			console.error("Error fetching connector schema:", error)
			message.error(`Failed to load schema for ${connectorType}`)
		} finally {
			setIsLoading(false)
		}
	}

	// Load initial data when provided (for job edit flow)
	useEffect(() => {
		if (initialData) {
			setDestinationName(initialData.name || "")

			if (initialData.type) {
				setConnector(initialData.type)
			}

			if (initialData.config) {
				setTimeout(() => {
					setFormData(initialData.config)
				}, 100)
			}
		}
	}, [initialData])

	useEffect(() => {
		if (initialData?.config && Object.keys(initialData.config).length > 0) {
			setFormData(initialData.config)
		} else if (connector && mockData[connector as keyof typeof mockData]) {
			// Fill in mock data when connector changes and there's no initialData
			setFormData(mockData[connector as keyof typeof mockData] || {})
		}
	}, [connector, initialData])

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
				setMockAssociatedJobs(destination?.associatedJobs || [])
				setCatalog(destination?.catalog || null)
				// Set mock data based on connector type
				setFormData(mockData[destination.type as keyof typeof mockData] || {})
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

	useEffect(() => {
		fetchConnectorSchema(connector)
	}, [connector])

	// Mock associated jobs for the destination
	const associatedJobs = jobs.slice(0, 5).map(job => ({
		...job,
		state: Math.random() > 0.7 ? "Inactive" : "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		source: job.source,
		paused: false,
	}))

	// Additional jobs that will be shown when "View all" is clicked
	const additionalJobs = jobs.slice(5, 10).map(job => ({
		...job,
		state: Math.random() > 0.7 ? "Inactive" : "Active",
		lastRuntime: "3 hours ago",
		lastRuntimeStatus: "Success",
		source: job.source,
		paused: false,
	}))

	const displayedJobs = showAllJobs
		? [...associatedJobs, ...additionalJobs]
		: associatedJobs

	const handleDelete = () => {
		// Create destination object from the current data
		const destinationToDelete = {
			id: destinationId || "",
			name: destinationName || "",
			type: connector,
			...formData,
			associatedJobs: mockAssociatedJobs,
		}
		setSelectedDestination(destinationToDelete as any)
		// Show the delete modal
		setShowDeleteModal(true)
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
						<span>S</span>
					</div>
					{source}
				</div>
			),
		},
		{
			title: "Pause job",
			dataIndex: "id",
			key: "pause",
			render: (_: string, record: any) => (
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
			{!fromJobFlow && (
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
			)}

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden border border-t border-[#D9D9D9]">
				{/* Left content */}
				<div
					className={`${
						docsMinimized ? "w-full" : "w-3/4"
					} mt-4 overflow-auto p-6 pt-0 transition-all duration-300`}
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
							<div className="flex">
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
								{!isNewDestination && (
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
					)}

					{activeTab === "config" ? (
						<div className="rounded-lg">
							<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
								<div className="mb-4 flex items-center gap-1 text-lg font-medium">
									<Notebook className="size-5" />
									Capture information
								</div>

								<div className="flex flex-col gap-6">
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
														if (value === "Amazon S3") {
															setFormData({
																...formData,
																type: "PARQUET",
															})
														} else if (
															mockData[value as keyof typeof mockData]
														) {
															setFormData(
																mockData[value as keyof typeof mockData] || {},
															)
														}
													}}
													className="w-full"
													options={[
														{ value: "Amazon S3", label: "Amazon S3" },
														{
															value: "Apache Iceberg",
															label: "Apache Iceberg",
														},
													]}
												/>
											</div>
										</div>

										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Catalog:
											</label>
											<Select
												className="w-full"
												placeholder="Select catalog"
												disabled={
													connector === "Amazon S3" || connector === "AWS S3"
												}
												options={[
													{
														value: "AWS Glue",
														label: "AWS Glue Catalog",
													},
													{ value: "REST Catalog", label: "REST Catalog" },
													{ value: "JDBC Catalog", label: "JDBC Catalog" },
												]}
												value={
													catalog ||
													(connector === "Amazon S3" || connector === "AWS S3"
														? "None"
														: undefined)
												}
												onChange={value => {
													setFormData({
														...formData,
														type: "	Apache Iceberg",
														catalog: value,
													})
													// Also apply mock data for the selected catalog
													if (mockData[value as keyof typeof mockData]) {
														setFormData(
															mockData[value as keyof typeof mockData],
														)
													}
												}}
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
								{isLoading ? (
									<div className="py-8 text-center">
										Loading configuration...
									</div>
								) : (
									Object.keys(connectorSchema).length > 0 && (
										<DynamicSchemaForm
											schema={connectorSchema}
											uiSchema={connectorUiSchema}
											formData={formData}
											onChange={setFormData}
											hideSubmit={true}
										/>
									)
								)}
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

				{/* Documentation panel */}
				<DocumentationPanel
					docUrl="https://olake.io/docs/category/aws-s3"
					isMinimized={docsMinimized}
					onToggle={toggleDocsPanel}
					showResizer={true}
				/>
			</div>
			{/* Delete Modal */}
			<DeleteModal fromSource={false} />

			{/* Footer with buttons */}
			{!fromJobFlow && (
				<div className="flex justify-between border-t border-gray-200 bg-white p-4">
					<div>
						{!isNewDestination && (
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
							Use destination
						</button>
					</div>
				</div>
			)}
		</div>
	)
}

export default DestinationEdit
