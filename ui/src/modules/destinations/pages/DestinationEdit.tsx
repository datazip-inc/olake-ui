import React, { useState, useEffect } from "react"
import { useParams, Link, useNavigate } from "react-router-dom"
import { Input, Button, Select, Switch, message, Spin, Table } from "antd"
import { useAppStore } from "../../../store"
import { ArrowLeft, Notebook } from "@phosphor-icons/react"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import FixedSchemaForm from "../../../utils/FormFix"
import { destinationService } from "../../../api/services/destinationService"
import { jobService } from "../../../api"
import StepTitle from "../../common/components/StepTitle"
import DeleteModal from "../../common/Modals/DeleteModal"
import {
	getCatalogInLowerCase,
	getConnectorImage,
	getConnectorName,
	getStatusClass,
	getStatusLabel,
} from "../../../utils/utils"
import { DestinationJob, Entity } from "../../../types"
import type { ColumnsType } from "antd/es/table"
import { formatDistanceToNow } from "date-fns"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import { connectorOptions } from "../components/connectorOptions"
import EntityEditModal from "../../common/Modals/EntityEditModal"
import { getStatusIcon } from "../../../utils/statusIcons"

interface CatalogOption {
	value: string
	label: string
}

interface DestinationEditProps {
	fromJobFlow?: boolean
	stepNumber?: string | number
	stepTitle?: string
	initialData?: any
	onNameChange?: (name: string) => void
	onConnectorChange?: (type: string) => void
	onVersionChange?: (version: string) => void
	onFormDataChange?: (config: Record<string, any>) => void
}

