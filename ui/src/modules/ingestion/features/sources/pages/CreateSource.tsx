import {
	ArrowLeftIcon,
	ArrowRightIcon,
	ArrowSquareOutIcon,
	InfoIcon,
	NotebookIcon,
} from "@phosphor-icons/react"
import Form from "@rjsf/antd"
import validator from "@rjsf/validator-ajv8"
import { message, Select, Spin, Tooltip } from "antd"
import { useState, useEffect, useRef } from "react"
import { Link } from "react-router-dom"

import ArrayFieldTemplate from "@/common/components/form/ArrayFieldTemplate"
import CustomFieldTemplate from "@/common/components/form/CustomFieldTemplate"
import ObjectFieldTemplate from "@/common/components/form/ObjectFieldTemplate"
import { widgets } from "@/common/components/form/widgets"
import {
	SpecFailedModal,
	TestConnectionFailureModal,
	TestConnectionModal,
	TestConnectionSuccessModal,
} from "@/common/components/modals"
import {
	transformErrors,
	TEST_CONNECTION_STATUS,
	OLAKE_LATEST_VERSION_URL,
} from "@/common/constants"
import { TestConnectionError } from "@/common/types"
import { trimFormDataStrings, handleSpecResponse } from "@/common/utils"
import {
	EntitySavedModal,
	EntityCancelModal,
} from "@/modules/ingestion/common/components"
import { sourceConnectorOptions as connectorOptions } from "@/modules/ingestion/common/components/connectorOptions"
import DocumentationPanel from "@/modules/ingestion/common/components/DocumentationPanel"
import EndpointTitle from "@/modules/ingestion/common/components/EndpointTitle"
import FormField from "@/modules/ingestion/common/components/FormField"
import { SetupTypeSelector } from "@/modules/ingestion/common/components/SetupTypeSelector"
import {
	CONNECTOR_TYPES,
	ENTITY_TYPES,
	SETUP_TYPES,
} from "@/modules/ingestion/common/constants"
import { validationService } from "@/modules/ingestion/common/services/validationService"
import { SetupType } from "@/modules/ingestion/common/types"
import { getConnectorInLowerCase } from "@/modules/ingestion/common/utils"

import {
	useSources,
	useSourceVersions,
	useSourceSpec,
	useCreateSource,
	useTestSourceConnection,
} from "../hooks"
import { Source } from "../types"
import { getConnectorLabel } from "../utils"

