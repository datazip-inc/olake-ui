import { useState, useEffect } from "react"
import { useNavigate, Link, useParams } from "react-router-dom"
import { message, Spin } from "antd"
import SourceEdit from "../../sources/pages/SourceEdit"
import DestinationEdit from "../../destinations/pages/DestinationEdit"
import { ArrowLeft, ArrowRight } from "@phosphor-icons/react"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import StepProgress from "../components/StepIndicator"
import SchemaConfiguration from "./SchemaConfiguration"
import JobConfiguration from "../components/JobConfiguration"
import { useAppStore } from "../../../store"
import { StreamData, Job, JobBase } from "../../../types"
import { jobService } from "../../../api"
import { sourceService } from "../../../api/services/sourceService"
import { destinationService } from "../../../api/services/destinationService"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"
import {
	getFrequencyValue,
	removeSavedJobFromLocalStorage,
} from "../../../utils/utils"

type Step = "source" | "destination" | "schema" | "config"

interface SourceData {
	id?: string
	name: string
	type: string
	config: Record<string, any>
	version?: string
}

interface DestinationData {
	id?: string
	name: string
	type: string
	config: Record<string, any>
	version?: string
}

// Define streams data interface
interface StreamsDataStructure {
	selected_streams: {
		[namespace: string]: {
			stream_name: string
			partition_regex: string
			split_column: string
		}[]
	}
	streams: StreamData[]
}

// Custom wrapper component for SourceEdit to use in job flow
const JobSourceEdit = ({
	sourceData,
	updateSourceData,
}: {
	sourceData: SourceData
	updateSourceData: (data: SourceData) => void
}) => (
	<div className="flex h-full flex-col">
		<div className="flex-1 overflow-auto">
			<SourceEdit
				fromJobFlow={true}
				stepNumber="1"
				stepTitle="Source config"
				initialData={sourceData}
				onNameChange={name => updateSourceData({ ...sourceData, name })}
				onConnectorChange={type => updateSourceData({ ...sourceData, type })}
				onVersionChange={version =>
					updateSourceData({ ...sourceData, version })
				}
				onFormDataChange={config => updateSourceData({ ...sourceData, config })}
			/>
		</div>
	</div>
)

// Custom wrapper component for DestinationEdit to use in job flow
const JobDestinationEdit = ({
	destinationData,
	updateDestinationData,
}: {
	destinationData: DestinationData
	updateDestinationData: (data: DestinationData) => void
}) => (
	<div className="flex h-full flex-col">
		<div
			className="flex-1 overflow-auto"
			style={{ paddingBottom: "80px" }}
		>
			<DestinationEdit
				fromJobFlow={true}
				stepNumber="2"
				stepTitle="Destination config"
				initialData={destinationData}
				onNameChange={name =>
					updateDestinationData({ ...destinationData, name })
				}
				onConnectorChange={type =>
					updateDestinationData({ ...destinationData, type })
				}
				onVersionChange={version =>
					updateDestinationData({ ...destinationData, version })
				}
				onFormDataChange={config =>
					updateDestinationData({ ...destinationData, config })
				}
			/>
		</div>
	</div>
)

