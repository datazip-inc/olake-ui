import { useState, useRef } from "react"
import { useNavigate, Link } from "react-router-dom"
import { message } from "antd"
import CreateSource, {
	CreateSourceHandle,
} from "../../sources/pages/CreateSource"
import CreateDestination, {
	CreateDestinationHandle,
} from "../../destinations/pages/CreateDestination"
import { ArrowLeft, ArrowRight, DownloadSimple } from "@phosphor-icons/react"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import StepProgress from "../components/StepIndicator"
import { useAppStore } from "../../../store"
import EntitySavedModal from "../../common/Modals/EntitySavedModal"
import SchemaConfiguration from "./SchemaConfiguration"
import JobConfiguration from "../components/JobConfiguration"
import EntityCancelModal from "../../common/Modals/EntityCancelModal"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import { JobBase, JobCreationSteps } from "../../../types"
import { getConnectorInLowerCase } from "../../../utils/utils"
import { destinationService, sourceService } from "../../../api"

const JobCreation: React.FC = () => {
	const navigate = useNavigate()
	const [currentStep, setCurrentStep] = useState<JobCreationSteps>("source")
	const [docsMinimized, setDocsMinimized] = useState(true)

	// Source and destination states
	const [sourceName, setSourceName] = useState("")
	const [sourceConnector, setSourceConnector] = useState("MongoDB")
	const [sourceFormData, setSourceFormData] = useState<any>({})
	const [sourceVersion, setSourceVersion] = useState("latest")
	const [destinationName, setDestinationName] = useState("")
	const [destinationConnector, setDestinationConnector] = useState("s3")
	const [destinationFormData, setDestinationFormData] = useState<any>({})
	const [destinationVersion, setDestinationVersion] = useState("latest")
	const [selectedStreams, setSelectedStreams] = useState<any>([])
	const [jobName, setJobName] = useState("")
	const [replicationFrequency, setReplicationFrequency] = useState("minutes")
	const [schemaChangeStrategy, setSchemaChangeStrategy] = useState("propagate")
	const [notifyOnSchemaChanges, setNotifyOnSchemaChanges] = useState(true)

	const {
		setShowEntitySavedModal,
		setShowSourceCancelModal,
		setShowTestingModal,
		setShowSuccessModal,
		addJob,
		setShowFailureModal,
	} = useAppStore()

	const sourceRef = useRef<CreateSourceHandle>(null)
	const destinationRef = useRef<CreateDestinationHandle>(null)

	const getReplicationFrequency = () => {
		if (replicationFrequency === "minutes") {
			return "minutes"
		} else if (replicationFrequency === "hours") {
			return "hourly"
		} else if (replicationFrequency === "days") {
			return "daily"
		} else if (replicationFrequency === "weeks") {
			return "weekly"
		}
	}

	const handleNext = async () => {
		if (currentStep === "source") {
			// Call the source component's validation
			if (sourceRef.current) {
				const isValid = await sourceRef.current.validateSource()
				if (!isValid) {
					message.error("Please fill in all required fields for the source")
					return
				}
			} else {
				// Fallback validation if ref isn't available
				if (!sourceName.trim()) {
					message.error("Source name is required")
					return
				}
			}

			const newSourceData = {
				name: sourceName,
				type: sourceConnector.toLowerCase(),
				version: sourceVersion,
				config: JSON.stringify(sourceFormData),
			}
			setShowTestingModal(true)
			const testResult = await sourceService.testSourceConnection(newSourceData)

			setTimeout(() => {
				setShowTestingModal(false)
				if (testResult.success) {
					setShowSuccessModal(true)
					setTimeout(() => {
						setShowSuccessModal(false)
						setCurrentStep("destination")
					}, 1000)
				} else {
					setShowFailureModal(true)
				}
			}, 1500)
		} else if (currentStep === "destination") {
			// Call the destination component's validation
			if (destinationRef.current) {
				const isValid = await destinationRef.current.validateDestination()
				if (!isValid) {
					message.error(
						"Please fill in all required fields for the destination",
					)
					return
				}
			} else {
				// Fallback validation if ref isn't available
				if (!destinationName.trim()) {
					message.error("Destination name is required")
					return
				}
			}

			const newDestinationData = {
				name: destinationName,
				type: destinationConnector,
				config: JSON.stringify(destinationFormData),
				version: destinationVersion,
			}
			setShowTestingModal(true)
			const testResult =
				await destinationService.testDestinationConnection(newDestinationData)

			// Ensure testing modal stays visible for at least 1.5 seconds
			setTimeout(() => {
				setShowTestingModal(false)
				if (testResult.success) {
					setShowSuccessModal(true)
					setTimeout(() => {
						setShowSuccessModal(false)
						setCurrentStep("schema")
					}, 1000)
				} else {
					setShowFailureModal(true)
				}
			}, 1500)
		} else if (currentStep === "schema") {
			setCurrentStep("config")
		} else if (currentStep === "config") {
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
				frequency: replicationFrequency
					? getReplicationFrequency() || "hourly"
					: "hourly",
			}
			addJob(newJobData)
				.then(() => {
					setShowEntitySavedModal(true)
				})
				.catch(error => {
					console.error("Error adding job:", error)
					message.error("Failed to create job")
				})
		}
	}

	const nextStep = () => {
		if (currentStep === "source") {
			setCurrentStep("destination")
		} else if (currentStep === "destination") {
			setCurrentStep("schema")
		} else if (currentStep === "schema") {
			setCurrentStep("config")
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

	const handleCancel = () => {
		if (currentStep === "source") {
			setShowSourceCancelModal(true)
		} else {
			message.info("Job creation cancelled")
			navigate("/jobs")
		}
	}

	const handleSaveJob = () => {
		const savedJob = {
			name: jobName || "-",
			source: {
				name: sourceName || "-",
				type: getConnectorInLowerCase(sourceConnector),
				version: sourceVersion,
				config: JSON.stringify(sourceFormData),
			},
			destination: {
				name: destinationName || "-",
				type: getConnectorInLowerCase(destinationConnector),
				version: destinationVersion,
				config: JSON.stringify(destinationFormData),
			},
			streams_config: JSON.stringify(selectedStreams),
			frequency: replicationFrequency
				? getReplicationFrequency() || "hourly"
				: "hourly",
			activate: false,
			created_at: new Date().toISOString(),
			updated_at: new Date().toISOString(),
			created_by: "user",
			updated_by: "user",
			last_run_state: "",
			last_run_time: "",
		}
		const existingSavedJobs = JSON.parse(
			localStorage.getItem("savedJobs") || "[]",
		)
		existingSavedJobs.push(savedJob)
		localStorage.setItem("savedJobs", JSON.stringify(existingSavedJobs))
		message.success("Job saved successfully!")
		navigate("/jobs")
	}

	const toggleDocsPanel = () => {
		setDocsMinimized(!docsMinimized)
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

						<div className="text-2xl font-bold"> Create job</div>
					</div>
					{/* Stepper */}
					<StepProgress currentStep={currentStep} />
				</div>
			</div>

			<div className="flex flex-1 overflow-hidden border-t border-gray-200">
				<div
					className={`${
						(currentStep === "schema" || currentStep === "config") &&
						!docsMinimized
							? "w-2/3"
							: "w-full"
					} pt-0 transition-all duration-300`}
				>
					{currentStep === "source" && (
						<div className="w-full">
							<CreateSource
								fromJobFlow={true}
								stepNumber={"I"}
								stepTitle="Set up your source"
								onSourceNameChange={setSourceName}
								onConnectorChange={setSourceConnector}
								onFormDataChange={data => {
									setSourceFormData(data)
								}}
								initialFormData={sourceFormData}
								onVersionChange={setSourceVersion}
								onComplete={() => {
									setCurrentStep("destination")
								}}
								ref={sourceRef}
							/>
						</div>
					)}

					{currentStep === "destination" && (
						<div className="w-full">
							<CreateDestination
								fromJobFlow={true}
								stepNumber={2}
								stepTitle="Set up your destination"
								onDestinationNameChange={setDestinationName}
								onConnectorChange={setDestinationConnector}
								onFormDataChange={data => {
									setDestinationFormData(data)
								}}
								initialFormData={destinationFormData}
								onVersionChange={setDestinationVersion}
								onComplete={() => {
									setCurrentStep("schema")
								}}
								ref={destinationRef}
							/>
						</div>
					)}

					{currentStep === "schema" && (
						<SchemaConfiguration
							selectedStreams={selectedStreams}
							setSelectedStreams={setSelectedStreams}
							stepNumber={3}
							stepTitle="Streams selection"
							useDirectForms={true}
							sourceName={sourceName}
							sourceConnector={sourceConnector.toLowerCase()}
							sourceVersion={sourceVersion}
							sourceConfig={JSON.stringify(sourceFormData)}
						/>
					)}

					{currentStep === "config" && (
						<JobConfiguration
							jobName={jobName}
							setJobName={setJobName}
							replicationFrequency={replicationFrequency}
							setReplicationFrequency={setReplicationFrequency}
							schemaChangeStrategy={schemaChangeStrategy}
							setSchemaChangeStrategy={setSchemaChangeStrategy}
							notifyOnSchemaChanges={notifyOnSchemaChanges}
							setNotifyOnSchemaChanges={setNotifyOnSchemaChanges}
							stepNumber={4}
							stepTitle="Job configuration"
						/>
					)}
				</div>

				{/* Documentation panel */}
				{currentStep === "schema" && (
					<DocumentationPanel
						docUrl={`https://olake.io/docs/connectors/${sourceConnector.toLowerCase()}/config`}
						isMinimized={docsMinimized}
						onToggle={toggleDocsPanel}
						showResizer={true}
					/>
				)}
			</div>

			{/* Footer */}
			<div className="flex justify-between border-t border-gray-200 bg-white p-4">
				<div className="flex space-x-4">
					<button
						className="rounded-[6px] border border-[#F5222D] px-4 py-1 text-[#F5222D] hover:bg-[#F5222D] hover:text-white"
						onClick={handleCancel}
					>
						Cancel
					</button>
					<button
						onClick={handleSaveJob}
						className="flex items-center justify-center gap-2 rounded-[6px] border border-[#D9D9D9] px-4 py-1 font-light hover:bg-[#EBEBEB]"
					>
						<DownloadSimple className="size-4" />
						Save Job
					</button>
				</div>
				<div className="flex items-center">
					{currentStep !== "source" && (
						<button
							onClick={handleBack}
							className="mr-4 rounded-[6px] border border-[#D9D9D9] px-4 py-1 font-light hover:bg-[#EBEBEB]"
						>
							Back
						</button>
					)}
					<button
						className="flex items-center justify-center gap-2 rounded-[6px] bg-[#203FDD] px-4 py-1 font-light text-white hover:bg-[#132685]"
						onClick={handleNext}
					>
						{currentStep === "config" ? "Create Job" : "Next"}
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
									: currentStep === "config"
										? jobName
										: ""
						}
					/>
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
