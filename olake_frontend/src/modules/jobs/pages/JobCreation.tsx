import { useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { message } from "antd"
import CreateSource from "../../sources/pages/CreateSource"
import CreateDestination from "../../destinations/pages/CreateDestination"
import { ArrowLeft, ArrowRight, DownloadSimple } from "@phosphor-icons/react"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import StepProgress from "../components/StepIndicator"
import { useAppStore } from "../../../store"
import EntitySavedModal from "../../common/components/EntitySavedModal"
import SchemaConfiguration from "../components/SchemaConfiguration"
import JobConfiguration from "../components/JobConfiguration"
import EntityCancelModal from "../../common/components/EntityCancelModal"
import TestConnectionSuccessModal from "../../common/components/TestConnectionSuccessModal"
import TestConnectionModal from "../../common/components/TestConnectionModal"

export type JobCreationSteps = "source" | "destination" | "schema" | "config"

const JobCreation: React.FC = () => {
	const navigate = useNavigate()
	const [currentStep, setCurrentStep] = useState<JobCreationSteps>("source")
	const [docsMinimized, setDocsMinimized] = useState(false)

	// Schema step states
	const [selectedStreams, setSelectedStreams] = useState<string[]>([
		"Payments",
		"public_raw_stream",
	])
	const [syncMode, setSyncMode] = useState("full")
	const [enableBackfill, setEnableBackfill] = useState(true)
	const [normalisation, setNormalisation] = useState(true)

	// Config step states
	const [jobName, setJobName] = useState("")
	const [replicationFrequency, setReplicationFrequency] = useState("daily")
	const [schemaChangeStrategy, setSchemaChangeStrategy] = useState("propagate")
	const [notifyOnSchemaChanges, setNotifyOnSchemaChanges] = useState(true)

	const {
		setShowEntitySavedModal,
		setShowSourceCancelModal,
		setShowTestingModal,
		setShowSuccessModal,
	} = useAppStore()

	const handleNext = () => {
		if (currentStep === "source") {
			dummyNetworkCall()
		} else if (currentStep === "destination") {
			dummyNetworkCall()
		} else if (currentStep === "schema") {
			setTimeout(() => {
				setCurrentStep("config")
			}, 1500)
		} else if (currentStep === "config") {
			setTimeout(() => {
				setShowEntitySavedModal(true)
			}, 1500)
		}
	}
	const nextStep = () => {
		if (currentStep === "source") {
			setCurrentStep("destination")
		} else if (currentStep === "destination") {
			setCurrentStep("schema")
		} else if (currentStep === "schema") {
			setCurrentStep("config")
		} else {
		}
	}

	const dummyNetworkCall = () => {
		setTimeout(() => {
			setShowTestingModal(true)
			setTimeout(() => {
				setShowTestingModal(false)
				setShowSuccessModal(true)
				setTimeout(() => {
					setShowSuccessModal(false)
					setShowEntitySavedModal(true)
				}, 2000)
			}, 2000)
		}, 1500)
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
					<Link
						to="/jobs"
						className="flex items-center gap-2"
					>
						<ArrowLeft className="mr-1 size-6" />
						<span className="text-2xl font-bold"> Create job</span>
					</Link>

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
							syncMode={syncMode}
							setSyncMode={setSyncMode}
							enableBackfill={enableBackfill}
							setEnableBackfill={setEnableBackfill}
							normalisation={normalisation}
							setNormalisation={setNormalisation}
							stepNumber={3}
							stepTitle="Schema evaluation"
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
						fromJobFlow={true}
						onComplete={nextStep}
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