const JobEdit: React.FC = () => {
	const navigate = useNavigate()
	const { jobId } = useParams<{ jobId: string }>()
	const {
		jobs,
		fetchJobs,
		fetchSources,
		fetchDestinations,
		setShowTestingModal,
		setShowSuccessModal,
		setShowFailureModal,
		setSourceTestConnectionError,
	} = useAppStore()

	const [currentStep, setCurrentStep] = useState<Step>("source")
	const [docsMinimized, setDocsMinimized] = useState(true)
	const [isSubmitting, setIsSubmitting] = useState(false)

	// Source and destination data for job
	const [sourceData, setSourceData] = useState<SourceData | null>(null)
	const [destinationData, setDestinationData] =
		useState<DestinationData | null>(null)

	// Schema step states
	const [selectedStreams, setSelectedStreams] = useState<StreamsDataStructure>({
		selected_streams: {},
		streams: [],
	})
	const [initialStreamsData, setInitialStreamsData] =
		useState<StreamsDataStructure | null>(null)

	// Config step states
	const [jobName, setJobName] = useState("")
	const [replicationFrequency, setReplicationFrequency] = useState("seconds")
	const [schemaChangeStrategy, setSchemaChangeStrategy] = useState("propagate")
	const [notifyOnSchemaChanges, setNotifyOnSchemaChanges] = useState(true)
	const [isSavedJob, setIsSavedJob] = useState(false)
	const [replicationFrequencyValue, setReplicationFrequencyValue] =
		useState("1")
	const [job, setJob] = useState<Job | null>(null)
	const [savedJobId, setSavedJobId] = useState<string | null>(null)

	// Find the job from the store
	// useEffect(() => {
	// 	if (!jobs.length) {
	// 		fetchJobs()
	// 	}
	// }, [fetchJobs, jobs])

	useEffect(() => {
		fetchJobs()
	}, [])

	// Load job data on component mount
	useEffect(() => {
		fetchJobs()
		fetchSources()
		fetchDestinations()
	}, [])

	const initializeFromExistingJob = (job: Job) => {
		setJobName(job.name)
		// Parse source config
		let sourceConfig = JSON.parse(job.source.config)

		// Set source data from job
		setSourceData({
			name: job.source.name,
			type: job.source.type,
			config: sourceConfig,
			version: job.source.version,
		})

		// Parse destination config
		let destConfig = JSON.parse(job.destination.config)

		// Set destination data from job
		setDestinationData({
			name: job.destination.name,
			type: job.destination.type,
			config: destConfig,
			version: job.destination.version,
		})

		// Set other job settings
		// Parse frequency with format value-unit (e.g. "1-minutes")
		if (job.frequency && job.frequency.includes("-")) {
			const [value, unit] = job.frequency.split("-")
			setReplicationFrequencyValue(value)
			setReplicationFrequency(unit)
		} else {
			setReplicationFrequency(getFrequencyValue(job.frequency) || "hours")
			setReplicationFrequencyValue("1")
		}

		// Parse streams config
		if (job.streams_config) {
			try {
				if (job.streams_config == "[]") {
					setSelectedStreams({
						selected_streams: {},
						streams: [],
					})
				} else {
					const parsedStreamsConfig = JSON.parse(job.streams_config)
					const streamsData = processStreamsConfig(parsedStreamsConfig)

					if (streamsData) {
						setSelectedStreams(streamsData)
						setInitialStreamsData(streamsData)
					}
				}
			} catch (e) {
				console.error("Error parsing streams config:", e)
			}
		}
	}

	// Initialize defaults for a new job
	const initializeForNewJob = () => {
		setSourceData({
			name: "New Source",
			type: "MongoDB",
			config: {
				hosts: ["localhost:27017"],
				username: "",
				password: "",
				authdb: "admin",
				database: "",
				collection: "",
			},
			version: "latest",
		})

		setDestinationData({
			name: "New Destination",
			type: "s3",
			config: {
				normalization: false,
				s3_bucket: "",
				s3_region: "us-east-1",
				type: "PARQUET",
			},
			version: "latest",
		})

		setJobName("New Job")
	}

	useEffect(() => {
		let job = jobs.find(j => j.id.toString() === jobId)
		if (job) {
			setJob(job)
		}
		if (!job) {
			const savedJobsFromStorage = JSON.parse(
				localStorage.getItem("savedJobs") || "[]",
			)
			job = savedJobsFromStorage.find((job: any) => job.id === jobId)
			if (job) {
				setJob(job)
				setIsSavedJob(true)
				setSavedJobId(job.id.toString())
			}
		}

		if (job) {
			initializeFromExistingJob(job)
		} else {
			initializeForNewJob()
		}
	}, [])

	// Process streams configuration into a consistent format
	const processStreamsConfig = (
		parsedConfig: any,
	): StreamsDataStructure | null => {
		if (parsedConfig.streams && parsedConfig.selected_streams) {
			return parsedConfig as StreamsDataStructure
		} else if (Array.isArray(parsedConfig)) {
			const streamsData: StreamsDataStructure = {
				selected_streams: { default: [] },
				streams: [],
			}
			parsedConfig.forEach((streamName: string) => {
				streamsData.selected_streams.default.push({
					stream_name: streamName,
					partition_regex: "",
					split_column: "",
				})
				streamsData.streams.push({
					stream: {
						name: streamName,
						namespace: "default",
					},
				} as StreamData)
			})

			return streamsData
		}

		// Handle case where streams_config is just selected_streams object
		else if (typeof parsedConfig === "object") {
			return {
				selected_streams: parsedConfig,
				streams: [],
			}
		}

		return null
	}

	// Helper function to safely parse JSON config strings
	// const parseConfig = (
	// 	config: string | Record<string, any>,
	// ): Record<string, any> => {
	// 	if (typeof config === "string") {
	// 		try {
	// 			return JSON.parse(config)
	// 		} catch (error) {
	// 			console.error("Error parsing config:", error)
	// 			return {}
	// 		}
	// 	}
	// 	return config as Record<string, any>
	// }

	// Handle job submission
	const handleJobSubmit = async () => {
		if (!sourceData || !destinationData) {
			message.error("Source and destination data are required")
			return
		}

		setIsSubmitting(true)

		try {
			// Create the job update payload
			const jobUpdatePayload: JobBase = {
				name: jobName,
				source: {
					name: sourceData.name,
					type: sourceData.type,
					config:
						typeof sourceData.config === "string"
							? sourceData.config
							: JSON.stringify(sourceData.config),
					version: sourceData.version || "latest",
				},
				destination: {
					name: destinationData.name,
					type: destinationData.type,
					config:
						typeof destinationData.config === "string"
							? destinationData.config
							: JSON.stringify(destinationData.config),
					version: destinationData.version || "latest",
				},
				streams_config:
					typeof selectedStreams === "string"
						? selectedStreams
						: JSON.stringify(selectedStreams),
				frequency: `${replicationFrequencyValue}-${replicationFrequency}`,
				activate: job?.activate || true,
			}

			if (jobId && !isSavedJob) {
				await jobService.updateJob(jobId, jobUpdatePayload)
				message.success("Job updated successfully!")
			} else {
				await jobService.createJob(jobUpdatePayload)
				message.success("Job created successfully!")
			}

			// Refresh jobs and navigate back to jobs list
			fetchJobs()
			navigate("/jobs")
		} catch (error) {
			console.error("Error saving job:", error)
			message.error("Failed to save job. Please try again.")
		} finally {
			setIsSubmitting(false)
		}
	}

	const handleNext = async () => {
		if (currentStep === "source") {
			if (isSavedJob && sourceData) {
				setShowTestingModal(true)
				try {
					const testData = {
						name: sourceData.name,
						type: sourceData.type.toLowerCase(),
						version: sourceData.version || "latest",
						config:
							typeof sourceData.config === "string"
								? sourceData.config
								: JSON.stringify(sourceData.config),
					}
					const testResult = await sourceService.testSourceConnection(testData)
					setTimeout(() => {
						setShowTestingModal(false)
					}, 1000)
					if (testResult.data?.status === "SUCCEEDED") {
						setTimeout(() => {
							setShowSuccessModal(true)
						}, 1000)
						setTimeout(() => {
							setShowSuccessModal(false)
							setCurrentStep("destination")
						}, 2000)
					} else {
						setSourceTestConnectionError(testResult.data?.message || "")
						setShowFailureModal(true)
					}
				} catch (error) {
					console.error("Error testing source connection:", error)
					setShowTestingModal(false)
					setShowFailureModal(true)
				}
			} else {
				setCurrentStep("destination")
			}
		} else if (currentStep === "destination") {
			if (isSavedJob && destinationData) {
				setShowTestingModal(true)
				try {
					const testData = {
						name: destinationData.name,
						type: destinationData.type.toLowerCase(),
						version: destinationData.version || "latest",
						config:
							typeof destinationData.config === "string"
								? destinationData.config
								: JSON.stringify(destinationData.config),
					}
					const testResult =
						await destinationService.testDestinationConnection(testData)

					setTimeout(() => {
						setShowTestingModal(false)
					}, 1000)
					if (testResult.success) {
						setTimeout(() => {
							setShowSuccessModal(true)
						}, 1000)
						setTimeout(() => {
							setShowSuccessModal(false)
							setCurrentStep("schema")
						}, 2000)
					} else {
						setShowFailureModal(true)
					}
				} catch (error) {
					console.error("Error testing destination connection:", error)
					setShowTestingModal(false)
					setShowFailureModal(true)
				}
			} else {
				setCurrentStep("schema")
			}
		} else if (currentStep === "schema") {
			setCurrentStep("config")
		} else if (currentStep === "config") {
			if (isSavedJob) {
				removeSavedJobFromLocalStorage(savedJobId || "")
			}
			handleJobSubmit()
		}
	}

	const handleBack = () => {
		if (currentStep === "destination") {
			setCurrentStep("source")
		} else if (currentStep === "schema") {
			setCurrentStep("destination")
		} else if (currentStep === "config") {
			setCurrentStep("schema")
		}
	}

	const toggleDocsPanel = () => {
		setDocsMinimized(!docsMinimized)
	}

	// Show loading while job data is loading
	if (!job && jobId) {
		return (
			<div className="flex h-screen items-center justify-center">
				<Spin tip="Loading job data..." />
			</div>
		)
	}

	return (
		<div className="flex h-screen flex-col">
			{/* Header */}
			<div className="bg-white px-6 pb-3 pt-6">
				<div className="flex items-center justify-between">
					<div className="flex items-center gap-2">
						<Link
							to="/jobs"
							className="flex items-center gap-2 p-1.5 hover:rounded-[6px] hover:bg-[#f6f6f6] hover:text-black"
						>
							<ArrowLeft className="mr-1 size-5" />
						</Link>
						<div className="text-2xl font-bold">
							{jobName ? (jobName === "-" ? " " : jobName) : "New Job"}
						</div>
					</div>
					{/* Stepper */}
					<StepProgress currentStep={currentStep} />
				</div>
			</div>

			{/* Main content */}
			<div
				className={`flex flex-1 overflow-auto border-gray-200 ${
					currentStep === "config" || currentStep === "schema" ? "border-t" : ""
				}`}
			>
				{/* Left content */}
				<div
					className={`${
						(currentStep === "schema" || currentStep === "config") &&
						!docsMinimized
							? "w-[calc(100%-30%)]"
							: "w-full"
					} ${currentStep === "schema" ? "" : "overflow-hidden"} relative flex flex-col`}
				>
					<div className="flex-1 pb-0">
						{currentStep === "source" && sourceData && (
							<JobSourceEdit
								sourceData={sourceData}
								updateSourceData={setSourceData}
							/>
						)}

						{currentStep === "destination" && destinationData && (
							<JobDestinationEdit
								destinationData={destinationData}
								updateDestinationData={setDestinationData}
							/>
						)}

						{currentStep === "schema" && (
							<div className="h-full overflow-scroll">
								<SchemaConfiguration
									selectedStreams={selectedStreams as any}
									setSelectedStreams={setSelectedStreams as any}
									stepNumber={3}
									stepTitle="Streams Selection"
									sourceName={sourceData?.name || ""}
									sourceConnector={sourceData?.type.toLowerCase() || ""}
									sourceVersion={sourceData?.version || "latest"}
									sourceConfig={JSON.stringify(sourceData?.config || {})}
									initialStreamsData={initialStreamsData as any}
								/>
							</div>
						)}

						{currentStep === "config" && (
							<JobConfiguration
								jobName={jobName}
								setJobName={setJobName}
								replicationFrequency={replicationFrequency}
								setReplicationFrequency={setReplicationFrequency}
								replicationFrequencyValue={replicationFrequencyValue}
								setReplicationFrequencyValue={setReplicationFrequencyValue}
								schemaChangeStrategy={schemaChangeStrategy}
								setSchemaChangeStrategy={setSchemaChangeStrategy}
								notifyOnSchemaChanges={notifyOnSchemaChanges}
								setNotifyOnSchemaChanges={setNotifyOnSchemaChanges}
								stepNumber={4}
								stepTitle="Job Configuration"
							/>
						)}
					</div>
				</div>

				{/* Documentation panel */}
				{currentStep === "schema" && (
					<DocumentationPanel
						isMinimized={docsMinimized}
						onToggle={toggleDocsPanel}
						docUrl={`https://olake.io/docs/connectors/${sourceData?.type.toLowerCase()}/config`}
					/>
				)}
			</div>

			{/* Footer */}
			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<div>
					<button
						className="rounded-[6px] border border-[#D9D9D9] px-4 py-1 font-light hover:bg-[#EBEBEB]"
						onClick={handleBack}
						disabled={currentStep === "source"}
						style={{
							opacity: currentStep === "source" ? 0.5 : 1,
							cursor: currentStep === "source" ? "not-allowed" : "pointer",
						}}
					>
						Back
					</button>
				</div>
				<div>
					<button
						className="flex items-center justify-center gap-2 rounded-[6px] bg-[#203FDD] px-4 py-1 font-light text-white hover:bg-[#132685]"
						onClick={handleNext}
						disabled={isSubmitting}
					>
						{currentStep === "config"
							? isSubmitting
								? "Saving..."
								: "Finish"
							: "Next"}
						{currentStep !== "config" && (
							<ArrowRight className="size-4 text-white" />
						)}
					</button>
				</div>
			</div>
			<TestConnectionModal />
			<TestConnectionSuccessModal />
			<TestConnectionFailureModal />
		</div>
	)
}

export default JobEdit
