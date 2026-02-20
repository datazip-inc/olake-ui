import { useState, useRef } from "react"
import { useNavigate, Link, useLocation } from "react-router-dom"
import { message } from "antd"
import {
	ArrowLeftIcon,
	ArrowRightIcon,
	DownloadSimpleIcon,
} from "@phosphor-icons/react"
import { v4 as uuidv4 } from "uuid"

import { useAppStore } from "../../../store"
import { destinationService, sourceService } from "../../../api"
import { validationService } from "../../../api/services/validationService"

import { JobBase, JobCreationSteps, AdvancedSettings } from "../../../types"
import {
	getConnectorInLowerCase,
	getSelectedStreams,
	validateCronExpression,
	validateStreams,
} from "../../../utils/utils"
import {
	DESTINATION_INTERNAL_TYPES,
	ENTITY_TYPES,
	JOB_CREATION_STEPS,
	JOB_STEP_NUMBERS,
	TEST_CONNECTION_STATUS,
} from "../../../utils/constants"

// Internal imports from components
import JobConfiguration from "../components/JobConfiguration"
import StepProgress from "../components/StepIndicator"
import CreateSource from "../../sources/pages/CreateSource"
import CreateDestination from "../../destinations/pages/CreateDestination"
import SchemaConfiguration from "./SchemaConfiguration"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import TestConnectionFailureModal from "../../common/Modals/TestConnectionFailureModal"
import EntitySavedModal from "../../common/Modals/EntitySavedModal"
import EntityCancelModal from "../../common/Modals/EntityCancelModal"
import ResetStreamsModal from "../../common/Modals/ResetStreamsModal"

// TODO: Preserve "use existing" mode for source and destination when navigating between steps.

