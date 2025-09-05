import { useState, useEffect, useRef } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { formatDistanceToNow } from "date-fns"
import { Input, Button, Select, Switch, message, Table, Spin } from "antd"
import type { ColumnsType } from "antd/es/table"
import {
	GenderNeuter,
	Notebook,
	ArrowLeft,
	PencilSimple,
	Info,
} from "@phosphor-icons/react"
import Form from "@rjsf/antd"
import validator from "@rjsf/validator-ajv8"

import { useAppStore } from "../../../store"
import { sourceService, jobService } from "../../../api"
import { Entity, SourceEditProps, SourceJob } from "../../../types"
import {
	getConnectorImage,
	getConnectorInLowerCase,
	getStatusClass,
	getStatusLabel,
} from "../../../utils/utils"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import StepTitle from "../../common/components/StepTitle"
import DeleteModal from "../../common/Modals/DeleteModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import EntityEditModal from "../../common/Modals/EntityEditModal"
import connectorOptions from "../components/connectorOptions"
import { getStatusIcon } from "../../../utils/statusIcons"
import {
	connectorTypeMap,
	DISPLAYED_JOBS_COUNT,
} from "../../../utils/constants"
import ObjectFieldTemplate from "../../common/components/Form/ObjectFieldTemplate"
import CustomFieldTemplate from "../../common/components/Form/CustomFieldTemplate"
import ArrayFieldTemplate from "../../common/components/Form/ArrayFieldTemplate"
import { widgets } from "../../common/components/Form/widgets"

