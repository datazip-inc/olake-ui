import { useState, useEffect } from "react"
import { useNavigate, Link, useLocation } from "react-router-dom"
import { message } from "antd"
import {
	ArrowLeftIcon,
	ArrowRightIcon,
	DownloadSimpleIcon,
} from "@phosphor-icons/react"
import { v4 as uuidv4 } from "uuid"

import { useCreateJob } from "../hooks"
import { useTestDestinationConnection } from "@/modules/ingestion/features/destinations/hooks"
import { useTestSourceConnection } from "@/modules/ingestion/features/sources/hooks"
import { useSources } from "@/modules/ingestion/features/sources/hooks"
import { useDestinations } from "@/modules/ingestion/features/destinations/hooks"
import {
	useJobStore,
	useStreamSelectionStore,
	useJobConfigurationStore,
} from "../stores"
import { validationService } from "@/modules/ingestion/common/services/validationService"

import { JobBase, JobCreationSteps } from "../types"
import type { TestConnectionError } from "@/common/types"
import {
	buildConnectorPayload,
	validateCronExpression,
	getSelectedStreams,
	validateStreams,
} from "../utils"
import { ENTITY_TYPES } from "@/modules/ingestion/common/constants"
import { TEST_CONNECTION_STATUS } from "@/common/constants"
import { JOB_CREATION_STEPS, JOB_STEP_NUMBERS } from "../constants"

// Internal imports from components
import {
	JobConfiguration,
	StepIndicator as StepProgress,
	SchemaConfiguration,
	ResetStreamsModal,
} from "../components"
import {
	TestConnectionModal,
	TestConnectionSuccessModal,
	TestConnectionFailureModal,
	EntitySavedModal,
	EntityCancelModal,
} from "@/modules/ingestion/common/components"

