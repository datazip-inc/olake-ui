import { useState, useEffect, useRef } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { formatDistanceToNow } from "date-fns"
import { Input, Button, Select, Switch, Table, Spin, Tooltip } from "antd"
import type { ColumnsType } from "antd/es/table"
import {
	GenderNeuterIcon,
	NotebookIcon,
	ArrowLeftIcon,
	InfoIcon,
	ArrowSquareOutIcon,
} from "@phosphor-icons/react"
import Form from "@rjsf/antd"
import validator from "@rjsf/validator-ajv8"

import { useSourceStore } from "../stores"
import { SourceEditProps, SourceJob } from "../types"
import type { TestConnectionError } from "@/common/types"
import { getConnectorLabel } from "../utils"
import {
	getConnectorImage,
	getStatusClass,
	getStatusLabel,
	handleSpecResponse,
	getConnectorInLowerCase,
} from "@/common/utils"
import { trimFormDataStrings } from "@/utils"
import {
	useSourceDetails,
	useSourceVersions,
	useSourceSpec,
} from "../hooks/queries/useSourceQueries"
import {
	useUpdateSource,
	useDeleteSource,
	useTestSourceConnection,
} from "../hooks/mutations/useSourceMutations"
import { useActivateJob } from "@/features/jobs/hooks/mutations/useJobMutations"
import { useQueryClient } from "@tanstack/react-query"
import { sourceKeys } from "../constants/queryKeys"
import DocumentationPanel from "@/common/components/DocumentationPanel"
import DeleteModal from "@/common/components/modals/DeleteModal"
import TestConnectionSuccessModal from "@/common/components/modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "@/common/components/modals/TestConnectionFailureModal"
import TestConnectionModal from "@/common/components/modals/TestConnectionModal"
import EntityEditModal from "@/common/components/modals/EntityEditModal"
import connectorOptions from "../components/connectorOptions"
import { getStatusIcon } from "@/common/components/statusIcons"
import {
	transformErrors,
	TEST_CONNECTION_STATUS,
} from "@/common/constants/constants"
import { DISPLAYED_JOBS_COUNT, OLAKE_LATEST_VERSION_URL } from "@/constants"
import ObjectFieldTemplate from "@/common/components/form/ObjectFieldTemplate"
import CustomFieldTemplate from "@/common/components/form/CustomFieldTemplate"
import ArrayFieldTemplate from "@/common/components/form/ArrayFieldTemplate"
import { widgets } from "@/common/components/form/widgets"
import SpecFailedModal from "@/common/components/modals/SpecFailedModal"