const SourceEdit: React.FC<SourceEditProps> = ({
	fromJobFlow = false,
	stepNumber,
	stepTitle,
	initialData,
	onNameChange,
	onConnectorChange,
	onVersionChange,
	docsMinimized = false,
	onDocsMinimizedChange,
}) => {
	const formRef = useRef<any>(null)
	const { sourceId } = useParams<{ sourceId: string }>()
	const navigate = useNavigate()
	const [activeTab, setActiveTab] = useState("config")
	const [connector, setConnector] = useState<string | null>(null)
	const [selectedVersion, setSelectedVersion] = useState("")
	const [availableVersions, setAvailableVersions] = useState<
		{ label: string; value: string }[]
	>([])
	const [sourceName, setSourceName] = useState("")
	const [showAllJobs, setShowAllJobs] = useState(false)
	const [formData, setFormData] = useState<Record<string, any>>({})
	const { setShowDeleteModal, setSelectedSource } = useAppStore()
	const [source, setSource] = useState<Entity | null>(null)
	const [loading, setLoading] = useState(false)
	const [loadingVersions, setLoadingVersions] = useState(false)
	const [schema, setSchema] = useState<any>(null)
	const [uiSchema, setUiSchema] = useState<any>(null)

	const {
		sources,
		fetchSources,
		updateSource,
		setShowEditSourceModal,
		setShowTestingModal,
		setShowSuccessModal,
		setShowFailureModal,
		setSourceTestConnectionError,
	} = useAppStore()

	useEffect(() => {
		fetchSources()
	}, [])

	useEffect(() => {
		if (sourceId) {
			const source = sources.find(s => s.id?.toString() === sourceId)
			if (source) {
				setSource(source)
				setSourceName(source.name)
				let normalizedType =
					connectorTypeMap[source.type.toLowerCase()] || source.type
				setConnector(normalizedType)
				setSelectedVersion(source.version)
				setFormData(
					typeof source.config === "string"
						? JSON.parse(source.config)
						: source.config,
				)
			} else {
				navigate("/sources")
			}
		}
	}, [sourceId, sources, fetchSources])

	useEffect(() => {
		if (initialData) {
			setSourceName(initialData.name || "")
			const connectorTypeMap: Record<string, string> = {
				mongodb: "MongoDB",
				postgres: "Postgres",
				mysql: "MySQL",
				oracle: "Oracle",
			}
			let normalizedType =
				connectorTypeMap[initialData.type.toLowerCase()] || initialData.type

			// Only set connector if it's not already set or if it's the same as initialData
			if (!connector || connector === normalizedType) {
				setConnector(normalizedType)
				setSelectedVersion(initialData.version || "")

				// Set form data from initialData only if connector matches
				if (initialData.config) {
					if (typeof initialData.config === "string") {
						try {
							const parsedConfig = JSON.parse(initialData.config)
							setFormData(parsedConfig)
						} catch (error) {
							console.error("Error parsing source config:", error)
							setFormData({})
						}
					} else {
						setFormData(initialData.config)
					}
				}
			}
		}
	}, [initialData])

	useEffect(() => {
		if (!selectedVersion || !connector) {
			setSchema(null)
			return
		}

		const fetchSourceSpec = async () => {
			try {
				setLoading(true)
				const response = await sourceService.getSourceSpec(
					connector as string,
					selectedVersion,
				)
				if (response.success && response.data?.jsonschema) {
					setSchema(response.data.jsonschema)
					if (typeof response.data.uischema === "string") {
						setUiSchema(JSON.parse(response.data.uischema))
					}
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

		return () => {
			setLoading(false)
		}
	}, [connector, selectedVersion])

	const resetVersionState = () => {
		setAvailableVersions([])
		setSelectedVersion("")
		setSchema(null)
		if (onVersionChange) {
			onVersionChange("")
		}
	}

	useEffect(() => {
		const fetchVersions = async () => {
			if (!connector) return
			setLoadingVersions(true)
			try {
				const response = await sourceService.getSourceVersions(
					getConnectorInLowerCase(connector),
				)
				if (response.success && response.data?.version) {
					const versions = response.data.version.map((version: string) => ({
						label: version,
						value: version,
					}))
					setAvailableVersions([...versions])
					if (
						source?.type !== getConnectorInLowerCase(connector) &&
						versions.length > 0 &&
						!initialData
					) {
						setSelectedVersion(versions[0].value)
						if (onVersionChange) {
							onVersionChange(versions[0].value)
						}
					} else if (initialData) {
						if (
							initialData?.type != getConnectorInLowerCase(connector) &&
							initialData.version
						) {
							setSelectedVersion(initialData.version)
							if (onVersionChange) {
								onVersionChange(initialData.version)
							}
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
	}, [connector])

	const transformJobs = (jobs: any[]): SourceJob[] => {
		return jobs.map(job => ({
			id: job.id,
			name: job.name || job.job_name,
			destination_type: job.destination_type || "",
			destination_name: job.destination_name || "",
			last_run_time: job.last_runtime || job.last_run_time || "-",
			last_run_state: job.last_run_state || "-",
			activate: job.activate || false,
		}))
	}

	const displayedJobs = showAllJobs
		? transformJobs(source?.jobs || [])
		: transformJobs((source?.jobs || []).slice(0, DISPLAYED_JOBS_COUNT))

	const getSourceData = () => {
		const configStr =
			typeof formData === "string" ? formData : JSON.stringify(formData)

		const sourceData = {
			id: source?.id || 0,
			name: sourceName,
			type: connector || "MongoDB",
			version: selectedVersion,
			status: "active" as const,
			config: configStr,
			created_at: source?.created_at || new Date().toISOString(),
			updated_at: source?.updated_at || new Date().toISOString(),
			created_by: source?.created_by || "",
			updated_by: source?.updated_by || "",
			jobs: source?.jobs || [],
		}
		return sourceData
	}

	const handleSave = async () => {
		if (!source) return

		if (displayedJobs.length > 0) {
			setSelectedSource(getSourceData())
			setShowEditSourceModal(true)
			return
		}

		setShowTestingModal(true)
		const testResult = await sourceService.testSourceConnection(getSourceData())
		if (testResult.data?.status === "SUCCEEDED") {
			setTimeout(() => {
				setShowTestingModal(false)
			}, 1000)

			setTimeout(() => {
				setShowSuccessModal(true)
			}, 1200)

			setTimeout(() => {
				setShowSuccessModal(false)
				saveSource()
			}, 2200)
		} else {
			setShowTestingModal(false)
			setSourceTestConnectionError(testResult.data?.message || "")
			setShowFailureModal(true)
		}
	}

	const saveSource = () => {
		if (sourceId) {
			updateSource(sourceId, getSourceData())
				.then(() => {
					message.success("Source updated successfully")
					navigate("/sources")
				})
				.catch(error => {
					message.error("Failed to update source")
					console.error(error)
				})
		}
	}

	const handleDelete = () => {
		if (!source) return

		const sourceToDelete = {
			...source,
			name: sourceName || source.name,
			type: connector || source.type,
		}

		setSelectedSource(sourceToDelete)
		setShowDeleteModal(true)
	}

	const handleViewAllJobs = () => {
		setShowAllJobs(true)
	}

	const handlePauseJob = async (jobId: string, checked: boolean) => {
		try {
			await jobService.activateJob(jobId, !checked)
			message.success(
				`Successfully ${checked ? "paused" : "resumed"} job ${jobId}`,
			)
			// Refetch sources to update the UI with the latest source details
			await fetchSources()
		} catch (error) {
			console.error("Error toggling job status:", error)
			message.error(`Failed to ${checked ? "pause" : "resume"} job ${jobId}`)
		}
	}

	// const handlePauseAllJobs = async (checked: boolean) => {
	// 	try {
	// 		// We're working with a custom job format, so we need to extract IDs
	// 		const allJobs = displayedJobs.map(job => String(job.id))
	// 		await Promise.all(
	// 			allJobs.map(jobId => jobService.activateJob(jobId, !checked)),
	// 		)
	// 		message.success(`Successfully ${checked ? "paused" : "resumed"} all jobs`)
	// 	} catch (error) {
	// 		console.error("Error toggling all jobs status:", error)
	// 		message.error(`Failed to ${checked ? "pause" : "resume"} all jobs`)
	// 	}
	// }

	const toggleDocsPanel = () => {
		if (onDocsMinimizedChange) {
			onDocsMinimizedChange(prev => !prev)
		}
	}

	const columns: ColumnsType<SourceJob> = [
		{
			title: "Name",
			dataIndex: "name",
			key: "name",
		},
		{
			title: "State",
			dataIndex: "activate",
			key: "activate",
			render: (activate: boolean) => (
				<span
					className={`rounded px-2 py-1 text-xs ${
						!activate
							? "bg-danger-light text-danger"
							: "bg-primary-200 text-primary-700"
					}`}
				>
					{activate ? "Active" : "Inactive"}
				</span>
			),
		},
		{
			title: "Last runtime",
			dataIndex: "last_run_time",
			render: (text: string) => {
				if (text != "-") {
					return formatDistanceToNow(new Date(text), { addSuffix: true })
				}
				return "-"
			},
		},
		{
			title: "Last runtime status",
			dataIndex: "last_run_state",
			key: "last_run_state",
			render: (last_run_state: string) => (
				<div
					className={`flex w-fit items-center justify-center gap-1 rounded-md px-4 py-1 ${getStatusClass(last_run_state)}`}
				>
					{getStatusIcon(last_run_state.toLowerCase())}
					<span>{getStatusLabel(last_run_state.toLowerCase())}</span>
				</div>
			),
		},
		{
			title: "Destination",
			dataIndex: "destination_name",
			key: "destination_name",
			render: (destination_name: string, record: SourceJob) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(record.destination_type || "")}
						alt={record.destination_type || ""}
						className="mr-2 size-6"
					/>
					{destination_name}
				</div>
			),
		},
		{
			title: "Running status",
			dataIndex: "activate",
			key: "pause",
			render: (activate: boolean, record: SourceJob) => (
				<Switch
					checked={activate}
					onChange={checked => handlePauseJob(record.id.toString(), !checked)}
					className={activate ? "bg-blue-600" : "bg-gray-200"}
				/>
			),
		},
	]

	return (
		<div className="flex h-screen">
			<div className="flex flex-1 flex-col">
				{!fromJobFlow && (
					<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
						<Link
							to="/sources"
							className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
						>
							<ArrowLeft className="size-5" />
						</Link>
						<div className="text-lg font-bold">{sourceName}</div>
					</div>
				)}

				<div className="flex flex-1 overflow-hidden">
					<div className="flex flex-1 flex-col">
						<div className="flex-1 overflow-auto p-6 pt-0">
							{fromJobFlow && stepNumber && stepTitle && (
								<div className="mb-4">
									<div className="flex items-center justify-between">
										<StepTitle
											stepNumber={stepNumber}
											stepTitle={stepTitle}
										/>
										<Link
											to={
												sourceId
													? `/sources/${sourceId}`
													: `/sources/${sources.find(s => s.name === sourceName)?.id || ""}`
											}
											className="flex items-center gap-2 rounded-md bg-primary px-4 py-2 text-white hover:bg-primary-600"
										>
											<PencilSimple className="size-4" />
											Edit Source
										</Link>
									</div>
								</div>
							)}

							{!fromJobFlow && (
								<div className="mb-4">
									<div className="mt-2 flex w-fit rounded-md bg-background-primary p-1">
										<button
											className={`mr-1 w-56 rounded-md px-3 py-1.5 text-center text-sm font-normal ${
												activeTab === "config"
													? "bg-primary text-neutral-light"
													: "bg-background-primary text-text-primary"
											}`}
											onClick={() => setActiveTab("config")}
										>
											Config
										</button>

										<button
											className={`mr-1 w-56 rounded-md px-3 py-1.5 text-center text-sm font-normal ${
												activeTab === "jobs"
													? "bg-primary text-neutral-light"
													: "bg-background-primary text-text-primary"
											}`}
											onClick={() => setActiveTab("jobs")}
										>
											Associated jobs
										</button>
									</div>
								</div>
							)}

							{activeTab === "config" ? (
								<div className="bg-white">
									<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
										<div className="mb-4 flex items-center gap-1 text-lg font-medium">
											<Notebook className="size-5" />
											Capture information
										</div>

										<div className="grid grid-cols-2 gap-6">
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Connector:
												</label>
												<div className="flex items-center">
													<Select
														value={connector}
														onChange={value => {
															setConnector(value)
															setFormData({})
															setSchema(null)
															if (onConnectorChange) {
																onConnectorChange(value)
															}
														}}
														className="h-8 w-full"
														options={connectorOptions}
														disabled={fromJobFlow}
													/>
												</div>
											</div>

											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Name of your source:
													<span className="text-red-500">*</span>
												</label>
												<Input
													placeholder="Enter the name of your source"
													value={sourceName}
													onChange={e => {
														setSourceName(e.target.value)
														if (onNameChange) {
															onNameChange(e.target.value)
														}
													}}
													className="h-8"
													disabled={fromJobFlow}
												/>
											</div>

											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													OLake Version:
													<span className="text-red-500">*</span>
												</label>
												{loadingVersions ? (
													<div className="flex h-8 items-center justify-center">
														<Spin size="small" />
													</div>
												) : availableVersions.length > 0 ? (
													<Select
														value={selectedVersion}
														onChange={value => {
															setSelectedVersion(value)
															if (onVersionChange) {
																onVersionChange(value)
															}
														}}
														disabled={fromJobFlow}
														className="h-8 w-full"
														options={availableVersions}
													/>
												) : (
													<div className="flex items-center gap-1 text-sm text-red-500">
														<Info />
														No versions available
													</div>
												)}
											</div>
										</div>
									</div>

									<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
										<div className="mb-2 flex items-center gap-1">
											<GenderNeuter className="size-6" />
											<div className="text-lg font-medium">Endpoint config</div>
										</div>
										{loading ? (
											<div className="flex h-32 items-center justify-center">
												<Spin tip="Loading schema..." />
											</div>
										) : (
											schema && (
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
													onChange={e => setFormData(e.formData)}
													onSubmit={() => handleSave()}
													uiSchema={uiSchema}
													validator={validator}
													disabled={fromJobFlow}
													showErrorList={false}
													omitExtraData
													liveOmit
												/>
											)
										)}
									</div>
								</div>
							) : (
								<div className="rounded-lg p-6">
									<h3 className="mb-4 text-lg font-medium">Associated jobs</h3>

									<Table
										columns={columns}
										dataSource={displayedJobs}
										pagination={false}
										rowKey={record => record.id}
										className="min-w-full"
										rowClassName={() => "custom-row"}
									/>

									{!showAllJobs && source?.jobs && source.jobs.length > 5 && (
										<div className="mt-6 flex justify-center">
											<Button
												type="default"
												onClick={handleViewAllJobs}
												className="w-full border-none bg-primary-100 font-medium text-primary"
											>
												View all associated jobs
											</Button>
										</div>
									)}
								</div>
							)}
						</div>

						{/* Footer */}
						{!fromJobFlow && (
							<div className="flex justify-between border-t border-gray-200 bg-white p-4 shadow-sm">
								<div>
									<button
										className="ml-1 rounded-md border border-danger px-4 py-2 text-danger transition-colors duration-200 hover:bg-danger hover:text-white"
										onClick={handleDelete}
									>
										Delete
									</button>
								</div>
								<div className="flex space-x-4">
									<button
										className="mr-1 flex items-center justify-center gap-1 rounded-md bg-primary px-4 py-2 font-light text-white shadow-sm transition-colors duration-200 hover:bg-primary-600"
										onClick={() => {
											if (formRef.current) {
												formRef.current.submit()
											}
										}}
									>
										Save changes
									</button>
								</div>
							</div>
						)}
					</div>

					<DocumentationPanel
						docUrl={`https://olake.io/docs/connectors/${connector?.toLowerCase()}/config`}
						isMinimized={docsMinimized}
						onToggle={toggleDocsPanel}
						showResizer={true}
					/>
				</div>
			</div>

			<TestConnectionModal />
			<TestConnectionSuccessModal />
			<TestConnectionFailureModal fromSources={true} />
			<DeleteModal fromSource={true} />
			<EntityEditModal entityType="source" />
		</div>
	)
}

export default SourceEdit
