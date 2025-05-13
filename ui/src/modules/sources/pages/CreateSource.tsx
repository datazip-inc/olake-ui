import { useState, useEffect, forwardRef, useImperativeHandle } from "react"
import { Link, useNavigate } from "react-router-dom"
import { Radio, Select, Spin } from "antd"
import { useAppStore } from "../../../store"
import {
	ArrowLeft,
	ArrowRight,
	GenderNeuter,
	Notebook,
} from "@phosphor-icons/react"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import EntitySavedModal from "../../common/Modals/EntitySavedModal"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import EntityCancelModal from "../../common/Modals/EntityCancelModal"
import StepTitle from "../../common/components/StepTitle"
import FixedSchemaForm, { validateFormData } from "../../../utils/FormFix"
import { sourceService } from "../../../api/services/sourceService"
import { getConnectorImage } from "../../../utils/utils"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"

type SetupType = "new" | "existing"

interface SourceConfig {
	name: string
	type: string
	config?: any
	version?: string
}

interface ConnectorOption {
	value: string
	label: React.ReactNode
}

interface Source {
	id: string | number
	name: string
	type: string
	version: string
	config?: any
}

interface CreateSourceProps {
	fromJobFlow?: boolean
	fromJobEditFlow?: boolean
	existingSourceId?: string
	onComplete?: () => void
	stepNumber?: string
	stepTitle?: string
	initialConfig?: SourceConfig
	initialFormData?: any
	onSourceNameChange?: (name: string) => void
	onConnectorChange?: (connector: string) => void
	onFormDataChange?: (formData: any) => void
	onVersionChange?: (version: string) => void
}

// Create ref handle interface
export interface CreateSourceHandle {
	validateSource: () => Promise<boolean>
}

interface FormFieldProps {
	label: string
	required?: boolean
	children: React.ReactNode
	error?: string | null
}

const FormField = ({ label, required, children, error }: FormFieldProps) => (
	<div className="w-full">
		<label className="mb-2 block text-sm font-medium text-gray-700">
			{label}
			{required && <span className="text-red-500">*</span>}
		</label>
		{children}
		{error && <div className="mt-1 text-sm text-red-500">{error}</div>}
	</div>
)

interface EndpointTitleProps {
	title?: string
}

const EndpointTitle = ({ title = "Endpoint config" }: EndpointTitleProps) => (
	<div className="mb-4 flex items-center gap-1">
		<div className="mb-2 flex items-center gap-2">
			<GenderNeuter className="size-5" />
			<div className="text-base font-medium">{title}</div>
		</div>
	</div>
)

