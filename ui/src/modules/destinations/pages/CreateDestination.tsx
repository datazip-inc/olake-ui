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

// Define an interface compatible with EntityBase but with more specific types for our component
interface Destination {
	id: string | number
	name: string
	type: string
	version: string
	config: string | Record<string, any>
}

// ExtendedDestination ensures config is always available
interface ExtendedDestination extends Destination {
	config: string | Record<string, any>
}

interface CreateDestinationProps {
	fromJobFlow?: boolean
	fromJobEditFlow?: boolean
	existingDestinationId?: string
	onComplete?: () => void
	stepNumber?: number
	stepTitle?: string
	initialConfig?: any
	initialFormData?: any
	onDestinationNameChange?: (name: string) => void
	onConnectorChange?: (connector: string) => void
	onFormDataChange?: (formData: any) => void
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
	const [setupType, setSetupType] = useState("new")
	const [connector, setConnector] = useState("Amazon S3")
	const [catalog, setCatalog] = useState<string | null>(null)
	const [destinationName, setDestinationName] = useState("")
	const [version, setVersion] = useState("")
	const [versions, setVersions] = useState<string[]>([])
	const [loadingVersions, setLoadingVersions] = useState(false)
	const [formData, setFormData] = useState<any>({})
	const [schema, setSchema] = useState<any>(null)
	const [loading, setLoading] = useState(false)
	const [uiSchema, setUiSchema] = useState<any>(null)
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

	useEffect(() => {
		if (!destinations.length) {
			fetchDestinations()
		}
	}, [destinations.length, fetchDestinations])

	// Initialize with initial config if provided
	useEffect(() => {
		if (initialConfig) {
			setDestinationName(initialConfig.name)
			setConnector(initialConfig.type)
			setFormData(initialConfig.config || {})
		}
	}, [initialConfig])

	// Update form data when initial form data changes
	useEffect(() => {
		if (initialFormData) {
			setFormData(initialFormData)
		}
	}, [initialFormData])

	useEffect(() => {
		if (fromJobEditFlow && existingDestinationId) {
			setSetupType("existing")
			const selectedDestination = destinations.find(
				d => d.id === existingDestinationId,
			) as ExtendedDestination
			if (selectedDestination) {
				setDestinationName(selectedDestination.name)
				setConnector(selectedDestination.type)
			}
		}
	}, [fromJobEditFlow, existingDestinationId, destinations])

	useEffect(() => {
		if (connector === "Apache Iceberg") {
			setCatalog("AWS Glue")
		} else {
			setCatalog(null)
		}
	}, [connector])

	useEffect(() => {
		if (setupType === "existing") {
			if (connector === "Apache Iceberg") {
				const catalogValue = catalog || "AWS Glue"
				const filtered = destinations.filter(destination => {
					if (destination.type !== getConnectorInLowerCase(connector))
						return false
					const extDestination = destination as any
					return (
						extDestination.config?.catalog ===
							getCatalogInLowerCase(catalogValue) ||
						extDestination.config?.catalog_type ===
							getCatalogInLowerCase(catalogValue)
					)
				})
				setFilteredDestinations(filtered as unknown as ExtendedDestination[])
			} else {
				const filtered = destinations.filter(
					destination =>
						destination.type === getConnectorInLowerCase(connector),
				)

				setFilteredDestinations(filtered as unknown as ExtendedDestination[])
			}
		}
	}, [connector, setupType, destinations, catalog])

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

					// Pass the version to parent component if onVersionChange is provided
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

