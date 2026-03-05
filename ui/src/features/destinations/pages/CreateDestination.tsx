import {
	useState,
	useEffect,
	forwardRef,
	useImperativeHandle,
	useRef,
} from "react"
import { Link } from "react-router-dom"
import { Input, message, Select, Spin } from "antd"
import {
	ArrowLeftIcon,
	ArrowRightIcon,
	ArrowSquareOutIcon,
	InfoIcon,
	NotebookIcon,
} from "@phosphor-icons/react"
import Form from "@rjsf/antd"

import { validationService } from "@/common/services/validationService"
import {
	CreateDestinationProps,
	DestinationConfig,
	ExtendedDestination,
} from "../types"
import { SetupType } from "@/common/types"
import type { TestConnectionError } from "@/common/types"
import { getConnectorInLowerCase } from "@/common/utils"
import { getConnectorDocumentationPath } from "../utils"
import { trimFormDataStrings } from "@/utils"
import { handleSpecResponse } from "@/common/utils"
import {
	useDestinations,
	useDestinationVersions,
	useDestinationSpec,
} from "../hooks/queries/useDestinationQueries"
import {
	useCreateDestination,
	useTestDestinationConnection,
} from "../hooks/mutations/useDestinationMutations"
import {
	CONNECTOR_TYPES,
	DESTINATION_INTERNAL_TYPES,
	ENTITY_TYPES,
	SETUP_TYPES,
	TEST_CONNECTION_STATUS,
	transformErrors,
} from "@/common/constants"
import { OLAKE_LATEST_VERSION_URL } from "@/constants"
import EndpointTitle from "../../../common/components/EndpointTitle"
import FormField from "../../../common/components/FormField"
import DocumentationPanel from "@/common/components/DocumentationPanel"
import { SetupTypeSelector } from "@/common/components/SetupTypeSelector"
import TestConnectionModal from "@/common/components/modals/TestConnectionModal"
import TestConnectionSuccessModal from "@/common/components/modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "@/common/components/modals/TestConnectionFailureModal"
import EntitySavedModal from "@/common/components/modals/EntitySavedModal"
import EntityCancelModal from "@/common/components/modals/EntityCancelModal"
import { connectorOptions } from "../components/connectorOptions"
import ObjectFieldTemplate from "@/common/components/form/ObjectFieldTemplate"
import CustomFieldTemplate from "@/common/components/form/CustomFieldTemplate"
import validator from "@rjsf/validator-ajv8"
import ArrayFieldTemplate from "@/common/components/form/ArrayFieldTemplate"
import { widgets } from "@/common/components/form/widgets"
import SpecFailedModal from "@/common/components/modals/SpecFailedModal"

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
			onComplete,
			initialConfig,
			initialFormData,
			initialName,
			initialConnector,
			initialVersion,
			initialCatalog,
			initialExistingDestinationId,
			onDestinationNameChange,
			onConnectorChange,
			onFormDataChange,
			onVersionChange,
			onExistingDestinationIdChange,
			docsMinimized = false,
			onDocsMinimizedChange,
		},
		ref,
	) => {
		const formRef = useRef<any>(null)
		const [setupType, setSetupType] = useState(
			initialExistingDestinationId ? SETUP_TYPES.EXISTING : SETUP_TYPES.NEW,
		)
		const [connector, setConnector] = useState<ConnectorType>(
			initialConnector === undefined
				? CONNECTOR_TYPES.AMAZON_S3
				: initialConnector === DESTINATION_INTERNAL_TYPES.S3
					? CONNECTOR_TYPES.AMAZON_S3
					: CONNECTOR_TYPES.APACHE_ICEBERG,
		)
		const [catalog, setCatalog] = useState<string | null>(
			initialCatalog || null,
		)
		const [destinationName, setDestinationName] = useState(initialName || "")
		const [version, setVersion] = useState(initialVersion || "")
		const [formData, setFormData] = useState<DestinationConfig>({})
		const [schema, setSchema] = useState<any>(null)
		const [uiSchema, setUiSchema] = useState<any>(null)
		const [existingDestination, setExistingDestination] = useState<
			string | null
		>(null)
		const [filteredDestinations, setFilteredDestinations] = useState<
			ExtendedDestination[]
		>([])
		const [destinationNameError, setDestinationNameError] = useState<
			string | null
		>(null)
		const [specError, setSpecError] = useState<string | null>(null)

		const [showTestingModal, setShowTestingModal] = useState(false)
		const [showSuccessModal, setShowSuccessModal] = useState(false)
		const [showFailureModal, setShowFailureModal] = useState(false)
		const [showEntitySavedModal, setShowEntitySavedModal] = useState(false)
		const [showSourceCancelModal, setShowSourceCancelModal] = useState(false)
		const [showSpecFailedModal, setShowSpecFailedModal] = useState(false)
		const [testConnectionError, setTestConnectionError] =
			useState<TestConnectionError | null>(null)
		const { data: destinations = [], isLoading: isLoadingDestinations } =
			useDestinations()
		const internalConnectorType =
			connector === CONNECTOR_TYPES.APACHE_ICEBERG
				? DESTINATION_INTERNAL_TYPES.ICEBERG
				: DESTINATION_INTERNAL_TYPES.S3
		const { data: versionsData, isLoading: loadingVersions } =
			useDestinationVersions(internalConnectorType)
		const versions = versionsData?.version ?? []
		const createDestinationMutation = useCreateDestination()
		const testDestinationMutation = useTestDestinationConnection()

		// Fetch spec via TanStack Query
		const {
			data: specData,
			isLoading: loadingSpec,
			error: specQueryError,
			refetch: refetchSpec,
		} = useDestinationSpec(
			setupType === SETUP_TYPES.NEW ? internalConnectorType : "",
			version,
			"",
			"",
		)

		useEffect(() => {
			if (specData) {
				handleSpecResponse(specData, setSchema, setUiSchema, "destination")
			}
		}, [specData])

		useEffect(() => {
			if (specQueryError) {
				setSchema({})
				setUiSchema({})
				const errMsg =
					specQueryError instanceof Error
						? specQueryError.message
						: "Failed to fetch spec, Please try again."
				setSpecError(errMsg)
				setShowSpecFailedModal(true)
			}
		}, [specQueryError])

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

		// Set existingDestination when destinations are loaded and we have an initialExistingDestinationId
		useEffect(() => {
			if (
				initialExistingDestinationId &&
				destinations.length > 0 &&
				setupType === SETUP_TYPES.EXISTING
			) {
				const destination = destinations.find(
					d => d.id === initialExistingDestinationId,
				)
				const connectorLowerCase = getConnectorInLowerCase(connector)
				if (destination && destination.type === connectorLowerCase) {
					setExistingDestination(destination.name)
				}
			}
		}, [initialExistingDestinationId, destinations.length])

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
				setCatalog(initialFormData?.writer?.catalog_type ?? null)
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

		// Auto-select version when versions are loaded
		useEffect(() => {
			if (versions.length > 0 && !version) {
				const defaultVersion =
					getConnectorInLowerCase(connector) === initialConnector &&
					initialVersion
						? initialVersion
						: versions[0]
				setVersion(defaultVersion)
				onVersionChange?.(defaultVersion)
			}
		}, [versions])

		useEffect(() => {
			setFormData({})
		}, [connector])

		const handleCancel = () => {
			setShowSourceCancelModal(true)
		}

		//makes sure user enters destination name and version and fills all the required fields in the form
		const validateDestination = async (): Promise<boolean> => {
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

					if (schema && formRef.current) {
						const validationResult = formRef.current.validateForm()
						return validationResult
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
			const isValid = await validateDestination()
			if (!isValid) return

			const isUnique = await validationService.checkUniqueName(
				destinationName,
				ENTITY_TYPES.DESTINATION,
			)
			if (!isUnique) return

			const newDestinationData = {
				name: destinationName,
				type:
					connector === CONNECTOR_TYPES.AMAZON_S3
						? DESTINATION_INTERNAL_TYPES.S3
						: DESTINATION_INTERNAL_TYPES.ICEBERG,
				version,
				config: JSON.stringify({ ...formData }),
			}

			setShowTestingModal(true)
			//test the connection and show either success or failure modal based on the result
			const testResult = await testDestinationMutation.mutateAsync({
				destination: newDestinationData,
			})
			setShowTestingModal(false)

			if (
				testResult.data?.connection_result.status ===
				TEST_CONNECTION_STATUS.SUCCEEDED
			) {
				setShowSuccessModal(true)
				setTimeout(() => {
					setShowSuccessModal(false)
					createDestinationMutation.mutate(newDestinationData, {
						onSuccess: () => setShowEntitySavedModal(true),
						onError: error => console.error("Error adding destination:", error),
					})
				}, 1000)
			} else {
				setTestConnectionError({
					message: testResult.data?.connection_result.message || "",
					logs: testResult.data?.logs || [],
				})
				setShowFailureModal(true)
			}
		}

		const handleDestinationNameChange = (
			e: React.ChangeEvent<HTMLInputElement>,
		) => {
			const newName = e.target.value.trim()
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
			if (setupType === SETUP_TYPES.EXISTING) {
				setExistingDestination(null)
				onExistingDestinationIdChange?.(null)
				setDestinationName("")
				onDestinationNameChange?.("")
			}
			setVersion("")
			setFormData({})
			setSchema(null)

			// Parent callbacks
			onConnectorChange?.(value)
			onVersionChange?.("")
			onFormDataChange?.({})
		}

		const handleSetupTypeChange = (type: SetupType) => {
			setSetupType(type)
			setDestinationName("")
			onExistingDestinationIdChange?.(null)
			onDestinationNameChange?.("")

			if (onDocsMinimizedChange) {
				if (type === SETUP_TYPES.EXISTING) {
					onDocsMinimizedChange(true)
				} else if (type === SETUP_TYPES.NEW) {
					onDocsMinimizedChange(false)
				}
			}
			// Clear form data when switching to new destination
			if (type === SETUP_TYPES.NEW) {
				setFormData({})
				setSchema(null)
				setConnector(CONNECTOR_TYPES.DESTINATION_DEFAULT_CONNECTOR) // Reset to default connector
				setExistingDestination(null)
				onExistingDestinationIdChange?.(null)
				// Schema will be automatically fetched due to useEffect when connector changes
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

			const configObj =
				selectedDestination.config &&
				typeof selectedDestination.config === "object"
					? selectedDestination.config
					: {}

			if (onFormDataChange) onFormDataChange(configObj)
			setDestinationName(selectedDestination.name)
			setFormData(configObj)
			setExistingDestination(value)
			onExistingDestinationIdChange?.(selectedDestination.id)
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
			/>
		)

		const newDestinationForm = () =>
			setupType === SETUP_TYPES.NEW ? (
				<>
					<div className="flex gap-6">
						<div className="flex-start flex w-1/2">
							<FormField label="Connector:">
								<Select
									data-testid="destination-connector-select"
									value={connector}
									onChange={handleConnectorChange}
									className="w-full"
									options={connectorOptions}
								/>
							</FormField>
						</div>
						<div className="w-1/2">
							<FormField
								label="OLake Version:"
								tooltip="Choose the OLake version for the destination"
								info={
									<a
										href={OLAKE_LATEST_VERSION_URL}
										target="_blank"
										rel="noopener noreferrer"
										className="flex items-center text-primary hover:text-primary/80"
									>
										<ArrowSquareOutIcon className="size-4" />
									</a>
								}
							>
								{loadingVersions ? (
									<div className="flex h-8 items-center justify-center">
										<Spin size="small" />
									</div>
								) : versions && versions.length > 0 ? (
									<Select
										value={version}
										data-testid="destination-version-select"
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
										<InfoIcon />
										No versions available
									</div>
								)}
							</FormField>
						</div>
					</div>

					<div className="mt-4 flex w-1/2 gap-6">
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
									data-testid="destination-connector-select"
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
							{isLoadingDestinations ? (
								<div className="flex h-8 items-center justify-center">
									<Spin size="small" />
								</div>
							) : (
								<Select
									placeholder="Select a destination"
									className="w-full"
									data-testid="existing-destination"
									onChange={handleExistingDestinationSelect}
									value={existingDestination}
									options={filteredDestinations.map(d => ({
										value: d.id,
										label: d.name,
									}))}
								/>
							)}
						</div>
					</div>
				</div>
			)

		// JSX for schema form
		const schemaFormSection = () =>
			setupType === SETUP_TYPES.NEW && (
				<>
					{loadingSpec ? (
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
									transformErrors={transformErrors}
									templates={{
										ObjectFieldTemplate,
										FieldTemplate: CustomFieldTemplate,
										ArrayFieldTemplate,
										ButtonTemplates: {
											SubmitButton: () => null,
										},
									}}
									widgets={widgets}
									formData={formData}
									onChange={e => {
										const trimmedData = trimFormDataStrings(e.formData)
										setFormData(trimmedData)
										if (onFormDataChange) onFormDataChange(trimmedData)
										const catalogValue = trimmedData?.writer?.catalog_type
										if (catalogValue) setCatalog(catalogValue)
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
					<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
						<Link
							to={"/destinations"}
							className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
						>
							<ArrowLeftIcon className="mr-1 size-5" />
						</Link>
						<div className="text-lg font-bold">Create destination</div>
					</div>

					<div className="flex flex-1 overflow-hidden">
						<div className="flex flex-1 flex-col">
							<div className="flex-1 overflow-auto p-6 pt-0">
								<div className="mb-6 mt-2 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
									<div>
										<div className="mb-4 flex items-center gap-2 text-base font-medium">
											<NotebookIcon className="size-5" />
											Capture information
										</div>

										{setupTypeSelector()}
										{newDestinationForm()}
									</div>
								</div>

								{schemaFormSection()}
							</div>

							{/* Footer */}
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
									<ArrowRightIcon className="size-4 text-white" />
								</button>
							</div>
						</div>

						<DocumentationPanel
							docUrl={`https://olake.io/docs/writers/${getConnectorDocumentationPath(connector, catalog)}`}
							showResizer={true}
							isMinimized={docsMinimized}
							onToggle={handleToggleDocPanel}
						/>
					</div>
				</div>

				<TestConnectionModal
					open={showTestingModal}
					connectionType="destination"
				/>
				<TestConnectionSuccessModal
					open={showSuccessModal}
					connectionType="destination"
				/>
				<TestConnectionFailureModal
					open={showFailureModal}
					onClose={() => setShowFailureModal(false)}
					connectionType="destination"
					testConnectionError={testConnectionError}
				/>
				<EntitySavedModal
					open={showEntitySavedModal}
					onClose={() => setShowEntitySavedModal(false)}
					type="destination"
					onComplete={onComplete}
					fromJobFlow={false}
					entityName={destinationName}
				/>
				<EntityCancelModal
					open={showSourceCancelModal}
					onClose={() => setShowSourceCancelModal(false)}
					type="destination"
					navigateTo="destinations"
				/>
				<SpecFailedModal
					open={showSpecFailedModal}
					onClose={() => setShowSpecFailedModal(false)}
					fromSource={false}
					error={specError ?? ""}
					onTryAgain={refetchSpec}
				/>
			</div>
		)
	},
)

CreateDestination.displayName = "CreateDestination"

export default CreateDestination
