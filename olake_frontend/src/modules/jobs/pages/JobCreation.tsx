import { useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { Input, Button, Radio, Select, Switch, message } from "antd"
import CreateSource from "../../sources/pages/CreateSource"
import CreateDestination from "../../destinations/pages/CreateDestination"
import { ArrowLeft, DownloadSimple } from "@phosphor-icons/react"
import DocumentationPanel from "../../common/components/DocumentationPanel"
import StepProgress from "../components/StepIndicator"

type Step = "source" | "destination" | "schema" | "config"

const JobCreation: React.FC = () => {
	const navigate = useNavigate()
	const [currentStep, setCurrentStep] = useState<Step>("source")
	const [docsMinimized, setDocsMinimized] = useState(false)
	const [searchText, setSearchText] = useState("")
	const [searchSchemaText, setSearchSchemaText] = useState("")
	const [syncAll, setSyncAll] = useState(false)
	const [selectAllColumns, setSelectAllColumns] = useState(true)

	// Schema step states
	const [selectedStreams, setSelectedStreams] = useState<string[]>([
		"Payments",
		"public_raw_stream",
	])
	const [activeTab, setActiveTab] = useState("config")
	const [syncMode, setSyncMode] = useState("full")
	const [enableBackfill, setEnableBackfill] = useState(true)
	const [normalisation, setNormalisation] = useState(true)
	const [partitionType, setPartitionType] = useState("set")
	const [granularity, setGranularity] = useState("Day")
	const [defaultValue, setDefaultValue] = useState("")

	// Config step states
	const [jobName, setJobName] = useState("")
	const [replicationFrequency, setReplicationFrequency] = useState("daily")
	const [schemaChangeStrategy, setSchemaChangeStrategy] = useState("propagate")
	const [notifyOnSchemaChanges, setNotifyOnSchemaChanges] = useState(true)

	const handleNext = () => {
		if (currentStep === "source") {
			setCurrentStep("destination")
		} else if (currentStep === "destination") {
			setCurrentStep("schema")
		} else if (currentStep === "schema") {
			setCurrentStep("config")
		} else if (currentStep === "config") {
			message.success("Job created successfully!")
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
		message.info("Job creation cancelled")
		navigate("/jobs")
	}

	const handleSaveJob = () => {
		message.success("Job saved successfully!")
		navigate("/jobs")
	}

	const handleAddPartition = () => {
		message.success("Partition added")
	}

	const handleSyncAllChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		setSyncAll(e.target.checked)
		if (e.target.checked) {
			setSelectedStreams([
				"Payments",
				"airbyte_destination_state_airbyte_destination_state_airbyte_destination_state_airbyte",
				"public_raw_stream",
				"job",
				"job_run_details",
				"cdc_test",
				"dz-stag-azure",
				"dz-stag-clients",
			])
		} else {
			setSelectedStreams([])
		}
	}

	const handleSelectAllColumns = (e: React.ChangeEvent<HTMLInputElement>) => {
		setSelectAllColumns(e.target.checked)
	}

	const streams = [
		"Payments",
		"airbyte_destination_state_airbyte_destination_state_airbyte_destination_state_airbyte",
		"public_raw_stream",
		"job",
		"job_run_details",
		"cdc_test",
		"dz-stag-azure",
		"dz-stag-clients",
	]

	const filteredStreams = streams.filter(stream =>
		stream.toLowerCase().includes(searchText.toLowerCase()),
	)

	const columns = [
		"Column_1",
		"Column_2",
		"Column_3",
		"Column_4",
		"Column_5",
		"Column_6",
		"Column_7",
		"Column_8",
		"Column_9",
	]

	const filteredColumns = columns.filter(column =>
		column.toLowerCase().includes(searchSchemaText.toLowerCase()),
	)

	const toggleDocsPanel = () => {
		setDocsMinimized(!docsMinimized)
	}

	return (
		<div className="flex h-screen flex-col">
			{/* Header */}
			<div className="bg-white px-6 pb-2 pt-6">
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
					} overflow-auto p-6 pt-0 transition-all duration-300`}
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
						<div className="mb-4">
							<div className="mb-4 flex items-center">
								<Input
									placeholder="Search streams"
									className="mr-4 w-64"
									value={searchText}
									onChange={e => setSearchText(e.target.value)}
								/>
								<div className="flex space-x-2">
									<Button className="border-gray-300">All tables</Button>
									<Button className="border-gray-300">CDC</Button>
									<Button className="border-gray-300">Full refresh</Button>
									<Button className="border-gray-300">Selected</Button>
									<Button className="border-gray-300">Not selected</Button>
								</div>
							</div>

							<div className="flex">
								<div className="w-1/2 overflow-hidden rounded-lg border border-gray-200">
									<div className="border-b border-gray-200 p-3">
										<label className="flex items-center">
											<input
												type="checkbox"
												className="mr-2"
												checked={syncAll}
												onChange={handleSyncAllChange}
											/>
											<span>Sync all</span>
										</label>
									</div>

									{filteredStreams.map((stream, index) => (
										<div
											key={index}
											className={`border-b border-gray-200 p-3 ${
												selectedStreams.includes(stream) ? "bg-blue-50" : ""
											}`}
										>
											<label className="flex items-center">
												<input
													type="checkbox"
													className="mr-2"
													checked={selectedStreams.includes(stream)}
													onChange={() => {
														if (selectedStreams.includes(stream)) {
															setSelectedStreams(
																selectedStreams.filter(s => s !== stream),
															)
														} else {
															setSelectedStreams([...selectedStreams, stream])
														}
													}}
												/>
												<span>{stream}</span>
											</label>
										</div>
									))}
								</div>

								<div className="ml-4 w-1/2 rounded-lg border border-gray-200 p-4">
									<h3 className="mb-4 text-lg font-medium">Payments</h3>

									<div className="mb-4 flex">
										<button
											className={`px-4 py-2 text-sm font-medium ${
												activeTab === "config"
													? "border-b-2 border-blue-600 text-blue-600"
													: "text-gray-500 hover:text-gray-700"
											}`}
											onClick={() => setActiveTab("config")}
										>
											Config
										</button>
										<button
											className={`px-4 py-2 text-sm font-medium ${
												activeTab === "schema"
													? "border-b-2 border-blue-600 text-blue-600"
													: "text-gray-500 hover:text-gray-700"
											}`}
											onClick={() => setActiveTab("schema")}
										>
											Schema
										</button>
										<button
											className={`px-4 py-2 text-sm font-medium ${
												activeTab === "partitioning"
													? "border-b-2 border-blue-600 text-blue-600"
													: "text-gray-500 hover:text-gray-700"
											}`}
											onClick={() => setActiveTab("partitioning")}
										>
											Partitioning
										</button>
									</div>

									{activeTab === "config" && (
										<div>
											<div className="mb-4">
												<p className="mb-2">Sync mode:</p>
												<Radio.Group
													value={syncMode}
													onChange={e => setSyncMode(e.target.value)}
												>
													<Radio value="full">Full refresh</Radio>
													<Radio value="cdc">CDC</Radio>
												</Radio.Group>
											</div>

											<div className="flex items-center justify-between border-t border-gray-200 py-3">
												<span>Enable backfill</span>
												<Switch
													checked={enableBackfill}
													onChange={setEnableBackfill}
													className={enableBackfill ? "bg-blue-600" : ""}
												/>
											</div>

											<div className="flex items-center justify-between border-t border-gray-200 py-3">
												<span>Normalisation</span>
												<Switch
													checked={normalisation}
													onChange={setNormalisation}
													className={normalisation ? "bg-blue-600" : ""}
												/>
											</div>
										</div>
									)}

									{activeTab === "schema" && (
										<div>
											<div className="mb-4">
												<Input
													placeholder="Search streams"
													className="mb-4 w-full"
													value={searchSchemaText}
													onChange={e => setSearchSchemaText(e.target.value)}
												/>

												<div className="mb-2">
													<label className="flex items-center">
														<input
															type="checkbox"
															className="mr-2"
															checked={selectAllColumns}
															onChange={handleSelectAllColumns}
														/>
														<span>Select all</span>
													</label>
												</div>

												{filteredColumns.map((column, index) => (
													<div
														key={index}
														className="mb-2 flex items-center justify-between"
													>
														<label className="flex items-center">
															<input
																type="checkbox"
																className="mr-2"
																checked={selectAllColumns}
															/>
															<span>{column}</span>
														</label>
														<span
															className={`rounded px-2 py-1 text-xs ${
																index === 0
																	? "bg-blue-100 text-blue-600"
																	: index === 1
																		? "bg-green-100 text-green-600"
																		: index === 2
																			? "bg-purple-100 text-purple-600"
																			: index === 3
																				? "bg-yellow-100 text-yellow-600"
																				: index === 4
																					? "bg-red-100 text-red-600"
																					: index === 5
																						? "bg-indigo-100 text-indigo-600"
																						: index === 6
																							? "bg-pink-100 text-pink-600"
																							: index === 7
																								? "bg-teal-100 text-teal-600"
																								: "bg-gray-100 text-gray-600"
															}`}
														>
															{index === 0
																? "STRING"
																: index === 1
																	? "INTEGER"
																	: index === 2
																		? "ARRAY"
																		: index === 3
																			? "BOOL"
																			: index === 4
																				? "OBJECT"
																				: index === 5
																					? "FLOAT"
																					: index === 6
																						? "DOUBLE"
																						: index === 7
																							? "TIMESTAMP"
																							: "NULL"}
														</span>
													</div>
												))}
											</div>
										</div>
									)}

									{activeTab === "partitioning" && (
										<div>
											<div className="mb-4">
												<Radio.Group
													value={partitionType}
													onChange={e => setPartitionType(e.target.value)}
													className="mb-4 flex"
												>
													<Radio value="set">Set partition</Radio>
													<Radio value="regex">Partitioning regex</Radio>
												</Radio.Group>

												<div className="mb-4">
													<label className="mb-2 block text-sm font-medium text-gray-700">
														Select column:
													</label>
													<div className="mb-2 flex">
														<Select
															placeholder="Select columns"
															className="mr-2 w-full"
															options={[
																{
																	value: "Column_name_1",
																	label: "Column_name_1",
																},
																{
																	value: "Column_name_2",
																	label: "Column_name_2",
																},
															]}
														/>
														<div className="flex">
															<Button
																className={`${
																	granularity === "Day"
																		? "bg-blue-600 text-white"
																		: "bg-white"
																}`}
																onClick={() => setGranularity("Day")}
															>
																Day
															</Button>
															<Button
																className={`${
																	granularity === "Month"
																		? "bg-blue-600 text-white"
																		: "bg-white"
																}`}
																onClick={() => setGranularity("Month")}
															>
																Month
															</Button>
															<Button
																className={`${
																	granularity === "Year"
																		? "bg-blue-600 text-white"
																		: "bg-white"
																}`}
																onClick={() => setGranularity("Year")}
															>
																Year
															</Button>
														</div>
													</div>
												</div>

												<div className="mb-4">
													<label className="mb-2 block text-sm font-medium text-gray-700">
														Default value:
													</label>
													<div className="flex">
														<Input
															placeholder="Enter default value for your column"
															value={defaultValue}
															onChange={e => setDefaultValue(e.target.value)}
															className="mr-2"
														/>
														<Button
															type="primary"
															className="bg-blue-600"
															onClick={handleAddPartition}
														>
															Add
														</Button>
													</div>
												</div>

												<table className="mb-4 min-w-full">
													<thead>
														<tr className="border-b border-gray-200">
															<th className="px-4 py-2 text-left font-medium text-gray-700">
																Column name
															</th>
															<th className="px-4 py-2 text-left font-medium text-gray-700">
																Granularity
															</th>
															<th className="px-4 py-2 text-left font-medium text-gray-700">
																Default
															</th>
														</tr>
													</thead>
													<tbody>
														<tr className="border-b border-gray-100">
															<td className="px-4 py-2">Column_name_1</td>
															<td className="px-4 py-2">Day</td>
															<td className="px-4 py-2">Table cell text</td>
														</tr>
														<tr className="border-b border-gray-100">
															<td className="px-4 py-2">Column_name_2</td>
															<td className="px-4 py-2">DD.MM</td>
															<td className="px-4 py-2">Table cell text</td>
														</tr>
														<tr className="border-b border-gray-100">
															<td className="px-4 py-2">Column_name_3</td>
															<td className="px-4 py-2">Nil</td>
															<td className="px-4 py-2">Table cell text</td>
														</tr>
													</tbody>
												</table>

												<div>
													<label className="mb-2 block text-sm font-medium text-gray-700">
														Regex preview:
													</label>
													<div className="break-all rounded bg-gray-100 p-2 text-sm text-gray-600">
														.../om/data/Column_name_1/Column_name_subs1/Column_name_paper
													</div>
												</div>
											</div>
										</div>
									)}
								</div>
							</div>
						</div>
					)}

					{currentStep === "config" && (
						<>
							<div>
								<div className="mt-4 flex items-center gap-2 text-sm text-[#203FDD]">
									<div className="size-2 rounded-full border border-[#203FDD] outline outline-2 outline-[#203FDD]"></div>
									<span className="text-sm text-[#8A8A8A]"> Step 4</span>
								</div>
							</div>
							<div className="mt-2 text-lg font-medium">Job configuration</div>
							<div className="mb-6 p-6">
								<div className="rounded-xl border border-[#D9D9D9] p-4">
									<div className="mb-2 grid grid-cols-2 gap-6">
										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Job name:
											</label>
											<Input
												placeholder="Enter your job name"
												value={jobName}
												onChange={e => setJobName(e.target.value)}
											/>
										</div>

										<div>
											<label className="mb-2 block text-sm font-medium text-gray-700">
												Replication frequency:
											</label>
											<Select
												placeholder="Data sync will repeat in?"
												className="w-full"
												options={[
													{ value: "hourly", label: "Hourly" },
													{ value: "daily", label: "Daily" },
													{ value: "weekly", label: "Weekly" },
													{ value: "monthly", label: "Monthly" },
												]}
												value={replicationFrequency}
												onChange={setReplicationFrequency}
											/>
										</div>
									</div>
								</div>

								<div className="mb-6">
									<label className="mb-2 block text-sm font-medium text-gray-700">
										When the source schema changes, I want to:
									</label>
									<div className="rounded-lg border border-gray-200 bg-gray-50 p-4">
										<Radio.Group
											value={schemaChangeStrategy}
											onChange={e => setSchemaChangeStrategy(e.target.value)}
										>
											<div className="mb-2">
												<Radio value="propagate">
													<div>
														<span className="font-medium">
															Propagate field changes only
														</span>
														<p className="mt-1 text-sm text-gray-500">
															Only column changes will be propagated.
															Incompatible schema changes will be detected, but
															not propagated.
														</p>
													</div>
												</Radio>
											</div>
											<div>
												<Radio value="ignore">
													<div>
														<span className="font-medium">
															Ignore schema changes
														</span>
														<p className="mt-1 text-sm text-gray-500">
															Schema changes will be ignored. Data will continue
															to sync with the existing schema.
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
							</div>
						</>
					)}
				</div>

				{/* Documentation panel */}
				{(currentStep === "schema" || currentStep === "config") && (
					<DocumentationPanel
						title="MongoDB"
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
					<Button
						danger
						onClick={handleCancel}
					>
						Cancel
					</Button>
					<Button
						onClick={handleSaveJob}
						className="flex items-center justify-center"
					>
						<DownloadSimple className="size-4" />
						Save Job
					</Button>
				</div>
				<div>
					{currentStep !== "source" && (
						<Button
							onClick={handleBack}
							className="mr-4"
						>
							Back
						</Button>
					)}
					<Button
						type="primary"
						onClick={handleNext}
					>
						{currentStep === "config" ? "Create Job →" : "Next →"}
					</Button>
				</div>
			</div>
		</div>
	)
}

export default JobCreation
