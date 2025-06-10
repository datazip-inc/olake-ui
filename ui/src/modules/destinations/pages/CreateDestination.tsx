import { useState, useEffect, forwardRef, useImperativeHandle } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Input, message, Select, Spin } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, ArrowRight, Notebook } from "@phosphor-icons/react"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import EntitySavedModal from "../../common/Modals/EntitySavedModal"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import EntityCancelModal from "../../common/Modals/EntityCancelModal"
import StepTitle from "../../common/components/StepTitle"
import FixedSchemaForm, { validateFormData } from "../../../utils/FormFix"
import { destinationService } from "../../../api/services/destinationService"
import {
	getCatalogInLowerCase,
	getConnectorInLowerCase,
	getConnectorName,
} from "../../../utils/utils"
import {
	CATALOG_TYPES,
	CONNECTOR_TYPES,
	IcebergCatalogTypes,
	mapCatalogValueToType,
	SETUP_TYPES,
} from "../../../utils/constants"
import {
	CatalogType,
	CreateDestinationProps,
	DestinationConfig,
	ExtendedDestination,
	SelectOption,
	SetupType,
} from "../../../types"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"
import EndpointTitle from "../../../utils/EndpointTitle"
import FormField from "../../../utils/FormField"
import { SetupTypeSelector } from "../../common/components/SetupTypeSelector"
import { connectorOptions } from "../components/connectorOptions"

type ConnectorType = (typeof CONNECTOR_TYPES)[keyof typeof CONNECTOR_TYPES]

// Create ref handle interface
export interface CreateDestinationHandle {
	validateDestination: () => Promise<boolean>
}

const CreateDestination = forwardRef<
	CreateDestinationHandle,
	CreateDestinationProps
