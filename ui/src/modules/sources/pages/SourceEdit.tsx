import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Select, Switch, message, Table } from "antd"
import { Check, GenderNeuter, Notebook } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { ArrowLeft } from "@phosphor-icons/react"
import type { ColumnsType } from "antd/es/table"
import { SourceJob } from "../../../types"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import DynamicSchemaForm from "../../common/components/DynamicSchemaForm"
import type { RJSFSchema } from "@rjsf/utils"
import StepTitle from "../../common/components/StepTitle"
import DeleteModal from "../../common/Modals/DeleteModal"
import { getConnectorImage } from "../../../utils/utils"
import EditSourceModal from "../../common/Modals/EditSourceModal"

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
	const isNewSource = sourceId === "new"
	const [activeTab, setActiveTab] = useState("config")
	const [connector, setConnector] = useState("MongoDB")
	const [sourceName, setSourceName] = useState("")
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [showAllJobs, setShowAllJobs] = useState(false)
	const [formData, setFormData] = useState<any>({})
	const { setShowDeleteModal, setSelectedSource } = useAppStore()
	const [mockAssociatedJobs, setMockAssociatedJobs] = useState<any[]>([])

	// Add a state for tracking paused job IDs
	const [pausedJobIds, setPausedJobIds] = useState<string[]>([])

	// Mock data for each connector type
	const mockData = {
		MongoDB: {
			hosts: ["localhost:27017"],
			username: "admin",
			password: "password",
			authdb: "admin",
			"replica-set": "rs01",
			"read-preference": "secondaryPreferred",
			srv: false,
			"server-ram": 16,
			database: "test_db",
			collection: "test_collection",
			max_threads: 50,
			default_mode: "cdc",
			backoff_retry_count: 3,
			partition_strategy: "",
		},
		Postgres: {
			host: "localhost",
			port: 5432,
			database: "test_db",
			username: "postgres",
			password: "password",
			jdbc_url_params: "connectTimeout=30",
			schema: "public",
			table: "test_table",
			ssl: {
				mode: "disable",
				client_cert: "",
				client_key: "",
				server_ca: "",
			},
			update_method: {
				replication_slot: "test_slot",
				intial_wait_time: 10,
			},
		},
		MySQL: {
			hosts: "localhost",
			port: 3306,
			database: "test_db",
			username: "root",
			password: "password",
			table: "test_table",
			tls_skip_verify: true,
			update_method: {
				intial_wait_time: 10,
			},
			server_id: 1,
			binlog_position: "mysql-bin.000001:1234",
			max_threads: 10,
			backoff_retry_count: 2,
			default_mode: "cdc",
		},
	}

	// Schema definitions for each connector
	const schemas = {
		MongoDB: {
			type: "object",
			properties: {
				hosts: {
					type: "array",
					title: "MongoDB Hosts",
					description: "List of MongoDB hosts. Use DNS SRV if srv = true.",
					items: {
						type: "string",
						title: null,
						pattern: "^[^\\s]+:\\d+$",
						errorMessage: {
							pattern: 'must match pattern "^[^\\s]+:\\d+$"',
						},
					},
					minItems: 1,
					uniqueItems: true,
					default: ["localhost:27017"],
				},
				username: {
					type: "string",
					title: "Username",
					description: "Credentials for MongoDB authentication",
				},
				password: {
					type: "string",
					title: "Password",
					description: "Credentials for MongoDB authentication",
					format: "password",
				},
				authdb: {
					type: "string",
					title: "Auth Database",
					description: "Authentication database (often admin)",
					default: "admin",
				},
				"replica-set": {
					type: "string",
					title: "Replica Set",
					description: "Name of the replica set, if applicable",
				},
				"read-preference": {
					type: "string",
					title: "Read Preference",
					description: "Which node to read from (e.g., secondaryPreferred)",
					enum: [
						"primary",
						"primaryPreferred",
						"secondary",
						"secondaryPreferred",
						"nearest",
					],
					default: "secondaryPreferred",
				},
				srv: {
					type: "boolean",
					title: "SRV",
					description:
						"If using DNS SRV connection strings, set to true. When true, there can only be 1 host in hosts field.",
					default: false,
				},
				"server-ram": {
					type: "integer",
					title: "Server RAM",
					description: "Memory management hint for the OLake container",
					default: 16,
				},
				database: {
					type: "string",
					title: "Database",
					description: "The MongoDB database name to replicate",
				},
				max_threads: {
					type: "integer",
					title: "Max Threads",
					description: "Maximum parallel threads for chunk-based snapshotting",
					default: 50,
				},
				default_mode: {
					type: "string",
					title: "Default Mode",
					description: "Default sync mode",
					enum: ["cdc", "full_refresh", "incremental"],
					default: "cdc",
				},
				backoff_retry_count: {
					type: "integer",
					title: "Backoff Retry Count",
					description:
						"Retries attempt to establish sync again if it fails, increases exponentially (in minutes - 1, 2,4,8,16... depending upon the backoff_retry_count value)",
					default: 3,
				},
				partition_strategy: {
					type: "string",
					title: "Partition Strategy",
					description: "The partition strategy for backfill",
					enum: ["timestamp", ""],
					default: "",
				},
			},
			required: ["hosts", "username", "password", "database", "collection"],
		},
		Postgres: {
			type: "object",
			properties: {
				host: {
					type: "string",
					title: "Host",
					description: "Postgres server host",
				},
				port: {
					type: "integer",
					title: "Port",
					description: "Postgres server port",
					default: 5432,
				},
				database: {
					type: "string",
					title: "Database",
					description: "Name of the database to connect to",
				},
				username: {
					type: "string",
					title: "Username",
					description: "Database user credentials",
				},
				password: {
					type: "string",
					title: "Password",
					description: "Database user password",
					format: "password",
				},
				jdbc_url_params: {
					type: "string",
					title: "JDBC URL Parameters",
					description: "Additional connection parameters",
					default: "",
				},
				mode: {
					type: "string",
					title: "SSL Mode",
					description: "SSL connection mode",
					enum: ["disable", "require", "verify-ca", "verify-full"],
					default: "disable",
				},
				client_cert: {
					type: "string",
					title: "Client Certificate",
					description: "Path to client certificate file",
				},
				client_key: {
					type: "string",
					title: "Client Key",
					description: "Path to client key file",
				},
				server_ca: {
					type: "string",
					title: "Server CA",
					description: "Path to server CA certificate",
				},
				schema: {
					type: "string",
					title: "Schema",
					description: "Database schema to use",
					default: "public",
				},
				table: {
					type: "string",
					title: "Table",
					description: "Table to sync",
				},
				replication_slot: {
					type: "string",
					title: "Replication Slot",
					description: "Name of the replication slot",
				},
				intial_wait_time: {
					type: "integer",
					title: "Initial Wait Time",
					description: "Time to wait before starting replication (seconds)",
					default: 10,
				},
				publication: {
					type: "string",
					title: "Publication",
					description: "Name of the publication",
				},
			},
			required: [
				"host",
				"port",
				"database",
				"username",
				"password",
				"schema",
				"table",
			],
		},
		MySQL: {
			type: "object",
			properties: {
				hosts: {
					type: "string",
					title: "Hosts",
					description: "MySQL server host",
				},
				port: {
					type: "integer",
					title: "Port",
					description: "MySQL server port",
					default: 3306,
				},
				database: {
					type: "string",
					title: "Database",
					description: "Name of the database to connect to",
				},
				username: {
					type: "string",
					title: "Username",
					description: "Database user credentials",
				},
				password: {
					type: "string",
					title: "Password",
					description: "Database user password",
					format: "password",
				},
				table: {
					type: "string",
					title: "Table",
					description: "Table to sync",
				},
				tls_skip_verify: {
					type: "boolean",
					title: "Skip TLS Verification",
					description: "Skip verification of TLS certificate",
					default: true,
				},
				intial_wait_time: {
					type: "integer",
					title: "Initial Wait Time",
					description: "Time to wait before starting replication (seconds)",
					default: 10,
				},
				server_id: {
					type: "integer",
					title: "Server ID",
					description: "MySQL server ID for replication",
				},
				binlog_position: {
					type: "string",
					title: "Binlog Position",
					description: "Position in binary log to start reading from",
				},
				max_threads: {
					type: "integer",
					title: "Max Threads",
					description: "Maximum number of threads for parallel operations",
					default: 10,
				},
				backoff_retry_count: {
					type: "integer",
					title: "Backoff Retry Count",
					description: "Number of retries before giving up",
					default: 2,
				},
				default_mode: {
					type: "string",
					title: "Default Mode",
					description: "Default synchronization mode",
					enum: ["cdc", "incremental", "full_refresh"],
					default: "cdc",
				},
			},
			required: ["hosts", "port", "database", "username", "password", "table"],
		},
	}

	// UI Schema for better form layout
	const uiSchema = {
		"ui:className":
			"mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm",
		password: {
			"ui:widget": "password",
		},
		hosts: {
			"ui:options": {
				orderable: false,
				label: false,
			},
			items: {
				"ui:title": "",
				"ui:placeholder": "Enter host:port (e.g. localhost:27017)",
			},
		},
		srv: {
			"ui:widget": "checkbox",
		},
		tls_skip_verify: {
			"ui:widget": "checkbox",
		},
		"read-preference": {
			"ui:widget": "select",
		},
		default_mode: {
			"ui:widget": "select",
		},
		partition_strategy: {
			"ui:widget": "select",
		},
		ssl: {
			"ui:options": {
				className: "grid grid-cols-1 gap-4",
			},
		},
		update_method: {
			"ui:options": {
				className: "grid grid-cols-1 gap-4",
			},
		},
		intial_wait_time: {
			"ui:options": {
				className: "grid grid-cols-1 gap-4",
			},
		},
		replication_slot: {
			"ui:options": {
				className: "grid grid-cols-1 gap-4",
			},
		},
	}

	// Connector-specific UI order
	const uiOrderByConnector = {
		MongoDB: [
			"hosts",
			"username",
			"password",
			"authdb",
			"replica-set",
			"read-preference",
			"srv",
			"server-ram",
			"database",
			"collection",
			"max_threads",
			"default_mode",
			"backoff_retry_count",
			"partition_strategy",
		],
		Postgres: [
			"host",
			"port",
			"database",
			"username",
			"password",
			"jdbc_url_params",
			"ssl",
			"schema",
			"table",
			"update_method",
			"publication",
			"replication_slot",
			"intial_wait_time",
			"mode",
			"client_cert",
			"client_key",
			"server_ca",
		],
		MySQL: [
			"hosts",
			"port",
			"database",
			"username",
			"password",
			"table",
			"tls_skip_verify",
			"update_method",
			"server_id",
			"binlog_position",
			"max_threads",
			"backoff_retry_count",
			"default_mode",
			"intial_wait_time",
		],
	}

	const {
		sources,
		jobs,
		fetchSources,
		fetchJobs,
		addSource,
		updateSource,
		setShowEditSourceModal,
	} = useAppStore()

	// Load source data when editing
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

				// Normalize connector type to match schema keys
				// This addresses case sensitivity issues
				let normalizedType = source.type
				if (source.type.toLowerCase() === "mongodb") normalizedType = "MongoDB"
				if (source.type.toLowerCase() === "postgres")
					normalizedType = "Postgres"
				if (source.type.toLowerCase() === "mysql") normalizedType = "MySQL"

				setConnector(normalizedType)
				setMockAssociatedJobs(source?.associatedJobs || [])

				// Use the actual config data from the source when editing
				if (source.config) {
					setFormData(source.config)
				} else {
					const mockFormData =
						mockData[normalizedType as keyof typeof mockData] || {}
					setFormData(mockFormData)
				}
			} else {
				// If source isn't found, redirect to sources page or show error
				message.error("Source not found")
				navigate("/sources")
			}
		}
	}, [
		sourceId,
		isNewSource,
		sources,
		fetchSources,
		jobs.length,
		fetchJobs,
		navigate,
	])

	// Load initial data when provided (for job edit flow)
	useEffect(() => {
		if (initialData) {
			setSourceName(initialData.name || "")

			// Make sure to set the connector before setting form data
			if (initialData.type) {
				setConnector(initialData.type)
			}

			// If we have config data, set it to form data
			if (initialData.config) {
				// Set a timeout to ensure the connector is set first
				setTimeout(() => {
					setFormData(initialData.config)
				}, 100)
			}
		}
	}, [initialData])

	// Use a more direct approach to update form data when connector changes
	useEffect(() => {
		// If initialData is present, update form data again
		if (initialData?.config && Object.keys(initialData.config).length > 0) {
			setFormData(initialData.config)
		} else if (!isNewSource && !initialData) {
			// When editing an existing source and connector changes,
			// load the appropriate mock data if actual data not available
			const mockFormData = mockData[connector as keyof typeof mockData] || {}
			setFormData(mockFormData)
		}
	}, [connector, initialData, isNewSource])

	// For new sources, set initial mock data based on the default connector
	useEffect(() => {
		if (isNewSource && Object.keys(formData).length === 0) {
			const mockFormData = mockData[connector as keyof typeof mockData] || {}
			setFormData(mockFormData)
		}
	}, [isNewSource, connector, formData])

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
		if (!isNewSource && mockAssociatedJobs.length > 0) {
			// Show warning modal if editing an existing source with jobs
			setShowEditSourceModal(true)
			return
		}

		saveSource()
	}

	const saveSource = () => {
		// Ensure we have the proper structure before saving
		let configToSave = { ...formData }

		const sourceData = {
			name:
				sourceName || `${connector}_source_${Math.floor(Math.random() * 1000)}`,
			type: connector,
			status: "active" as const,
			config: configToSave,
		}

		if (isNewSource) {
			addSource(sourceData)
				.then(() => {
					message.success("Source created successfully")
					// Show the entity saved modal
					useAppStore.getState().setShowEntitySavedModal(true)
					navigate("/sources")
				})
				.catch(error => {
					message.error("Failed to create source")
					console.error(error)
				})
		} else if (sourceId) {
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
		// Create source object from the current data
		const sourceToDelete = {
			id: sourceId || "",
			name: sourceName || "",
			type: connector,
			...formData,
			associatedJobs: mockAssociatedJobs,
		}
		// Set the current source as selected in the store
		setSelectedSource(sourceToDelete as any)
		// Show the delete modal
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
			// Pause all jobs - add all job IDs to pausedJobIds
			const allJobIds = displayedJobs.map(job => job.id)
			setPausedJobIds(allJobIds)
		} else {
			// Resume all jobs - clear pausedJobIds
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
			dataIndex: "id",
			key: "state",
			render: (jobId: string) => (
				<span
					className={`rounded px-2 py-1 text-xs ${
						pausedJobIds.includes(jobId)
							? "bg-[#FFF1F0] text-[#F5222D]"
							: "bg-[#E6F4FF] text-[#0958D9]"
					}`}
				>
					{pausedJobIds.includes(jobId) ? "Inactive" : "Active"}
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
			render: (destination: any) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(destination.type)}
						alt={destination.type}
						className="mr-2 size-6"
					/>
					{destination.name}
				</div>
			),
		},
		{
			title: "Running status",
			dataIndex: "id",
			key: "pause",
			render: (jobId: string) => (
				<Switch
					checked={!pausedJobIds.includes(jobId)}
					onChange={checked => handlePauseJob(jobId, !checked)}
					className={
						!pausedJobIds.includes(jobId) ? "bg-blue-600" : "bg-gray-200"
					}
				/>
			),
		},
	]

	return (
		<div className="flex h-screen flex-col">
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
						<h1 className="text-2xl font-bold">
							{isNewSource
								? "Create New Source"
								: sourceName || "MongoDB_Source_1"}
						</h1>
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
													// Load mock data for the selected connector if not editing existing
													if (isNewSource) {
														const mockFormData =
															mockData[value as keyof typeof mockData] || {}
														setFormData(mockFormData)
													}
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

								<DynamicSchemaForm
									schema={
										schemas[connector as keyof typeof schemas] as RJSFSchema
									}
									uiSchema={{
										...uiSchema,
										"ui:order":
											uiOrderByConnector[
												connector as keyof typeof uiOrderByConnector
											],
									}}
									formData={formData}
									onChange={setFormData}
									hideSubmit={true}
								/>
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
					docUrl={`https://olake.io/docs/connectors/${connector.toLowerCase()}/config`}
					isMinimized={docsMinimized}
					onToggle={toggleDocsPanel}
					showResizer={true}
				/>
			</div>

			<EditSourceModal mockAssociatedJobs={mockAssociatedJobs} />

			{/* Delete Modal */}
			<DeleteModal fromSource={true} />

			{/* Footer */}
			{!fromJobFlow && (
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
							onClick={handleSave}
						>
							{isNewSource ? "Create source" : "Save changes"}
						</button>
					</div>
				</div>
			)}
		</div>
	)
}

export default SourceEdit
