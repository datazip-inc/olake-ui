import { useState, useEffect, useRef } from "react"
import clsx from "clsx"
import { useNavigate, Link, useParams } from "react-router-dom"
import { message } from "antd"
import { ArrowLeftIcon, ArrowRightIcon } from "@phosphor-icons/react"

import { useJobStore, useStreamSelectionStore } from "../stores"
import { useJobDetails } from "../hooks/queries/useJobQueries"
import { useUpdateJob } from "../hooks/mutations/useJobMutations"
import { jobService } from "../services"
import { Job, JobBase, JobCreationSteps, AdvancedSettings } from "../types"
import { StreamData, StreamsDataStructure } from "@/common/types"
import JobConfiguration from "../components/JobConfiguration"
import StepProgress from "../components/StepIndicator"
import SchemaConfiguration from "../components/SchemaConfiguration"
import { getConnectorInLowerCase } from "@/common/utils"
import {
	validateCronExpression,
	getSelectedStreams,
	validateStreams,
} from "../utils"
import {
	JOB_CREATION_STEPS,
	JOB_STEP_NUMBERS,
	STREAM_DEFAULTS,
} from "../constants"
import ResetStreamsModal from "@/features/jobs/components/modals/ResetStreamsModal"
import StreamDifferenceModal from "@/features/jobs/components/modals/StreamDifferenceModal"
import StreamEditDisabledModal from "@/features/jobs/components/modals/StreamEditDisabledModal"