	useEffect(() => {
		const fetchDestinationSpec = async () => {
			try {
				setLoading(true)
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
					console.error("Failed to get source spec:", response.message)
				}
			} catch (error) {
				console.error("Error fetching source spec:", error)
			} finally {
				setLoading(false)
			}
		}

		fetchDestinationSpec()
	}, [connector, catalog, version])

	const handleCancel = () => {
		setShowEntitySavedModal(false)
	}

	const handleCreate = async () => {
		let catalogInLowerCase
		if (catalog) {
			catalogInLowerCase = getCatalogInLowerCase(catalog)
		}
		const newDestinationData = {
			name: destinationName,
			type: connector === "Amazon S3" ? "s3" : "iceberg",
			version: version,
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
		setConnector(value)
		if (onConnectorChange) {
			onConnectorChange(value)
		}
		if (value === "Apache Iceberg") {
			setCatalog("AWS Glue")
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
		if (selectedDestination) {
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
			let configObj: Record<string, any> = {}
			try {
				if (typeof selectedDestination.config === "string") {
					configObj = JSON.parse(selectedDestination.config)
				} else if (
					selectedDestination.config &&
					typeof selectedDestination.config === "object"
				) {
					configObj = selectedDestination.config
				}
			} catch (e) {
				console.error("Error parsing destination config:", e)
			}

			if (onFormDataChange) {
				onFormDataChange(configObj)
			}

			// Check for catalog properties
			if (configObj.catalog || configObj.catalog_type) {
				let catalogValue = "none"
				if (configObj.catalog) {
					catalogValue = configObj.catalog
				} else if (configObj.catalog_type) {
					catalogValue = configObj.catalog_type
				}

				if (catalogValue == "None") {
					setCatalog(catalogValue)
				} else if (catalogValue === "glue") {
					setCatalog("AWS Glue")
				} else if (catalogValue === "rest") {
					setCatalog("REST Catalog")
				} else if (catalogValue === "jdbc") {
					setCatalog("JDBC Catalog")
				} else if (catalogValue === "hive") {
					setCatalog("Hive Catalog")
				}
			}
			setFormData(configObj)
		}
	}

	const handleFormChange = (newFormData: any) => {
		setFormData(newFormData)
		if (onFormDataChange) {
			onFormDataChange(newFormData)
		}
	}

	// Add a handler for version changes
	const handleVersionChange = (value: string) => {
		setVersion(value)
		if (onVersionChange) {
			onVersionChange(value)
		}
	}

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
										onChange={e => setSetupType(e.target.value)}
										className="flex"
									>
										<Radio
											value="new"
											className="mr-8"
										>
											Set up a new destination
										</Radio>
										<Radio value="existing">Use an existing destination</Radio>
									</Radio.Group>
								</div>
							)}

							{setupType === "new" && !fromJobEditFlow ? (
								<div className="flex-start flex w-full gap-12">
									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Connector:
										</label>
										<Select
											value={connector}
											onChange={handleConnectorChange}
											className="w-full"
											options={[
												{
													value: "Amazon S3",
													label: (
														<div className="flex items-center">
															<img
																src={AWSS3Icon}
																alt="AWS S3"
																className="mr-2 size-5"
															/>
															<span>Amazon S3</span>
														</div>
													),
												},
												{
													value: "Apache Iceberg",
													label: (
														<div className="flex items-center">
															<img
																src={ApacheIceBerg}
																alt="Apache Iceberg"
																className="mr-2 size-5"
															/>
															<span>Apache Iceberg</span>
														</div>
													),
												},
											]}
										/>
									</div>

									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Catalog :
										</label>
										{connector === "Apache Iceberg" ? (
											<Select
												value={catalog}
												onChange={handleCatalogChange}
												className="w-full"
												options={[
													{ value: "AWS Glue", label: "AWS Glue" },
													{ value: "REST Catalog", label: "REST catalog" },
													{ value: "JDBC Catalog", label: "JDBC" },
													{ value: "Hive Catalog", label: "Hive catalog" },
												]}
											/>
										) : (
											<Select
												value="None"
												className="w-full"
												disabled
												options={[{ value: "None", label: "None" }]}
											/>
										)}
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
												options={[
													{
														value: "Amazon S3",
														label: (
															<div className="flex items-center">
																<img
																	src={AWSS3Icon}
																	alt="AWS S3"
																	className="mr-2 size-5"
																/>
																<span>Amazon S3</span>
															</div>
														),
													},
													{
														value: "Apache Iceberg",
														label: (
															<div className="flex items-center">
																<img
																	src={ApacheIceBerg}
																	alt="Apache Iceberg"
																	className="mr-2 size-5"
																/>
																<span>Apache Iceberg</span>
															</div>
														),
													},
												]}
											/>
										</div>
										<div className="w-1/3">
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Catalog:
											</label>
											{connector === "Apache Iceberg" ? (
												<Select
													value={catalog}
													onChange={handleCatalogChange}
													className="h-8 w-full"
													disabled={fromJobEditFlow}
													options={[
														{ value: "AWS Glue", label: "AWS Glue" },
														{ value: "REST Catalog", label: "REST catalog" },
														{ value: "JDBC Catalog", label: "JDBC" },
														{ value: "Hive Catalog", label: "Hive catalog" },
													]}
												/>
											) : (
												<Select
													value="None"
													className="w-full"
													disabled
													options={[{ value: "None", label: "None" }]}
												/>
											)}
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

							{setupType === "new" && !fromJobEditFlow && (
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
											options={versions.map(version => ({
												value: version,
												label: version,
											}))}
										/>
									</div>
								</div>
							)}
						</div>
					</div>

					{setupType === "new" && (
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
