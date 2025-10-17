import {
	useState,
	useEffect,
	forwardRef,
	useImperativeHandle,
	useRef,
} from "react"
import { Link, useNavigate } from "react-router-dom"
import { message, Select, Spin, Tooltip } from "antd"
import {
	ArrowLeftIcon,
	ArrowRightIcon,
	ArrowSquareOutIcon,
	InfoIcon,
	NotebookIcon,
} from "@phosphor-icons/react"
import Form from "@rjsf/antd"
import validator from "@rjsf/validator-ajv8"

import { useAppStore } from "../../../store"
import { sourceService } from "../../../api/services/sourceService"
import { SetupType, Source, CreateSourceProps } from "../../../types"
import {
	getConnectorLabel,
	handleSpecResponse,
	withAbortController,
} from "../../../utils/utils"
import {
	CONNECTOR_TYPES,
	OLAKE_LATEST_VERSION_URL,
	transformErrors,
	TEST_CONNECTION_STATUS,
} from "../../../utils/constants"
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
import connectorOptions from "../components/connectorOptions"
import { SETUP_TYPES } from "../../../utils/constants"
import ObjectFieldTemplate from "../../common/components/Form/ObjectFieldTemplate"
import CustomFieldTemplate from "../../common/components/Form/CustomFieldTemplate"
import ArrayFieldTemplate from "../../common/components/Form/ArrayFieldTemplate"
import { widgets } from "../../common/components/Form/widgets"

// Create ref handle interface
export interface CreateSourceHandle {
	validateSource: () => Promise<boolean>
}