const JobCreation: React.FC = () => {
	const navigate = useNavigate()
	const location = useLocation()
	const initialData = location.state?.initialData || {}
	const savedJobId = location.state?.savedJobId

	const [currentStep, setCurrentStep] = useState<JobCreationSteps>(
		JOB_CREATION_STEPS.CONFIG as JobCreationSteps,
	)

	// Config step states
	const {
		jobName,
		cronExpression,
		advancedSettings,
		selectedSource,
		selectedDestination,
		setJobName,
		setCronExpression,
		setSelectedSource,
		setSelectedDestination,
		setAdvancedSettings,
		setIsEditMode,
		reset: resetJobConfig,
	} = useJobConfigurationStore()

	const streamsData = useStreamSelectionStore(state => state.streamsData)

	// Initialize the store exactly once using initialData if navigating
	useEffect(() => {
		if (initialData.jobName) setJobName(initialData.jobName)
		if (initialData.cronExpression)
			setCronExpression(initialData.cronExpression)
		if (initialData.advanced_settings)
			setAdvancedSettings(initialData.advanced_settings)
		setIsEditMode(false)

		return () => {
			resetJobConfig()
		}
	}, [])

	// Load sources and destinations lists for dropdown resolution
	const { data: sourcesData } = useSources()
	const { data: destinationsData } = useDestinations()

	// Pre-fill full source entity from URL param ID once sources load
	useEffect(() => {
		if (initialData.sourceId && sourcesData && !selectedSource) {
			const source = (sourcesData as any[]).find(
				(s: any) => s.id === parseInt(initialData.sourceId!),
			)
			if (source) setSelectedSource(source)
		}
	}, [initialData.sourceId, sourcesData, selectedSource, setSelectedSource])

	// Pre-fill full destination entity from URL param ID once destinations load
	useEffect(() => {
		if (initialData.destinationId && destinationsData && !selectedDestination) {
			const dest = (destinationsData as any[]).find(
				(d: any) => d.id === parseInt(initialData.destinationId!),
			)
			if (dest) setSelectedDestination(dest)
		}
	}, [
		initialData.destinationId,
		destinationsData,
		selectedDestination,
		setSelectedDestination,
	])

	const { setShowResetStreamsModal } = useJobStore()
	const isDiscovering = useStreamSelectionStore(state => state.isDiscovering)

	const [showTestingModal, setShowTestingModal] = useState(false)
	const [showSuccessModal, setShowSuccessModal] = useState(false)
	const [showFailureModal, setShowFailureModal] = useState(false)
	const [showEntitySavedModal, setShowEntitySavedModal] = useState(false)
	const [showCancelModal, setShowCancelModal] = useState(false)
	const { mutateAsync: addJob } = useCreateJob()
	const testSourceMutation = useTestSourceConnection()
	const testDestinationMutation = useTestDestinationConnection()
	const [testConnectionError, setTestConnectionError] =
		useState<TestConnectionError | null>(null)
	// Track which entity the connection test is for (source vs destination)
	const [connectionTestType, setConnectionTestType] = useState<
		"source" | "destination"
	>("source")

	const sourceConnectorPayload = buildConnectorPayload(selectedSource)
	const destinationConnectorPayload = buildConnectorPayload(selectedDestination)

	const validateConfig = (): boolean => {
		if (!jobName.trim()) {
			message.error("Job name is required")
			return false
		}
		if (!validateCronExpression(cronExpression)) return false
		if (!selectedSource?.id) {
			message.error("Please select a source")
			return false
		}
		if (!selectedDestination?.id) {
			message.error("Please select a destination")
			return false
		}
		return true
	}

	const runConnectionTest = async (isSource: boolean): Promise<boolean> => {
		if (isSource && !selectedSource) return false
		if (!isSource && !selectedDestination) return false
		setConnectionTestType(isSource ? "source" : "destination")
		setShowTestingModal(true)
		try {
			const testResult = isSource
				? await testSourceMutation.mutateAsync({
						source: {
							type: sourceConnectorPayload.type,
							version: sourceConnectorPayload.version,
							config: sourceConnectorPayload.config,
						},
						existing: true,
					})
				: await testDestinationMutation.mutateAsync({
						destination: {
							type: destinationConnectorPayload.type,
							version: destinationConnectorPayload.version,
							config: destinationConnectorPayload.config,
						},
						existing: true,
						sourceType: sourceConnectorPayload.type,
						sourceVersion: sourceConnectorPayload.version,
					})
			setShowTestingModal(false)
			if (
				testResult.data?.connection_result.status ===
				TEST_CONNECTION_STATUS.SUCCEEDED
			) {
				setShowSuccessModal(true)
				await new Promise(resolve => setTimeout(resolve, 1000))
				setShowSuccessModal(false)
				return true
			}
			setTestConnectionError({
				message: testResult.data?.connection_result.message || "",
				logs: testResult.data?.logs || [],
			})
			setShowFailureModal(true)
			return false
		} catch {
			setShowTestingModal(false)
			message.error(
				isSource
					? "Source connection test failed"
					: "Destination connection test failed",
			)
			return false
		}
	}

	// Job creation handler
	const handleJobCreation = async () => {
		const newJobData: JobBase = {
			name: jobName,
			source: {
				...(selectedSource?.id && { id: selectedSource.id }),
				name: selectedSource?.name ?? "",
				type: sourceConnectorPayload.type,
				version: sourceConnectorPayload.version,
				config: sourceConnectorPayload.config,
			},
			destination: {
				...(selectedDestination?.id && { id: selectedDestination.id }),
				name: selectedDestination?.name ?? "",
				type: destinationConnectorPayload.type,
				version: destinationConnectorPayload.version,
				config: destinationConnectorPayload.config,
			},
			streams_config: JSON.stringify({
				...streamsData,
				selected_streams: getSelectedStreams(streamsData!.selected_streams),
			}),
			frequency: cronExpression,
			advanced_settings: advancedSettings,
		}

		try {
			await addJob(newJobData)
			if (savedJobId) {
				const savedJobs = JSON.parse(localStorage.getItem("savedJobs") || "[]")
				const updatedSavedJobs = savedJobs.filter(
					(job: any) => job.id !== savedJobId,
				)
				localStorage.setItem("savedJobs", JSON.stringify(updatedSavedJobs))
			}
			setShowEntitySavedModal(true)
		} catch (error) {
			console.error("Error adding job:", error)
		}
	}

	// Main next handler — 2-step flow: CONFIG → STREAMS
	const handleNext = async () => {
		switch (currentStep) {
			case JOB_CREATION_STEPS.CONFIG: {
				if (!validateConfig()) return

				// Check job name uniqueness
				const isUnique = await validationService.checkUniqueName(
					jobName,
					ENTITY_TYPES.JOB,
				)
				if (!isUnique) return

				const sourceOk = await runConnectionTest(true)
				if (!sourceOk) return

				const destOk = await runConnectionTest(false)
				if (!destOk) return

				setCurrentStep(JOB_CREATION_STEPS.STREAMS)
				break
			}
			case JOB_CREATION_STEPS.STREAMS: {
				if (
					!streamsData ||
					!validateStreams(getSelectedStreams(streamsData.selected_streams))
				) {
					message.error("Filter Value cannot be empty")
					return
				}
				await handleJobCreation()
				break
			}
			default:
				console.warn("Unknown step:", currentStep)
		}
	}

	const handleConfirmResetStreams = () => {
		useStreamSelectionStore.getState().reset()
		setCurrentStep(JOB_CREATION_STEPS.CONFIG)
	}

	const handleBack = () => {
		if (currentStep === JOB_CREATION_STEPS.STREAMS) {
			// Warn user configured streams will be lost
			setShowResetStreamsModal(true)
		}
	}

	const handleCancel = () => {
		setShowCancelModal(true)
	}

	const handleSaveJob = () => {
		const draftStreams = streamsData ?? { selected_streams: {}, streams: [] }
		const jobData = {
			id: savedJobId || uuidv4(),
			name: jobName,
			source: {
				name: selectedSource?.name ?? "",
				type: selectedSource?.type ?? "",
				id: selectedSource?.id ?? undefined,
			},
			destination: {
				name: selectedDestination?.name ?? "",
				type: selectedDestination?.type ?? "",
				id: selectedDestination?.id ?? undefined,
			},
			streams_config: JSON.stringify(draftStreams),
			frequency: cronExpression,
			advanced_settings: advancedSettings,
		}

		const savedJobs = JSON.parse(localStorage.getItem("savedJobs") || "[]")

		if (savedJobId) {
			const updatedSavedJobs = savedJobs.map((job: any) =>
				job.id === savedJobId ? jobData : job,
			)
			localStorage.setItem("savedJobs", JSON.stringify(updatedSavedJobs))
		} else {
			savedJobs.push(jobData)
			localStorage.setItem("savedJobs", JSON.stringify(savedJobs))
		}

		navigate("/jobs")
	}

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

						<div className="text-2xl font-bold"> Create Job</div>
					</div>
					{/* Stepper */}
					<StepProgress currentStep={currentStep} />
				</div>
			</div>

			<div className="flex flex-1 overflow-hidden border-t border-gray-200">
				<div
					className={`w-full ${currentStep === JOB_CREATION_STEPS.STREAMS ? "" : "overflow-auto"} pt-0 transition-all duration-300`}
				>
					{currentStep === JOB_CREATION_STEPS.CONFIG && (
						<JobConfiguration
							stepNumber={JOB_STEP_NUMBERS.CONFIG}
							stepTitle="Job Configuration"
						/>
					)}

					{currentStep === JOB_CREATION_STEPS.STREAMS && (
						<div className="h-full overflow-scroll">
							<SchemaConfiguration
								stepNumber={JOB_STEP_NUMBERS.STREAMS}
								stepTitle="Streams Selection"
								sourceName={selectedSource?.name ?? ""}
								sourceConnector={sourceConnectorPayload.type}
								sourceVersion={sourceConnectorPayload.version}
								sourceConfig={sourceConnectorPayload.config}
								destinationType={destinationConnectorPayload.type}
								jobName={jobName}
								advancedSettings={advancedSettings}
							/>
						</div>
					)}
				</div>
			</div>

			{/* Footer */}
			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<div className="flex space-x-4">
					<button
						className="rounded-md border border-danger px-4 py-1 text-danger hover:bg-danger hover:text-white"
						onClick={handleCancel}
					>
						Cancel
					</button>
					<button
						onClick={handleSaveJob}
						className="flex items-center justify-center gap-2 rounded-md border border-gray-400 px-4 py-1 font-light hover:bg-[#ebebeb]"
					>
						<DownloadSimpleIcon className="size-4" />
						Save Job
					</button>
				</div>
				<div className="mr-[4%] flex items-center transition-[margin] duration-500 ease-in-out">
					{currentStep !== JOB_CREATION_STEPS.CONFIG && (
						<button
							onClick={handleBack}
							className="mr-4 rounded-md border border-gray-400 px-4 py-1 font-light hover:bg-[#ebebeb] disabled:cursor-not-allowed disabled:opacity-50"
							disabled={
								currentStep === JOB_CREATION_STEPS.STREAMS && isDiscovering
							}
						>
							Back
						</button>
					)}
					<button
						className="flex items-center justify-center gap-2 rounded-md bg-primary px-4 py-1 font-light text-white hover:bg-primary-600"
						onClick={handleNext}
					>
						{currentStep === JOB_CREATION_STEPS.STREAMS ? "Create Job" : "Next"}
						<ArrowRightIcon className="size-4 text-white" />
					</button>
					<TestConnectionModal
						open={showTestingModal}
						connectionType={connectionTestType}
					/>
					<TestConnectionSuccessModal
						open={showSuccessModal}
						connectionType={connectionTestType}
					/>
					<EntitySavedModal
						open={showEntitySavedModal}
						onClose={() => setShowEntitySavedModal(false)}
						type={JOB_CREATION_STEPS.STREAMS}
						fromJobFlow={true}
						entityName={jobName}
					/>
					<TestConnectionFailureModal
						open={showFailureModal}
						onClose={() => setShowFailureModal(false)}
						connectionType={connectionTestType}
						testConnectionError={testConnectionError}
					/>
					<EntityCancelModal
						open={showCancelModal}
						onClose={() => setShowCancelModal(false)}
						type="job"
						navigateTo="jobs"
					/>
				</div>
			</div>
			<ResetStreamsModal onConfirm={handleConfirmResetStreams} />
		</div>
	)
}

export default JobCreation
