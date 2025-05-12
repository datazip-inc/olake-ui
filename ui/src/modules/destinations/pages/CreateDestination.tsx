import { useState, useEffect } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Input, Radio, Select, Spin } from "antd"
import { useAppStore } from "../../../store"
import {
	ArrowLeft,
	ArrowRight,
	GenderNeuter,
	Notebook,
} from "@phosphor-icons/react"
import AWSS3Icon from "../../../assets/AWSS3.svg"
import ApacheIceBerg from "../../../assets/ApacheIceBerg.svg"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import EntitySavedModal from "../../common/Modals/EntitySavedModal"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import EntityCancelModal from "../../common/Modals/EntityCancelModal"
import StepTitle from "../../common/components/StepTitle"
import FixedSchemaForm from "../../../utils/FormFix"
import { destinationService } from "../../../api/services/destinationService"
import {
	getCatalogInLowerCase,
	getConnectorInLowerCase,
	getConnectorName,
} from "../../../utils/utils"

// Constants
const CONNECTOR_TYPES = {
	AMAZON_S3: "Amazon S3",
	APACHE_ICEBERG: "Apache Iceberg",
} as const

const CATALOG_TYPES = {
	AWS_GLUE: "AWS Glue",
	REST_CATALOG: "REST Catalog",
	JDBC_CATALOG: "JDBC Catalog",
	HIVE_CATALOG: "Hive Catalog",
	NONE: "None",
} as const

const SETUP_TYPES = {
	NEW: "new",
	EXISTING: "existing",
} as const

// Type aliases for our constants
type ConnectorType = (typeof CONNECTOR_TYPES)[keyof typeof CONNECTOR_TYPES]
type SetupType = (typeof SETUP_TYPES)[keyof typeof SETUP_TYPES]

// Types
interface DestinationConfig {
	[key: string]: any
	catalog?: string
	catalog_type?: string
	writer?: {
		catalog?: string
		catalog_type?: string
	}
}

interface Destination {
	id: string | number
	name: string
	type: string
	version: string
	config: string | DestinationConfig
}

// ExtendedDestination ensures config is always available
interface ExtendedDestination extends Destination {
	config: DestinationConfig
}

interface CreateDestinationProps {
	fromJobFlow?: boolean
	fromJobEditFlow?: boolean
	existingDestinationId?: string
	onComplete?: () => void
	stepNumber?: number
	stepTitle?: string
	initialConfig?: {
		name: string
		type: string
		config?: DestinationConfig
	}
	initialFormData?: DestinationConfig
	onDestinationNameChange?: (name: string) => void
	onConnectorChange?: (connector: string) => void
	onFormDataChange?: (formData: DestinationConfig) => void
	onVersionChange?: (version: string) => void
}