const SourceEdit: React.FC<SourceEditProps> = ({
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
	const [sourceName, setSourceName] = useState("")
	const [showAllJobs, setShowAllJobs] = useState(false)
	const [formData, setFormData] = useState<Record<string, any>>({})
	const { setSelectedSource } = useSourceStore()
	const [showDeleteModal, setShowDeleteModal] = useState(false)
	const [showEditModal, setShowEditModal] = useState(false)
	const [schema, setSchema] = useState<any>(null)
	const [uiSchema, setUiSchema] = useState<any>(null)
	const [specError, setSpecError] = useState<string | null>(null)
	const [testConnectionError, setTestConnectionError] =
		useState<TestConnectionError | null>(null)

	const normalizedSourceConnector = getConnectorInLowerCase(connector)

	const [showTestingModal, setShowTestingModal] = useState(false)
	const [showSuccessModal, setShowSuccessModal] = useState(false)
	const [showFailureModal, setShowFailureModal] = useState(false)
	const [showSpecFailedModal, setShowSpecFailedModal] = useState(false)
	const queryClient = useQueryClient()

	// TanStack Query hooks
	const { data: source, isLoading: isLoadingSource } = useSourceDetails(
		sourceId ?? "",
	)
	const { data: versionsData, isLoading: loadingVersions } = useSourceVersions(
		normalizedSourceConnector,
	)
	const availableVersions = (versionsData?.version ?? []).map((v: string) => ({
		label: v,
		value: v,
	}))

	const updateSourceMutation = useUpdateSource(sourceId ?? "")
	const deleteSourceMutation = useDeleteSource()
	const testSourceMutation = useTestSourceConnection()
	const { mutate: activateJob } = useActivateJob()

	useEffect(() => {
		if (!sourceId) {
			navigate("/sources")
		}
	}, [sourceId])

	// Initialize form when source is loaded from query
	useEffect(() => {
		if (source && sourceId) {
			setSourceName(source.name)
			const normalizedType = getConnectorLabel(source.type)
			setConnector(normalizedType)
			setSelectedVersion(source.version)
			setFormData(
				typeof source.config === "string"
					? JSON.parse(source.config)
					: source.config,
			)
		}
	}, [source, sourceId])

	useEffect(() => {
		if (initialData) {
			setSourceName(initialData.name || "")
			const normalizedType = getConnectorLabel(initialData.type)

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

	// Fetch spec via TanStack Query
	const {
		data: specData,
		isLoading: loadingSpec,
		error: specQueryError,
		refetch: refetchSpec,
	} = useSourceSpec(normalizedSourceConnector, selectedVersion)

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

	const transformJobs = (jobs: any[]): SourceJob[] => {
		return jobs.map(job => ({
			id: job.id,
			name: job.name,
			destination_type: job.destination_type || "",
			destination_name: job.destination_name || "",
			last_run_time: job.last_run_time || "-",
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
			type: getConnectorInLowerCase(connector || "MongoDB"),
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

	const handleConfirmEdit = async () => {
		setShowEditModal(false)
		setShowTestingModal(true)
		const testResult = await testSourceMutation.mutateAsync({
			source: getSourceData(),
		})
		if (
			testResult.data?.connection_result.status ===
			TEST_CONNECTION_STATUS.SUCCEEDED
		) {
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
			setTestConnectionError({
				message: testResult.data?.connection_result.message || "",
				logs: testResult.data?.logs || [],
			})
			setShowFailureModal(true)
		}
	}

	const handleSave = async () => {
		if (!source) return

		if (displayedJobs.length > 0) {
			setShowEditModal(true)
			return
		}

		setShowTestingModal(true)
		const testResult = await testSourceMutation.mutateAsync({
			source: getSourceData(),
		})
		if (
			testResult.data?.connection_result.status ===
			TEST_CONNECTION_STATUS.SUCCEEDED
		) {
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
			setTestConnectionError({
				message: testResult.data?.connection_result.message || "",
				logs: testResult.data?.logs || [],
			})
			setShowFailureModal(true)
		}
	}

	const saveSource = () => {
		if (sourceId) {
			updateSourceMutation.mutate(getSourceData(), {
				onSuccess: () => navigate("/sources"),
				onError: error => console.error(error),
			})
		}
	}

	const handleDelete = () => {
		if (!source) return

		const sourceToDelete = {
			...source,
			name: sourceName || source.name,
			type: getConnectorInLowerCase(connector || source.type),
		}

		setSelectedSource(sourceToDelete)
		setShowDeleteModal(true)
	}

	const handleViewAllJobs = () => {
		setShowAllJobs(true)
	}

	const handlePauseJob = (jobId: string, checked: boolean) => {
		activateJob(
			{ jobId, activate: !checked },
			{
				onSuccess: () => {
					if (sourceId) {
						queryClient.invalidateQueries({
							queryKey: sourceKeys.detail(sourceId),
						})
					}
				},
				onError: error => console.error("Error toggling job status:", error),
			},
		)
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

	if (isLoadingSource) {
		return (
			<div className="flex h-screen items-center justify-center">
				<Spin size="large" />
			</div>
		)
	}

	return (
		<div className="flex h-screen">
			<div className="flex flex-1 flex-col">
				<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
					<Link
						to="/sources"
						className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
					>
						<ArrowLeftIcon className="size-5" />
					</Link>
					<div className="text-lg font-bold">{sourceName}</div>
				</div>

				<div className="flex flex-1 overflow-hidden">
					<div className="flex flex-1 flex-col">
						<div className="flex-1 overflow-auto p-6 pt-0">
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

							{activeTab === "config" ? (
								<div className="bg-white">
									<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
										<div className="mb-4 flex items-center gap-1 text-lg font-medium">
											<NotebookIcon className="size-5" />
											Capture information
										</div>

										<div className="grid grid-cols-2 gap-6">
											<div>
												<label className="mb-2 block text-sm font-medium text-gray-700">
													Connector:
												</label>
												<div className="flex items-center">
													<Select
														data-testid="source-connector-select"
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
														disabled
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
													disabled
												/>
											</div>

											<div>
												<label className="mb-2 flex items-center gap-1 text-sm font-medium text-gray-700">
													OLake Version:
													<span className="text-red-500">*</span>
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
												) : availableVersions.length > 0 ? (
													<Select
														data-testid="source-version-select"
														value={selectedVersion}
														onChange={value => {
															setSelectedVersion(value)
															if (onVersionChange) {
																onVersionChange(value)
															}
														}}
														className="h-8 w-full"
														options={availableVersions}
													/>
												) : (
													<div className="flex items-center gap-1 text-sm text-red-500">
														<InfoIcon />
														No versions available
													</div>
												)}
											</div>
										</div>
									</div>

									<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
										<div className="mb-2 flex items-center gap-1">
											<GenderNeuterIcon className="size-6" />
											<div className="text-lg font-medium">Endpoint config</div>
										</div>
										{loadingSpec ? (
											<div className="flex h-32 items-center justify-center">
												<Spin tip="Loading schema..." />
											</div>
										) : (
											schema && (
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
													onSubmit={() => handleSave()}
													uiSchema={uiSchema}
													validator={validator}
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
								{activeTab === "config" && (
									<button
										className="mr-1 flex items-center justify-center gap-1 rounded-md bg-primary px-4 py-2 font-light text-white shadow-sm transition-colors duration-200"
										onClick={() => {
											if (formRef.current) {
												formRef.current.submit()
											}
										}}
									>
										Save changes
									</button>
								)}
							</div>
						</div>
					</div>

					<DocumentationPanel
						docUrl={`https://olake.io/docs/connectors/${getConnectorInLowerCase(connector)}`}
						isMinimized={docsMinimized}
						onToggle={toggleDocsPanel}
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
			<DeleteModal
				open={showDeleteModal}
				onClose={() => setShowDeleteModal(false)}
				entity={source}
				fromSource={true}
				onDelete={() => {
					if (source)
						deleteSourceMutation.mutate(String(source.id), {
							onSuccess: () => navigate("/sources"),
						})
				}}
			/>
			<EntityEditModal
				entityType="source"
				open={showEditModal}
				jobs={displayedJobs}
				onConfirm={handleConfirmEdit}
				onCancel={() => setShowEditModal(false)}
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

export default SourceEdit
