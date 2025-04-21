import { useState, useEffect } from "react"
import { Link } from "react-router-dom"
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
import DynamicSchemaForm from "../../common/components/DynamicSchemaForm"
import { sourceService } from "../../../api/services/sourceService"
import { getConnectorImage } from "../../../utils/utils"

interface CreateSourceProps {
	fromJobFlow?: boolean
	fromJobEditFlow?: boolean
	existingSourceId?: string
	onComplete?: () => void
	stepNumber?: string
	stepTitle?: string
	initialConfig?: any
	initialFormData?: any
	onSourceNameChange?: (name: string) => void
	onConnectorChange?: (connector: string) => void
	onFormDataChange?: (formData: any) => void
}

// Custom components for section titles
const EndpointTitleComp = ({ title }: { title: string }) => (
	<div className="mb-4 flex items-center gap-1">
		<div className="mb-2 flex items-center gap-2">
			<GenderNeuter className="size-5" />
			<div className="text-base font-medium">{title || "Endpoint config"}</div>
		</div>
	</div>
)

const CreateSource: React.FC<CreateSourceProps> = ({
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
}) => {
	const [setupType, setSetupType] = useState("new")
	const [connector, setConnector] = useState("MongoDB")
	const [sourceName, setSourceName] = useState("")
	const [formData, setFormData] = useState<any>({})
	const [schema, setSchema] = useState<any>(null)
	const [loading, setLoading] = useState(false)
	const [isDocPanelCollapsed, setIsDocPanelCollapsed] = useState(false)
	const [filteredSources, setFilteredSources] = useState<any[]>([])

	const {
		sources,
		fetchSources,
		setShowEntitySavedModal,
		setShowTestingModal,
		setShowSuccessModal,
		setShowSourceCancelModal,
		addSource,
	} = useAppStore()

	useEffect(() => {
		if (!sources.length) {
			fetchSources()
		}
	}, [sources.length, fetchSources])

	// Initialize with initial config if provided
	useEffect(() => {
		if (initialConfig) {
			setSourceName(initialConfig.name)
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
		if (fromJobEditFlow && existingSourceId) {
			setSetupType("existing")

			const selectedSource = sources.find(s => s.id === existingSourceId)

			if (selectedSource) {
				setSourceName(selectedSource.name)
				setConnector(selectedSource.type)
			}
		}
	}, [fromJobEditFlow, existingSourceId, sources])

	useEffect(() => {
		if (setupType === "existing") {
			setFilteredSources(sources.filter(source => source.type === connector))
		}
	}, [connector, setupType, sources])

	// Fetch schema when connector changes
	useEffect(() => {
		const fetchSchema = async () => {
			try {
				setLoading(true)
				const schemaData = await sourceService.getConnectorSchema(connector)
				setSchema(schemaData)

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
	}, [connector, setupType])

	const handleCancel = () => {
		setShowSourceCancelModal(true)
	}

	const handleCreate = () => {
		// Add the new source to the store state
		const newSourceData = {
			name: sourceName,
			type: connector,
			status: "active" as const,
			config: formData,
		}

		addSource(newSourceData)
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
				console.error("Error adding source:", error)
			})
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
		const selectedSource = sources.find(s => s.id === value)

		if (selectedSource) {
			setSourceName(selectedSource.name)
			setConnector(selectedSource.type)
		}
	}

	const handleFormChange = (newFormData: any) => {
		setFormData(newFormData)
		if (onFormDataChange) {
			onFormDataChange(newFormData)
		}
	}

	const toggleDocPanel = () => {
		setIsDocPanelCollapsed(!isDocPanelCollapsed)
	}

	return (
		<div className="flex h-screen flex-col">
			{!fromJobFlow && (
				<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
					<Link
						to={"/sources"}
						className="flex items-center text-lg font-bold"
					>
						<ArrowLeft className="mr-1 size-6 font-bold" />
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
											Set up a new source
										</Radio>
										<Radio value="existing">Use an existing source</Radio>
									</Radio.Group>
								</div>
							)}

							{setupType === "new" && !fromJobEditFlow ? (
								<div className="flex-start flex w-full gap-6">
									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Connector:
										</label>
										<div className="flex items-center">
											<Select
												value={connector}
												onChange={handleConnectorChange}
												className="h-8 w-full"
												options={[
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
																	alt="PostgreSQL"
																	className="mr-2 size-5"
																/>
																<span>PostgreSQL</span>
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
												]}
												style={{ boxShadow: "0 1px 2px 0 rgba(0, 0, 0, 0.05)" }}
											/>
										</div>
									</div>

									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Name of your source:
										</label>
										<input
											type="text"
											className="h-8 w-full rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
											placeholder="Enter the name of your source"
											value={sourceName}
											onChange={handleSourceNameChange}
										/>
									</div>
								</div>
							) : (
								<div className="flex-start flex w-full gap-6">
									<div className="w-1/3">
										<label className="mb-2 block text-sm font-medium text-gray-700">
											Connector:
										</label>
										<div className="flex items-center">
											<Select
												value={connector}
												onChange={handleConnectorChange}
												className="w-full"
												disabled={fromJobEditFlow}
												options={[
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
														value: "PostgreSQL",
														label: (
															<div className="flex items-center">
																<img
																	src={getConnectorImage("Postgres")}
																	alt="PostgreSQL"
																	className="mr-2 size-5"
																/>
																<span>PostgreSQL</span>
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
												]}
												style={{ boxShadow: "0 1px 2px 0 rgba(0, 0, 0, 0.05)" }}
											/>
										</div>
									</div>

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
										<div className="mb-6 rounded-xl border border-gray-200 bg-white p-6 shadow-sm">
											<EndpointTitleComp title="Endpoint config" />
											<DynamicSchemaForm
												schema={schema}
												uiSchema={schema.uiSchema}
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

				<DocumentationPanel
					docUrl={`https://olake.io/docs/category/${connector.toLowerCase()}`}
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
						Next
						<ArrowRight className="size-4 text-white" />
					</button>
				</div>
			)}

			<TestConnectionModal />
			<TestConnectionSuccessModal />
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
}

export default CreateSource