const JobCreation: React.FC = () => {
	const navigate = useNavigate()
	const location = useLocation()
	const initialData = location.state?.initialData || {}
	const savedJobId = location.state?.savedJobId

	const [currentStep, setCurrentStep] = useState<JobCreationSteps>(
		JOB_CREATION_STEPS.CONFIG as JobCreationSteps,
	)
	// state to toggle documentation panel
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [sourceName, setSourceName] = useState(initialData.sourceName || "")
	const [sourceConnector, setSourceConnector] = useState(
		initialData.sourceConnector || "MongoDB",
	)
	const [sourceFormData, setSourceFormData] = useState<any>(
		initialData.sourceFormData || {},
	)
	const [sourceVersion, setSourceVersion] = useState(
		initialData.sourceVersion || "",
	)
	const [destinationName, setDestinationName] = useState(
		initialData.destinationName || "",
	)
	const [existingSourceId, setExistingSourceId] = useState<number | null>(
		initialData.sourceId ? parseInt(initialData.sourceId) : null,
	)

	//state to hold catalog value to open documentation panel
	const [destinationCatalogType, setDestinationCatalogType] = useState<
		string | null
	>(null)

	const [destinationConnector, setDestinationConnector] = useState(
		initialData.destinationConnector || DESTINATION_INTERNAL_TYPES.S3,
	)
	const [destinationFormData, setDestinationFormData] = useState<any>(
		initialData.destinationFormData || {},
	)
	const [destinationVersion, setDestinationVersion] = useState(
		initialData.destinationVersion || "",
	)
	const [existingDestinationId, setExistingDestinationId] = useState<
		number | null
	>(initialData.destinationId ? parseInt(initialData.destinationId) : null)

	const [selectedStreams, setSelectedStreams] = useState<any>(
		initialData.selectedStreams || [],
	)
	const [jobName, setJobName] = useState(initialData.jobName || "")
	const [cronExpression, setCronExpression] = useState(
		initialData.cronExpression || "* * * * *",
	)
	const [advancedSettings, setAdvancedSettings] =
		useState<AdvancedSettings | null>(initialData.advancedSettings || null)

	//once the job name is filled we will set this to true so the job name will be disabled
	const [jobNameFilled, setJobNameFilled] = useState(
		initialData.isJobNameFilled || false,
	)
	//when streams are loading we will disable back button
	const [isStreamsLoading, setIsStreamsLoading] = useState(false)
	const [isFromSources, setIsFromSources] = useState(true)

	const {
		setShowEntitySavedModal,
		setShowSourceCancelModal,
		setShowTestingModal,
		setShowSuccessModal,
		addJob,
		setShowFailureModal,
		setSourceTestConnectionError,
		setDestinationTestConnectionError,
		setShowResetStreamsModal,
	} = useAppStore()

	const sourceRef = useRef<any>(null)
	const destinationRef = useRef<any>(null)

	const normalizedSourceConnector = getConnectorInLowerCase(sourceConnector)

	// Validation functions
	const validateSource = async (): Promise<boolean> => {
		if (sourceRef.current) {
			const isValid = await sourceRef.current.validateSource()
			if (!isValid) {
				message.error("Please fill in all required fields for the source")
				return false
			}
		} else if (!sourceName.trim() && sourceVersion.trim() != "") {
			message.error("Source name is required")
			return false
		}

		// Check uniqueness only if creating a new source (not using existing)
		if (!existingSourceId && sourceName.trim()) {
			const isUnique = await validationService.checkUniqueName(
				sourceName,
				ENTITY_TYPES.SOURCE,
			)
			if (!isUnique) return false
		}

		return true
	}

	const validateDestination = async (): Promise<boolean> => {
		if (destinationRef.current) {
			const isValid = await destinationRef.current.validateDestination()
			if (!isValid) {
				message.error("Please fill in all required fields for the destination")
				return false
			}
		} else if (!destinationName.trim() && destinationVersion.trim() != "") {
			message.error("Destination name is required")
			return false
		}

		// Check uniqueness only if creating a new destination (not using existing)
		if (!existingDestinationId && destinationName.trim()) {
			const isUnique = await validationService.checkUniqueName(
				destinationName,
				ENTITY_TYPES.DESTINATION,
			)
			if (!isUnique) return false
		}

		return true
	}

	const validateConfig = (): boolean => {
		if (!jobName.trim()) {
			message.error("Job name is required")
			return false
		}
		return validateCronExpression(cronExpression)
	}

	// Connection test handler
	const handleConnectionTest = async (
		isSource: boolean,
		data: any,
		nextStep: JobCreationSteps,
	): Promise<void> => {
		setShowTestingModal(true)
		try {
			const testResult = isSource
				? await sourceService.testSourceConnection(data, !!existingSourceId)
				: await destinationService.testDestinationConnection(
						data,
						!!existingDestinationId,
						normalizedSourceConnector,
						sourceVersion,
					)

			setTimeout(() => {
				setShowTestingModal(false)
				if (
					testResult.data?.connection_result.status ===
					TEST_CONNECTION_STATUS.SUCCEEDED
				) {
					setShowSuccessModal(true)
					setTimeout(() => {
						setShowSuccessModal(false)
						setCurrentStep(nextStep)
					}, 1000)
				} else {
					const testConnectionError = {
						message: testResult.data?.connection_result.message || "",
						logs: testResult.data?.logs || [],
					}
					setIsFromSources(isSource)
					if (isSource) {
						setSourceTestConnectionError(testConnectionError)
					} else {
						setDestinationTestConnectionError(testConnectionError)
					}
					setShowFailureModal(true)
				}
			}, 1500)
		} catch {
			setShowTestingModal(false)
			message.error(
				isSource
					? "Source connection test failed"
					: "Destination connection test failed",
			)
		}
	}

	// Job creation handler
	const handleJobCreation = async () => {
		const newJobData: JobBase = {
			name: jobName,
			source: {
				...(existingSourceId && { id: existingSourceId }),
				name: sourceName,
				type: normalizedSourceConnector,
				version: sourceVersion,
				config: JSON.stringify(sourceFormData),
			},
			destination: {
				...(existingDestinationId && { id: existingDestinationId }),
				name: destinationName,
				type: getConnectorInLowerCase(destinationConnector),
				version: destinationVersion,
				config: JSON.stringify(destinationFormData),
			},
			streams_config: JSON.stringify({
				...selectedStreams,
				selected_streams: getSelectedStreams(selectedStreams.selected_streams),
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

	// Main handler
	const handleNext = async () => {
		switch (currentStep) {
			case JOB_CREATION_STEPS.SOURCE: {
				if (!(await validateSource())) return
				const sourceData = {
					name: sourceName,
					type: normalizedSourceConnector,
					version: sourceVersion,
					config:
						typeof sourceFormData === "string"
							? sourceFormData
							: JSON.stringify(sourceFormData),
				}
				await handleConnectionTest(
					true,
					sourceData,
					JOB_CREATION_STEPS.DESTINATION,
				)
				break
			}
			case JOB_CREATION_STEPS.DESTINATION: {
				if (!(await validateDestination())) return
				const destinationData = {
					name: destinationName,
					type: getConnectorInLowerCase(destinationConnector),
					config:
						typeof destinationFormData === "string"
							? destinationFormData
							: JSON.stringify(destinationFormData),
					version: destinationVersion,
				}
				await handleConnectionTest(
					false,
					destinationData,
					JOB_CREATION_STEPS.STREAMS,
				)
				break
			}
			case JOB_CREATION_STEPS.STREAMS:
				if (
					!validateStreams(getSelectedStreams(selectedStreams.selected_streams))
				) {
					message.error("Filter Value cannot be empty")
					return
				}
				await handleJobCreation()
				break
			case JOB_CREATION_STEPS.CONFIG:
				if (!validateConfig()) return

				const isUnique = await validationService.checkUniqueName(
					jobName,
					ENTITY_TYPES.JOB,
				)
				if (!isUnique) return

				//TODO : Job name is disabled once filled and moved to next step , need to be handled later
				setJobNameFilled(true)
				setCurrentStep(JOB_CREATION_STEPS.SOURCE)
				break
			default:
				console.warn("Unknown step:", currentStep)
		}
	}

	//TODO: Handle steps properly

	const handleConfirmResetStreams = () => {
		setSelectedStreams([])
		setCurrentStep(JOB_CREATION_STEPS.DESTINATION)
	}

	const nextStep = () => {
		if (currentStep === JOB_CREATION_STEPS.SOURCE) {
			setCurrentStep(JOB_CREATION_STEPS.DESTINATION)
		} else if (currentStep === JOB_CREATION_STEPS.DESTINATION) {
			setCurrentStep(JOB_CREATION_STEPS.STREAMS)
		} else if (currentStep === JOB_CREATION_STEPS.CONFIG) {
			setCurrentStep(JOB_CREATION_STEPS.SOURCE)
		}
	}

	const handleBack = () => {
		if (currentStep === JOB_CREATION_STEPS.DESTINATION) {
			setCurrentStep(JOB_CREATION_STEPS.SOURCE)
		} else if (currentStep === JOB_CREATION_STEPS.STREAMS) {
			// when we go back from streams we need to show configured streams will be lost modal
			setShowResetStreamsModal(true)
		} else if (currentStep === JOB_CREATION_STEPS.SOURCE) {
			setCurrentStep(JOB_CREATION_STEPS.CONFIG)
		}
	}

	const handleCancel = () => {
		if (currentStep === JOB_CREATION_STEPS.SOURCE) {
			setShowSourceCancelModal(true)
		} else {
			message.info("Job creation cancelled")
			navigate("/jobs")
		}
	}

	const handleSaveJob = () => {
		const jobData = {
			id: savedJobId || uuidv4(),
			name: jobName,
			source: {
				name: sourceName,
				type: normalizedSourceConnector,
				version: sourceVersion,
				config: JSON.stringify(sourceFormData),
				...(existingSourceId && { id: existingSourceId }),
			},
			destination: {
				name: destinationName,
				type: getConnectorInLowerCase(destinationConnector),
				version: destinationVersion,
				config: JSON.stringify(destinationFormData),
				...(existingDestinationId && { id: existingDestinationId }),
			},
			streams_config: JSON.stringify(selectedStreams),
			frequency: cronExpression,
			advancedSettings: advancedSettings,
		}

		const savedJobs = JSON.parse(localStorage.getItem("savedJobs") || "[]")

		if (savedJobId) {
			// Update existing saved job
			const updatedSavedJobs = savedJobs.map((job: any) =>
				job.id === savedJobId ? jobData : job,
			)
			localStorage.setItem("savedJobs", JSON.stringify(updatedSavedJobs))
		} else {
			// Create new saved job
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
					className={`w-full ${currentStep === JOB_CREATION_STEPS.STREAMS ? "" : "overflow-hidden"} pt-0 transition-all duration-300`}
				>
					{currentStep === JOB_CREATION_STEPS.SOURCE && (
						<div className="h-full w-full overflow-auto">
							<CreateSource
								fromJobFlow={true}
								stepNumber={JOB_STEP_NUMBERS.SOURCE}
								stepTitle="Set up your source"
								initialExistingSourceId={existingSourceId}
								onSourceNameChange={setSourceName}
								onConnectorChange={setSourceConnector}
								onExistingSourceIdChange={setExistingSourceId}
								initialConnector={sourceConnector}
								onFormDataChange={data => {
									setSourceFormData(data)
								}}
								initialFormData={sourceFormData}
								initialName={sourceName}
								initialVersion={sourceVersion}
								onVersionChange={setSourceVersion}
								onComplete={() => {
									setCurrentStep(JOB_CREATION_STEPS.DESTINATION)
								}}
								ref={sourceRef}
								docsMinimized={docsMinimized}
								onDocsMinimizedChange={setDocsMinimized}
							/>
						</div>
					)}

					{currentStep === JOB_CREATION_STEPS.DESTINATION && (
						<div className="h-full w-full overflow-auto">
							<CreateDestination
								fromJobFlow={true}
								stepNumber={JOB_STEP_NUMBERS.DESTINATION}
								stepTitle="Set up your destination"
								initialExistingDestinationId={existingDestinationId}
								onExistingDestinationIdChange={setExistingDestinationId}
								onDestinationNameChange={setDestinationName}
								onConnectorChange={setDestinationConnector}
								initialConnector={getConnectorInLowerCase(destinationConnector)}
								initialVersion={destinationVersion}
								onFormDataChange={data => {
									setDestinationFormData(data)
								}}
								initialFormData={destinationFormData}
								initialName={destinationName}
								initialCatalog={destinationCatalogType}
								onCatalogTypeChange={setDestinationCatalogType}
								onVersionChange={setDestinationVersion}
								onComplete={() => {
									setCurrentStep(JOB_CREATION_STEPS.STREAMS)
								}}
								ref={destinationRef}
								docsMinimized={docsMinimized}
								onDocsMinimizedChange={setDocsMinimized}
								sourceConnector={sourceConnector}
								sourceVersion={sourceVersion}
							/>
						</div>
					)}

					{currentStep === JOB_CREATION_STEPS.STREAMS && (
						<div className="h-full overflow-scroll">
							<SchemaConfiguration
								selectedStreams={selectedStreams}
								setSelectedStreams={setSelectedStreams}
								stepNumber={JOB_STEP_NUMBERS.STREAMS}
								stepTitle="Streams Selection"
								useDirectForms={true}
								sourceName={sourceName}
								sourceConnector={normalizedSourceConnector}
								sourceVersion={sourceVersion}
								sourceConfig={
									typeof sourceFormData === "string"
										? sourceFormData
										: JSON.stringify(sourceFormData)
								}
								initialStreamsData={
									selectedStreams &&
									selectedStreams.selected_streams &&
									Object.keys(selectedStreams.selected_streams).length > 0
										? selectedStreams
										: undefined
								}
								destinationType={getConnectorInLowerCase(destinationConnector)}
								jobName={jobName}
								onLoadingChange={setIsStreamsLoading}
								advancedSettings={advancedSettings}
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
							jobNameFilled={jobNameFilled}
							advancedSettings={advancedSettings}
							setAdvancedSettings={setAdvancedSettings}
						/>
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
				<div
					className={`flex items-center transition-[margin] duration-500 ease-in-out ${!docsMinimized && (currentStep === JOB_CREATION_STEPS.SOURCE || currentStep === JOB_CREATION_STEPS.DESTINATION) ? "mr-[40%]" : "mr-[4%]"}`}
				>
					{currentStep !== JOB_CREATION_STEPS.CONFIG && (
						<button
							onClick={handleBack}
							className="mr-4 rounded-md border border-gray-400 px-4 py-1 font-light hover:bg-[#ebebeb] disabled:cursor-not-allowed disabled:opacity-50"
							disabled={
								currentStep === JOB_CREATION_STEPS.STREAMS && isStreamsLoading
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
					<TestConnectionModal />
					<TestConnectionSuccessModal />
					<EntitySavedModal
						type={currentStep}
						onComplete={nextStep}
						fromJobFlow={true}
						entityName={
							currentStep === JOB_CREATION_STEPS.SOURCE
								? sourceName
								: currentStep === JOB_CREATION_STEPS.DESTINATION
									? destinationName
									: currentStep === JOB_CREATION_STEPS.STREAMS
										? jobName
										: ""
						}
					/>
					<TestConnectionFailureModal fromSources={isFromSources} />
					<EntityCancelModal
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
