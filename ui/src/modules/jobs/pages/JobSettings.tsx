import { useState, useEffect } from "react"
import { useParams, useNavigate, Link } from "react-router-dom"
import { Input, Button, Radio, Switch, Dropdown, message, Divider } from "antd"
import { ArrowRight } from "@phosphor-icons/react"
import { useAppStore } from "../../../store"
import { ArrowLeft, CaretDown } from "@phosphor-icons/react"
import DocumentationPanel from "../../common/components/DocumentationPanel"

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
		<>
			<div className="flex h-screen flex-col">
				<div className="overflow-scroll">
					<div className="px-6 pt-6">
						<div className="flex items-center justify-between">
							<div>
								<div className="flex items-center gap-2">
									<Link
										to="/jobs"
										className="items-cente mt-[2px] flex"
									>
										<ArrowLeft size={20} />
									</Link>

									<div className="text-2xl font-bold">
										{job?.name || "<Job_name>"}
									</div>
								</div>
								<span className="ml-6 mt-2 rounded bg-blue-100 px-2 py-1 text-xs text-[#0958D9]">
									{job?.status || "Active"}
								</span>
							</div>

							<div className="flex items-center gap-2">
								<Button className="rounded-full bg-green-500 text-white">
									S
								</Button>
								<span className="text-gray-500">--------------</span>
								<Button className="rounded-full bg-red-500 text-white">
									D
								</Button>
							</div>
						</div>
					</div>

					<Divider className="h-[2px]" />

					<div className="flex px-6">
						{/* Main content */}
						<div
							className={`${
								docsMinimized ? "w-full" : "w-3/4"
							} pr-6 transition-all duration-300`}
						>
							<h2 className="mb-4 text-xl font-medium">Job settings</h2>

							<div className="mb-6">
								<div className="flex w-full flex-row justify-between rounded-xl border border-[#D9D9D9] bg-white p-6">
									<div className="mb-6 w-1/2">
										<label className="mb-2 block text-sm text-gray-700">
											Job name:
										</label>
										<Input
											placeholder="Enter your job name"
											defaultValue={job?.name}
											className="max-w-md"
										/>
									</div>

									<div className="mb-6 w-1/2">
										<label className="mb-2 block text-sm text-gray-700">
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
								</div>

								<div className="mb-6">
									<div className="mb-2 mt-4 text-sm text-[#000]">
										When the source schema changes, I want to:
									</div>
									<div className="rounded-xl border border-[#D9D9D9] px-6 pt-6">
										<Radio.Group defaultValue="propagate">
											<div className="mb-2">
												<Radio value="propagate">
													<div>
														<span className="font-medium">
															Propagate field changes only
														</span>
													</div>
												</Radio>
												<p className="mb-4 mt-1 pl-5 text-sm text-[#575757]">
													Only column changes will be propagated. Incompatible
													schema changes will be detected, but not propagated.
												</p>
											</div>
										</Radio.Group>
									</div>
								</div>

								<div className="flex items-center justify-between rounded-xl border border-[#D9D9D9] px-6 py-6">
									<span className="font-medium">
										Be notified when schema changes occur
									</span>
									<Switch
										checked={notifyOnSchemaChanges}
										onChange={setNotifyOnSchemaChanges}
										className={notifyOnSchemaChanges ? "bg-[#203FDD]" : ""}
									/>
								</div>

								<div className="mt-6 flex items-center justify-between rounded-xl border border-[#D9D9D9] px-6 py-6">
									<span className="font-medium">Pause your job</span>
									<Switch
										checked={pauseJob}
										onChange={setPauseJob}
										className={pauseJob ? "bg-blue-600" : ""}
									/>
								</div>
							</div>

							<div className="mb-6 rounded-xl border border-gray-200 bg-white p-6">
								<div className="mb-3">
									<div className="mb-2 flex items-center justify-between">
										<span className="font-medium">Clear your data:</span>
										<Button
											onClick={handleClearData}
											className="py-4"
										>
											Clear data
										</Button>
									</div>
									<p className="text-sm text-[#8A8A8A]">
										Clearing data will delete all the data in your destination
									</p>
								</div>

								<div className="mb-3 border-gray-200 pt-4">
									<div className="mb-2 flex items-center justify-between">
										<span className="font-medium">
											Clear destination and sync:
										</span>
										<Button
											onClick={handleClearDestinationAndSync}
											className="py-4"
										>
											Clear destination and sync
										</Button>
									</div>
									<p className="text-sm text-[#8A8A8A]">
										It will delete all the data in the destination and then sync
										the data from the source
									</p>
								</div>

								<div className="border-gray-200 pt-4">
									<div className="mb-2 flex items-center justify-between">
										<span className="font-medium">Delete the job:</span>
										<button
											onClick={handleDeleteJob}
											className="rounded-[6px] border bg-[#F5222D] px-4 py-1 font-light text-white hover:bg-[#b81922]"
										>
											Delete this job
										</button>
									</div>
									<p className="text-sm text-[#8A8A8A]">
										No data will be deleted in your source and destination.
									</p>
								</div>
							</div>
						</div>

						{/* Documentation panel with iframe */}
						<DocumentationPanel
							docUrl="https://olake.io/docs/category/mongodb"
							isMinimized={docsMinimized}
							onToggle={toggleDocsPanel}
							showResizer={true}
						/>
					</div>
				</div>

				<div className="flex justify-end border-t border-gray-200 p-4">
					<Button
						type="primary"
						className="bg-blue-600"
						onClick={handleSaveSettings}
					>
						<ArrowRight size={16} />
						Save
					</Button>
				</div>
			</div>
		</>
	)
}

export default JobSettings
