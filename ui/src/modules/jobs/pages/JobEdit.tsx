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
import { mockStreamData } from "../../../api/mockData"

type Step = "source" | "destination" | "schema" | "config"

interface SourceData {
	id: string
	name: string
	type: string
	config: Record<string, any>
}

interface DestinationData {
	id: string
	name: string
	type: string
	config: Record<string, any>
}

// Custom wrapper components for SourceEdit and DestinationEdit to use in job flow
const JobSourceEdit = ({ sourceData }: { sourceData: SourceData }) => {
	return (
		<div className="flex h-full flex-col">
			<div className="flex-1 overflow-auto">
				<SourceEdit
					fromJobFlow={true}
					stepNumber="1"
					stepTitle="Source config"
					initialData={sourceData}
				/>
			</div>
		</div>
	)
}

const JobDestinationEdit = ({
	destinationData,
}: {
	destinationData: DestinationData
}) => {
	return (
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
				/>
			</div>
		</div>
	)
}

const JobEdit: React.FC = () => {
	const navigate = useNavigate()
	const { jobId } = useParams<{ jobId: string }>()
	const { jobs, fetchJobs, fetchSources, fetchDestinations } = useAppStore()

	const [currentStep, setCurrentStep] = useState<Step>("source")
	const [docsMinimized, setDocsMinimized] = useState(true)

	// Source and destination data for job
	const [sourceData, setSourceData] = useState<SourceData | null>(null)
	const [destinationData, setDestinationData] =
		useState<DestinationData | null>(null)

	// Schema step states
	const [selectedStreams, setSelectedStreams] = useState<string[]>(
		mockStreamData.map(stream => stream.stream.name),
	)

	// Config step states
	const [jobName, setJobName] = useState("")
	const [replicationFrequency, setReplicationFrequency] = useState("seconds")
	const [schemaChangeStrategy, setSchemaChangeStrategy] = useState("propagate")
	const [notifyOnSchemaChanges, setNotifyOnSchemaChanges] = useState(true)

	// Find the job from the store

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}
	}, [fetchJobs, jobs])

	const job = jobs.find(j => j.id.toString() === jobId)
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
			const sourceDataObj: SourceData = {
				name: job.source.name,
				type: job.source.type,
				config: JSON.parse(job.source.config),
				// {
				// 	hosts: ["localhost:27017"],
				// 	username: "admin",
				// 	password: "password",
				// 	authdb: "admin",
				// 	database: "test_db",
				// 	collection: "test_collection",
				// },
			}
			setSourceData(sourceDataObj)

			// Load mock destination data
			const destinationDataObj: DestinationData = {
				id: "mock-destination-id",
				name: job.destination || "AWS S3 Destination",
				type: "Amazon S3",
				config: {
					normalization: false,
					s3_bucket: "my-test-bucket",
					s3_region: "ap-south-1",
					s3_access_key: "AKIAXXXXXXXXXXXXXXXX",
					s3_secret_key: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
					s3_path: "/data/test",
					type: "PARQUET",
				},
			}
			setDestinationData(destinationDataObj)
			setSelectedStreams(["Payments", "public_raw_stream"])
			setReplicationFrequency("daily")
			setSchemaChangeStrategy("propagate")
			setNotifyOnSchemaChanges(true)
		} else {
			setSourceData({
				id: "new-source",
				name: "New MongoDB Source",
				type: "MongoDB",
				config: {
					hosts: ["localhost:27017"],
					username: "admin",
					password: "password",
					authdb: "admin",
					database: "my_database",
					collection: "users",
				},
			})

			setDestinationData({
				id: "new-destination",
				name: "New S3 Destination",
				type: "Amazon S3",
				config: {
					normalization: false,
					s3_bucket: "my-new-bucket",
					s3_region: "us-east-1",
					s3_access_key: "AKIAXXXXXXXXXXXXXXXX",
					s3_secret_key: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
					s3_path: "/data/new",
					type: "PARQUET",
				},
			})
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
						<div className="text-2xl font-bold">{jobName}</div>
					</div>
					{/* Stepper */}
					<StepProgress currentStep={currentStep} />
				</div>
			</div>

			{/* Main content */}
			<div
				className={`flex flex-1 overflow-hidden border-gray-200 ${
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
					} relative flex flex-col`}
				>
					<div className="flex-1 pb-0">
						{currentStep === "source" && sourceData && (
							<div className="w-full">
								<JobSourceEdit sourceData={sourceData} />
							</div>
						)}

						{currentStep === "destination" && destinationData && (
							<div className="w-full">
								<JobDestinationEdit destinationData={destinationData} />
							</div>
						)}

						{currentStep === "schema" && (
							<div className="w-full">
								<SchemaConfiguration
									selectedStreams={selectedStreams}
									setSelectedStreams={setSelectedStreams}
									stepNumber={3}
									stepTitle="Schema evaluation"
								/>
							</div>
						)}

						{currentStep === "config" && (
							<div className="w-full">
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
							</div>
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
					>
						{currentStep === "config" ? "Finish" : "Next"}
						{currentStep !== "config" && (
							<ArrowRight className="size-4 text-white" />
						)}
					</button>
				</div>
			</div>
		</div>
	)
}

export default JobEdit
