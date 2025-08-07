import { useState, useEffect, forwardRef, useImperativeHandle } from "react"
import { Link, useNavigate } from "react-router-dom"
import { message, Select, Spin } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, ArrowRight, Notebook } from "@phosphor-icons/react"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import EntitySavedModal from "../../common/Modals/EntitySavedModal"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import EntityCancelModal from "../../common/Modals/EntityCancelModal"
import StepTitle from "../../common/components/StepTitle"
import FixedSchemaForm, { validateFormData } from "../../../utils/FormFix"
import { sourceService } from "../../../api/services/sourceService"
import { getConnectorLabel } from "../../../utils/utils"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"
import { SetupType, Source, CreateSourceProps } from "../../../types"
import EndpointTitle from "../../../utils/EndpointTitle"
import FormField from "../../../utils/FormField"
import connectorOptions from "../components/connectorOptions"
import { SetupTypeSelector } from "../../common/components/SetupTypeSelector"
import { CONNECTOR_TYPES } from "../../../utils/constants"

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
		const [setupType, setSetupType] = useState<SetupType>("new")
		const [connector, setConnector] = useState(initialConnector || "MongoDB")
		const [sourceName, setSourceName] = useState(initialName || "")
		const [selectedVersion, setSelectedVersion] = useState(
			initialVersion || "latest",
		)
		const [versions, setVersions] = useState<string[]>([])
		const [loadingVersions, setLoadingVersions] = useState(false)
		const [formData, setFormData] = useState<any>({})
		const [schema, setSchema] = useState<any>(null)
		const [loading, setLoading] = useState(false)
		const [filteredSources, setFilteredSources] = useState<Source[]>([])
		const [formErrors, setFormErrors] = useState<Record<string, string>>({})
		const [sourceNameError, setSourceNameError] = useState<string | null>(null)
		const [validating, setValidating] = useState(false)

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
			if (setupType === "existing") {
				fetchSources()
				setFilteredSources(
					sources.filter(source => source.type === connector.toLowerCase()),
				)
			}
		}, [connector, setupType, fetchSources])

		useEffect(() => {
			if (
				initialVersion &&
				initialVersion !== "latest" &&
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
					if (response.data && response.data.version) {
						setVersions(response.data.version)
						if (
							response.data.version.length > 0 &&
							(!initialVersion ||
								connector !== initialConnector ||
								initialVersion === "latest")
						) {
							const defaultVersion = response.data.version[0]
							setSelectedVersion(defaultVersion)
							if (onVersionChange) {
								onVersionChange(defaultVersion)
							}
						}
					}
				} catch (error) {
					console.error("Error fetching versions:", error)
				} finally {
					setLoadingVersions(false)
				}
			}

			fetchVersions()
		}, [connector, onVersionChange, initialVersion, initialConnector])

		useEffect(() => {
			const fetchSourceSpec = async () => {
				try {
					setLoading(true)
					const response = await sourceService.getSourceSpec(
						connector,
						selectedVersion,
					)
					if (response.success && response.data?.spec) {
						setSchema(response.data.spec)
					} else {
						console.error("Failed to get source spec:", response.message)
					}
				} catch (error) {
					console.error("Error fetching source spec:", error)
				} finally {
					setLoading(false)
				}
			}

			fetchSourceSpec()
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
			setValidating(true)
			let isValid = true
			if (setupType === "new") {
				if (!sourceName.trim()) {
					setSourceNameError("Source name is required")
					message.error("Source name is required")
					isValid = false
				} else {
					setSourceNameError(null)
				}
			}
			if (setupType === "new" && schema) {
				const schemaErrors = validateFormData(formData, schema)
				setFormErrors(schemaErrors)
				isValid = isValid && Object.keys(schemaErrors).length === 0
			}
			return isValid
		}

		useImperativeHandle(ref, () => ({
			validateSource,
		}))

		const handleCreate = async () => {
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
				if (testResult.data?.status === "SUCCEEDED") {
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
					setSourceTestConnectionError(testResult.data?.message || "")
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
			setFormData({})
			setSchema(null)
			setConnector(value)
			if (onConnectorChange) {
				onConnectorChange(value)
			}
		}

		const handleSetupTypeChange = (type: SetupType) => {
			setSetupType(type)
			// Clear form data when switching to new source
			if (type === "new") {
				setSourceName("")
				setFormData({})
				setSchema(null)
				setConnector(CONNECTOR_TYPES.SOURCE_DEFAULT_CONNECTOR) // Reset to default connector

				// Schema will be automatically fetched due to useEffect when connector changes
				if (onSourceNameChange) onSourceNameChange("")
				if (onConnectorChange) onConnectorChange("MongoDB")
				if (onFormDataChange) onFormDataChange({})
				if (onVersionChange) onVersionChange("latest")
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
				setSourceName(selectedSource.name)
				setConnector(getConnectorLabel(selectedSource.type))
				setSelectedVersion(selectedSource.version)
			}
		}

		const handleFormChange = (newFormData: any) => {
			setFormData(newFormData)

			if (onFormDataChange) {
				onFormDataChange(newFormData)
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

		// UI component renderers
		const renderConnectorSelection = () => (
			<div className="w-1/3">
				<label className="mb-2 block text-sm font-medium text-gray-700">
					Connector:
				</label>
				<div className="flex items-center">
					<Select
						value={connector}
						onChange={handleConnectorChange}
						className={setupType === "new" ? "h-8 w-full" : "w-full"}
						options={connectorOptions}
						{...(setupType !== "new"
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

					<div className="w-1/3">
						<label className="mb-2 block text-sm font-medium text-gray-700">
							OLake Version:
						</label>
						<Select
							value={selectedVersion}
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

				<div className="w-2/3">
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

				<div className="w-1/3">
					<label className="mb-2 block text-sm font-medium text-gray-700">
						Select existing source:
					</label>
					<Select
						placeholder="Select a source"
						className="w-full"
						onChange={handleExistingSourceSelect}
						value={undefined}
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
			setupType === "new" && (
				<>
					{loading ? (
						<div className="flex h-32 items-center justify-center">
							<Spin tip="Loading schema..." />
						</div>
					) : (
						schema && (
							<div className="mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
								<EndpointTitle title="Endpoint config" />
								<FixedSchemaForm
									schema={schema}
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
			<div className={`flex h-screen`}>
				<div className="flex flex-1 flex-col">
					{!fromJobFlow && (
						<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
							<Link
								to={"/sources"}
								className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
							>
								<ArrowLeft className="mr-1 size-5" />
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
											<Notebook className="size-5" />
											Capture information
										</div>

										{renderSetupTypeSelector()}

										{setupType === "new"
											? renderNewSourceForm()
											: renderExistingSourceForm()}
									</div>
								</div>

								{renderSchemaForm()}
							</div>

							{/* Footer - moved inside center section */}
							{!fromJobFlow && (
								<div className="flex justify-between border-t border-gray-200 bg-white p-4 shadow-sm">
									<button
										onClick={handleCancel}
										className="border-danger text-danger hover:bg-danger ml-1 rounded-md border px-4 py-2 transition-colors duration-200 hover:text-white"
									>
										Cancel
									</button>
									<button
										className="bg-primary hover:bg-primary-600 mr-1 flex items-center justify-center gap-1 rounded-md px-4 py-2 font-light text-white shadow-sm transition-colors duration-200"
										onClick={handleCreate}
									>
										Create
										<ArrowRight className="size-4 text-white" />
									</button>
								</div>
							)}
						</div>

						<DocumentationPanel
							docUrl={`https://olake.io/docs/connectors/${connector.toLowerCase()}/config`}
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
