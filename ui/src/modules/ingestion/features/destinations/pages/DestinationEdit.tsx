import {
	ArrowLeftIcon,
	ArrowSquareOutIcon,
	InfoIcon,
	NotebookIcon,
} from "@phosphor-icons/react"
import Form from "@rjsf/antd"
import validator from "@rjsf/validator-ajv8"
import { useQueryClient } from "@tanstack/react-query"
import { Input, Button, Select, Switch, Spin, Table, Tooltip } from "antd"
import type { ColumnsType } from "antd/es/table"
import { formatDistanceToNow } from "date-fns"
import React, { useState, useEffect, useRef } from "react"
import { useParams, Link, useNavigate } from "react-router-dom"

import ArrayFieldTemplate from "@/common/components/form/ArrayFieldTemplate"
import CustomFieldTemplate from "@/common/components/form/CustomFieldTemplate"
import ObjectFieldTemplate from "@/common/components/form/ObjectFieldTemplate"
import { widgets } from "@/common/components/form/widgets"
import {
	ErrorLogsModal,
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
import {
	getStatusClass,
	getStatusLabel,
	handleSpecResponse,
} from "@/common/utils"
import { trimFormDataStrings } from "@/common/utils"
import {
	DeleteModal,
	EntityEditModal,
} from "@/modules/ingestion/common/components"
import { destinationConnectorOptions as connectorOptions } from "@/modules/ingestion/common/components/connectorOptions"
import DocumentationPanel from "@/modules/ingestion/common/components/DocumentationPanel"
import { getStatusIcon } from "@/modules/ingestion/common/components/statusIcons"
import {
	CONNECTOR_TYPES,
	DESTINATION_INTERNAL_TYPES,
	ENTITY_TYPES,
	DISPLAYED_JOBS_COUNT,
} from "@/modules/ingestion/common/constants"
import { Entity, EntityType } from "@/modules/ingestion/common/types"
import {
	getConnectorImage,
	getConnectorInLowerCase,
} from "@/modules/ingestion/common/utils"
import { TAB_TYPES } from "@/modules/ingestion/features/jobs/constants"
import { useActivateJob } from "@/modules/ingestion/features/jobs/hooks"

import { destinationKeys } from "../constants/queryKeys"
import {
	useDestinationDetails,
	useDestinationVersions,
	useDestinationSpec,
	useUpdateDestination,
	useDeleteDestination,
	useTestDestinationConnection,
} from "../hooks"
import { useDestinationStore } from "../stores"
import { DestinationJob } from "../types"
import { getConnectorDocumentationPath } from "../utils"

const DestinationEdit: React.FC = () => {
	// Local component state.
	const formRef = useRef<any>(null)
	const { destinationId } = useParams<{ destinationId: string }>()
	const [activeTab, setActiveTab] = useState(TAB_TYPES.CONFIG)
	const [connector, setConnector] = useState<string | null>(null)
	const [catalog, setCatalog] = useState<string | null>(null)
	const [destinationName, setDestinationName] = useState("")
	const [selectedVersion, setSelectedVersion] = useState("")
	const [showAllJobs, setShowAllJobs] = useState(false)
	const [schema, setSchema] = useState<any>(null)
	const [uiSchema, setUiSchema] = useState<any>(null)
	const [formData, setFormData] = useState<Record<string, any>>({})
	const [specError, setSpecError] = useState<string | null>(null)
	const [docsMinimized, setDocsMinimized] = useState(false)

	const [showEditModal, setShowEditModal] = useState(false)

	const { setSelectedDestination } = useDestinationStore()
	const [showDeleteModal, setShowDeleteModal] = useState(false)
	const [showTestingModal, setShowTestingModal] = useState(false)
	const [showSuccessModal, setShowSuccessModal] = useState(false)
	const [showFailureModal, setShowFailureModal] = useState(false)
	const [showSpecFailedModal, setShowSpecFailedModal] = useState(false)
	const [testConnectionError, setTestConnectionError] =
		useState<TestConnectionError | null>(null)

	// Data fetching and mutation hooks.
	const queryClient = useQueryClient()
	const { data: destination, isLoading: isLoadingDestination } =
		useDestinationDetails(destinationId ?? "")
	const internalConnectorType =
		destination?.type ?? getConnectorInLowerCase(connector)
	const { data: versionsData, isLoading: loadingVersions } =
		useDestinationVersions(internalConnectorType)
	const versions = versionsData?.version ?? []
	const updateDestinationMutation = useUpdateDestination(destinationId ?? "")
	const deleteDestinationMutation = useDeleteDestination()
	const testDestinationMutation = useTestDestinationConnection()
	const { mutate: activateJob } = useActivateJob()

	// Spec query for selected connector and version.
	const {
		data: specData,
		isLoading: loadingSpec,
		error: specQueryError,
		refetch: refetchSpec,
	} = useDestinationSpec(internalConnectorType, selectedVersion, "", "")

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

	const navigate = useNavigate()

	// Transform jobs to the format needed for our interface
	const transformJobs = (jobs: any[]): DestinationJob[] => {
		return jobs.map(job => ({
			id: job.id,
			name: job.name,
			source_type: job.source_type || "",
			source_name: job.source_name || "N/A",
			last_run_time: job.last_run_time || "-",
			last_run_state: job.last_run_state || "-",
			activate: job.activate || false,
			destination_name: job.destination_name || "",
			destination_type: job.destination_type || "",
		}))
	}

	const displayedJobs = showAllJobs
		? transformJobs(destination?.jobs || [])
		: transformJobs((destination?.jobs || []).slice(0, DISPLAYED_JOBS_COUNT))

	useEffect(() => {
		if (!destinationId) {
			navigate("/destinations")
		}
	}, [destinationId])

	useEffect(() => {
		if (destination && destinationId) {
			setDestinationName(destination.name)
			const connectorType =
				destination.type === DESTINATION_INTERNAL_TYPES.ICEBERG
					? CONNECTOR_TYPES.APACHE_ICEBERG
					: CONNECTOR_TYPES.AMAZON_S3
			setConnector(connectorType)
			setSelectedVersion(destination.version || "")

			const config =
				typeof destination.config === "string"
					? JSON.parse(destination.config)
					: destination.config
			setFormData(config)
		}
	}, [destination, destinationId])

	const handleVersionChange = (value: string) => {
		setSelectedVersion(value)
	}

	const getDestinationData = () => {
		const configStr =
			typeof formData === "string" ? formData : JSON.stringify(formData)

		const destinationData = {
			...(destination || {}),
			name: destinationName,
			type:
				connector === CONNECTOR_TYPES.APACHE_ICEBERG
					? DESTINATION_INTERNAL_TYPES.ICEBERG
					: DESTINATION_INTERNAL_TYPES.S3,
			version: selectedVersion,
			config: configStr,
		}
		return destinationData
	}

	const handleDelete = () => {
		if (!destination && !destinationId) return

		const destinationToDelete = destination
			? {
					...destination,
					name: destinationName || destination.name,
					type: connector || destination.type,
				}
			: {
					id: destinationId || "",
					name: destinationName || "",
					type: connector,
					jobs: [],
				}

		setSelectedDestination(destinationToDelete as Entity)
		setShowDeleteModal(true)
	}

	const handleConfirmEdit = async () => {
		setShowEditModal(false)
		setShowTestingModal(true)
		const testResult = await testDestinationMutation.mutateAsync({
			destination: getDestinationData(),
		})
		if (
			testResult.data?.connection_result.status ===
			TEST_CONNECTION_STATUS.SUCCEEDED
		) {
			setTimeout(() => {
				setShowTestingModal(false)
				setShowSuccessModal(true)
			}, 1000)
			setTimeout(() => {
				setShowSuccessModal(false)
				saveDestination()
			}, 2000)
		} else {
			setShowTestingModal(false)
			setTestConnectionError({
				message: testResult.data?.connection_result.message || "",
				logs: testResult.data?.logs || [],
			})
			setShowFailureModal(true)
		}
	}

	const handleSaveChanges = async () => {
		if (!destination && !destinationId) return

		if (displayedJobs.length > 0) {
			setShowEditModal(true)
			return
		}

		setShowTestingModal(true)
		const testResult = await testDestinationMutation.mutateAsync({
			destination: getDestinationData(),
		})
		if (
			testResult.data?.connection_result.status ===
			TEST_CONNECTION_STATUS.SUCCEEDED
		) {
			setTimeout(() => {
				setShowTestingModal(false)
				setShowSuccessModal(true)
			}, 1000)

			setTimeout(() => {
				setShowSuccessModal(false)
				saveDestination()
			}, 2000)
		} else {
			setShowTestingModal(false)
			setTestConnectionError({
				message: testResult.data?.connection_result.message || "",
				logs: testResult.data?.logs || [],
			})
			setShowFailureModal(true)
		}
	}

	const handleViewAllJobs = () => {
		setShowAllJobs(true)
	}

	const saveDestination = () => {
		if (destinationId) {
			updateDestinationMutation.mutate(getDestinationData() as any, {
				onSuccess: () => navigate("/destinations"),
				onError: error => console.error(error),
			})
		}
	}

	const handlePauseJob = (jobId: string, checked: boolean) => {
		activateJob(
			{ jobId, activate: !checked },
			{
				onSuccess: () => {
					if (destinationId) {
						queryClient.invalidateQueries({
							queryKey: destinationKeys.detail(destinationId),
						})
					}
				},
				onError: error => console.error("Error toggling job status:", error),
			},
		)
	}

	const toggleDocsPanel = () => {
		setDocsMinimized(prev => !prev)
	}

	const updateConnector = (value: string) => {
		setFormData({})
		setSchema(null)
		setUiSchema(null)
		setConnector(value)
	}

	const updateDestinationName = (value: string) => {
		setDestinationName(value)
	}

	const columns: ColumnsType<DestinationJob> = [
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
			key: "last_run_time",
			render: (text: string) => {
				return text !== "-"
					? formatDistanceToNow(new Date(text), { addSuffix: true })
					: "-"
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
			title: "Source",
			dataIndex: "source_name",
			key: "source_name",
			render: (source_name: string, record: DestinationJob) => (
				<div className="flex items-center">
					<img
						src={getConnectorImage(record.source_type || "")}
						alt={record.source_type || ""}
						className="mr-2 size-6"
					/>
					{source_name || "N/A"}
				</div>
			),
		},
		{
			title: "Running status",
			dataIndex: "activate",
			key: "pause",
			render: (activate: boolean, record: DestinationJob) => (
				<Switch
					checked={activate}
					onChange={checked => handlePauseJob(record.id.toString(), !checked)}
					className={activate ? "bg-blue-600" : "bg-gray-200"}
				/>
			),
		},
	]

	const renderConfigTab = () => (
		<div className="rounded-lg">
			<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
				<div className="mb-4 flex items-center gap-1 text-lg font-medium">
					<NotebookIcon className="size-5" />
					Capture information
				</div>

				<div className="flex flex-col gap-6">
					<div className="flex gap-12">
						<div className="w-1/2">
							<label className="mb-2 block text-sm font-medium text-gray-700">
								Connector:
							</label>
							<div className="flex items-center">
								<Select
									data-testid="destination-connector-select"
									value={connector}
									onChange={updateConnector}
									className="h-8 w-full"
									options={connectorOptions}
									disabled
								/>
							</div>
						</div>
						<div className="w-1/2">
							<label className="mb-2 flex items-center gap-1 text-sm font-medium text-gray-700">
								OLake Version:
								<Tooltip title="Choose the OLake version for the destination">
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
							) : versions.length > 0 ? (
								<Select
									value={selectedVersion}
									data-testid="destination-version-select"
									onChange={handleVersionChange}
									className="w-full"
									placeholder="Select version"
									options={versions.map(version => ({
										value: version,
										label: version,
									}))}
								/>
							) : (
								<div className="flex items-center gap-1 text-sm text-red-500">
									<InfoIcon />
									No versions available
								</div>
							)}
						</div>
					</div>

					<div className="flex w-full gap-6">
						<div className="w-1/2">
							<label className="mb-2 block text-sm font-medium text-gray-700">
								Name of your destination:
								<span className="text-red-500">*</span>
							</label>
							<Input
								placeholder="Enter the name of your destination"
								value={destinationName}
								onChange={e => updateDestinationName(e.target.value)}
								className="h-8"
								disabled
							/>
						</div>
					</div>
				</div>
			</div>

			<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
				<h3 className="mb-4 text-lg font-medium">Endpoint config</h3>
				{loadingSpec ? (
					<div className="flex h-32 items-center justify-center">
						<Spin tip="Loading schema..." />
					</div>
				) : schema ? (
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
							const catalogValue = trimmedData?.writer?.catalog_type
							if (catalogValue) setCatalog(catalogValue)
						}}
						transformErrors={transformErrors}
						onSubmit={handleSaveChanges}
						uiSchema={uiSchema}
						validator={validator}
						showErrorList={false}
						omitExtraData
						liveOmit
					/>
				) : null}
			</div>
		</div>
	)

	const renderJobsTab = () => (
		<div className="">
			<h3 className="mb-4 text-base font-medium">Associated jobs</h3>

			<Table
				columns={columns}
				dataSource={displayedJobs}
				pagination={false}
				rowKey={record => record.id}
				className="min-w-full"
				rowClassName="no-hover"
			/>

			{!showAllJobs && destination?.jobs && destination.jobs.length > 5 && (
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

			{/* <div className="mt-6 flex items-center justify-between rounded-xl border border-[#D9D9D9] p-4">
				<span className="font-medium">Pause all associated jobs</span>
				<Switch
					onChange={handlePauseAllJobs}
					className="bg-gray-200"
				/>
			</div> */}
		</div>
	)

	if (isLoadingDestination) {
		return (
			<div className="flex h-screen items-center justify-center">
				<Spin size="large" />
			</div>
		)
	}

	return (
		<div className="flex h-full">
			<div className="flex flex-1 flex-col">
				<div className="flex items-center gap-2 border-b border-[#D9D9D9] px-6 py-4">
					<Link
						to="/destinations"
						className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
					>
						<ArrowLeftIcon className="size-5" />
					</Link>
					<div className="text-lg font-bold">{destinationName}</div>
				</div>

				<div className="flex flex-1 overflow-hidden">
					<div className="flex flex-1 flex-col">
						<div className="flex-1 overflow-auto p-6 pt-0">
							<div className="mb-4 mt-2">
								<div className="flex w-fit rounded-md bg-background-primary p-1">
									<button
										className={`mr-1 w-56 rounded-md px-3 py-1.5 text-center text-sm font-normal ${
											activeTab === TAB_TYPES.CONFIG
												? "bg-primary text-neutral-light"
												: "bg-background-primary text-text-primary"
										}`}
										onClick={() => setActiveTab(TAB_TYPES.CONFIG)}
									>
										Config
									</button>
									<button
										className={`mr-1 w-56 rounded-md px-3 py-1.5 text-center text-sm font-normal ${
											activeTab === TAB_TYPES.JOBS
												? "bg-primary text-neutral-light"
												: "bg-background-primary text-text-primary"
										}`}
										onClick={() => setActiveTab(TAB_TYPES.JOBS)}
									>
										Associated jobs
									</button>
								</div>
							</div>

							{activeTab === TAB_TYPES.CONFIG
								? renderConfigTab()
								: renderJobsTab()}
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
								{activeTab === TAB_TYPES.CONFIG && (
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
						docUrl={`https://olake.io/docs/writers/${getConnectorDocumentationPath(connector || "", catalog ? catalog : "glue")}`}
						isMinimized={docsMinimized}
						onToggle={toggleDocsPanel}
						showResizer={true}
					/>
				</div>
			</div>

			<DeleteModal
				open={showDeleteModal}
				onClose={() => setShowDeleteModal(false)}
				entity={destination}
				fromSource={false}
				onDelete={() => {
					if (destination)
						deleteDestinationMutation.mutate(String(destination.id), {
							onSuccess: () => navigate("/destinations"),
						})
				}}
			/>
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
				onEdit={() => setShowFailureModal(false)}
				connectionType="destination"
				testConnectionError={testConnectionError}
			/>
			<EntityEditModal
				entityType={ENTITY_TYPES.DESTINATION as EntityType}
				open={showEditModal}
				jobs={displayedJobs}
				onConfirm={handleConfirmEdit}
				onCancel={() => setShowEditModal(false)}
			/>
			<ErrorLogsModal
				open={showSpecFailedModal}
				onClose={() => setShowSpecFailedModal(false)}
				title="Destination Spec Load Failed"
				error={specError ?? ""}
				onAction={() => {
					refetchSpec()
					setShowSpecFailedModal(false)
				}}
				actionButtonText="Try Again"
			/>
		</div>
	)
}

export default DestinationEdit
