import { useState, useEffect } from "react"
import { useNavigate, Link, useParams } from "react-router-dom"
import { Button, message } from "antd"
import CreateSource from "../../sources/pages/CreateSource"
import CreateDestination from "../../destinations/pages/CreateDestination"
import { ArrowLeft, ArrowRight } from "@phosphor-icons/react"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import StepProgress from "../components/StepIndicator"
import SchemaConfiguration from "../components/SchemaConfiguration"
import JobConfiguration from "../components/JobConfiguration"
import { useAppStore } from "../../../store"
import EntityCancelModal from "../../common/components/EntityCancelModal"

type Step = "source" | "destination" | "schema" | "config"

const JobEdit: React.FC = () => {
	const navigate = useNavigate()
	const { jobId } = useParams<{ jobId: string }>()
	const {
		jobs,
		sources,
		destinations,
		fetchJobs,
		fetchSources,
		fetchDestinations,
	} = useAppStore()

	const [currentStep, setCurrentStep] = useState<Step>("source")
	const [docsMinimized, setDocsMinimized] = useState(false)

	// Schema step states
	const [selectedStreams, setSelectedStreams] = useState<string[]>([])
	const [syncMode, setSyncMode] = useState("full")
	const [enableBackfill, setEnableBackfill] = useState(true)
	const [normalisation, setNormalisation] = useState(true)

	// Config step states
	const [jobName, setJobName] = useState("")
	const [replicationFrequency, setReplicationFrequency] = useState("daily")
	const [schemaChangeStrategy, setSchemaChangeStrategy] = useState("propagate")
	const [notifyOnSchemaChanges, setNotifyOnSchemaChanges] = useState(true)

	// Find the job from the store
	const job = jobs.find(j => j.id === jobId)
	const sourceObj = sources.find(s => s.name === job?.source)
	const destinationObj = destinations.find(d => d.name === job?.destination)

	// Get the value cancel modal from the store
	const { setShowSourceCancelModal } = useAppStore()

	// Load job data on component mount
	useEffect(() => {
		// Make sure we have the jobs, sources, and destinations data
		fetchJobs()
		fetchSources()
		fetchDestinations()
	}, [fetchJobs, fetchSources, fetchDestinations])

	// Set initial form values when job data is available
	useEffect(() => {
		if (job) {
			setJobName(job.name)

			// For demo purposes, we're using mock data for the additional fields
			// In a real application, you would fetch these details from the API
			setSelectedStreams(["Payments", "public_raw_stream"])
			setSyncMode("full")
			setEnableBackfill(true)
			setNormalisation(true)
			setReplicationFrequency("daily")
			setSchemaChangeStrategy("propagate")
			setNotifyOnSchemaChanges(true)
		}
	}, [job])

	const handleNext = () => {
		if (currentStep === "source") {
			setCurrentStep("destination")
		} else if (currentStep === "destination") {
			setCurrentStep("schema")
		} else if (currentStep === "schema") {
			setCurrentStep("config")
		} else if (currentStep === "config") {
			message.success("Job updated successfully!")
			navigate("/jobs")
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
			message.info("Job edit cancelled")
			navigate("/jobs")
		}
	}

	const handleSaveJob = () => {
		message.success("Job updated successfully!")
		navigate("/jobs")
	}

	const toggleDocsPanel = () => {
		setDocsMinimized(!docsMinimized)
	}

	// Show loading while job data is loading
	if (!job) {
		return (
			<div className="flex h-screen items-center justify-center">
				<div className="text-lg">Loading job data...</div>
			</div>
		)
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
						<span className="text-2xl font-bold">Edit job: {jobName}</span>
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
							? "w-[calc(100%-30%)]"
							: "w-full"
					} flex flex-col overflow-auto transition-all duration-300 ease-in-out`}
				>
					<div className="flex-1 overflow-auto pt-0">
						{currentStep === "source" && (
							<div className="w-full">
								<CreateSource
									fromJobFlow={true}
									fromJobEditFlow={true}
									existingSourceId={sourceObj?.id}
									stepNumber={"I"}
									stepTitle="Source configuration"
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
									fromJobEditFlow={true}
									existingDestinationId={destinationObj?.id}
									stepNumber={2}
									stepTitle="Destination configuration"
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
								stepTitle="Job Configuration"
							/>
						)}
					</div>
				</div>

				{/* Right documentation panel */}
				{(currentStep === "schema" || currentStep === "config") && (
					<DocumentationPanel
						docUrl="https://olake.io/docs/category/mongodb"
						isMinimized={docsMinimized}
						onToggle={toggleDocsPanel}
						showResizer={true}
						initialWidth={30}
					/>
				)}
			</div>

			{/* Footer placed outside the flex container to ensure it's always at the bottom */}
			<div className="border-t border-gray-200 bg-white p-4">
				<div className="flex justify-between">
					{currentStep !== "source" ? (
						<Button
							onClick={handleBack}
							size="large"
							className="flex items-center gap-1 border-gray-300 text-gray-600"
						>
							<ArrowLeft className="size-4" />
							Back
						</Button>
					) : (
						<Button
							onClick={handleCancel}
							danger
							className="border-red-500 px-6 text-red-500 hover:border-red-600 hover:text-red-600"
						>
							Cancel
						</Button>
					)}

					{currentStep === "config" ? (
						<div className="flex gap-2">
							<Button
								onClick={handleSaveJob}
								size="large"
								className="border-blue-600 text-blue-600"
							>
								Save Job
							</Button>
							<Button
								onClick={handleNext}
								type="primary"
								size="large"
								className="flex items-center gap-1 bg-blue-600"
							>
								Next <ArrowRight className="size-4" />
							</Button>
						</div>
					) : (
						<Button
							onClick={handleNext}
							type="primary"
							size="large"
							className="flex items-center gap-1 bg-blue-600"
						>
							Next <ArrowRight className="size-4" />
						</Button>
					)}
				</div>
			</div>
			<EntityCancelModal
				type="job-edit"
				navigateTo="jobs"
			/>
		</div>
	)
}

export default JobEdit
