import {
	useState,
	useEffect,
	forwardRef,
	useImperativeHandle,
	useRef,
} from "react"
import { Link, useNavigate } from "react-router-dom"
import { Input, message, Select, Spin } from "antd"
import { ArrowLeft, ArrowRight, Info, Notebook } from "@phosphor-icons/react"

import { useAppStore } from "../../../store"
import { destinationService } from "../../../api/services/destinationService"
import {
	CatalogType,
	CreateDestinationProps,
	DestinationConfig,
	ExtendedDestination,
	SetupType,
} from "../../../types"
import { getConnectorInLowerCase, getConnectorName } from "../../../utils/utils"
import {
	CONNECTOR_TYPES,
	DESTINATION_INTERNAL_TYPES,
	mapCatalogValueToType,
	SETUP_TYPES,
	widgets,
} from "../../../utils/constants"
import Form from "@rjsf/antd"
import EndpointTitle from "../../../utils/EndpointTitle"
import FormField from "../../../utils/FormField"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import StepTitle from "../../common/components/StepTitle"
import { SetupTypeSelector } from "../../common/components/SetupTypeSelector"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"
import EntitySavedModal from "../../common/Modals/EntitySavedModal"
import EntityCancelModal from "../../common/Modals/EntityCancelModal"
import { connectorOptions } from "../components/connectorOptions"
import ObjectFieldTemplate from "../../common/components/Form/ObjectFieldTemplate"
import CustomFieldTemplate from "../../common/components/Form/CustomFieldTemplate"
import validator from "@rjsf/validator-ajv8"
import ArrayFieldTemplate from "../../common/components/Form/ArrayFieldTemplate"
import { validateFormData } from "../../../utils/validateFormData"

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
			docsMinimized = false,
			onDocsMinimizedChange,
		},
		ref,
	) => {
		const formRef = useRef<any>(null)
		const [setupType, setSetupType] = useState(SETUP_TYPES.NEW)
		const [connector, setConnector] = useState<ConnectorType>(
			initialConnector === undefined
				? CONNECTOR_TYPES.AMAZON_S3
				: initialConnector === DESTINATION_INTERNAL_TYPES.S3
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
		const [schema, setSchema] = useState<any>(null)
		const [loading, setLoading] = useState(false)
		const [uiSchema, setUiSchema] = useState<any>(null)
		const [filteredDestinations, setFilteredDestinations] = useState<
			ExtendedDestination[]
		>([])
		const [destinationNameError, setDestinationNameError] = useState<
			string | null
		>(null)
		const navigate = useNavigate()

		const resetVersionState = () => {
			setVersions([])
			setVersion("")
			setSchema(null)
			if (onVersionChange) {
				onVersionChange("")
			}
		}

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
			if (setupType === SETUP_TYPES.EXISTING) {
				fetchDestinations()
			}
		}, [setupType, fetchDestinations])

		useEffect(() => {
			if (initialConfig) {
				setDestinationName(initialConfig.name)
				setConnector(initialConfig.type as ConnectorType)
				setFormData(initialConfig.config || {})
			}
		}, [initialConfig])

		useEffect(() => {
			if (onDocsMinimizedChange) {
				onDocsMinimizedChange(false)
			}
		}, [])

		useEffect(() => {
			if (initialFormData) {
				setFormData(initialFormData)
				setCatalog(initialFormData?.catalog_type ?? null)
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
					initialConnector === DESTINATION_INTERNAL_TYPES.S3
						? CONNECTOR_TYPES.AMAZON_S3
						: CONNECTOR_TYPES.APACHE_ICEBERG,
				)
			}
		}, [initialConnector])

		useEffect(() => {
			if (setupType !== SETUP_TYPES.EXISTING) return

			const filterDestinationsByConnector = () => {
				const connectorLowerCase = getConnectorInLowerCase(connector)

				return destinations
					.filter(destination => destination.type === connectorLowerCase)
					.map(dest => ({
						...dest,
						config: parseDestinationConfig(dest.config),
					}))
			}

			setFilteredDestinations(filterDestinationsByConnector())
		}, [connector, setupType, destinations])

		useEffect(() => {
			const fetchVersions = async () => {
				setLoadingVersions(true)
				try {
					const response = await destinationService.getDestinationVersions(
						connector.toLowerCase(),
					)
					if (response.data?.version) {
						const receivedVersions = response.data.version
						setVersions(receivedVersions)
						if (receivedVersions.length > 0) {
							const defaultVersion = receivedVersions[0]
							setVersion(defaultVersion)
							if (onVersionChange) {
								onVersionChange(defaultVersion)
							}
						}
					} else {
						resetVersionState()
					}
				} catch (error) {
					resetVersionState()
					console.error("Error fetching versions:", error)
				} finally {
					setLoadingVersions(false)
				}
			}

			fetchVersions()
		}, [connector, onVersionChange, setupType])

		useEffect(() => {
			if (!version) {
				setSchema(null)
				setUiSchema(null)
				return
			}

			const fetchDestinationSpec = async () => {
				try {
					setLoading(true)
					const response = await destinationService.getDestinationSpec(
						connector,
						version,
					)
					if (response.success && response.data?.spec) {
						setSchema(response.data.spec)
						setUiSchema(JSON.parse(response.data.uischema))
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
		}, [connector, version, setupType])

		useEffect(() => {
			if (!fromJobFlow) {
				setFormData({})
			}
		}, [connector])

		const handleCancel = () => {
			setShowSourceCancelModal(true)
		}

		const validateDestination = async (): Promise<boolean> => {
			// setValidating(true)

			try {
				if (setupType === SETUP_TYPES.NEW) {
					if (!destinationName.trim() && version.trim() !== "") {
						setDestinationNameError("Destination name is required")
						message.error("Destination name is required")
						return false
					} else {
						setDestinationNameError(null)
					}

					if (version.trim() === "") {
						message.error("No versions available")
						return false
					}

					// Trigger RJSF validation UI to show red borders on invalid fields in job flow
					if (
						fromJobFlow &&
						schema &&
						formRef.current &&
						formRef.current.submit
					) {
						try {
							formRef.current.submit()
						} catch {}
					}

					// Block flow if schema validation fails
					if (schema) {
						const schemaErrors = validateFormData(formData, schema)
						if (Object.keys(schemaErrors).length > 0) {
							return false
						}
					}
				}

				if (setupType === SETUP_TYPES.EXISTING) {
					// Name required always for "existing"
					if (destinationName.trim() === "") {
						message.error("Destination name is required")
						return false
					} else {
						setDestinationNameError(null)
					}
				}

				return true
			} catch (error) {
				console.error("Error validating destination:", error)
				return false
			}
		}

		useImperativeHandle(ref, () => ({
			validateDestination,
		}))

		const handleCreate = async () => {
			// In job flow, submission is only used to surface validation errors; avoid side effects
			if (fromJobFlow) {
				return
			}
			const isValid = await validateDestination()
			if (!isValid) return

			const newDestinationData = {
				name: destinationName,
				type:
					connector === CONNECTOR_TYPES.AMAZON_S3
						? DESTINATION_INTERNAL_TYPES.S3
						: DESTINATION_INTERNAL_TYPES.ICEBERG,
				version,
				config: JSON.stringify({ ...formData }),
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
			setFormData({})
			setSchema(null)
			setConnector(value as ConnectorType)
			if (onConnectorChange) {
				onConnectorChange(value)
			}
		}

		const handleSetupTypeChange = (type: SetupType) => {
			setSetupType(type)
			if (onDocsMinimizedChange) {
				if (type === SETUP_TYPES.EXISTING) {
					onDocsMinimizedChange(true)
				} else if (type === SETUP_TYPES.NEW) {
					onDocsMinimizedChange(false)
				}
			}
			// Clear form data when switching to new destination
			if (type === SETUP_TYPES.NEW) {
				setDestinationName("")
				setFormData({})
				setSchema(null)
				setConnector(CONNECTOR_TYPES.DESTINATION_DEFAULT_CONNECTOR) // Reset to default connector
				// Schema will be automatically fetched due to useEffect when connector changes
				if (onDestinationNameChange) onDestinationNameChange("")
				if (onConnectorChange) onConnectorChange(CONNECTOR_TYPES.AMAZON_S3)
				if (onFormDataChange) onFormDataChange({})
				if (onVersionChange) onVersionChange("")
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

			let configObj: any = {}
			if (selectedDestination.config) {
				let config = selectedDestination.config
				if (typeof config === "string") {
					try {
						config = JSON.parse(config)
					} catch (e) {
						console.error("Error parsing config string:", e)
						config = ""
					}
				}
				if (
					config &&
					typeof config === "object" &&
					config !== null &&
					Object.prototype.hasOwnProperty.call(config, "writer") &&
					typeof (config as any).writer === "object"
				) {
					configObj = (config as any).writer || {}
				} else {
					configObj = {}
				}
			}

			if (onFormDataChange) onFormDataChange(configObj)
			setDestinationName(selectedDestination.name)
			setFormData(configObj)
		}

		const handleVersionChange = (value: string) => {
			setVersion(value)
			if (onVersionChange) {
				onVersionChange(value)
			}
		}

		const setupTypeSelector = () => (
			<SetupTypeSelector
				value={setupType as SetupType}
				onChange={handleSetupTypeChange}
				newLabel="Set up a new destination"
				existingLabel="Use an existing destination"
				fromJobFlow={fromJobFlow}
			/>
		)

		const newDestinationForm = () =>
			setupType === SETUP_TYPES.NEW ? (
				<>
					<div className="flex gap-6">
						<div className="flex-start flex w-1/2">
							<FormField label="Connector:">
								<Select
									value={connector}
									onChange={handleConnectorChange}
									className="w-full"
									options={connectorOptions}
								/>
							</FormField>
						</div>
						<div className="w-1/2">
							<FormField label="Version:">
								{loadingVersions ? (
									<div className="flex h-8 items-center justify-center">
										<Spin size="small" />
									</div>
								) : versions && versions.length > 0 ? (
									<Select
										value={version}
										onChange={handleVersionChange}
										className="w-full"
										placeholder="Select version"
										options={versions.map(v => ({
											value: v,
											label: v,
										}))}
									/>
								) : (
									<div className="flex items-center gap-1 text-sm text-red-500">
										<Info />
										No versions available
									</div>
								)}
							</FormField>
						</div>
					</div>

					<div className="mt-4 flex w-2/3 gap-6">
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
					</div>
				</>
			) : (
				<div className="flex flex-col gap-8">
					<div className="flex w-full gap-6">
						<div className="w-1/2">
							<FormField label="Connector:">
								<Select
									value={connector}
									onChange={handleConnectorChange}
									className="h-8 w-full"
									options={connectorOptions}
								/>
							</FormField>
						</div>

						<div className="w-1/2">
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

								<Form
									ref={formRef}
									schema={schema}
									templates={{
										ObjectFieldTemplate: ObjectFieldTemplate,
										FieldTemplate: CustomFieldTemplate,
										ArrayFieldTemplate: ArrayFieldTemplate,
										ButtonTemplates: {
											SubmitButton: () => null,
										},
									}}
									widgets={widgets}
									formData={formData}
									onChange={e => {
										setFormData(e.formData)
										if (onFormDataChange) onFormDataChange(e.formData)
										const catalogValue = e.formData?.catalog_type
										if (catalogValue) {
											const mappedCatalogType =
												mapCatalogValueToType(catalogValue)
											if (mappedCatalogType) {
												setCatalog(mappedCatalogType)
											}
										}
									}}
									onSubmit={handleCreate}
									uiSchema={uiSchema}
									validator={validator}
									showErrorList={false}
									omitExtraData
									liveOmit
								/>
							</div>
						)
					)}
				</>
			)

		const handleToggleDocPanel = () => {
			if (onDocsMinimizedChange) {
				onDocsMinimizedChange(prev => !prev)
			}
		}

		return (
			<div className="flex h-screen">
				<div className="flex flex-1 flex-col">
					{!fromJobFlow && (
						<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
							<Link
								to={"/destinations"}
								className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
							>
								<ArrowLeft className="mr-1 size-5" />
							</Link>
							<div className="text-lg font-bold">Create destination</div>
						</div>
					)}

					<div className="flex flex-1 overflow-hidden">
						<div className="flex flex-1 flex-col">
							<div className="flex-1 overflow-auto p-6 pt-0">
								{stepNumber && stepTitle && (
									<StepTitle
										stepNumber={stepNumber}
										stepTitle={stepTitle}
									/>
								)}
								<div className="mb-6 mt-2 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
									<div>
										<div className="mb-4 flex items-center gap-2 text-base font-medium">
											<Notebook className="size-5" />
											Capture information
										</div>

										{setupTypeSelector()}
										{newDestinationForm()}
									</div>
								</div>

								{schemaFormSection()}
							</div>

							{/* Footer */}
							{!fromJobFlow && (
								<div className="flex justify-between border-t border-gray-200 bg-white p-4 shadow-sm">
									<button
										onClick={handleCancel}
										className="ml-1 rounded-md border border-danger px-4 py-2 text-danger transition-colors duration-200 hover:bg-danger hover:text-white"
									>
										Cancel
									</button>
									<button
										className="mr-1 flex items-center justify-center gap-1 rounded-md bg-primary px-4 py-2 font-light text-white shadow-sm transition-colors duration-200 hover:bg-primary-600"
										onClick={() => {
											if (formRef.current) {
												formRef.current.submit()
											}
										}}
									>
										Create
										<ArrowRight className="size-4 text-white" />
									</button>
								</div>
							)}
						</div>

						<DocumentationPanel
							docUrl={`https://olake.io/docs/writers/${getConnectorName(connector, catalog)}`}
							showResizer={true}
							isMinimized={docsMinimized}
							onToggle={handleToggleDocPanel}
						/>
					</div>
				</div>

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
