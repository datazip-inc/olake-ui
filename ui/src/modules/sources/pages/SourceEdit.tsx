import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Select, Switch, message, Table, Spin } from "antd"
import { GenderNeuter, Notebook, ArrowLeft } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import type { ColumnsType } from "antd/es/table"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import FixedSchemaForm from "../../../utils/FormFix"
import StepTitle from "../../common/components/StepTitle"
import DeleteModal from "../../common/Modals/DeleteModal"
import {
	getConnectorImage,
	getStatusClass,
	getStatusLabel,
} from "../../../utils/utils"
import { sourceService } from "../../../api"
import { formatDistanceToNow } from "date-fns"
import { jobService } from "../../../api"
import { Entity, SourceJob } from "../../../types"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import connectorOptions from "../components/connectorOptions"
import EntityEditModal from "../../common/Modals/EntityEditModal"
import { getStatusIcon } from "../../../utils/statusIcons"

interface SourceEditProps {
	fromJobFlow?: boolean
	stepNumber?: string | number
	stepTitle?: string
	initialData?: any
	onNameChange?: (name: string) => void
	onConnectorChange?: (type: string) => void
	onVersionChange?: (version: string) => void
	onFormDataChange?: (config: Record<string, any>) => void
}

const SourceEdit: React.FC<SourceEditProps> = ({
	fromJobFlow = false,
	stepNumber,
	stepTitle,
	initialData,
	onNameChange,
	onConnectorChange,
	onFormDataChange,
}) => {
	const { sourceId } = useParams<{ sourceId: string }>()
	const navigate = useNavigate()
	const [activeTab, setActiveTab] = useState("config")
	const [connector, setConnector] = useState<string | null>(null)
	const [selectedVersion, setSelectedVersion] = useState("")
	const [sourceName, setSourceName] = useState("")
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [showAllJobs, setShowAllJobs] = useState(false)
	const [formData, setFormData] = useState<Record<string, any>>({})
	const { setShowDeleteModal, setSelectedSource } = useAppStore()
	const [source, setSource] = useState<Entity | null>(null)
	const [loading, setLoading] = useState(false)
	const [schema, setSchema] = useState<Record<string, any> | null>(null)

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
				let normalizedType = source.type
				if (source.type.toLowerCase() === "mongodb") normalizedType = "MongoDB"
				if (source.type.toLowerCase() === "postgres")
					normalizedType = "Postgres"
				if (source.type.toLowerCase() === "mysql") normalizedType = "MySQL"
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
			let normalizedType = initialData.type
			if (initialData.type?.toLowerCase() === "mongodb")
				normalizedType = "MongoDB"
			if (initialData.type?.toLowerCase() === "postgres")
				normalizedType = "Postgres"
			if (initialData.type?.toLowerCase() === "mysql") normalizedType = "MySQL"
			setConnector(normalizedType)
			setSelectedVersion(initialData.version || "latest")

			// Set form data from initialData
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
	}, [initialData])

	useEffect(() => {
		const fetchSourceSpec = async () => {
			try {
				setLoading(true)
				const response = await sourceService.getSourceSpec(
					connector as string,
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

		if (connector) {
			fetchSourceSpec()
		}

		return () => {
			setLoading(false)
		}
	}, [connector, selectedVersion])

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
		: transformJobs((source?.jobs || []).slice(0, 5))

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
				setShowSuccessModal(true)
			}, 1000)

			setTimeout(() => {
				setShowSuccessModal(false)
				saveSource()
			}, 2000)
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
		setDocsMinimized(!docsMinimized)
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
					className={`flex w-fit items-center justify-center gap-1 rounded-[6px] px-4 py-1 ${getStatusClass(last_run_state)}`}
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
		<div className={`flex h-screen flex-col ${fromJobFlow ? "pb-32" : ""}`}>
			{/* Header */}
			{!fromJobFlow && (
				<div className="flex items-center gap-2 px-6 pb-0 pt-6">
					<Link
						to="/sources"
						className="flex items-center gap-2 p-1.5 hover:rounded-[6px] hover:bg-[#f6f6f6] hover:text-black"
					>
						<ArrowLeft className="size-5" />
					</Link>

					<div className="flex items-center">
						<h1 className="text-2xl font-bold">{sourceName}</h1>
					</div>
				</div>
			)}

			<div className="mt-2 flex flex-1 overflow-hidden border border-t border-[#D9D9D9]">
				{/* Left content */}
				<div
					className={`${
						docsMinimized ? "w-full" : "w-3/4"
					} overflow-auto p-6 pt-4 transition-all duration-300`}
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
													if (onConnectorChange) {
														onConnectorChange(value)
													}
												}}
												className="h-8 w-full"
												options={connectorOptions}
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
										/>
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
										<FixedSchemaForm
											schema={schema}
											formData={formData}
											onChange={(updatedFormData: Record<string, any>) => {
												setFormData(updatedFormData)
												if (onFormDataChange) {
													onFormDataChange(updatedFormData)
												}
											}}
											hideSubmit={true}
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
					)}
				</div>

				<DocumentationPanel
					docUrl={`https://olake.io/docs/connectors/${connector?.toLowerCase()}/config`}
					isMinimized={docsMinimized}
					onToggle={toggleDocsPanel}
					showResizer={true}
				/>
			</div>

			<TestConnectionModal />
			<TestConnectionSuccessModal />
			<TestConnectionFailureModal />

			{/* Delete Modal */}
			<DeleteModal fromSource={true} />

			{/* Footer */}
			{!fromJobFlow && (
				<div className="flex justify-between border-t border-gray-200 bg-white p-4">
					<div>
						<button
							className="rounded-[6px] border border-[#F5222D] px-4 py-1 text-[#F5222D] hover:bg-[#F5222D] hover:text-white"
							onClick={handleDelete}
						>
							Delete
						</button>
					</div>
					<div className="flex space-x-4">
						<button
							className="flex items-center justify-center gap-1 rounded-[6px] bg-[#203FDD] px-4 py-1 font-light text-white hover:bg-[#132685]"
							onClick={handleSave}
						>
							Save changes
						</button>
					</div>
				</div>
			)}

			{/* Entity Edit Modal */}
			<EntityEditModal entityType="source" />
		</div>
	)
}

export default SourceEdit
