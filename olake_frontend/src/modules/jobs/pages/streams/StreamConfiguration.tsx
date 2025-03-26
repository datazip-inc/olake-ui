import { useState } from "react"
import { StreamConfigurationProps } from "../../../../types"
import { Button, Input, Radio, Select, Switch, Table } from "antd"
import StreamsSchema from "./StreamsSchema"
import {
	ColumnsPlusRight,
	GridFour,
	SlidersHorizontal,
} from "@phosphor-icons/react"
import { PARTITIONING_COLUMNS } from "../../../../utils/constants"

const StreamConfiguration = ({ stream }: StreamConfigurationProps) => {
	const [activeTab, setActiveTab] = useState("config")
	const [syncMode, setSyncMode] = useState(stream.sync_mode)
	const [enableBackfill, setEnableBackfill] = useState(false)
	const [normalisation, setNormalisation] = useState(false)
	const [partitioningValue, setPartitioningValue] = useState("set_partition")
	const [selectedColumn, setSelectedColumn] = useState<string | null>(null)
	const [defaultValue, setDefaultValue] = useState("")
	const [selectedGranularity, setSelectedGranularity] = useState<string | null>(
		null,
	)
	const [tableData, setTableData] = useState<
		Array<{ name: string; granularity: string; default: string }>
	>([])
	const [partitionRegex, setPartitionRegex] = useState("")
	const [partitionInfo, setPartitionInfo] = useState<string[]>([])

	// Transform properties into Select options
	const propertyOptions = stream.stream.json_schema?.properties
		? Object.entries(stream.stream.json_schema.properties).map(
				([key, value]) => ({
					value: key,
					label: key,
					format: (value as any).format,
				}),
			)
		: []

	const isDateTimeColumn = selectedColumn
		? propertyOptions.find(opt => opt.value === selectedColumn)?.format ===
			"date-time"
		: false

	const handleAddClick = () => {
		if (!selectedColumn) return

		let granularity = "	Nil"
		if (selectedGranularity === "day") {
			granularity = "DD"
		} else if (selectedGranularity === "month") {
			granularity = "MM"
		} else if (selectedGranularity === "year") {
			granularity = "YYYY"
		}

		setTableData([
			...tableData,
			{
				name: selectedColumn,
				granularity: granularity,
				default: defaultValue,
			},
		])

		// Reset form
		setSelectedColumn(null)
		setSelectedGranularity(null)
		setDefaultValue("")
	}

	return (
		<div>
			<div className="pb-4 font-medium capitalize">{stream.stream.name}</div>
			<div className="mb-4 flex w-full items-center">
				<Button
					className={`${activeTab === "config" ? "border border-[#203FDD] text-blue-600" : "border-none bg-[#F5F5F5] text-slate-900"} w-1/3`}
					onClick={() => setActiveTab("config")}
					icon={<SlidersHorizontal className="size-4" />}
				>
					Config
				</Button>
				<Button
					className={`${activeTab === "schema" ? "text-blue-600" : "border-none bg-[#F5F5F5] text-slate-900"} w-1/3`}
					onClick={() => setActiveTab("schema")}
					icon={<ColumnsPlusRight className="size-4" />}
				>
					Schema
				</Button>
				<Button
					className={` ${activeTab === "partitioning" ? "text-blue-600" : "border-none bg-[#F5F5F5] text-slate-900"} w-1/3`}
					onClick={() => setActiveTab("partitioning")}
					icon={<GridFour className="size-4" />}
				>
					Partitioning
				</Button>
			</div>
			{activeTab === "config" && (
				<>
					<div className="flex flex-col gap-4">
						<div className="rounded-xl border border-[#E3E3E3] p-3">
							<div className="mb-4">
								<label className="mb-3 block w-full font-medium text-[#575757]">
									Sync mode:
								</label>
								<Radio.Group
									className="mb-4 flex w-full items-center"
									value={syncMode}
									onChange={e => setSyncMode(e.target.value)}
								>
									<Radio
										value="full"
										className="w-1/3"
									>
										Full refresh
									</Radio>
									<Radio
										value="cdc"
										className="w-1/3"
									>
										CDC
									</Radio>
								</Radio.Group>
							</div>
						</div>
						<div className="rounded-xl border border-[#E3E3E3] p-3">
							<div className="flex items-center justify-between">
								<label className="font-medium">Enable backfill</label>
								<Switch
									checked={enableBackfill}
									onChange={setEnableBackfill}
								/>
							</div>
						</div>

						<div className="mb-4 rounded-xl border border-[#E3E3E3] p-3">
							<div className="flex items-center justify-between">
								<label className="font-medium">Normalisation</label>
								<Switch
									checked={normalisation}
									onChange={setNormalisation}
								/>
							</div>
						</div>
					</div>
				</>
			)}
			{activeTab === "schema" && (
				<StreamsSchema initialData={stream.stream.json_schema?.properties} />
			)}

			{activeTab === "partitioning" && (
				<div className="flex flex-col gap-4">
					<div>
						<Radio.Group
							className="mb-4 flex w-full items-center"
							value={partitioningValue}
							onChange={e => setPartitioningValue(e.target.value)}
						>
							<Radio
								value="set_partition"
								className="w-1/2"
							>
								Set partition
							</Radio>
							<Radio
								value="partitioning_regex"
								className="w-1/2"
							>
								Partitioning regex
							</Radio>
						</Radio.Group>
					</div>

					{partitioningValue === "set_partition" && (
						<>
							<div>Select column:</div>
							<div className="flex w-full justify-between gap-2">
								<Select
									showSearch
									placeholder="Select columns"
									optionFilterProp="label"
									className="w-2/4"
									options={propertyOptions}
									onChange={value => setSelectedColumn(value)}
									value={selectedColumn}
								/>
								{isDateTimeColumn && (
									<div className="flex gap-2">
										<Button
											className={`text-[#575757] ${selectedGranularity === "day" ? "border-none bg-[#E9EBFC]" : ""}`}
											onClick={() => setSelectedGranularity("day")}
										>
											Day
										</Button>
										<Button
											className={`text-[#575757] ${selectedGranularity === "month" ? "border-none bg-[#E9EBFC]" : ""}`}
											onClick={() => setSelectedGranularity("month")}
										>
											Month
										</Button>
										<Button
											className={`text-[#575757] ${selectedGranularity === "year" ? "border-none bg-[#E9EBFC]" : ""}`}
											onClick={() => setSelectedGranularity("year")}
										>
											Year
										</Button>
									</div>
								)}
							</div>
							<div>Default value:</div>
							<Input
								placeholder="Enter default value for your column"
								className="w-2/3"
								value={defaultValue}
								onChange={e => setDefaultValue(e.target.value)}
							/>
							<Button
								className="w-16 bg-[#203FDD] py-3 font-light text-white"
								onClick={handleAddClick}
							>
								Add
							</Button>
							<Table
								dataSource={tableData}
								columns={PARTITIONING_COLUMNS}
								pagination={false}
							/>
							<div className="text-sm text-[#575757]">Regex preview:</div>
						</>
					)}

					{partitioningValue === "partitioning_regex" && (
						<>
							<div className="text-[#575757]">Partitioning regex:</div>
							<Input
								placeholder="Enter your partition regex"
								className="w-full"
								value={partitionRegex}
								onChange={e => setPartitionRegex(e.target.value)}
							/>
							<Button
								className="w-20 bg-[#203FDD] py-3 font-light text-white"
								onClick={() => {
									if (partitionRegex) {
										setPartitionInfo([...partitionInfo, partitionRegex])
										setPartitionRegex("")
									}
								}}
							>
								Partition
							</Button>
							{partitionInfo.length > 0 && (
								<div className="mt-4">
									<div className="text-sm text-[#575757]">
										Added partitions:
									</div>
									{partitionInfo.map((regex, index) => (
										<div
											key={index}
											className="mt-2 text-sm"
										>
											{regex}
										</div>
									))}
								</div>
							)}
						</>
					)}
				</div>
			)}
		</div>
	)
}

export default StreamConfiguration
