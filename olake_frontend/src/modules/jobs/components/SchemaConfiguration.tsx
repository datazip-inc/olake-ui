import { useState } from "react"
import { Input, Button, Radio, Switch } from "antd"

interface SchemaConfigurationProps {
	selectedStreams: string[]
	setSelectedStreams: React.Dispatch<React.SetStateAction<string[]>>
	syncMode: string
	setSyncMode: React.Dispatch<React.SetStateAction<string>>
	enableBackfill: boolean
	setEnableBackfill: React.Dispatch<React.SetStateAction<boolean>>
	normalisation: boolean
	setNormalisation: React.Dispatch<React.SetStateAction<boolean>>
	stepNumber?: number | string
	stepTitle?: string
}

const SchemaConfiguration: React.FC<SchemaConfigurationProps> = ({
	selectedStreams,
	setSelectedStreams,
	syncMode,
	setSyncMode,
	enableBackfill,
	setEnableBackfill,
	normalisation,
	setNormalisation,
	stepNumber = 3,
	stepTitle = "Schema evaluation",
}) => {
	const [searchText, setSearchText] = useState("")
	const [activeTab, setActiveTab] = useState("config")
	const [syncAll, setSyncAll] = useState(false)
	const [searchSchemaText, setSearchSchemaText] = useState("")
	const [selectAllColumns, setSelectAllColumns] = useState(true)

	// Sample data - in a real app this would come from an API
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

	const filteredStreams = streams.filter(stream =>
		stream.toLowerCase().includes(searchText.toLowerCase()),
	)

	const filteredColumns = columns.filter(column =>
		column.toLowerCase().includes(searchSchemaText.toLowerCase()),
	)

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

	return (
		<div className="mb-4 p-6">
			{stepNumber && stepTitle && (
				<div className="mb-4 flex flex-col gap-2">
					<div className="flex items-center gap-2">
						<div className="size-2 rounded-full border border-[#203FDD] outline outline-2 outline-[#203FDD]"></div>
						<span className="text-gray-600">Step {stepNumber}</span>
					</div>
					<h1 className="text-xl font-medium">{stepTitle}</h1>
				</div>
			)}

			<div className="mb-4 flex items-center">
				<Input
					placeholder="Search streams"
					className="mr-4 w-64"
					value={searchText}
					onChange={e => setSearchText(e.target.value)}
					prefix={<span className="text-gray-400">🔍</span>}
				/>
				<div className="flex space-x-2">
					<Button className="border-gray-200 hover:bg-gray-50">
						All tables
					</Button>
					<Button className="border-gray-200 hover:bg-gray-50">CDC</Button>
					<Button className="border-gray-200 hover:bg-gray-50">
						Full refresh
					</Button>
					<Button className="border-gray-200 hover:bg-gray-50">Selected</Button>
					<Button className="border-gray-200 hover:bg-gray-50">
						Not selected
					</Button>
				</div>
			</div>

			<div className="flex">
				<div className="w-1/2 overflow-hidden rounded-lg border border-gray-200">
					<div className="border-b border-gray-200 p-3">
						<label className="flex items-center">
							<input
								type="checkbox"
								className="mr-2 h-4 w-4 rounded border-gray-300 text-blue-600"
								checked={syncAll}
								onChange={handleSyncAllChange}
							/>
							<span>Sync all</span>
						</label>
					</div>

					{filteredStreams.map((stream, index) => (
						<div
							key={index}
							className={`border-b border-gray-200 p-3 ${selectedStreams.includes(stream) ? "bg-blue-50" : ""}`}
						>
							<label className="flex items-center truncate">
								<input
									type="checkbox"
									className="mr-2 h-4 w-4 rounded border-gray-300 text-blue-600"
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
								<span className="truncate">{stream}</span>
							</label>
						</div>
					))}
				</div>

				<div className="ml-4 w-1/2 rounded-lg border border-gray-200 p-4">
					<h3 className="mb-4 text-lg font-medium">Payments</h3>

					<div className="mb-4 flex items-center gap-2">
						<Button
							className={`border-gray-200 ${activeTab === "config" ? "bg-blue-50 text-blue-600" : "hover:bg-gray-50"}`}
							onClick={() => setActiveTab("config")}
						>
							Config
						</Button>
						<Button
							className={`border-gray-200 ${activeTab === "schema" ? "bg-blue-50 text-blue-600" : "hover:bg-gray-50"}`}
							onClick={() => setActiveTab("schema")}
						>
							Schema
						</Button>
						<Button
							className={`border-gray-200 ${activeTab === "partitioning" ? "bg-blue-50 text-blue-600" : "hover:bg-gray-50"}`}
							onClick={() => setActiveTab("partitioning")}
						>
							Partitioning
						</Button>
					</div>

					{activeTab === "config" && (
						<div>
							<div className="mb-4">
								<label className="mb-1 block font-medium">Sync mode:</label>
								<Radio.Group
									className="mb-4"
									value={syncMode}
									onChange={e => setSyncMode(e.target.value)}
								>
									<Radio
										value="full"
										className="font-medium"
									>
										Full refresh
									</Radio>
									<Radio
										value="cdc"
										className="font-medium"
									>
										CDC
									</Radio>
								</Radio.Group>
							</div>

							<div className="mb-4">
								<div className="flex items-center justify-between">
									<label className="font-medium">Enable backfill</label>
									<Switch
										checked={enableBackfill}
										onChange={setEnableBackfill}
									/>
								</div>
							</div>

							<div className="mb-4">
								<div className="flex items-center justify-between">
									<label className="font-medium">Normalisation</label>
									<Switch
										checked={normalisation}
										onChange={setNormalisation}
									/>
								</div>
							</div>
						</div>
					)}

					{activeTab === "schema" && (
						<div>
							<div className="mb-4">
								<Input
									placeholder="Search columns"
									className="mb-4 w-full"
									value={searchSchemaText}
									onChange={e => setSearchSchemaText(e.target.value)}
								/>

								<div className="mb-2">
									<label className="flex items-center">
										<input
											type="checkbox"
											className="mr-2 h-4 w-4 rounded border-gray-300 text-blue-600"
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
												className="mr-2 h-4 w-4 rounded border-gray-300 text-blue-600"
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
						<div className="p-4 text-center text-gray-500">
							Partitioning settings would go here
						</div>
					)}
				</div>
			</div>
		</div>
	)
}

export default SchemaConfiguration