const CreateDestination: React.FC<CreateDestinationProps> = ({
	fromJobFlow = false,
	fromJobEditFlow = false,
	existingDestinationId,
	onComplete,
	stepNumber,
	stepTitle,
	initialConfig,
	initialFormData,
	onDestinationNameChange,
	onConnectorChange,
	onFormDataChange,
	onVersionChange,
}) => {
	const [setupType, setSetupType] = useState<SetupType>(SETUP_TYPES.NEW)
	const [connector, setConnector] = useState<ConnectorType>(
		CONNECTOR_TYPES.AMAZON_S3,
	)
	const [catalog, setCatalog] = useState<string | null>(null)
	const [destinationName, setDestinationName] = useState("")
	const [version, setVersion] = useState("")
	const [versions, setVersions] = useState<string[]>([])
	const [loadingVersions, setLoadingVersions] = useState(false)
	const [formData, setFormData] = useState<DestinationConfig>({})
	const [schema, setSchema] = useState<Record<string, any> | null>(null)
	const [loading, setLoading] = useState(false)
	const [uiSchema, setUiSchema] = useState<Record<string, any> | null>(null)
	const [filteredDestinations, setFilteredDestinations] = useState<
		ExtendedDestination[]
	>([])
	const navigate = useNavigate()

	const {
		destinations,
		fetchDestinations,
		setShowEntitySavedModal,
		setShowTestingModal,
		setShowSuccessModal,
		addDestination,
	} = useAppStore()

	// Fetch destinations if needed
	useEffect(() => {
		if (!destinations.length) {
			fetchDestinations()
		}
	}, [destinations.length, fetchDestinations])

	// Initialize with initial config if provided
	useEffect(() => {
		if (initialConfig) {
			setDestinationName(initialConfig.name)
			setConnector(initialConfig.type as ConnectorType)
			setFormData(initialConfig.config || {})
		}
	}, [initialConfig])

	// Update form data when initial form data changes
	useEffect(() => {
		if (initialFormData) {
			setFormData(initialFormData)
		}
	}, [initialFormData])

	// Handle edit flow initialization
	useEffect(() => {
		if (fromJobEditFlow && existingDestinationId) {
			setSetupType(SETUP_TYPES.EXISTING)
			const selectedDestination = destinations.find(
				d => d.id.toString() === existingDestinationId,
			) as ExtendedDestination | undefined

			if (selectedDestination) {
				setDestinationName(selectedDestination.name)
				setConnector(selectedDestination.type as ConnectorType)
			}
		}
	}, [fromJobEditFlow, existingDestinationId, destinations])

	// Set default catalog for Apache Iceberg
	useEffect(() => {
		if (connector === CONNECTOR_TYPES.APACHE_ICEBERG) {
			setCatalog(CATALOG_TYPES.AWS_GLUE)
		} else {
			setCatalog(null)
		}
	}, [connector])

	// Filter destinations based on connector and catalog
	useEffect(() => {
		if (setupType === SETUP_TYPES.EXISTING) {
			if (connector === CONNECTOR_TYPES.APACHE_ICEBERG) {
				const catalogValue = catalog || CATALOG_TYPES.AWS_GLUE
				const filtered = destinations
					.filter(destination => {
						if (destination.type !== getConnectorInLowerCase(connector))
							return false

						let config: DestinationConfig = {}
						try {
							if (typeof destination.config === "string") {
								config = JSON.parse(destination.config)
							} else {
								config = destination.config as DestinationConfig
							}

							return (
								config?.writer?.catalog ===
									getCatalogInLowerCase(catalogValue) ||
								config?.writer?.catalog_type ===
									getCatalogInLowerCase(catalogValue)
							)
						} catch {
							return false
						}
					})
					.map(dest => {
						// Transform destination to ensure config is a DestinationConfig object
						const config =
							typeof dest.config === "string"
								? JSON.parse(dest.config)
								: (dest.config as DestinationConfig)

						return {
							...dest,
							config,
						} as ExtendedDestination
					})

				setFilteredDestinations(filtered)
			} else {
				const filtered = destinations
					.filter(
						destination =>
							destination.type === getConnectorInLowerCase(connector),
					)
					.map(dest => {
						// Transform destination to ensure config is a DestinationConfig object
						const config =
							typeof dest.config === "string"
								? JSON.parse(dest.config)
								: (dest.config as DestinationConfig)

						return {
							...dest,
							config,
						} as ExtendedDestination
					})

				setFilteredDestinations(filtered)
			}
		}
	}, [connector, setupType, destinations, catalog])

	// Fetch connector versions
	useEffect(() => {
		const fetchVersions = async () => {
			setLoadingVersions(true)
			try {
				const response = await destinationService.getDestinationVersions(
					connector.toLowerCase(),
				)
				if (response.data && response.data.version) {
					setVersions(response.data.version)
					const defaultVersion = response.data.version[0] || ""
					setVersion(defaultVersion)

					if (onVersionChange) {
						onVersionChange(defaultVersion)
					}
				}
			} catch (error) {
				console.error("Error fetching versions:", error)
			} finally {
				setLoadingVersions(false)
			}
		}

		fetchVersions()
	}, [connector, onVersionChange])

	// Fetch destination spec
	useEffect(() => {
		const fetchDestinationSpec = async () => {
			setLoading(true)
			try {
				const response = await destinationService.getDestinationSpec(
					connector,
					catalog,
				)
				if (response.success && response.data?.spec) {
					setSchema(response.data.spec)
					if (response.data?.uiSchema) {
						setUiSchema(response.data.uiSchema)
					}
				} else {
					console.error("Failed to get destination spec:", response.message)
				}
			} catch (error) {
				console.error("Error fetching destination spec:", error)
			} finally {
				setLoading(false)
			}
		}

		fetchDestinationSpec()
	}, [connector, catalog, version])

	// Handlers
	const handleCancel = () => {
		setShowEntitySavedModal(false)
	}

	const handleCreate = async () => {
		const catalogInLowerCase = catalog
			? getCatalogInLowerCase(catalog)
			: undefined
		const newDestinationData = {
			name: destinationName,
			type: connector === CONNECTOR_TYPES.AMAZON_S3 ? "s3" : "iceberg",
			version,
			config: JSON.stringify({ ...formData, catalog: catalogInLowerCase }),
		}

		try {
			setShowTestingModal(true)
			const testResult =
				await destinationService.testDestinationConnection(newDestinationData)
			setShowTestingModal(false)

			if (testResult.success) {
				setShowSuccessModal(true)
				setTimeout(() => {
					setShowSuccessModal(false)
					addDestination(newDestinationData)
						.then(() => {
							setShowEntitySavedModal(true)
						})
						.catch(error => {
							console.error("Error adding destination:", error)
						})
				}, 1000)
			} else {
				console.error("Connection test failed:", testResult.message)
				navigate("/destinations")
			}
		} catch (error) {
			setShowTestingModal(false)
			console.error("Error testing connection:", error)
			navigate("/destinations")
		}
	}

	const handleDestinationNameChange = (
		e: React.ChangeEvent<HTMLInputElement>,
	) => {
		const newName = e.target.value
		setDestinationName(newName)
		if (onDestinationNameChange) {
			onDestinationNameChange(newName)
		}
	}

	const handleConnectorChange = (value: string) => {
		setConnector(value as ConnectorType)
		if (onConnectorChange) {
			onConnectorChange(value)
		}

		// Set default catalog for Apache Iceberg
		if (value === CONNECTOR_TYPES.APACHE_ICEBERG) {
			setCatalog(CATALOG_TYPES.AWS_GLUE)
		} else {
			setCatalog(null)
		}
	}

	const handleCatalogChange = (value: string) => {
		setCatalog(value)
	}

	const handleExistingDestinationSelect = (value: string) => {
		const selectedDestination = destinations.find(
			d => d.id.toString() === value.toString(),
		)

		if (!selectedDestination) return

		if (onDestinationNameChange) {
			onDestinationNameChange(selectedDestination.name)
		}

		if (onConnectorChange) {
			onConnectorChange(selectedDestination.type)
		}

		if (onVersionChange) {
			onVersionChange(selectedDestination.version)
		}

		// Parse the config properly if it's a string
		let configObj: DestinationConfig = {}
		try {
			if (typeof selectedDestination.config === "string") {
				configObj = JSON.parse(selectedDestination.config)
			} else if (
				selectedDestination.config &&
				typeof selectedDestination.config === "object"
			) {
				configObj = selectedDestination.config as DestinationConfig
			}
		} catch (e) {
			console.error("Error parsing destination config:", e)
		}

		if (onFormDataChange) {
			onFormDataChange(configObj)
		}

		// Check for catalog properties
		if (configObj.catalog || configObj.catalog_type) {
			const catalogValue = configObj.catalog || configObj.catalog_type || "none"

			if (catalogValue === "None") {
				setCatalog(catalogValue)
			} else if (catalogValue === "glue") {
				setCatalog(CATALOG_TYPES.AWS_GLUE)
			} else if (catalogValue === "rest") {
				setCatalog(CATALOG_TYPES.REST_CATALOG)
			} else if (catalogValue === "jdbc") {
				setCatalog(CATALOG_TYPES.JDBC_CATALOG)
			} else if (catalogValue === "hive") {
				setCatalog(CATALOG_TYPES.HIVE_CATALOG)
			}
		}

		setFormData(configObj)
	}

	const handleFormChange = (newFormData: DestinationConfig) => {
		setFormData(newFormData)
		if (onFormDataChange) {
			onFormDataChange(newFormData)
		}
	}

	const handleVersionChange = (value: string) => {
		setVersion(value)
		if (onVersionChange) {
			onVersionChange(value)
		}
	}

	// Helper function to render connector option
	const renderConnectorOption = (
		value: string,
		icon: string,
		label: string,
	) => ({
		value,
		label: (
			<div className="flex items-center">
				<img
					src={icon}
					alt={label}
					className="mr-2 size-5"
				/>
				<span>{label}</span>
			</div>
		),
	})

	// Connector options
	const connectorOptions = [
		renderConnectorOption(CONNECTOR_TYPES.AMAZON_S3, AWSS3Icon, "Amazon S3"),
		renderConnectorOption(
			CONNECTOR_TYPES.APACHE_ICEBERG,
			ApacheIceBerg,
			"Apache Iceberg",
		),
	]

	// Type for catalog options
	type CatalogOption = { value: string; label: string }

	// Catalog options
	const catalogOptions: CatalogOption[] =
		connector === CONNECTOR_TYPES.APACHE_ICEBERG
			? [
					{ value: CATALOG_TYPES.AWS_GLUE, label: "AWS Glue" },
					{ value: CATALOG_TYPES.REST_CATALOG, label: "REST catalog" },
					{ value: CATALOG_TYPES.JDBC_CATALOG, label: "JDBC" },
					{ value: CATALOG_TYPES.HIVE_CATALOG, label: "Hive catalog" },
				]
			: [{ value: CATALOG_TYPES.NONE, label: "None" }]

	return (
		<div className={`flex h-screen flex-col ${fromJobFlow ? "pb-32" : ""}`}>
			{/* Header */}
			{!fromJobFlow && (
				<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
					<Link
						to={"/destinations"}
						className="flex items-center gap-2 p-1.5 hover:rounded-[6px] hover:bg-[#f6f6f6] hover:text-black"
					>
						<ArrowLeft className="mr-1 size-5" />
					</Link>
					<div className="text-xl font-bold">Create destination</div>
				</div>
			)}

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden">
				{/* Left content */}
				<div className="w-full overflow-auto p-6 pt-6">
					{stepNumber && stepTitle && (
						<StepTitle
							stepNumber={stepNumber}
							stepTitle={stepTitle}
						/>
					)}
					<div className="mb-6 mt-6 rounded-xl border border-gray-200 bg-white p-6">
						<div>
							<div className="mb-4 flex items-center gap-1 text-base font-medium">
								<Notebook className="size-5" />
								Capture information
							</div>

							{!fromJobEditFlow && (
								<div className="mb-4 flex">
									<Radio.Group
										value={setupType}
										onChange={e => setSetupType(e.target.value as SetupType)}
										className="flex"
									>
										<Radio
											value={SETUP_TYPES.NEW}
											className="mr-8"
										>
											Set up a new destination
										</Radio>
										<Radio value={SETUP_TYPES.EXISTING}>
											Use an existing destination
										</Radio>
									</Radio.Group>
								</div>
							)}

							{setupType === SETUP_TYPES.NEW && !fromJobEditFlow ? (
								<div className="flex-start flex w-full gap-12">
									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Connector:
										</label>
										<Select
											value={connector}
											onChange={handleConnectorChange}
											className="w-full"
											options={connectorOptions}
										/>
									</div>

									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Catalog:
										</label>
										<Select
											value={catalog || CATALOG_TYPES.NONE}
											onChange={handleCatalogChange}
											className="w-full"
											disabled={connector !== CONNECTOR_TYPES.APACHE_ICEBERG}
											options={catalogOptions}
										/>
									</div>
								</div>
							) : (
								<div className="flex flex-col gap-8">
									<div className="flex w-full gap-6">
										<div className="w-1/3">
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Connector:
											</label>
											<Select
												value={connector}
												onChange={handleConnectorChange}
												className="h-8 w-full"
												disabled={fromJobEditFlow}
												options={connectorOptions}
											/>
										</div>
										<div className="w-1/3">
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Catalog:
											</label>
											<Select
												value={catalog || CATALOG_TYPES.NONE}
												onChange={handleCatalogChange}
												className="h-8 w-full"
												disabled={
													fromJobEditFlow ||
													connector !== CONNECTOR_TYPES.APACHE_ICEBERG
												}
												options={catalogOptions}
											/>
										</div>
									</div>

									<div className="w-3/5">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											{fromJobEditFlow
												? "Destination:"
												: "Select existing destination:"}
										</label>
										<Select
											placeholder="Select a destination"
											className="w-full"
											onChange={handleExistingDestinationSelect}
											value={
												fromJobEditFlow ? existingDestinationId : undefined
											}
											disabled={fromJobEditFlow}
											options={filteredDestinations.map(d => ({
												value: d.id,
												label: d.name,
											}))}
										/>
									</div>
								</div>
							)}

							{setupType === SETUP_TYPES.NEW && !fromJobEditFlow && (
								<div className="mt-4 flex w-full gap-12">
									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Name of your destination:
											<span className="text-red-500">*</span>
										</label>
										<Input
											placeholder="Enter the name of your destination"
											value={destinationName}
											onChange={handleDestinationNameChange}
										/>
									</div>
									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Version:
										</label>
										<Select
											value={version}
											onChange={handleVersionChange}
											className="w-full"
											loading={loadingVersions}
											placeholder="Select version"
											options={versions.map(v => ({
												value: v,
												label: v,
											}))}
										/>
									</div>
								</div>
							)}
						</div>
					</div>

					{setupType === SETUP_TYPES.NEW && (
						<>
							{loading ? (
								<div className="flex h-32 items-center justify-center">
									<Spin tip="Loading schema..." />
								</div>
							) : (
								<>
									{schema && (
										<div className="mb-6 rounded-xl border border-gray-200 bg-white p-6">
											<div className="mb-4 flex items-center">
												<div className="mb-2 flex items-center gap-1">
													<GenderNeuter className="size-5" />
													<div className="text-base font-medium">
														Endpoint config
													</div>
												</div>
											</div>
											{schema && (
												<FixedSchemaForm
													schema={schema}
													{...(uiSchema ? { uiSchema } : {})}
													formData={formData}
													onChange={handleFormChange}
													hideSubmit={true}
												/>
											)}
										</div>
									)}
								</>
							)}
						</>
					)}
				</div>

				{/* Documentation panel */}
				<DocumentationPanel
					docUrl={`https://olake.io/docs/writers/${getConnectorName(connector, catalog)}`}
					showResizer={true}
				/>
			</div>

			{/* Footer */}
			{!fromJobFlow && (
				<div className="flex justify-between border-t border-gray-200 bg-white p-4">
					<button
						onClick={handleCancel}
						className="rounded-[6px] border border-[#F5222D] px-4 py-1 text-[#F5222D] hover:bg-[#F5222D] hover:text-white"
					>
						Cancel
					</button>
					<button
						className="flex items-center justify-center gap-1 rounded-[6px] bg-[#203FDD] px-4 py-1 font-light text-white hover:bg-[#132685]"
						onClick={handleCreate}
					>
						Create
						<ArrowRight className="size-4 text-white" />
					</button>
				</div>
			)}

			<TestConnectionModal />
			<TestConnectionSuccessModal />
			<EntitySavedModal
				type="destination"
				onComplete={onComplete}
				fromJobFlow={fromJobFlow || false}
				entityName={destinationName}
			/>
			<EntityCancelModal
				type="destination"
				navigateTo={
					fromJobEditFlow ? "jobs" : fromJobFlow ? "jobs/new" : "destinations"
				}
			/>
		</div>
	)
}

export default CreateDestination