const JobEdit: React.FC = () => {
	const navigate = useNavigate()
	const { jobId } = useParams<{ jobId: string }>()
	const {
		setSelectedJobId,
		setShowResetStreamsModal,
		setShowStreamDifferenceModal,
	} = useJobStore()
	const isDiscovering = useStreamSelectionStore(state => state.isDiscovering)
	const streamsData = useStreamSelectionStore(state => state.streamsData)
	const { data: job, isError: isJobError } = useJobDetails(jobId || "", {
		staleTime: Infinity,
	})

	const { mutateAsync: updateJob } = useUpdateJob()

	// Set selected job from route param (source of truth is the URL)
	useEffect(() => {
		if (jobId) setSelectedJobId(jobId)
	}, [jobId])

	const [currentStep, setCurrentStep] = useState<JobCreationSteps>(
		JOB_CREATION_STEPS.STREAMS,
	)
	const [isSubmitting, setIsSubmitting] = useState(false)

	const [selectedSourceId, setSelectedSourceId] = useState<number | null>(null)
	const [selectedDestinationId, setSelectedDestinationId] = useState<
		number | null
	>(null)

	const [sourceSnapshot, setSourceSnapshot] = useState<{
		id?: number
		name: string
		type: string
		config: string
		version: string
	} | null>(null)
	const [destinationSnapshot, setDestinationSnapshot] = useState<{
		id?: number
		name: string
		type: string
		config: string
		version: string
	} | null>(null)

	const [nextStep, setNextStep] = useState<JobCreationSteps | null>(null)
	const [streamDifference, setStreamDifference] =
		useState<StreamsDataStructure | null>(null)

	// Config step states
	const [jobName, setJobName] = useState("")
	const [cronExpression, setCronExpression] = useState("* * * * *")
	const [advancedSettings, setAdvancedSettings] =
		useState<AdvancedSettings | null>(null)

	const initialStreamsData = useRef<StreamsDataStructure | null>(null)

	const normalizedSourceConnector = sourceSnapshot
		? getConnectorInLowerCase(sourceSnapshot.type)
		: ""

	const initializeFromExistingJob = (job: Job) => {
		setJobName(job.name)

		// Resolve source snapshot
		const sourceConfig =
			typeof job.source.config === "string"
				? job.source.config
				: JSON.stringify(job.source.config)
		setSelectedSourceId(job.source.id ?? null)
		setSourceSnapshot({
			id: job.source.id,
			name: job.source.name,
			type: job.source.type,
			config: sourceConfig,
			version: job.source.version,
		})

		// Resolve destination snapshot
		const destConfig =
			typeof job.destination.config === "string"
				? job.destination.config
				: JSON.stringify(job.destination.config)
		setSelectedDestinationId(job.destination.id ?? null)
		setDestinationSnapshot({
			id: job.destination.id,
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
				if (job.streams_config === "[]") {
					initialStreamsData.current = {
						selected_streams: {},
						streams: [],
					}
				} else {
					const parsedStreamsConfig = JSON.parse(job.streams_config)
					const streamsConfig = processStreamsConfig(parsedStreamsConfig)
					if (streamsConfig) {
						initialStreamsData.current = streamsConfig
					}
				}
			} catch (e) {
				console.error("Error parsing streams config:", e)
			}
		}

		setAdvancedSettings(job.advanced_settings ?? null)
	}

	// Initialize from fetched job data
	const hasInitialized = useRef(false)
	useEffect(() => {
		if (job && !hasInitialized.current) {
			hasInitialized.current = true
			initializeFromExistingJob(job)
		}
	}, [job])

	// Navigate to jobs list on fetch error
	useEffect(() => {
		if (isJobError) navigate("/jobs")
	}, [isJobError])

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
					...STREAM_DEFAULTS,
					stream_name: streamName,
				})
				streamsData.streams.push({
					stream: {
						name: streamName,
						namespace: "default",
					},
				} as StreamData)
			})
			return streamsData
		} else if (typeof parsedConfig === "object") {
			return {
				selected_streams: parsedConfig,
				streams: [],
			}
		}
		return null
	}

	const getJobUpdatePayLoad = (
		streamsConfig: StreamsDataStructure,
		diff: StreamsDataStructure | null,
	): JobBase => ({
		name: jobName,
		source: {
			...(sourceSnapshot?.id && { id: sourceSnapshot.id }),
			name: sourceSnapshot?.name || "",
			type: normalizedSourceConnector,
			config: sourceSnapshot?.config || "{}",
			version: sourceSnapshot?.version || "",
		},
		destination: {
			...(destinationSnapshot?.id && { id: destinationSnapshot.id }),
			name: destinationSnapshot?.name || "",
			type: getConnectorInLowerCase(destinationSnapshot?.type),
			config: destinationSnapshot?.config || "{}",
			version: destinationSnapshot?.version || "",
		},
		streams_config: JSON.stringify({
			...streamsConfig,
			selected_streams: getSelectedStreams(streamsConfig.selected_streams),
		}),
		frequency: cronExpression,
		activate: job?.activate,
		...(diff && { difference_streams: JSON.stringify(diff) }),
		advanced_settings: advancedSettings,
	})

	const handleStreamDifference = async () => {
		if (!sourceSnapshot || !destinationSnapshot || !jobId) {
			message.error("Source and destination data are required")
			return
		}

		if (!streamsData) {
			message.error("No streams data available")
			return
		}

		if (!validateStreams(getSelectedStreams(streamsData.selected_streams))) {
			message.error("Filter Value cannot be empty")
			return
		}

		const streamDifferenceResponse = (
			await jobService.getStreamDifference(
				jobId,
				JSON.stringify({
					...streamsData,
					selected_streams: getSelectedStreams(streamsData.selected_streams),
				}),
			)
		)?.difference_streams

		const diff =
			typeof streamDifferenceResponse === "string"
				? JSON.parse(streamDifferenceResponse || "{}")
				: streamDifferenceResponse || {}
		const hasDiff = Object.keys(diff?.selected_streams ?? diff).length > 0
		// if there is a stream difference, show the stream difference modal
		if (hasDiff) {
			setStreamDifference(streamDifferenceResponse)
			setShowStreamDifferenceModal(true)
			return
		}

		// No difference - clear state and submit with null stream difference
		setStreamDifference(null)
		handleJobSubmit(null)
	}

	// Handle job submission
	const handleJobSubmit = async (diff: StreamsDataStructure | null) => {
		if (!sourceSnapshot || !destinationSnapshot || !jobId) {
			message.error("Source and destination data are required")
			return
		}

		// Use store data if available; fall back to initial data from job API
		const streamsConfig = streamsData ?? initialStreamsData.current

		if (!streamsConfig) {
			message.error("No valid streams configuration found")
			return
		}

		if (!validateStreams(getSelectedStreams(streamsConfig.selected_streams))) {
			message.error("Filter Value cannot be empty")
			return
		}
		setIsSubmitting(true)
		try {
			// Create the job update payload
			const jobUpdatePayload = getJobUpdatePayLoad(streamsConfig, diff)

			await updateJob({ jobId, job: jobUpdatePayload })

			// wait for 1 second before refreshing jobs to avoid fetching old state
			await new Promise(resolve => setTimeout(resolve, 1000))
			navigate("/jobs")
		} catch (error) {
			console.error("Error saving job:", error)
		} finally {
			setIsSubmitting(false)
		}
	}

	const handleNext = async () => {
		if (currentStep === JOB_CREATION_STEPS.CONFIG) {
			if (!jobName.trim()) {
				message.error("Job name is required")
				return
			}
			if (!validateCronExpression(cronExpression)) return
			setCurrentStep(JOB_CREATION_STEPS.STREAMS)
		} else if (currentStep === JOB_CREATION_STEPS.STREAMS) {
			handleStreamDifference()
		}
	}

	const handleBack = () => {
		if (currentStep === JOB_CREATION_STEPS.STREAMS) {
			setShowResetStreamsModal(true)
		}
	}

	const handleStepClick = (step: string) => {
		if (step === currentStep) return
		if (currentStep === JOB_CREATION_STEPS.STREAMS) {
			setNextStep(step as JobCreationSteps)
			setShowResetStreamsModal(true)
			return
		}
		setCurrentStep(step as JobCreationSteps)
	}

	const handleConfirmResetStreams = () => {
		useStreamSelectionStore.getState().reset()
		setNextStep(null)
		setCurrentStep(nextStep || JOB_CREATION_STEPS.CONFIG)
	}

	const isBackDisabled =
		currentStep === JOB_CREATION_STEPS.CONFIG ||
		(currentStep === JOB_CREATION_STEPS.STREAMS && isDiscovering)

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
							<ArrowLeftIcon className="mr-1 size-5" />
						</Link>
						<div className="text-2xl font-bold">
							{jobName ? (jobName === "-" ? " " : jobName) : "Edit Job"}
						</div>
					</div>
					{/* Stepper */}
					<StepProgress
						currentStep={currentStep}
						onStepClick={handleStepClick}
						isEditMode={!!jobId}
						disabled={isDiscovering}
					/>
				</div>
			</div>

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden border-t border-gray-200">
				<div
					className={clsx(
						"w-full pt-0 transition-all duration-300",
						currentStep !== JOB_CREATION_STEPS.STREAMS && "overflow-auto",
					)}
				>
					<div className="h-full">
						{currentStep === JOB_CREATION_STEPS.CONFIG && (
							<JobConfiguration
								jobName={jobName}
								setJobName={setJobName}
								cronExpression={cronExpression}
								setCronExpression={setCronExpression}
								stepNumber={JOB_STEP_NUMBERS.CONFIG}
								stepTitle="Job Configuration"
								advancedSettings={advancedSettings}
								setAdvancedSettings={setAdvancedSettings}
								selectedSourceId={selectedSourceId}
								setSelectedSourceId={setSelectedSourceId}
								selectedDestinationId={selectedDestinationId}
								setSelectedDestinationId={setSelectedDestinationId}
								isEditMode={true}
							/>
						)}

						{currentStep === JOB_CREATION_STEPS.STREAMS && (
							<div className="h-full overflow-auto">
								<SchemaConfiguration
									stepNumber={JOB_STEP_NUMBERS.STREAMS}
									stepTitle="Streams Selection"
									sourceName={sourceSnapshot?.name || ""}
									sourceConnector={normalizedSourceConnector}
									sourceVersion={sourceSnapshot?.version || ""}
									sourceConfig={sourceSnapshot?.config || "{}"}
									fromJobEditFlow={true}
									jobId={jobId ? parseInt(jobId) : -1}
									destinationType={
										destinationSnapshot
											? getConnectorInLowerCase(destinationSnapshot.type)
											: ""
									}
									jobName={jobName}
									advancedSettings={advancedSettings}
								/>
							</div>
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
				<div className="mr-[4%] flex gap-2 transition-[margin] duration-500 ease-in-out">
					{currentStep === JOB_CREATION_STEPS.CONFIG && jobId && (
						<button
							className="flex items-center justify-center gap-2 rounded-md border border-primary px-4 py-1 font-light text-primary hover:bg-primary-50"
							onClick={() => handleJobSubmit(null)}
							disabled={isSubmitting || isDiscovering}
						>
							{isSubmitting ? "Saving..." : "Save"}
						</button>
					)}
					<button
						className="flex items-center justify-center gap-2 rounded-md bg-primary px-4 py-1 font-light text-white hover:bg-primary-600"
						onClick={handleNext}
						disabled={isSubmitting || isDiscovering}
					>
						{currentStep === JOB_CREATION_STEPS.STREAMS
							? isSubmitting
								? "Saving..."
								: "Finish"
							: "Next"}
						{currentStep !== JOB_CREATION_STEPS.STREAMS && (
							<ArrowRightIcon className="size-4 text-white" />
						)}
					</button>
				</div>
			</div>
			<ResetStreamsModal onConfirm={handleConfirmResetStreams} />
			{streamDifference && (
				<StreamDifferenceModal
					streamDifference={streamDifference}
					onConfirm={() => handleJobSubmit(streamDifference)}
				/>
			)}
			<StreamEditDisabledModal from="jobEdit" />
		</div>
	)
}

export default JobEdit