const DestinationEdit: React.FC<DestinationEditProps> = ({
	fromJobFlow = false,
	stepNumber,
	stepTitle,
	initialData,
	onNameChange,
	onConnectorChange,
	onVersionChange,
	onFormDataChange,
}) => {
	const { destinationId } = useParams<{ destinationId: string }>()
	const isNewDestination = destinationId === "new"
	const [activeTab, setActiveTab] = useState("config")
	const [connector, setConnector] = useState<string | null>(null)
	const [catalog, setCatalog] = useState<string | null>(null)
	const catalogName = "AWS Glue"
	const [destinationName, setDestinationName] = useState("")
	const [selectedVersion, setSelectedVersion] = useState("latest")
	const [versions, setVersions] = useState<string[]>([])
	const [loadingVersions, setLoadingVersions] = useState(false)
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [showAllJobs, setShowAllJobs] = useState(false)
	const [schema, setSchema] = useState<Record<string, any> | null>(null)
	const [uiSchema, setUiSchema] = useState<Record<string, any> | null>(null)
	const [formData, setFormData] = useState<Record<string, any>>({})
	const [isLoading, setIsLoading] = useState(false)
	const [destination, setDestination] = useState<Entity | null>(null)
	const [initialCatalog, setInitialCatalog] = useState<string | null>(null)
	const [initialFormData, setInitialFormData] = useState<Record<string, any>>(
		{},
	)

	const {
		destinations,
		fetchDestinations,
		setSelectedDestination,
		setShowDeleteModal,
		setShowEditDestinationModal,
		setShowTestingModal,
		setShowSuccessModal,
		setShowFailureModal,
		setDestinationTestConnectionError,
		updateDestination,
	} = useAppStore()

	const navigate = useNavigate()

	// Define catalog options
	const catalogOptions: CatalogOption[] = [
		{ value: "AWS Glue", label: "AWS Glue" },
		{ value: "REST Catalog", label: "REST Catalog" },
		{ value: "JDBC Catalog", label: "JDBC Catalog" },
		{ value: "HIVE Catalog", label: "Hive Catalog" },
	]

	// Transform jobs to the format needed for our interface
	const transformJobs = (jobs: any[]): DestinationJob[] => {
		return jobs.map(job => ({
			id: job.id,
			name: job.name || job.job_name,
			source_type: job.source_type || "",
			source_name: job.source_name || "N/A",
			last_run_time: job.last_runtime || job.last_run_time || "-",
			last_run_state: job.last_run_state || "-",
			activate: job.activate || false,
			job_name: job.job_name || job.name,
			destination_name: job.destination_name || "",
			destination_type: job.destination_type || "",
		}))
	}

	const displayedJobs = showAllJobs
		? transformJobs(destination?.jobs || [])
		: transformJobs((destination?.jobs || []).slice(0, 5))

	const getCatalogName = (catalogType: string) => {
		switch (catalogType?.toLowerCase()) {
			case "glue":
			case "aws glue":
				return "AWS Glue"
			case "rest":
			case "rest catalog":
				return "REST Catalog"
			case "jdbc":
			case "jdbc catalog":
				return "JDBC Catalog"
			case "hive":
			case "hive catalog":
				return "Hive Catalog"
			default:
				return null
		}
	}

	useEffect(() => {
		fetchDestinations()
	}, [])

	useEffect(() => {
		if (destinationId && destinationId !== "new") {
			const destination = destinations.find(
				d => d.id.toString() === destinationId,
			)
			if (destination) {
				setDestination(destination)
				setDestinationName(destination.name)
				const connectorType =
					destination.type === "iceberg" ? "Apache Iceberg" : "Amazon S3"
				setConnector(connectorType)
				setSelectedVersion(destination.version || "")

				const config =
					typeof destination.config === "string"
						? JSON.parse(destination.config)
						: destination.config

				setFormData(config)
				setInitialFormData(config)

				if (destination.type === "iceberg") {
					try {
						const catalogType = config.writer.catalog_type || "AWS Glue"
						setCatalog(getCatalogName(catalogType) || null)
					} catch (error) {
						console.error("Error parsing config for catalog:", error)
						setCatalog("AWS Glue")
					}
				} else {
					setInitialCatalog("s3")
				}
			}
		}
	}, [destinationId, destinations, fetchDestinations])

	useEffect(() => {
		if (initialData) {
			setDestinationName(initialData.name || "")
			let connectorType = initialData.type
			if (
				connectorType?.toLowerCase() === "s3" ||
				connectorType?.toLowerCase() === "amazon s3"
			) {
				connectorType = "Amazon S3"
			} else if (
				connectorType?.toLowerCase() === "iceberg" ||
				connectorType?.toLowerCase() === "apache iceberg"
			) {
				connectorType = "Apache Iceberg"
			}
			setConnector(connectorType)
			setSelectedVersion(initialData.version || "latest")
			if (initialData.config) {
				let parsedConfig = initialData.config
				if (typeof initialData.config === "string") {
					try {
						parsedConfig = JSON.parse(initialData.config)
					} catch (error) {
						console.error("Error parsing destination config:", error)
						parsedConfig = {}
					}
				}
				setFormData(parsedConfig)
				setInitialFormData(parsedConfig)
				if (connectorType === "Apache Iceberg") {
					let writerCatalogType = parsedConfig.writer.catalog_type
					setCatalog(getCatalogName(writerCatalogType) || "AWS Glue")
					setInitialCatalog(getCatalogName(writerCatalogType) || "AWS Glue")
				} else {
					setInitialCatalog("s3")
				}
			}
		}
	}, [initialData])

	useEffect(() => {
		const fetchVersions = async () => {
			if (!connector) return

			setLoadingVersions(true)
			try {
				let connectorType = connector
				if (connector === "Apache Iceberg") {
					connectorType = "iceberg"
				} else {
					connectorType = "s3"
				}

				const response = await destinationService.getDestinationVersions(
					connectorType.toLowerCase(),
				)

				if (response.data && response.data.version) {
					setVersions(response.data.version)

					// If no version is selected, set the first one as default
					if (!selectedVersion && response.data.version.length > 0) {
						setSelectedVersion(response.data.version[0])
						if (onVersionChange) {
							onVersionChange(response.data.version[0])
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
		if (!initialData) {
			if (connector === "Apache Iceberg") {
				setCatalog("AWS Glue")
			} else {
				setCatalog("None")
			}
		}
	}, [connector])

	useEffect(() => {
		const fetchDestinationSpec = async () => {
			if (!connector) return
			try {
				setIsLoading(true)
				const response = await destinationService.getDestinationSpec(
					connector,
					catalog,
					selectedVersion,
				)

				if (response.success && response.data?.spec) {
					setSchema(response.data.spec)
					if (response.data?.uiSchema) {
						setUiSchema(response.data.uiSchema)
					}
					if (initialCatalog) {
						if (
							initialFormData &&
							getCatalogInLowerCase(catalog || "") !=
								getCatalogInLowerCase(initialCatalog)
						) {
							setFormData({})
						} else {
							if (initialFormData) {
								setFormData(initialFormData)
							}
						}
					}
				} else {
					console.error("Failed to get destination spec:", response.message)
				}
			} catch (error) {
				console.error("Error fetching destination spec:", error)
			} finally {
				setIsLoading(false)
			}
		}

		fetchDestinationSpec()
	}, [connector, selectedVersion, catalog])

	const handleVersionChange = (value: string) => {
		setSelectedVersion(value)
		if (onVersionChange) {
			onVersionChange(value)
		}
	}

	const getDestinationData = () => {
		const configStr =
			typeof formData === "string" ? formData : JSON.stringify(formData)

		const destinationData = {
			...(destination || {}),
			name: destinationName,
			type: connector === "Apache Iceberg" ? "iceberg" : "s3",
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

	const handleSaveChanges = async () => {
		if (!destination && !destinationId) return

		if (displayedJobs.length > 0) {
			setSelectedDestination(getDestinationData() as Entity)
			setShowEditDestinationModal(true)
			return
		}

		setShowTestingModal(true)
		const testResult =
			await destinationService.testDestinationConnection(getDestinationData())
		if (testResult.data?.status === "SUCCEEDED") {
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
			setDestinationTestConnectionError(testResult.data?.message || "")
			setShowFailureModal(true)
		}
	}

	const handleViewAllJobs = () => {
		setShowAllJobs(true)
	}

	const saveDestination = () => {
		if (destinationId) {
			updateDestination(destinationId, getDestinationData())
				.then(() => {
					message.success("Destination updated successfully")
					navigate("/destinations")
				})
				.catch(error => {
					message.error("Failed to update source")
					console.error(error)
				})
		}
	}

	// const handlePauseAllJobs = async (checked: boolean) => {
	// 	try {
	// 		const allJobs = displayedJobs.map(job => job.id.toString())
	// 		await Promise.all(
	// 			allJobs.map(jobId => jobService.activateJob(jobId, !checked)),
	// 		)
	// 		message.success(`Successfully ${checked ? "paused" : "resumed"} all jobs`)
	// 	} catch (error) {
	// 		console.error("Error toggling all jobs status:", error)
	// 		message.error(`Failed to ${checked ? "pause" : "resume"} all jobs`)
	// 	}
	// }

	const handlePauseJob = async (jobId: string, checked: boolean) => {
		try {
			await jobService.activateJob(jobId, !checked)
			message.success(
				`Successfully ${checked ? "paused" : "resumed"} job ${jobId}`,
			)
			await fetchDestinations()
		} catch (error) {
			console.error("Error toggling job status:", error)
			message.error(`Failed to ${checked ? "pause" : "resume"} job ${jobId}`)
		}
	}

	const toggleDocsPanel = () => {
		setDocsMinimized(!docsMinimized)
	}

	const updateConnector = (value: string) => {
		setConnector(value)
		if (onConnectorChange) {
			onConnectorChange(value)
		}
	}

	const updateDestinationName = (value: string) => {
		setDestinationName(value)
		if (onNameChange) {
			onNameChange(value)
		}
	}

	const updateFormData = (data: Record<string, any>) => {
		setFormData(data)
		if (onFormDataChange) {
			onFormDataChange(data)
		}
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
							? "bg-[#FFF1F0] text-[#F5222D]"
							: "bg-[#E6F4FF] text-[#0958D9]"
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
					className={`flex w-fit items-center justify-center gap-1 rounded-[6px] px-4 py-1 ${getStatusClass(last_run_state)}`}
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
					<Notebook className="size-5" />
					Capture information
				</div>

				<div className="flex flex-col gap-6">
					<div className="flex gap-12">
						<div className="w-1/3">
							<label className="mb-2 block text-sm font-medium text-gray-700">
								Connector:
							</label>
							<div className="flex items-center">
								<Select
									value={connector}
									onChange={updateConnector}
									className="h-8 w-full"
									options={connectorOptions}
								/>
							</div>
						</div>

						<div className="w-1/3">
							<label className="mb-2 block text-sm font-medium text-gray-700">
								Catalog:
							</label>
							<Select
								className="h-8 w-full"
								placeholder="Select catalog"
								disabled={connector === "Amazon S3" || connector === "AWS S3"}
								options={catalogOptions}
								value={
									catalog ||
									(connector === "Amazon S3" || connector === "AWS S3"
										? "None"
										: undefined)
								}
								onChange={value => {
									setCatalog(value)
								}}
							/>
						</div>
					</div>

					<div className="flex w-full gap-12">
						<div className="w-1/3">
							<label className="mb-2 block text-sm font-medium text-gray-700">
								Name of your destination:
								<span className="text-red-500">*</span>
							</label>
							<Input
								placeholder="Enter the name of your destination"
								value={destinationName}
								onChange={e => updateDestinationName(e.target.value)}
								className="h-8"
							/>
						</div>

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
				</div>
			</div>

			<div className="mb-6 rounded-xl border border-[#D9D9D9] p-6">
				<h3 className="mb-4 text-lg font-medium">Endpoint config</h3>
				{isLoading ? (
					<div className="flex h-32 items-center justify-center">
						<Spin tip="Loading schema..." />
					</div>
				) : schema ? (
					<FixedSchemaForm
						schema={schema}
						formData={formData}
						onChange={updateFormData}
						hideSubmit={true}
						{...(uiSchema ? { uiSchema } : {})}
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
						className="w-full border-none bg-[#E9EBFC] font-medium text-[#203FDD]"
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

	return (
		<div className={`flex h-screen flex-col ${fromJobFlow ? "pb-32" : ""}`}>
			{/* Header */}
			{!fromJobFlow && (
				<div className="flex gap-2 px-6 pb-0 pt-6">
					<Link
						to="/destinations"
						className="mb-4 flex items-center gap-2 p-1.5 hover:rounded-[6px] hover:bg-[#f6f6f6] hover:text-black"
					>
						<ArrowLeft className="size-5" />
					</Link>

					<div className="mb-4 flex items-center">
						<h1 className="text-2xl font-bold">
							{isNewDestination
								? "Create New Destination"
								: destinationName || "<Destination_name>"}
						</h1>
					</div>
				</div>
			)}

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden border border-t border-[#D9D9D9]">
				{/* Left content */}
				<div
					className={`${
						docsMinimized ? "w-full" : "w-3/4"
					} mt-4 overflow-auto p-6 pt-0 transition-all duration-300`}
				>
					{fromJobFlow && stepNumber && stepTitle && (
						<div>
							<StepTitle
								stepNumber={stepNumber}
								stepTitle={stepTitle}
							/>
						</div>
					)}

					{!fromJobFlow && (
						<div className="mb-4">
							<div className="flex w-fit rounded-[6px] bg-[#f5f5f5] p-1">
								<button
									className={`w-56 rounded-[6px] px-3 py-1.5 text-sm font-normal ${
										activeTab === "config"
											? "mr-1 bg-[#203fdd] text-center text-[#F0F0F0]"
											: "mr-1 bg-[#F5F5F5] text-center text-[#0A0A0A]"
									}`}
									onClick={() => setActiveTab("config")}
								>
									Config
								</button>
								{!isNewDestination && (
									<button
										className={`w-56 rounded-[6px] px-3 py-1.5 text-sm font-normal ${
											activeTab === "jobs"
												? "mr-1 bg-[#203fdd] text-center text-[#F0F0F0]"
												: "mr-1 bg-[#F5F5F5] text-center text-[#0A0A0A]"
										}`}
										onClick={() => setActiveTab("jobs")}
									>
										Associated jobs
									</button>
								)}
							</div>
						</div>
					)}

					{activeTab === "config" ? renderConfigTab() : renderJobsTab()}
				</div>

				{/* Documentation panel */}
				<DocumentationPanel
					docUrl={`https://olake.io/docs/writers/${getConnectorName(connector || "", catalog ? catalog : catalogName)}`}
					isMinimized={docsMinimized}
					onToggle={toggleDocsPanel}
					showResizer={true}
				/>
			</div>
			{/* Delete Modal */}
			<DeleteModal fromSource={false} />
			<TestConnectionModal />
			<TestConnectionSuccessModal />
			<TestConnectionFailureModal />
			<EntityEditModal entityType="destination" />

			{/* Footer with buttons */}
			{!fromJobFlow && (
				<div className="flex justify-between border-t border-gray-200 bg-white p-4">
					<div>
						{!isNewDestination && (
							<button
								className="rounded-[6px] border border-[#F5222D] px-4 py-1 text-[#F5222D] hover:bg-[#F5222D] hover:text-white"
								onClick={handleDelete}
							>
								Delete
							</button>
						)}
					</div>
					<div className="flex space-x-4">
						<button
							className="flex items-center justify-center gap-1 rounded-[6px] bg-[#203FDD] px-4 py-1 font-light text-white hover:bg-[#132685]"
							onClick={handleSaveChanges}
						>
							Save Changes
						</button>
					</div>
				</div>
			)}
		</div>
	)
}

export default DestinationEdit