const CreateSource = forwardRef<CreateSourceHandle, CreateSourceProps>(
	(
		{
			fromJobFlow = false,
			fromJobEditFlow = false,
			existingSourceId,
			onComplete,
			stepNumber,
			stepTitle,
			initialConfig,
			initialFormData,
			onSourceNameChange,
			onConnectorChange,
			onFormDataChange,
			onVersionChange,
		},
		ref,
	) => {
		const [setupType, setSetupType] = useState<SetupType>("new")
		const [connector, setConnector] = useState("MongoDB")
		const [sourceName, setSourceName] = useState("")
		const [selectedVersion, setSelectedVersion] = useState("")
		const [versions, setVersions] = useState<string[]>([])
		const [loadingVersions, setLoadingVersions] = useState(false)
		const [formData, setFormData] = useState<any>({})
		const [schema, setSchema] = useState<any>(null)
		const [loading, setLoading] = useState(false)
		const [isDocPanelCollapsed, setIsDocPanelCollapsed] = useState(false)
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
		} = useAppStore()

		const connectorOptions: ConnectorOption[] = [
			{
				value: "MongoDB",
				label: (
					<div className="flex items-center">
						<img
							src={getConnectorImage("MongoDB")}
							alt="MongoDB"
							className="mr-2 size-5"
						/>
						<span>MongoDB</span>
					</div>
				),
			},
			{
				value: "Postgres",
				label: (
					<div className="flex items-center">
						<img
							src={getConnectorImage("Postgres")}
							alt="Postgres"
							className="mr-2 size-5"
						/>
						<span>Postgres</span>
					</div>
				),
			},
			{
				value: "MySQL",
				label: (
					<div className="flex items-center">
						<img
							src={getConnectorImage("MySQL")}
							alt="MySQL"
							className="mr-2 size-5"
						/>
						<span>MySQL</span>
					</div>
				),
			},
		]

		useEffect(() => {
			if (!sources.length) {
				fetchSources()
			}
		}, [sources.length, fetchSources])

		useEffect(() => {
			if (initialConfig) {
				setSourceName(initialConfig.name)
				setConnector(initialConfig.type)
				setFormData(initialConfig.config || {})
			}
		}, [initialConfig])

		useEffect(() => {
			if (initialFormData) {
				setFormData(initialFormData)
			}
		}, [initialFormData])

		useEffect(() => {
			if (fromJobEditFlow && existingSourceId) {
				setSetupType("existing")

				const selectedSource = sources.find(
					s => String(s.id) === existingSourceId,
				)

				if (selectedSource) {
					setSourceName(selectedSource.name)
					setConnector(selectedSource.type)
					setSelectedVersion(selectedSource.version)
				}
			}
		}, [fromJobEditFlow, existingSourceId, sources])

		useEffect(() => {
			if (setupType === "existing") {
				setFilteredSources(
					sources.filter(source => source.type === connector.toLowerCase()),
				)
			}
		}, [connector, setupType, sources])

		useEffect(() => {
			const fetchVersions = async () => {
				setLoadingVersions(true)
				try {
					const response = await sourceService.getSourceVersions(
						connector.toLowerCase(),
					)
					if (response.data && response.data.version) {
						setVersions(response.data.version)
						if (response.data.version.length > 0) {
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
		}, [connector, onVersionChange])

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
		}, [connector, selectedVersion])

		const handleCancel = () => {
			setShowSourceCancelModal(true)
		}

		const validateSource = async (): Promise<boolean> => {
			setValidating(true)

			let isValid = true

			if (setupType === "new") {
				if (!sourceName.trim()) {
					setSourceNameError("Source name is required")
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
				if (testResult.success) {
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
			setSourceName(newName)

			if (onSourceNameChange) {
				onSourceNameChange(newName)
			}
		}

		const handleConnectorChange = (value: string) => {
			setConnector(value)
			if (onConnectorChange) {
				onConnectorChange(value)
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
				setConnector(selectedSource.type)
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

		const toggleDocPanel = () => {
			setIsDocPanelCollapsed(!isDocPanelCollapsed)
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
						disabled={fromJobEditFlow}
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
							Version:
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
							className={`h-8 w-full rounded-[6px] border ${sourceNameError ? "border-red-500" : "border-gray-300"} px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500`}
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
						{fromJobEditFlow ? "Source:" : "Select existing source:"}
					</label>
					<Select
						placeholder="Select a source"
						className="w-full"
						onChange={handleExistingSourceSelect}
						value={fromJobEditFlow ? existingSourceId : undefined}
						disabled={fromJobEditFlow}
						options={filteredSources.map(s => ({
							value: s.id,
							label: s.name,
						}))}
					/>
				</div>
			</div>
		)

		const renderSetupTypeSelector = () =>
			!fromJobEditFlow && (
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
							Set up a new source
						</Radio>
						<Radio value="existing">Use an existing source</Radio>
					</Radio.Group>
				</div>
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
			<div className={`flex h-screen flex-col ${fromJobFlow ? "pb-32" : ""}`}>
				{!fromJobFlow && (
					<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
						<Link
							to={"/sources"}
							className="flex items-center gap-2 p-1.5 hover:rounded-[6px] hover:bg-[#f6f6f6] hover:text-black"
						>
							<ArrowLeft className="mr-1 size-5" />
						</Link>
						<div className="text-lg font-bold">Create source</div>
					</div>
				)}

				<div className="flex flex-1 overflow-hidden">
					<div className="w-full overflow-auto p-6 pt-0">
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

								{setupType === "new" && !fromJobEditFlow
									? renderNewSourceForm()
									: renderExistingSourceForm()}
							</div>
						</div>

						{renderSchemaForm()}
					</div>

					<DocumentationPanel
						docUrl={`https://olake.io/docs/connectors/${connector.toLowerCase()}/config`}
						isMinimized={isDocPanelCollapsed}
						onToggle={toggleDocPanel}
						showResizer={true}
					/>
				</div>

				{/* Footer */}
				{!fromJobFlow && !fromJobEditFlow && (
					<div className="flex justify-between border-t border-gray-200 bg-white p-4 shadow-sm">
						<button
							onClick={handleCancel}
							className="rounded-[6px] border border-[#F5222D] px-4 py-2 text-[#F5222D] transition-colors duration-200 hover:bg-[#F5222D] hover:text-white"
						>
							Cancel
						</button>
						<button
							className="flex items-center justify-center gap-1 rounded-[6px] bg-[#203FDD] px-4 py-2 font-light text-white shadow-sm transition-colors duration-200 hover:bg-[#132685]"
							onClick={handleCreate}
						>
							Create
							<ArrowRight className="size-4 text-white" />
						</button>
					</div>
				)}

				<TestConnectionModal />
				<TestConnectionSuccessModal />
				<TestConnectionFailureModal />
				<EntitySavedModal
					type="source"
					onComplete={onComplete}
					fromJobFlow={fromJobFlow || false}
					entityName={sourceName}
				/>
				<EntityCancelModal
					type="source"
					navigateTo={
						fromJobEditFlow ? "jobs" : fromJobFlow ? "jobs/new" : "sources"
					}
				/>
			</div>
		)
	},
)

CreateSource.displayName = "CreateSource"

export default CreateSource
