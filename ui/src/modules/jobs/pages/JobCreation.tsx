import { useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { message } from "antd"
import CreateSource from "../../sources/pages/CreateSource"
import CreateDestination from "../../destinations/pages/CreateDestination"
import { ArrowLeft, ArrowRight, DownloadSimple } from "@phosphor-icons/react"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import StepProgress from "../components/StepIndicator"
import { useAppStore } from "../../../store"
import EntitySavedModal from "../../common/Modals/EntitySavedModal"
import SchemaConfiguration from "./SchemaConfiguration"
import JobConfiguration from "../components/JobConfiguration"
import EntityCancelModal from "../../common/Modals/EntityCancelModal"
import { mockStreamData } from "../../../api/mockData"
import TestConnectionSuccessModal from "../../common/Modals/TestConnectionSuccessModal"
import TestConnectionModal from "../../common/Modals/TestConnectionModal"
import { JobCreationSteps } from "../../../types"

const JobCreation: React.FC = () => {
	const navigate = useNavigate()
	const [currentStep, setCurrentStep] = useState<JobCreationSteps>("source")
	const [docsMinimized, setDocsMinimized] = useState(false)

	// Source and destination states
	const [sourceName, setSourceName] = useState("")
	const [sourceConnector, setSourceConnector] = useState("")
	const [sourceFormData, setSourceFormData] = useState<any>({})
	const [destinationName, setDestinationName] = useState("")
	const [destinationConnector, setDestinationConnector] = useState("")
	const [destinationFormData, setDestinationFormData] = useState<any>({})

	// Schema step states
	const [selectedStreams, setSelectedStreams] = useState<string[]>(
		mockStreamData.map(stream => stream.stream.name),
	)

	// Config step states
	const [jobName, setJobName] = useState("")
	const [replicationFrequency, setReplicationFrequency] = useState("seconds")
	const [schemaChangeStrategy, setSchemaChangeStrategy] = useState("propagate")
	const [notifyOnSchemaChanges, setNotifyOnSchemaChanges] = useState(true)

	const {
		setShowEntitySavedModal,
		setShowSourceCancelModal,
		setShowTestingModal,
		setShowSuccessModal,
		addSource,
		addDestination,
		addJob,
	} = useAppStore()

	const handleNext = () => {
		if (currentStep === "source") {
			// Create source
			const newSourceData = {
				name: sourceName,
				type: sourceConnector,
				status: "active" as const,
				config: sourceFormData,
			}

			addSource(newSourceData)
				.then(() => {
					setShowTestingModal(true)
					setTimeout(() => {
						setShowTestingModal(false)
						setShowSuccessModal(true)
						setTimeout(() => {
							setShowSuccessModal(false)
							setShowEntitySavedModal(true)
							setTimeout(() => {
								setShowEntitySavedModal(false)
								setCurrentStep("destination")
							}, 2000)
						}, 2000)
					}, 2000)
				})
				.catch(error => {
					console.error("Error adding source:", error)
					message.error("Failed to create source")
				})
		} else if (currentStep === "destination") {
			// Create destination
			const newDestinationData = {
				name: destinationName,
				type: destinationConnector,
				status: "active" as const,
				config: destinationFormData,
			}

			addDestination(newDestinationData)
				.then(() => {
					setShowTestingModal(true)
					setTimeout(() => {
						setShowTestingModal(false)
						setShowSuccessModal(true)
						setTimeout(() => {
							setShowSuccessModal(false)
							setShowEntitySavedModal(true)
							setTimeout(() => {
								setShowEntitySavedModal(false)
								setCurrentStep("schema")
							}, 2000)
						}, 2000)
					}, 2000)
				})
				.catch(error => {
					console.error("Error adding destination:", error)
					message.error("Failed to create destination")
				})
		} else if (currentStep === "schema") {
			setCurrentStep("config")
		} else if (currentStep === "config") {
			// Create job
			const newJobData = {
				name: jobName,
				source: sourceName,
				destination: destinationName,
				streams: selectedStreams,
				replicationFrequency,
				schemaChangeStrategy,
				notifyOnSchemaChanges,
				status: "active" as const,
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

			{/* Main content */}
			<div className="flex flex-1 overflow-hidden border-t border-gray-200">
				{/* Left content */}
				<div
					className={`${
						(currentStep === "schema" || currentStep === "config") &&
						!docsMinimized
							? "w-2/3"
							: "w-full"
					} overflow-auto pt-0 transition-all duration-300`}
				>
					{currentStep === "source" && (
						<div className="w-full">
							<CreateSource
								fromJobFlow={true}
								stepNumber={"I"}
								stepTitle="Set up your source"
								onSourceNameChange={setSourceName}
								onConnectorChange={setSourceConnector}
								onFormDataChange={setSourceFormData}
								onComplete={() => {
									setCurrentStep("destination")
								}}
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
								onFormDataChange={setDestinationFormData}
								onComplete={() => {
									setCurrentStep("schema")
								}}
							/>
						</div>
					)}

					{currentStep === "schema" && (
						<SchemaConfiguration
							selectedStreams={selectedStreams}
							setSelectedStreams={setSelectedStreams}
							stepNumber={3}
							stepTitle="Streams selection"
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
				{(currentStep === "schema" || currentStep === "config") && (
					<DocumentationPanel
						docUrl="https://olake.io/docs/category/mongodb"
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
