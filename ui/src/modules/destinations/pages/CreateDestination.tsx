import { useState, useEffect, forwardRef, useImperativeHandle } from "react"
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
	SETUP_TYPES,
} from "../../../utils/constants"
import { CatalogType } from "../../../types"

type ConnectorType = (typeof CONNECTOR_TYPES)[keyof typeof CONNECTOR_TYPES]
type SetupType = (typeof SETUP_TYPES)[keyof typeof SETUP_TYPES]

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

interface ExtendedDestination extends Destination {
	config: DestinationConfig
}

interface CreateDestinationProps {
	fromJobFlow?: boolean
	hitBack?: boolean
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
	initialName?: string
	initialConnector?: string
	initialCatalog?: CatalogType | null
	onDestinationNameChange?: (name: string) => void
	onConnectorChange?: (connector: string) => void
	onFormDataChange?: (formData: DestinationConfig) => void
	onVersionChange?: (version: string) => void
	onCatalogTypeChange?: (catalog: CatalogType | null) => void
}

// Create ref handle interface
export interface CreateDestinationHandle {
	validateDestination: () => Promise<boolean>
}

type FormFieldProps = {
	label: string
	required?: boolean
	children: React.ReactNode
	error?: string | null
}

const FormField = ({ label, required, children, error }: FormFieldProps) => (
	<div className="w-1/3">
		<label className="mb-2 block text-sm font-medium text-gray-700">
			{label}
			{required && <span className="text-red-500">*</span>}
		</label>
		{children}
		{error && <div className="mt-1 text-sm text-red-500">{error}</div>}
	</div>
)

type SelectOption = { value: string; label: React.ReactNode | string }

const CreateDestination = forwardRef<
	CreateDestinationHandle,
	CreateDestinationProps
>(
	(
		{
			fromJobFlow = false,
			hitBack = false,
			fromJobEditFlow = false,
			existingDestinationId,
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
		const [setupType, setSetupType] = useState<SetupType>(SETUP_TYPES.NEW)
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

		const mapCatalogValueToType = (
			catalogValue: string,
		): CatalogType | null => {
			if (catalogValue === "none") return CATALOG_TYPES.NONE
			if (catalogValue === "glue") return CATALOG_TYPES.AWS_GLUE
			if (catalogValue === "rest") return CATALOG_TYPES.REST_CATALOG
			if (catalogValue === "jdbc") return CATALOG_TYPES.JDBC_CATALOG
			if (catalogValue === "hive") return CATALOG_TYPES.HIVE_CATALOG
			return null
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
							return (
								config?.writer?.catalog === catalogLowerCase ||
								config?.writer?.catalog_type === catalogLowerCase
							)
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

				if (testResult.success) {
					setShowSuccessModal(true)
					setTimeout(() => {
						setShowSuccessModal(false)
						addDestination(newDestinationData)
							.then(() => setShowEntitySavedModal(true))
							.catch(error => console.error("Error adding destination:", error))
					}, 1000)
				} else {
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

		const renderConnectorOption = (
			value: string,
			icon: string,
			label: string,
		): SelectOption => ({
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

		const connectorOptions: SelectOption[] = [
			renderConnectorOption(CONNECTOR_TYPES.AMAZON_S3, AWSS3Icon, "Amazon S3"),
			renderConnectorOption(
				CONNECTOR_TYPES.APACHE_ICEBERG,
				ApacheIceBerg,
				"Apache Iceberg",
			),
		]

		const catalogOptions: SelectOption[] =
			connector === CONNECTOR_TYPES.APACHE_ICEBERG
				? [
						{ value: CATALOG_TYPES.AWS_GLUE, label: "AWS Glue" },
						{ value: CATALOG_TYPES.REST_CATALOG, label: "REST catalog" },
						{ value: CATALOG_TYPES.JDBC_CATALOG, label: "JDBC Catalog" },
						{ value: CATALOG_TYPES.HIVE_CATALOG, label: "Hive Catalog" },
					]
				: [{ value: CATALOG_TYPES.NONE, label: "None" }]

		const setupTypeSelector = () =>
			!fromJobEditFlow && (
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
			)

		const newDestinationForm = () =>
			setupType === SETUP_TYPES.NEW && !fromJobEditFlow ? (
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
								disabled={fromJobEditFlow}
								options={connectorOptions}
							/>
						</FormField>

						<FormField label="Catalog:">
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
						</FormField>
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
							value={fromJobEditFlow ? existingDestinationId : undefined}
							disabled={fromJobEditFlow}
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
								<div className="mb-4 flex items-center">
									<div className="mb-2 flex items-center gap-1">
										<GenderNeuter className="size-5" />
										<div className="text-base font-medium">Endpoint config</div>
									</div>
								</div>
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
	},
)

CreateDestination.displayName = "CreateDestination"

export default CreateDestination
