import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Radio, Switch, Dropdown, message } from "antd"
import { CornersIn, CornersOut } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { ArrowLeft, CaretDown } from "@phosphor-icons/react"

const JobSettings: React.FC = () => {
	const { jobId } = useParams<{ jobId: string }>()
	const navigate = useNavigate()
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [replicationFrequency, setReplicationFrequency] = useState("daily")
	const [notifyOnSchemaChanges, setNotifyOnSchemaChanges] = useState(true)
	const [pauseJob, setPauseJob] = useState(false)

	const { jobs, fetchJobs } = useAppStore()

	useEffect(() => {
		if (!jobs.length) {
			fetchJobs()
		}
	}, [fetchJobs, jobs.length])

	const job = jobs.find(j => j.id === jobId)

	const handleClearData = () => {
		message.success("Data cleared successfully")
	}

	const handleClearDestinationAndSync = () => {
		message.success("Destination cleared and sync initiated")
	}

	const handleDeleteJob = () => {
		message.success("Job deleted successfully")
		navigate("/jobs")
	}

	const handleSaveSettings = () => {
		message.success("Job settings saved successfully")
	}

	const toggleDocsPanel = () => {
		setDocsMinimized(!docsMinimized)
	}

	const frequencyOptions = [
		{ label: "Hourly", value: "hourly" },
		{ label: "Daily", value: "daily" },
		{ label: "Weekly", value: "weekly" },
		{ label: "Monthly", value: "monthly" },
	]

	return (
		<div className="flex p-6">
			{/* Main content */}
			<div
				className={`${
					docsMinimized ? "w-full" : "w-3/4"
				} pr-6 transition-all duration-300`}
			>
				<div className="mb-6">
					<Link
						to="/jobs"
						className="mb-4 flex items-center text-blue-600"
					>
						<ArrowLeft
							size={16}
							className="mr-1"
						/>{" "}
						Back to Jobs
					</Link>

					<div className="mb-2 flex items-center">
						<h1 className="text-2xl font-bold">
							{job?.name || "Job Settings"}
						</h1>
						<span className="ml-2 rounded bg-blue-100 px-2 py-1 text-xs text-blue-600">
							{job?.status || "Active"}
						</span>
					</div>

					<div className="mb-6 flex items-center justify-between">
						<div className="flex items-center">
							<div className="mr-8 flex items-center">
								<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-green-500 text-white">
									S
								</div>
								<span>Source</span>
							</div>
							<div className="w-16 border-t-2 border-dashed border-gray-300"></div>
							<div className="ml-8 flex items-center">
								<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-red-500 text-white">
									D
								</div>
								<span>Destination</span>
							</div>
						</div>
					</div>
				</div>

				<h2 className="mb-4 text-xl font-bold">Job settings</h2>

				<div className="mb-6 rounded-lg border border-gray-200 bg-white p-6">
					<div className="mb-6">
						<label className="mb-2 block text-sm font-medium text-gray-700">
							Job name:
						</label>
						<Input
							placeholder="Enter your job name"
							defaultValue={job?.name}
							className="max-w-md"
						/>
					</div>

					<div className="mb-6">
						<label className="mb-2 block text-sm font-medium text-gray-700">
							Replication frequency:
						</label>
						<Dropdown
							menu={{
								items: frequencyOptions.map(option => ({
									key: option.value,
									label: option.label,
									onClick: () => setReplicationFrequency(option.value),
								})),
							}}
						>
							<Button className="flex w-64 items-center justify-between">
								<span>
									{frequencyOptions.find(
										option => option.value === replicationFrequency,
									)?.label || "Select frequency"}
								</span>
								<CaretDown size={16} />
							</Button>
						</Dropdown>
					</div>

					<div className="mb-6">
						<label className="mb-2 block text-sm font-medium text-gray-700">
							When the source schema changes, I want to:
						</label>
						<div className="rounded-lg border border-gray-200 bg-gray-50 p-4">
							<Radio.Group defaultValue="propagate">
								<div className="mb-2">
									<Radio value="propagate">
										<div>
											<span className="font-medium">
												Propagate field changes only
											</span>
											<p className="mt-1 text-sm text-gray-500">
												Only column changes will be propagated. Incompatible
												schema changes will be detected, but not propagated.
											</p>
										</div>
									</Radio>
								</div>
								<div>
									<Radio value="ignore">
										<div>
											<span className="font-medium">Ignore schema changes</span>
											<p className="mt-1 text-sm text-gray-500">
												Schema changes will be ignored. Data will continue to
												sync with the existing schema.
											</p>
										</div>
									</Radio>
								</div>
							</Radio.Group>
						</div>
					</div>

					<div className="flex items-center justify-between border-t border-gray-200 py-3">
						<span className="font-medium">
							Be notified when schema changes occur
						</span>
						<Switch
							checked={notifyOnSchemaChanges}
							onChange={setNotifyOnSchemaChanges}
							className={notifyOnSchemaChanges ? "bg-blue-600" : ""}
						/>
					</div>

					<div className="flex items-center justify-between border-t border-gray-200 py-3">
						<span className="font-medium">Pause your job</span>
						<Switch
							checked={pauseJob}
							onChange={setPauseJob}
							className={pauseJob ? "bg-blue-600" : ""}
						/>
					</div>
				</div>

				<div className="mb-6 rounded-lg border border-gray-200 bg-white p-6">
					<div className="mb-6">
						<div className="mb-2 flex items-center justify-between">
							<span className="font-medium">Clear your data:</span>
							<Button onClick={handleClearData}>Clear data</Button>
						</div>
						<p className="text-sm text-gray-500">
							Clearing data will delete all the data in your destination
						</p>
					</div>

					<div className="mb-6 border-t border-gray-200 pt-4">
						<div className="mb-2 flex items-center justify-between">
							<span className="font-medium">Clear destination and sync:</span>
							<Button onClick={handleClearDestinationAndSync}>
								Clear destination and sync
							</Button>
						</div>
						<p className="text-sm text-gray-500">
							It will delete all the data in the destination and then sync the
							data from the source
						</p>
					</div>

					<div className="border-t border-gray-200 pt-4">
						<div className="mb-2 flex items-center justify-between">
							<span className="font-medium">Delete the job:</span>
							<Button
								danger
								onClick={handleDeleteJob}
							>
								Delete this job
							</Button>
						</div>
						<p className="text-sm text-gray-500">
							No data will be deleted in your source and destination.
						</p>
					</div>
				</div>

				<div className="flex justify-end">
					<Button
						type="primary"
						className="bg-blue-600"
						onClick={handleSaveSettings}
					>
						Save settings
					</Button>
				</div>
			</div>

			{/* Documentation panel with iframe */}
			{!docsMinimized && (
				<div className="h-[calc(100vh-120px)] w-1/4 overflow-hidden rounded-lg border border-gray-200 bg-white">
					<div className="flex items-center justify-between border-b border-gray-200 p-4">
						<div className="flex items-center">
							<div className="mr-2 flex h-8 w-8 items-center justify-center rounded-full bg-blue-600 text-white">
								<span className="font-bold">M</span>
							</div>
							<span className="text-lg font-bold">MongoDB</span>
						</div>
						<Button
							type="text"
							icon={<CornersIn size={16} />}
							onClick={toggleDocsPanel}
							className="hover:bg-gray-100"
						/>
					</div>

					<div className="h-[calc(100%-60px)] w-full">
						<iframe
							src="https://olake.io/docs/category/mongodb"
							className="h-full w-full border-0"
							title="MongoDB Documentation"
							sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
						/>
					</div>
				</div>
			)}

			{/* Minimized docs panel button */}
			{docsMinimized && (
				<div className="fixed bottom-6 right-6">
					<Button
						type="primary"
						className="flex items-center bg-blue-600"
						onClick={toggleDocsPanel}
						icon={
							<CornersOut
								size={16}
								className="mr-2"
							/>
						}
					>
						Show Documentation
					</Button>
				</div>
			)}
		</div>
	)
}

export default JobSettings
