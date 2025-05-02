import { useState, useEffect } from "react"
import { Link } from "react-router-dom"
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
import DynamicSchemaForm from "../../common/components/DynamicSchemaForm"
import { Destination } from "../../../types"
import { destinationService } from "../../../api/services/destinationService"
import { getConnectorName } from "../../../utils/utils"

interface ExtendedDestination extends Destination {
	config?: any
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
}) => {
	const [setupType, setSetupType] = useState("new")
	const [connector, setConnector] = useState("Amazon S3")
	const [catalog, setCatalog] = useState<string | null>(null)
	const [destinationName, setDestinationName] = useState("")
	const [formData, setFormData] = useState<any>({})
	const [schema, setSchema] = useState<any>(null)
	const [loading, setLoading] = useState(false)
	const [filteredDestinations, setFilteredDestinations] = useState<
		ExtendedDestination[]
	>([])

	const {
		destinations,
		fetchDestinations,
		setShowEntitySavedModal,
		setShowTestingModal,
		setShowSuccessModal,
		addDestination,
	} = useAppStore()

	useEffect(() => {
		const fetchSchema = async () => {
			try {
				setLoading(true)
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
					setSchema(schemaData)
				} else {
					schemaData = await destinationService.getConnectorSchema(connector)
					setSchema(schemaData)
				}

				// Initialize with default values from schema
				if (schemaData.properties) {
					const initialData: any = {}

					// Apply default values from schema properties
					Object.entries(schemaData.properties).forEach(
						([key, value]: [string, any]) => {
							if (value.default !== undefined) {
								initialData[key] = value.default
							}
						},
					)

					// Only set initial data if we don't have existing form data
					if (Object.keys(formData).length === 0) {
						setFormData(initialData)
					}
				}
			} catch (error) {
				console.error("Error fetching schema:", error)
			} finally {
				setLoading(false)
			}
		}

		// Fetch schema for both new and existing sources
		fetchSchema()
	}, [connector, setupType, catalog])

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

	// Make sure catalog is immediately set when connector changes
	useEffect(() => {
		if (connector === "Apache Iceberg") {
			setCatalog("AWS Glue")
		} else {
			setCatalog(null)
		}
	}, [connector])

	// Update useEffect for filtered destinations to remove redundant catalog check
	useEffect(() => {
		if (setupType === "existing") {
			// Only filter by catalog if it's Apache Iceberg
			if (connector === "Apache Iceberg") {
				// Make sure we have a catalog value
				const catalogValue = catalog || "AWS Glue"

				// Create a safe version of the filter that checks for config existence
				const filtered = destinations.filter(destination => {
					// First check if it's the right connector type
					if (destination.type !== connector) return false

					// For Apache Iceberg, also check the catalog value
					const extDestination = destination as ExtendedDestination
					return extDestination.catalog === catalogValue
				})
				setFilteredDestinations(filtered as ExtendedDestination[])
			} else {
				const filtered = destinations.filter(
					destination => destination.type === connector,
				) as ExtendedDestination[]

				setFilteredDestinations(filtered)
			}
		}
	}, [connector, setupType, destinations, catalog])

	const handleCancel = () => {
		setShowEntitySavedModal(false)
	}

	const handleCreate = () => {
		// Add the new destination to the store state
		const newDestinationData = {
			name: destinationName,
			type: connector,
			status: "active" as const,
			config: { ...formData, catalog },
		}

		addDestination(newDestinationData)
			.then(() => {
				// Continue with the existing flow
				setShowTestingModal(true)
				setTimeout(() => {
					setShowTestingModal(false)
					setShowSuccessModal(true)
					setTimeout(() => {
						setShowSuccessModal(false)
						setShowEntitySavedModal(true)
					}, 2000)
				}, 2000)
			})
			.catch(error => {
				console.error("Error adding destination:", error)
			})
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
	}

	const handleCatalogChange = (value: string) => {
		setCatalog(value)
	}

	const handleExistingDestinationSelect = (value: string) => {
		const selectedDestination = destinations.find(
			d => d.id === value,
		) as ExtendedDestination
		if (selectedDestination) {
			setDestinationName(selectedDestination.name)
			setConnector(selectedDestination.type)
			if (selectedDestination.config?.catalog) {
				setCatalog(selectedDestination.config.catalog)
			}
			setFormData(selectedDestination.config || formData)
		}
	}

	const handleFormChange = (newFormData: any) => {
		setFormData(newFormData)
		if (onFormDataChange) {
			onFormDataChange(newFormData)
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
								<div className="mt-4 w-[70%]">
									<label className="mb-2 block text-sm font-medium text-gray-700">
										Name of your destination :
									</label>
									<Input
										placeholder="Enter the name of your destination"
										value={destinationName}
										onChange={handleDestinationNameChange}
									/>
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

											<DynamicSchemaForm
												schema={schema}
												uiSchema={schema?.uiSchema}
												formData={formData}
												onChange={handleFormChange}
												hideSubmit={true}
											/>
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
