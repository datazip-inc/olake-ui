import { useState, useEffect } from "react"
import clsx from "clsx"
import { useNavigate, Link, useParams } from "react-router-dom"
import { message } from "antd"
import { ArrowLeft, ArrowRight } from "@phosphor-icons/react"

import { useAppStore } from "@store/index"
import { jobService } from "@api/index"
import {
	StreamData,
	Job,
	JobBase,
	JobCreationSteps,
	SourceData,
	DestinationData,
	StreamsDataStructure,
} from "@app-types/index"
import JobConfiguration from "../components/JobConfiguration"
import StepProgress from "../components/StepIndicator"
import SourceEdit from "@modules/sources/pages/SourceEdit"
import DestinationEdit from "@modules/destinations/pages/DestinationEdit"
import SchemaConfiguration from "./SchemaConfiguration"
import TestConnectionModal from "@modules/common/Modals/TestConnectionModal"
import TestConnectionSuccessModal from "@modules/common/Modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "@modules/common/Modals/TestConnectionFailureModal"
import {
	getConnectorInLowerCase,
	getSelectedStreams,
	validateCronExpression,
	validateStreams,
} from "../../../utils/utils"
import {
	DESTINATION_INTERNAL_TYPES,
	JOB_CREATION_STEPS,
	JOB_STEP_NUMBERS,
} from "../../../utils/constants"
import ResetStreamsModal from "../../common/Modals/ResetStreamsModal"

// Custom wrapper component for SourceEdit to use in job flow
const JobSourceEdit = ({
	sourceData,
	updateSourceData,
	docsMinimized,
	onDocsMinimizedChange,
}: {
	sourceData: SourceData
	updateSourceData: (data: SourceData) => void
	docsMinimized: boolean
	onDocsMinimizedChange: React.Dispatch<React.SetStateAction<boolean>>
}) => (
	<div className="flex h-full flex-col">
		<div className="flex-1 overflow-auto">
			<SourceEdit
				fromJobFlow={true}
				stepNumber={JOB_STEP_NUMBERS.SOURCE}
				stepTitle="Source Config"
				initialData={sourceData}
				onNameChange={name => updateSourceData({ ...sourceData, name })}
				onConnectorChange={type => {
					updateSourceData({
						...sourceData,
						type,
						config: {},
					})
				}}
				onVersionChange={version =>
					updateSourceData({ ...sourceData, version })
				}
				onFormDataChange={config => updateSourceData({ ...sourceData, config })}
				docsMinimized={docsMinimized}
				onDocsMinimizedChange={onDocsMinimizedChange}
			/>
		</div>
	</div>
)