>(
	(
		{
			fromJobFlow = false,
			hitBack = false,
			onComplete,
			stepNumber,
			stepTitle,
			initialConfig,
			initialFormData,
			initialName,
			initialConnector,
			initialCatalog,
			onDestinationNameChange,
			onConnectorChange,
			onFormDataChange,
			onVersionChange,
			onCatalogTypeChange,
		},
		ref,
	) => {
		const [setupType, setSetupType] = useState(SETUP_TYPES.NEW)
		const [connector, setConnector] = useState<ConnectorType>(
			initialConnector === undefined
				? CONNECTOR_TYPES.AMAZON_S3
				: initialConnector === "s3"
					? CONNECTOR_TYPES.AMAZON_S3
					: CONNECTOR_TYPES.APACHE_ICEBERG,
		)
		const [catalog, setCatalog] = useState<CatalogType | null>(
			initialCatalog || null,
		)
		const [destinationName, setDestinationName] = useState(initialName || "")
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
		const [formErrors, setFormErrors] = useState<Record<string, string>>({})
		const [destinationNameError, setDestinationNameError] = useState<
			string | null
		>(null)
		const [validating, setValidating] = useState(false)
		const navigate = useNavigate()

		const {
			destinations,
			fetchDestinations,
			setShowEntitySavedModal,
			setShowTestingModal,
			setShowSuccessModal,
			addDestination,
			setShowFailureModal,
			setShowSourceCancelModal,
			setDestinationTestConnectionError,
		} = useAppStore()

		const parseDestinationConfig = (
			config: string | DestinationConfig,
		): DestinationConfig => {
			if (typeof config === "string") {
				try {
					return JSON.parse(config)
				} catch (e) {
					console.error("Error parsing destination config:", e)
					return {}
				}
			}
			return config as DestinationConfig
		}

		useEffect(() => {
			if (!destinations.length) {
				fetchDestinations()
			}
		}, [destinations.length, fetchDestinations])

		useEffect(() => {
			if (initialConfig) {
				setDestinationName(initialConfig.name)
				setConnector(initialConfig.type as ConnectorType)
				setFormData(initialConfig.config || {})
			}
		}, [initialConfig])

		useEffect(() => {
			if (initialFormData) {
				setFormData(initialFormData)
			}
		}, [initialFormData])

		useEffect(() => {
			if (initialName) {
				setDestinationName(initialName)
			}
		}, [initialName])

		useEffect(() => {
			if (initialConnector) {
				setConnector(
					initialConnector === "s3"
						? CONNECTOR_TYPES.AMAZON_S3
						: CONNECTOR_TYPES.APACHE_ICEBERG,
				)
			}
		}, [initialConnector])

		useEffect(() => {
			if (connector === CONNECTOR_TYPES.APACHE_ICEBERG) {
				setCatalog(CATALOG_TYPES.AWS_GLUE)
			} else {
				setCatalog(null)
			}
		}, [connector])

		useEffect(() => {
			if (initialCatalog) {
				setCatalog(initialCatalog)
				if (onCatalogTypeChange) {
					onCatalogTypeChange(initialCatalog)
				}
			}
		}, [initialCatalog, onCatalogTypeChange])

		useEffect(() => {
			if (setupType !== SETUP_TYPES.EXISTING) return

			const filterDestinationsByConnectorAndCatalog = () => {
				const connectorLowerCase = getConnectorInLowerCase(connector)
				const isIceberg = connector === CONNECTOR_TYPES.APACHE_ICEBERG
				const catalogValue = isIceberg
					? catalog || CATALOG_TYPES.AWS_GLUE
					: null
				const catalogLowerCase = catalogValue
					? getCatalogInLowerCase(catalogValue)
					: null

				return destinations
					.filter(destination => {
						if (destination.type !== connectorLowerCase) return false

						if (!isIceberg) return true

						try {
							const config = parseDestinationConfig(destination.config)
							return config?.writer?.catalog_type === catalogLowerCase
						} catch {
							return false
						}
					})
					.map(dest => ({
						...dest,
						config: parseDestinationConfig(dest.config),
					}))
			}

			setFilteredDestinations(filterDestinationsByConnectorAndCatalog())
		}, [connector, setupType, destinations, catalog])

		useEffect(() => {
			const fetchVersions = async () => {
				setLoadingVersions(true)
				try {
					const response = await destinationService.getDestinationVersions(
						connector.toLowerCase(),
					)
					if (response.data?.version) {
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

		useEffect(() => {
			const fetchDestinationSpec = async () => {
				setLoading(true)
				try {
					const response = await destinationService.getDestinationSpec(
						connector,
						catalog,
						version,
					)
					if (response.success && response.data?.spec) {
						setSchema(response.data.spec)
						setUiSchema(response.data.uiSchema || null)
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

		useEffect(() => {
			if (!fromJobFlow) {
				setFormData({})
			}
			if (fromJobFlow && !hitBack) {
				setFormData({})
			}
		}, [connector, catalog])

		const handleCancel = () => {
			setShowSourceCancelModal(true)
		}

		const validateDestination = async (): Promise<boolean> => {
			setValidating(true)
			let isValid = true

			if (setupType === SETUP_TYPES.NEW) {
				if (!destinationName.trim()) {
					setDestinationNameError("Destination name is required")
					message.error("Destination name is required")
					isValid = false
				} else {
					setDestinationNameError(null)
				}
			}

			if (setupType === SETUP_TYPES.NEW && schema) {
				const enrichedFormData = { ...formData }
				if (schema.properties) {
					Object.entries(schema.properties).forEach(
						([key, propValue]: [string, any]) => {
							if (
								propValue.default !== undefined &&
								(enrichedFormData[key] === undefined ||
									enrichedFormData[key] === null)
							) {
								enrichedFormData[key] = propValue.default
							}
						},
					)
				}

				const schemaErrors = validateFormData(enrichedFormData, schema)
				setFormErrors(schemaErrors)
				isValid = isValid && Object.keys(schemaErrors).length === 0
			}

			return isValid
		}

		useImperativeHandle(ref, () => ({
			validateDestination,
		}))

		const handleCreate = async () => {
			const isValid = await validateDestination()
			if (!isValid) return

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

				if (testResult.data?.status === "SUCCEEDED") {
					setShowSuccessModal(true)
					setTimeout(() => {
						setShowSuccessModal(false)
						addDestination(newDestinationData)
							.then(() => setShowEntitySavedModal(true))
							.catch(error => console.error("Error adding destination:", error))
					}, 1000)
				} else {
					setDestinationTestConnectionError(testResult.data?.message || "")
					setShowFailureModal(true)
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
			if (newName.length >= 1) {
				setDestinationNameError(null)
			}
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
		}

		const handleCatalogChange = (value: string) => {
			setCatalog(value as CatalogType)
			if (onCatalogTypeChange) {
				onCatalogTypeChange(value as CatalogType)
			}
		}

		const handleExistingDestinationSelect = (value: string) => {
			const selectedDestination = destinations.find(
				d => d.id.toString() === value.toString(),
			)
			if (!selectedDestination) return

			if (onDestinationNameChange)
				onDestinationNameChange(selectedDestination.name)
			if (onConnectorChange) onConnectorChange(selectedDestination.type)
			if (onVersionChange) onVersionChange(selectedDestination.version)
			const configObj = parseDestinationConfig(selectedDestination.config)
			if (onFormDataChange) onFormDataChange(configObj)

			setDestinationName(selectedDestination.name)

			if (configObj.catalog || configObj.catalog_type) {
				const catalogValue =
					configObj.catalog || configObj.catalog_type || "none"
				const catalogType = mapCatalogValueToType(catalogValue)
				if (catalogType) setCatalog(catalogType)
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

		const catalogOptions: SelectOption[] =
			connector === CONNECTOR_TYPES.APACHE_ICEBERG
				? IcebergCatalogTypes
				: [{ value: CATALOG_TYPES.NONE, label: "None" }]

		const setupTypeSelector = () => (
			<SetupTypeSelector
				value={setupType as SetupType}
				onChange={value => setSetupType(value)}
				newLabel="Set up a new destination"
				existingLabel="Use an existing destination"
				fromJobFlow={fromJobFlow}
			/>
		)

		const newDestinationForm = () =>
			setupType === SETUP_TYPES.NEW ? (
				<>
					<div className="flex-start flex w-full gap-12">
						<FormField label="Connector:">
							<Select
								value={connector}
								onChange={handleConnectorChange}
								className="w-full"
								options={connectorOptions}
							/>
						</FormField>

						<FormField label="Catalog:">
							<Select
								value={catalog || CATALOG_TYPES.NONE}
								onChange={handleCatalogChange}
								className="w-full"
								disabled={connector !== CONNECTOR_TYPES.APACHE_ICEBERG}
								options={catalogOptions}
							/>
						</FormField>
					</div>

					<div className="mt-4 flex w-full gap-12">
						<FormField
							label="Name of your destination:"
							required
							error={destinationNameError}
						>
							<Input
								placeholder="Enter the name of your destination"
								value={destinationName}
								onChange={handleDestinationNameChange}
								status={destinationNameError ? "error" : ""}
							/>
						</FormField>

						<FormField label="Version:">
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
						</FormField>
					</div>
				</>
			) : (
				<div className="flex flex-col gap-8">
					<div className="flex w-full gap-6">
						<FormField label="Connector:">
							<Select
								value={connector}
								onChange={handleConnectorChange}
								className="h-8 w-full"
								options={connectorOptions}
							/>
						</FormField>

						<FormField label="Catalog:">
							<Select
								value={catalog || CATALOG_TYPES.NONE}
								onChange={handleCatalogChange}
								className="h-8 w-full"
								disabled={connector !== CONNECTOR_TYPES.APACHE_ICEBERG}
								options={catalogOptions}
							/>
						</FormField>
					</div>

					<div className="w-3/5">
						<label className="mb-2 block text-sm font-medium text-gray-700">
							Select existing destination:
						</label>
						<Select
							placeholder="Select a destination"
							className="w-full"
							onChange={handleExistingDestinationSelect}
							value={undefined}
							options={filteredDestinations.map(d => ({
								value: d.id,
								label: d.name,
							}))}
						/>
					</div>
				</div>
			)

		// JSX for schema form
		const schemaFormSection = () =>
			setupType === SETUP_TYPES.NEW && (
				<>
					{loading ? (
						<div className="flex h-32 items-center justify-center">
							<Spin tip="Loading schema..." />
						</div>
					) : (
						schema && (
							<div className="mb-6 rounded-xl border border-gray-200 bg-white p-6">
								<EndpointTitle title="Endpoint config" />
								<FixedSchemaForm
									schema={schema}
									{...(uiSchema ? { uiSchema } : {})}
									formData={formData}
									onChange={handleFormChange}
									hideSubmit={true}
									errors={formErrors}
									validate={validating}
								/>
							</div>
						)
					)}
				</>
			)

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

								{setupTypeSelector()}
								{newDestinationForm()}
							</div>
						</div>

						{schemaFormSection()}
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
				<TestConnectionFailureModal fromSources={false} />
				<EntitySavedModal
					type="destination"
					onComplete={onComplete}
					fromJobFlow={fromJobFlow || false}
					entityName={destinationName}
				/>
				<EntityCancelModal
					type="destination"
					navigateTo={fromJobFlow ? "jobs/new" : "destinations"}
				/>
			</div>
		)
	},
)

CreateDestination.displayName = "CreateDestination"

export default CreateDestination