const CreateSource = forwardRef<CreateSourceHandle, CreateSourceProps>(
	(
		{
			fromJobFlow = false,
			onComplete,
			stepNumber,
			stepTitle,
			initialFormData,
			initialName,
			initialConnector,
			initialVersion,
			onSourceNameChange,
			onConnectorChange,
			onFormDataChange,
			onVersionChange,
			docsMinimized = false,
			onDocsMinimizedChange,
		},
		ref,
	) => {
		const formRef = useRef<any>(null)
		const [setupType, setSetupType] = useState<SetupType>("new")
		const [connector, setConnector] = useState(initialConnector || "MongoDB")
		const [sourceName, setSourceName] = useState(initialName || "")
		const [selectedVersion, setSelectedVersion] = useState(initialVersion || "")
		const [versions, setVersions] = useState<string[]>([])
		const [loadingVersions, setLoadingVersions] = useState(false)
		const [formData, setFormData] = useState<any>({})
		const [schema, setSchema] = useState<any>(null)
		const [uiSchema, setUiSchema] = useState<any>(null)
		const [loading, setLoading] = useState(false)
		const [filteredSources, setFilteredSources] = useState<Source[]>([])
		const [sourceNameError, setSourceNameError] = useState<string | null>(null)
		const [existingSource, setExistingSource] = useState<string | null>(null)

		const navigate = useNavigate()

		const {
			sources,
			fetchSources,
			setShowEntitySavedModal,
			setShowTestingModal,
			setShowSuccessModal,
			setShowSourceCancelModal,
			addSource,
			setShowFailureModal,
			setSourceTestConnectionError,
		} = useAppStore()

		useEffect(() => {
			if (!sources.length) {
				fetchSources()
			}
		}, [sources.length, fetchSources])

		useEffect(() => {
			if (initialName) {
				setSourceName(initialName)
			}
		}, [initialName])

		useEffect(() => {
			if (initialFormData) {
				setFormData(initialFormData)
			}
		}, [initialFormData])

		useEffect(() => {
			if (setupType === SETUP_TYPES.EXISTING) {
				fetchSources()
				setFilteredSources(
					sources.filter(source => source.type === connector.toLowerCase()),
				)
			}
		}, [connector, setupType, fetchSources])

		const resetVersionState = () => {
			setVersions([])
			setSelectedVersion("")
			setSchema(null)
			if (onVersionChange) {
				onVersionChange("")
			}
		}

		useEffect(() => {
			if (
				initialVersion &&
				initialVersion !== "" &&
				initialConnector === connector
			) {
				setSelectedVersion(initialVersion)
			}
		}, [initialVersion, initialConnector, connector])

		useEffect(() => {
			const fetchVersions = async () => {
				setLoadingVersions(true)
				try {
					const response = await sourceService.getSourceVersions(
						connector.toLowerCase(),
					)
					if (response.data?.version) {
						setVersions(response.data.version)
						if (
							response.data.version.length > 0 &&
							(!initialVersion ||
								connector !== initialConnector ||
								initialVersion === "")
						) {
							let defaultVersion = response.data.version[0]
							if (
								connector.toLowerCase() === initialConnector &&
								initialVersion
							) {
								defaultVersion = initialVersion
							}
							setSelectedVersion(defaultVersion)
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
		}, [connector, initialConnector])

		useEffect(() => {
			if (!selectedVersion) {
				setSchema(null)
				return
			}

			if (setupType === SETUP_TYPES.EXISTING) return

			setLoading(true)
			return withAbortController(
				signal =>
					sourceService.getSourceSpec(connector, selectedVersion, signal),
				response =>
					handleSpecResponse(response, setSchema, setUiSchema, "source"),
				error => {
					setSchema({})
					setUiSchema({})
					console.error("Error fetching source spec:", error)
				},
				() => setLoading(false),
			)
		}, [connector, selectedVersion, setupType])

		useEffect(() => {
			if (initialConnector) {
				setConnector(getConnectorLabel(initialConnector))
			}
		}, [])

		const handleCancel = () => {
			setShowSourceCancelModal(true)
		}

		const validateSource = async (): Promise<boolean> => {
			try {
				if (setupType === SETUP_TYPES.NEW) {
					if (!sourceName.trim() && selectedVersion.trim() !== "") {
						setSourceNameError("Source name is required")
						message.error("Source name is required")
						return false
					} else {
						setSourceNameError(null)
					}

					if (selectedVersion.trim() === "") {
						message.error("No versions available")
						return false
					}

					// Use RJSF's built-in validation - validate returns validation state
					if (schema && formRef.current) {
						const validationResult = formRef.current.validateForm()
						return validationResult
					}
				}

				if (setupType === SETUP_TYPES.EXISTING) {
					if (sourceName.trim() === "") {
						message.error("Source name is required")
						return false
					} else {
						setSourceNameError(null)
					}
				}

				return true
			} catch (error) {
				console.error("Error validating source:", error)
				return false
			}
		}

		useImperativeHandle(ref, () => ({
			validateSource,
		}))

		const handleCreate = async () => {
			if (fromJobFlow) {
				return
			}
			const isValid = await validateSource()
			if (!isValid) return

			const newSourceData = {
				name: sourceName,
				type: connector.toLowerCase(),
				version: selectedVersion,
				config: JSON.stringify(formData),
			}

			try {
				setShowTestingModal(true)
				const testResult =
					await sourceService.testSourceConnection(newSourceData)
				setShowTestingModal(false)
				if (
					testResult.data?.connection_result.status ===
					TEST_CONNECTION_STATUS.SUCCEEDED
				) {
					setShowSuccessModal(true)
					setTimeout(() => {
						setShowSuccessModal(false)
						addSource(newSourceData)
							.then(() => {
								setShowEntitySavedModal(true)
							})
							.catch(error => {
								console.error("Error adding source:", error)
							})
					}, 1000)
				} else {
					const testConnectionError = {
						message: testResult.data?.connection_result.message || "",
						logs: testResult.data?.logs || [],
					}
					setSourceTestConnectionError(testConnectionError)
					setShowFailureModal(true)
				}
			} catch (error) {
				setShowTestingModal(false)
				console.error("Error testing connection:", error)
				navigate("/sources")
			}
		}

		const handleSourceNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
			const newName = e.target.value
			if (newName.length >= 1) {
				setSourceNameError(null)
			}
			setSourceName(newName)

			if (onSourceNameChange) {
				onSourceNameChange(newName)
			}
		}

		const handleConnectorChange = (value: string) => {
			setConnector(value)
			if (setupType === SETUP_TYPES.EXISTING) {
				setExistingSource(null)
				setSourceName("")
				onSourceNameChange?.("")
			}
			setSelectedVersion("")
			setFormData({})
			setSchema(null)

			// Parent callbacks
			onConnectorChange?.(value)
			onVersionChange?.("")
			onFormDataChange?.({})
		}

		const handleSetupTypeChange = (type: SetupType) => {
			setSetupType(type)
			setSourceName("")
			onSourceNameChange?.("")
			// show documentation only in the case of new
			if (onDocsMinimizedChange) {
				if (type === SETUP_TYPES.EXISTING) {
					onDocsMinimizedChange(true) // Close doc panel
				} else if (type === SETUP_TYPES.NEW) {
					onDocsMinimizedChange(false)
				}
			}
			// Clear form data when switching to new source
			if (type === SETUP_TYPES.NEW) {
				setFormData({})
				setSchema(null)
				setConnector(CONNECTOR_TYPES.SOURCE_DEFAULT_CONNECTOR) // Reset to default connector
				setExistingSource(null)
				// Schema will be automatically fetched due to useEffect when connector changes
				if (onConnectorChange) onConnectorChange(CONNECTOR_TYPES.MONGODB)
				if (onFormDataChange) onFormDataChange({})
				if (onVersionChange) onVersionChange("")
			}
		}

		const handleExistingSourceSelect = (value: string) => {
			const selectedSource = sources.find(
				s => s.id.toString() === value.toString(),
			)

			if (selectedSource) {
				if (onSourceNameChange) {
					onSourceNameChange(selectedSource.name)
				}
				if (onConnectorChange) {
					onConnectorChange(selectedSource.type)
				}
				if (onVersionChange) {
					onVersionChange(selectedSource.version)
				}
				if (onFormDataChange) {
					onFormDataChange(selectedSource.config)
				}
				setExistingSource(value)
				setSourceName(selectedSource.name)
				setConnector(getConnectorLabel(selectedSource.type))
				setSelectedVersion(selectedSource.version)
			}
		}

		const handleVersionChange = (value: string) => {
			setSelectedVersion(value)
			if (onVersionChange) {
				onVersionChange(value)
			}
		}

		const handleToggleDocPanel = () => {
			if (onDocsMinimizedChange) {
				onDocsMinimizedChange(prev => !prev)
			}
		}

		const renderConnectorSelection = () => (
			<div className="w-1/2">
				<label className="mb-2 block text-sm font-medium text-gray-700">
					Connector:
				</label>
				<div className="flex items-center">
					<Select
						value={connector}
						onChange={handleConnectorChange}
						data-testid="source-connector-select"
						className={setupType === SETUP_TYPES.NEW ? "h-8 w-full" : "w-full"}
						options={connectorOptions}
						{...(setupType !== SETUP_TYPES.NEW
							? { style: { boxShadow: "0 1px 2px 0 rgba(0, 0, 0, 0.05)" } }
							: {})}
					/>
				</div>
			</div>
		)

		const renderNewSourceForm = () => (
			<div className="flex flex-col gap-6">
				<div className="flex w-full gap-6">
					{renderConnectorSelection()}

					<div className="w-1/2">
						<label className="mb-2 flex items-center gap-1 text-sm font-medium text-gray-700">
							OLake Version:
							<Tooltip title="Choose the OLake version for the source">
								<InfoIcon
									size={16}
									className="cursor-help text-slate-900"
								/>
							</Tooltip>
							<a
								href={OLAKE_LATEST_VERSION_URL}
								target="_blank"
								rel="noopener noreferrer"
								className="flex items-center text-primary hover:text-primary/80"
							>
								<ArrowSquareOutIcon className="size-4" />
							</a>
						</label>
						{loadingVersions ? (
							<div className="flex h-8 items-center justify-center">
								<Spin size="small" />
							</div>
						) : versions && versions.length > 0 ? (
							<>
								<Select
									value={selectedVersion}
									onChange={handleVersionChange}
									className="w-full"
									placeholder="Select version"
									data-testid="source-version-select"
									options={versions.map(version => ({
										value: version,
										label: version,
									}))}
								/>
							</>
						) : (
							<div className="flex items-center gap-1 text-sm text-red-500">
								<InfoIcon />
								No versions available
							</div>
						)}
					</div>
				</div>

				<div className="w-1/2">
					<FormField
						label="Name of your source"
						required
						error={sourceNameError}
					>
						<input
							type="text"
							className={`h-8 w-full rounded-md border ${sourceNameError ? "border-red-500" : "border-gray-400"} px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500`}
							placeholder="Enter the name of your source"
							value={sourceName}
							onChange={handleSourceNameChange}
						/>
					</FormField>
				</div>
			</div>
		)

		const renderExistingSourceForm = () => (
			<div className="flex-start flex w-full gap-6">
				{renderConnectorSelection()}

				<div className="w-1/2">
					<label className="mb-2 block text-sm font-medium text-gray-700">
						Select existing source:
					</label>
					<Select
						placeholder="Select a source"
						className="w-full"
						data-testid="existing-source"
						onChange={handleExistingSourceSelect}
						value={existingSource}
						options={filteredSources.map(s => ({
							value: s.id,
							label: s.name,
						}))}
					/>
				</div>
			</div>
		)

		const renderSetupTypeSelector = () => (
			<SetupTypeSelector
				value={setupType as SetupType}
				onChange={handleSetupTypeChange}
				newLabel="Set up a new source"
				existingLabel="Use an existing source"
				fromJobFlow={fromJobFlow}
			/>
		)

		const renderSchemaForm = () =>
			setupType === SETUP_TYPES.NEW && (
				<>
					{loading ? (
						<div className="flex h-32 items-center justify-center">
							<Spin tip="Loading schema..." />
						</div>
					) : (
						schema && (
							<div className="mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
								<EndpointTitle title="Endpoint config" />
								<Form
									ref={formRef}
									schema={schema}
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
										setFormData(e.formData)
										if (onFormDataChange) onFormDataChange(e.formData)
									}}
									transformErrors={transformErrors}
									uiSchema={uiSchema}
									validator={validator}
									omitExtraData
									liveOmit
									showErrorList={false} // adding this will not show error list
									onSubmit={handleCreate}
								/>
							</div>
						)
					)}
				</>
			)

		return (
			<div className={`flex h-screen`}>
				<div className="flex flex-1 flex-col">
					{!fromJobFlow && (
						<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
							<Link
								to={"/sources"}
								className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
							>
								<ArrowLeftIcon className="mr-1 size-5" />
							</Link>
							<div className="text-lg font-bold">Create source</div>
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
									<div className="mb-6">
										<div className="mb-4 flex items-center gap-2 text-base font-medium">
											<NotebookIcon className="size-5" />
											Capture information
										</div>

										{renderSetupTypeSelector()}

										{setupType === SETUP_TYPES.NEW
											? renderNewSourceForm()
											: renderExistingSourceForm()}
									</div>
								</div>

								{renderSchemaForm()}
							</div>

							{/* Footer  */}
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
										<ArrowRightIcon className="size-4 text-white" />
									</button>
								</div>
							)}
						</div>

						<DocumentationPanel
							docUrl={`https://olake.io/docs/connectors/${connector.toLowerCase()}`}
							isMinimized={docsMinimized}
							onToggle={handleToggleDocPanel}
							showResizer={true}
						/>
					</div>
				</div>

				<TestConnectionModal />
				<TestConnectionSuccessModal />
				<TestConnectionFailureModal fromSources={true} />
				<EntitySavedModal
					type="source"
					onComplete={onComplete}
					fromJobFlow={fromJobFlow || false}
					entityName={sourceName}
				/>
				<EntityCancelModal
					type="source"
					navigateTo={fromJobFlow ? "jobs/new" : "sources"}
				/>
			</div>
		)
	},
)

CreateSource.displayName = "CreateSource"

export default CreateSource