// Custom wrapper component for DestinationEdit to use in job flow
const JobDestinationEdit = ({
	destinationData,
	sourceData,
	updateDestinationData,
	docsMinimized,
	onDocsMinimizedChange,
}: {
	destinationData: DestinationData
	sourceData: SourceData | null
	updateDestinationData: (data: DestinationData) => void
	docsMinimized: boolean
	onDocsMinimizedChange: React.Dispatch<React.SetStateAction<boolean>>
}) => (
	<div className="flex h-full flex-col">
		<div
			className="flex-1 overflow-auto"
			style={{ paddingBottom: "80px" }}
		>
			<DestinationEdit
				fromJobFlow={true}
				stepNumber={JOB_STEP_NUMBERS.DESTINATION}
				stepTitle="Destination Config"
				initialData={destinationData}
				onNameChange={name =>
					updateDestinationData({ ...destinationData, name })
				}
				onConnectorChange={type => {
					updateDestinationData({
						...destinationData,
						type,
						config: {},
					})
				}}
				onVersionChange={version =>
					updateDestinationData({ ...destinationData, version })
				}
				onFormDataChange={config =>
					updateDestinationData({ ...destinationData, config })
				}
				docsMinimized={docsMinimized}
				onDocsMinimizedChange={onDocsMinimizedChange}
				sourceConnector={getConnectorInLowerCase(sourceData?.type || "")}
				sourceVersion={sourceData?.version || ""}
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
		setShowResetStreamsModal,
	} = useAppStore()

	const [currentStep, setCurrentStep] = useState<JobCreationSteps>(
		JOB_CREATION_STEPS.STREAMS,
	)
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [isSubmitting, setIsSubmitting] = useState(false)

	const [sourceData, setSourceData] = useState<SourceData | null>(null)
	const [destinationData, setDestinationData] =
		useState<DestinationData | null>(null)
	const [nextStep, setNextStep] = useState<JobCreationSteps | null>(null)

	// Streams step states
	const [selectedStreams, setSelectedStreams] = useState<StreamsDataStructure>({
		selected_streams: {},
		streams: [],
	})

	// Config step states
	const [jobName, setJobName] = useState("")
	const [cronExpression, setCronExpression] = useState("* * * * *")
	const [job, setJob] = useState<Job | null>(null)
	const [isFromSources, setIsFromSources] = useState(true)
	const [streamsModified, setStreamsModified] = useState(false)
	const [isStreamsLoading, setIsStreamsLoading] = useState(false)

	// Load job data on component mount
	useEffect(() => {
		const loadData = async () => {
			try {
				await Promise.all([fetchJobs(), fetchSources(), fetchDestinations()])
			} catch (error) {
				console.error("Error loading data:", error)
				message.error("Failed to load job data. Please try again.")
			}
		}
		loadData()
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
		if (job.frequency) {
			setCronExpression(job.frequency)
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
				hosts: [],
				username: "",
				password: "",
				authdb: "",
				database: "",
				collection: "",
			},
			version: "",
		})

		setDestinationData({
			name: "New Destination",
			type: DESTINATION_INTERNAL_TYPES.S3,
			config: {
				normalization: false,
				s3_bucket: "",
				s3_region: "",
				type: "PARQUET",
			},
			version: "",
		})

		setJobName("New Job")
	}

	useEffect(() => {
		// TODO: when user refreshes specifc data should be retained
		let job = jobs.find(j => j.id.toString() === jobId)
		if (job) {
			setJob(job)
			initializeFromExistingJob(job)
		} else if (jobId) {
			navigate("/jobs")
		} else {
			initializeForNewJob()
		}
	}, [jobs, jobId])

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
					normalization: false,
					filter: "",
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

	const getjobUpdatePayLoad = () => {
		const jobUpdateRequestPayload: JobBase = {
			name: jobName,
			source: {
				name: sourceData?.name || "",
				type: getConnectorInLowerCase(sourceData?.type || ""),
				config:
					typeof sourceData?.config === "string"
						? sourceData?.config
						: JSON.stringify(sourceData?.config),
				version: sourceData?.version || "",
			},
			destination: {
				name: destinationData?.name || "",
				type: getConnectorInLowerCase(destinationData?.type || ""),
				config:
					typeof destinationData?.config === "string"
						? destinationData?.config
						: JSON.stringify(destinationData?.config),
				version: destinationData?.version || "",
			},
			streams_config:
				typeof selectedStreams === "string"
					? selectedStreams
					: JSON.stringify({
							...selectedStreams,
							selected_streams: getSelectedStreams(
								selectedStreams.selected_streams,
							),
						}),
			frequency: cronExpression,
			activate: job?.activate,
		}
		return jobUpdateRequestPayload
	}

	// Handle job submission
	const handleJobSubmit = async () => {
		if (!sourceData || !destinationData || !jobId) {
			message.error("Source and destination data are required")
			return
		}

		if (
			!validateStreams(getSelectedStreams(selectedStreams.selected_streams))
		) {
			message.error("Filter Value cannot be empty")
			return
		}

		setIsSubmitting(true)

		try {
			// Create the job update payload
			const jobUpdatePayload = getjobUpdatePayLoad()

			await jobService.updateJob(jobId, jobUpdatePayload)
			message.success("Job updated successfully!")

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
		if (currentStep === JOB_CREATION_STEPS.SOURCE) {
			if (sourceData) {
				setIsFromSources(true)
				setCurrentStep(JOB_CREATION_STEPS.DESTINATION)
			}
		} else if (currentStep === JOB_CREATION_STEPS.DESTINATION) {
			if (destinationData) {
				setIsFromSources(false)
				setCurrentStep(JOB_CREATION_STEPS.STREAMS)
			}
		} else if (currentStep === JOB_CREATION_STEPS.STREAMS) {
			handleJobSubmit()
		} else if (currentStep === JOB_CREATION_STEPS.CONFIG) {
			if (!jobName.trim()) {
				message.error("Job name is required")
				return
			}
			if (!validateCronExpression(cronExpression)) {
				return
			}
			setCurrentStep(JOB_CREATION_STEPS.SOURCE)
		}
	}

	const handleBack = async () => {
		if (currentStep === JOB_CREATION_STEPS.DESTINATION) {
			setCurrentStep(JOB_CREATION_STEPS.SOURCE)
		} else if (currentStep === JOB_CREATION_STEPS.STREAMS) {
			setShowResetStreamsModal(true)
		} else if (currentStep === JOB_CREATION_STEPS.SOURCE) {
			setCurrentStep(JOB_CREATION_STEPS.CONFIG)
		}
	}

	const handleStepClick = async (step: string) => {
		if (currentStep === JOB_CREATION_STEPS.STREAMS) {
			setNextStep(step as JobCreationSteps)
			setShowResetStreamsModal(true)
			return
		}
		setCurrentStep(step as JobCreationSteps)
	}

	const handleStreamsChange = (newStreams: any) => {
		setSelectedStreams(newStreams)
		setStreamsModified(true)
	}

	const handleConfirmResetStreams = () => {
		setSelectedStreams({
			selected_streams: {},
			streams: [],
		})
		setNextStep(null)
		setCurrentStep(nextStep || JOB_CREATION_STEPS.DESTINATION)
	}

	const isBackDisabled =
		currentStep === JOB_CREATION_STEPS.CONFIG ||
		(currentStep === JOB_CREATION_STEPS.STREAMS && isStreamsLoading)

	return (
		<div className="flex h-screen flex-col">
			{/* Header */}
			<div className="bg-white px-6 pb-3 pt-6">
				<div className="flex items-center justify-between">
					<div className="flex items-center gap-2">
						<Link
							to="/jobs"
							className="flex items-center gap-2 p-1.5 hover:rounded-md hover:bg-gray-100 hover:text-black"
						>
							<ArrowLeft className="mr-1 size-5" />
						</Link>
						<div className="text-2xl font-bold">
							{jobName ? (jobName === "-" ? " " : jobName) : "New Job"}
						</div>
					</div>
					{/* Stepper */}
					<StepProgress
						currentStep={currentStep}
						onStepClick={handleStepClick}
						isEditMode={!!jobId}
						disabled={isStreamsLoading}
					/>
				</div>
			</div>

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden border-t border-gray-200">
				{/* Left content */}
				<div
					className={clsx(
						"w-full pt-0 transition-all duration-300",
						currentStep !== JOB_CREATION_STEPS.STREAMS && "overflow-hidden",
					)}
				>
					<div className="h-full">
						{currentStep === JOB_CREATION_STEPS.SOURCE && sourceData && (
							<JobSourceEdit
								sourceData={sourceData}
								updateSourceData={setSourceData}
								docsMinimized={docsMinimized}
								onDocsMinimizedChange={setDocsMinimized}
							/>
						)}

						{currentStep === JOB_CREATION_STEPS.DESTINATION &&
							destinationData && (
								<JobDestinationEdit
									destinationData={destinationData}
									sourceData={sourceData}
									updateDestinationData={setDestinationData}
									docsMinimized={docsMinimized}
									onDocsMinimizedChange={setDocsMinimized}
								/>
							)}

						{currentStep === JOB_CREATION_STEPS.STREAMS && (
							<div className="h-full overflow-auto">
								<SchemaConfiguration
									selectedStreams={selectedStreams as any}
									setSelectedStreams={handleStreamsChange}
									stepNumber={JOB_STEP_NUMBERS.STREAMS}
									stepTitle="Streams Selection"
									sourceName={sourceData?.name || ""}
									sourceConnector={sourceData?.type.toLowerCase() || ""}
									sourceVersion={sourceData?.version || ""}
									sourceConfig={JSON.stringify(sourceData?.config || {})}
									fromJobEditFlow={true}
									jobId={jobId ? parseInt(jobId) : -1}
									destinationType={destinationData?.type.toLowerCase() || ""}
									initialStreamsData={
										streamsModified ? selectedStreams : undefined
									}
									jobName={jobName}
									onLoadingChange={setIsStreamsLoading}
								/>
							</div>
						)}

						{currentStep === JOB_CREATION_STEPS.CONFIG && (
							<JobConfiguration
								jobName={jobName}
								setJobName={setJobName}
								cronExpression={cronExpression}
								setCronExpression={setCronExpression}
								stepNumber={JOB_STEP_NUMBERS.CONFIG}
								stepTitle="Job Configuration"
							/>
						)}
					</div>
				</div>
			</div>

			{/* Footer */}
			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<div>
					<button
						className="rounded-md border border-gray-400 px-4 py-1 font-light hover:bg-[#ebebeb]"
						onClick={handleBack}
						disabled={isBackDisabled}
						style={{
							opacity: isBackDisabled ? 0.5 : 1,
							cursor: isBackDisabled ? "not-allowed" : "pointer",
						}}
					>
						Back
					</button>
				</div>
				<div
					className={clsx(
						"flex gap-2 transition-[margin] duration-500 ease-in-out",
						!docsMinimized &&
							(currentStep === JOB_CREATION_STEPS.SOURCE ||
								currentStep === JOB_CREATION_STEPS.DESTINATION)
							? "mr-[40%]"
							: "mr-[4%]",
					)}
				>
					{currentStep === JOB_CREATION_STEPS.CONFIG && jobId && (
						<button
							className="flex items-center justify-center gap-2 rounded-md border border-primary px-4 py-1 font-light text-primary hover:bg-primary-50"
							onClick={handleJobSubmit}
							disabled={isSubmitting}
						>
							{isSubmitting ? "Saving..." : "Save"}
						</button>
					)}
					<button
						className="flex items-center justify-center gap-2 rounded-md bg-primary px-4 py-1 font-light text-white hover:bg-primary-600"
						onClick={handleNext}
						disabled={isSubmitting}
					>
						{currentStep === JOB_CREATION_STEPS.STREAMS
							? isSubmitting
								? "Saving..."
								: "Finish"
							: "Next"}
						{currentStep !== JOB_CREATION_STEPS.STREAMS && (
							<ArrowRight className="size-4 text-white" />
						)}
					</button>
				</div>
			</div>
			<TestConnectionModal />
			<TestConnectionSuccessModal />
			<TestConnectionFailureModal fromSources={isFromSources} />
			<ResetStreamsModal onConfirm={handleConfirmResetStreams} />
		</div>
	)
}

export default JobEdit
