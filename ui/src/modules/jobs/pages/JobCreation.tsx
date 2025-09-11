import { useState, useRef } from "react"
import { useNavigate, Link, useLocation } from "react-router-dom"
import { message } from "antd"
import { ArrowLeft, ArrowRight, DownloadSimple } from "@phosphor-icons/react"
import { v4 as uuidv4 } from "uuid"

import { useAppStore } from "../../../store"
import { destinationService, sourceService, jobService } from "../../../api"

import { JobBase, JobCreationSteps } from "../../../types"
import {
	getConnectorInLowerCase,
	validateCronExpression,
} from "../../../utils/utils"
import { DESTINATION_INTERNAL_TYPES } from "../../../utils/constants"

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

const JobCreation: React.FC = () => {
	const navigate = useNavigate()
	const location = useLocation()
	const initialData = location.state?.initialData || {}
	const savedJobId = location.state?.savedJobId

	const [currentStep, setCurrentStep] = useState<JobCreationSteps>("config")
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
	const [selectedStreams, setSelectedStreams] = useState<any>(
		initialData.selectedStreams || [],
	)
	const [jobName, setJobName] = useState(initialData.jobName || "")
	const [cronExpression, setCronExpression] = useState(
		initialData.cronExpression || "* * * * *",
	)
	const [jobNameFilled, setJobNameFilled] = useState(false)
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
	} = useAppStore()

	const sourceRef = useRef<any>(null)
	const destinationRef = useRef<any>(null)

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
		return true
	}

	const validateConfig = (): boolean => {
		if (!jobName.trim()) {
			message.error("Job name is required")
			return false
		}
		return validateCronExpression(cronExpression)
	}

	const checkJobNameUnique = async (): Promise<boolean> => {
		try {
			const response = await jobService.checkJobNameUnique(jobName)
			return response.unique
		} catch (error) {
			console.error("Error checking job name uniqueness:", error)
			message.error("Failed to check job name uniqueness. Please try again.")
			return false
		}
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
				? await sourceService.testSourceConnection(data)
				: await destinationService.testDestinationConnection(
						data,
						getConnectorInLowerCase(sourceConnector),
						sourceVersion,
					)

			setTimeout(() => {
				setShowTestingModal(false)
				if (testResult.data?.status === "SUCCEEDED") {
					setShowSuccessModal(true)
					setTimeout(() => {
						setShowSuccessModal(false)
						setCurrentStep(nextStep)
					}, 1000)
				} else {
					setIsFromSources(isSource)
					if (isSource) {
						setSourceTestConnectionError(testResult.data?.message || "")
					} else {
						setDestinationTestConnectionError(testResult.data?.message || "")
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
				name: sourceName,
				type: getConnectorInLowerCase(sourceConnector),
				version: sourceVersion,
				config: JSON.stringify(sourceFormData),
			},
			destination: {
				name: destinationName,
				type: getConnectorInLowerCase(destinationConnector),
				version: destinationVersion,
				config: JSON.stringify(destinationFormData),
			},
			streams_config: JSON.stringify(selectedStreams),
			frequency: cronExpression,
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
			message.error("Failed to create job")
		}
	}

	// Main handler
	const handleNext = async () => {
		switch (currentStep) {
			case "source": {
				if (!(await validateSource())) return
				const sourceData = {
					name: sourceName,
					type: getConnectorInLowerCase(sourceConnector),
					version: sourceVersion,
					config:
						typeof sourceFormData === "string"
							? sourceFormData
							: JSON.stringify(sourceFormData),
				}
				await handleConnectionTest(true, sourceData, "destination")
				break
			}
			case "destination": {
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
				await handleConnectionTest(false, destinationData, "streams")
				break
			}
			case "streams":
				await handleJobCreation()
				break
			case "config":
				if (!validateConfig()) return

				// Check job name uniqueness before proceeding
				const isUnique = await checkJobNameUnique()
				if (!isUnique) {
					message.error(
						"Job name already exists. Please choose a different name.",
					)
					return
				}

				setJobNameFilled(true)
				setCurrentStep("source")
				break
			default:
				console.warn("Unknown step:", currentStep)
		}
	}

	const nextStep = () => {
		if (currentStep === "source") {
			setCurrentStep("destination")
		} else if (currentStep === "destination") {
			setCurrentStep("streams")
		} else if (currentStep === "config") {
			setCurrentStep("source")
		}
	}

	const handleBack = () => {
		if (currentStep === "destination") {
			setCurrentStep("source")
		} else if (currentStep === "streams") {
			setCurrentStep("destination")
		} else if (currentStep === "source") {
			setCurrentStep("config")
		}
	}

	const handleCancel = () => {
		if (currentStep === "source") {
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
				type: getConnectorInLowerCase(sourceConnector),
				version: sourceVersion,
				config: JSON.stringify(sourceFormData),
			},
			destination: {
				name: destinationName,
				type: getConnectorInLowerCase(destinationConnector),
				version: destinationVersion,
				config: JSON.stringify(destinationFormData),
			},
			streams_config: JSON.stringify(selectedStreams),
			frequency: cronExpression,
		}

		const savedJobs = JSON.parse(localStorage.getItem("savedJobs") || "[]")

		if (savedJobId) {
			// Update existing saved job
			const updatedSavedJobs = savedJobs.map((job: any) =>
				job.id === savedJobId ? jobData : job,
			)
			localStorage.setItem("savedJobs", JSON.stringify(updatedSavedJobs))
			message.success("Job saved successfully!")
		} else {
			// Create new saved job
			savedJobs.push(jobData)
			localStorage.setItem("savedJobs", JSON.stringify(savedJobs))
			message.success("Job saved successfully!")
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
							<ArrowLeft className="mr-1 size-5" />
						</Link>

						<div className="text-2xl font-bold"> Create Job</div>
					</div>
					{/* Stepper */}
					<StepProgress currentStep={currentStep} />
				</div>
			</div>

			<div className="flex flex-1 overflow-hidden border-t border-gray-200">
				<div
					className={`w-full ${currentStep === "streams" ? "" : "overflow-hidden"} pt-0 transition-all duration-300`}
				>
					{currentStep === "source" && (
						<div className="h-full w-full overflow-auto">
							<CreateSource
								fromJobFlow={true}
								stepNumber={2}
								stepTitle="Set up your source"
								onSourceNameChange={setSourceName}
								onConnectorChange={setSourceConnector}
								initialConnector={sourceConnector}
								onFormDataChange={data => {
									setSourceFormData(data)
								}}
								initialFormData={sourceFormData}
								initialName={sourceName}
								initialVersion={sourceVersion}
								onVersionChange={setSourceVersion}
								onComplete={() => {
									setCurrentStep("destination")
								}}
								ref={sourceRef}
								docsMinimized={docsMinimized}
								onDocsMinimizedChange={setDocsMinimized}
							/>
						</div>
					)}

					{currentStep === "destination" && (
						<div className="h-full w-full overflow-auto">
							<CreateDestination
								fromJobFlow={true}
								stepNumber={3}
								stepTitle="Set up your destination"
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
									setCurrentStep("streams")
								}}
								ref={destinationRef}
								docsMinimized={docsMinimized}
								onDocsMinimizedChange={setDocsMinimized}
								sourceConnector={sourceConnector}
								sourceVersion={sourceVersion}
							/>
						</div>
					)}

					{currentStep === "streams" && (
						<div className="h-full overflow-scroll">
							<SchemaConfiguration
								selectedStreams={selectedStreams}
								setSelectedStreams={setSelectedStreams}
								stepNumber={4}
								stepTitle="Streams Selection"
								useDirectForms={true}
								sourceName={sourceName}
								sourceConnector={getConnectorInLowerCase(sourceConnector)}
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
							/>
						</div>
					)}

					{currentStep === "config" && (
						<JobConfiguration
							jobName={jobName}
							setJobName={setJobName}
							cronExpression={cronExpression}
							setCronExpression={setCronExpression}
							stepNumber={1}
							stepTitle="Job Configuration"
							jobNameFilled={jobNameFilled}
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
						<DownloadSimple className="size-4" />
						Save Job
					</button>
				</div>
				<div
					className={`flex items-center transition-[margin] duration-500 ease-in-out ${!docsMinimized && (currentStep === "source" || currentStep === "destination") ? "mr-[40%]" : "mr-[4%]"}`}
				>
					{currentStep !== "config" && (
						<button
							onClick={handleBack}
							className="mr-4 rounded-md border border-gray-400 px-4 py-1 font-light hover:bg-[#ebebeb]"
						>
							Back
						</button>
					)}
					<button
						className="flex items-center justify-center gap-2 rounded-md bg-primary px-4 py-1 font-light text-white hover:bg-primary-600"
						onClick={handleNext}
					>
						{currentStep === "streams" ? "Create Job" : "Next"}
						<ArrowRight className="size-4 text-white" />
					</button>
					<TestConnectionModal />
					<TestConnectionSuccessModal />
					<EntitySavedModal
						type={currentStep}
						onComplete={nextStep}
						fromJobFlow={true}
						entityName={
							currentStep === "source"
								? sourceName
								: currentStep === "destination"
									? destinationName
									: currentStep === "streams"
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
		</div>
	)
}

export default JobCreation