const CreateSource: React.FC = () => {
	// Local component state.
	const formRef = useRef<any>(null)
	const [setupType, setSetupType] = useState<SetupType>(SETUP_TYPES.NEW)
	const [connector, setConnector] = useState(CONNECTOR_TYPES.MONGODB)
	const [sourceName, setSourceName] = useState("")
	const [selectedVersion, setSelectedVersion] = useState("")
	const [formData, setFormData] = useState<any>({})
	const [schema, setSchema] = useState<any>(null)
	const [uiSchema, setUiSchema] = useState<any>(null)
	const [filteredSources, setFilteredSources] = useState<Source[]>([])
	const [sourceNameError, setSourceNameError] = useState<string | null>(null)
	const [existingSource, setExistingSource] = useState<string | null>(null)
	const [specError, setSpecError] = useState<string | null>(null)
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [showTestingModal, setShowTestingModal] = useState(false)
	const [showSuccessModal, setShowSuccessModal] = useState(false)
	const [showFailureModal, setShowFailureModal] = useState(false)
	const [showEntitySavedModal, setShowEntitySavedModal] = useState(false)
	const [showSourceCancelModal, setShowSourceCancelModal] = useState(false)
	const [showSpecFailedModal, setShowSpecFailedModal] = useState(false)
	const [testConnectionError, setTestConnectionError] =
		useState<TestConnectionError | null>(null)

	// Derived constants from current state.
	const normalizedConnector = getConnectorInLowerCase(connector)

	// Data fetching and mutation hooks.
	const { data: sources = [], isLoading: isLoadingSources } = useSources()
	const { data: versionsData, isLoading: loadingVersions } =
		useSourceVersions(normalizedConnector)
	const versions = versionsData?.version ?? []
	const createSourceMutation = useCreateSource()
	const testSourceMutation = useTestSourceConnection()

	useEffect(() => {
		if (setupType === SETUP_TYPES.EXISTING) {
			setFilteredSources(
				sources.filter(source => source.type === normalizedConnector),
			)
		}
	}, [normalizedConnector, setupType, sources])

	// Auto-select version when versions are loaded
	useEffect(() => {
		if (versions.length > 0 && !selectedVersion) {
			setSelectedVersion(versions[0])
		}
	}, [versions])

	// Spec query for selected connector and version.
	const {
		data: specData,
		isLoading: loadingSpec,
		error: specQueryError,
		refetch: refetchSpec,
	} = useSourceSpec(
		setupType === SETUP_TYPES.NEW ? normalizedConnector : "",
		selectedVersion,
	)

	useEffect(() => {
		if (specData) {
			handleSpecResponse(specData, setSchema, setUiSchema, "source")
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

	const handleCreate = async () => {
		const isValid = await validateSource()
		if (!isValid) return

		const isUnique = await validationService.checkUniqueName(
			sourceName,
			ENTITY_TYPES.SOURCE,
		)
		if (!isUnique) return

		const newSourceData = {
			name: sourceName,
			type: normalizedConnector,
			version: selectedVersion,
			config: JSON.stringify(formData),
		}

		setShowTestingModal(true)
		const testResult = await testSourceMutation.mutateAsync({
			source: newSourceData,
		})
		setShowTestingModal(false)
		if (
			testResult.data?.connection_result.status ===
			TEST_CONNECTION_STATUS.SUCCEEDED
		) {
			setShowSuccessModal(true)
			setTimeout(() => {
				setShowSuccessModal(false)
				createSourceMutation.mutate(newSourceData, {
					onSuccess: () => setShowEntitySavedModal(true),
					onError: error => console.error("Error adding source:", error),
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

	const handleSourceNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const newName = e.target.value.trim()
		if (newName.length >= 1) {
			setSourceNameError(null)
		}
		setSourceName(newName)
	}

	const handleConnectorChange = (value: string) => {
		setConnector(value)
		if (setupType === SETUP_TYPES.EXISTING) {
			setExistingSource(null)
			setSourceName("")
		}
		setSelectedVersion("")
		setFormData({})
		setSchema(null)
	}

	const handleSetupTypeChange = (type: SetupType) => {
		setSetupType(type)
		setSourceName("")
		setDocsMinimized(type === SETUP_TYPES.EXISTING) // Close doc panel for existing
		// Clear form data when switching to new source
		if (type === SETUP_TYPES.NEW) {
			setFormData({})
			setSchema(null)
			setConnector(CONNECTOR_TYPES.SOURCE_DEFAULT_CONNECTOR) // Reset to default connector
			setExistingSource(null)
		}
	}

	const handleExistingSourceSelect = (value: string) => {
		const selectedSource = sources.find(
			s => s.id.toString() === value.toString(),
		)

		if (selectedSource) {
			setExistingSource(value)
			setSourceName(selectedSource.name)
			setConnector(getConnectorLabel(selectedSource.type))
			setSelectedVersion(selectedSource.version)
		}
	}

	const handleVersionChange = (value: string) => {
		setSelectedVersion(value)
	}

	const handleToggleDocPanel = () => {
		setDocsMinimized(prev => !prev)
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
				{isLoadingSources ? (
					<div className="flex h-8 items-center justify-center">
						<Spin size="small" />
					</div>
				) : (
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
				)}
			</div>
		</div>
	)

	const renderSetupTypeSelector = () => (
		<SetupTypeSelector
			value={setupType as SetupType}
			onChange={handleSetupTypeChange}
			newLabel="Set up a new source"
		/>
	)

	const renderSchemaForm = () =>
		setupType === SETUP_TYPES.NEW && (
			<>
				{loadingSpec ? (
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
									const trimmedData = trimFormDataStrings(e.formData)
									setFormData(trimmedData)
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
				<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
					<Link
						to={"/sources"}
						className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
					>
						<ArrowLeftIcon className="mr-1 size-5" />
					</Link>
					<div className="text-lg font-bold">Create source</div>
				</div>

				<div className="flex flex-1 overflow-hidden">
					<div className="flex flex-1 flex-col">
						<div className="flex-1 overflow-auto p-6 pt-0">
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
						docUrl={`https://olake.io/docs/connectors/${normalizedConnector}`}
						isMinimized={docsMinimized}
						onToggle={handleToggleDocPanel}
						showResizer={true}
					/>
				</div>
			</div>

			<TestConnectionModal
				open={showTestingModal}
				connectionType="source"
			/>
			<TestConnectionSuccessModal
				open={showSuccessModal}
				connectionType="source"
			/>
			<TestConnectionFailureModal
				open={showFailureModal}
				onClose={() => setShowFailureModal(false)}
				connectionType="source"
				testConnectionError={testConnectionError}
			/>
			<EntitySavedModal
				open={showEntitySavedModal}
				onClose={() => setShowEntitySavedModal(false)}
				type="source"
				fromJobFlow={false}
				entityName={sourceName}
			/>
			<EntityCancelModal
				open={showSourceCancelModal}
				onClose={() => setShowSourceCancelModal(false)}
				type="source"
				navigateTo="sources"
			/>
			<SpecFailedModal
				open={showSpecFailedModal}
				onClose={() => setShowSpecFailedModal(false)}
				fromSource
				error={specError ?? ""}
				onTryAgain={refetchSpec}
			/>
		</div>
	)
}

CreateSource.displayName = "CreateSource"

export default CreateSource
