import { useEffect, useState } from "react"
import { StreamConfigurationProps } from "../../../../types"
import { Button, Input, Radio, Select, Switch, Table } from "antd"
import StreamsSchema from "./StreamsSchema"
import {
	ColumnsPlusRight,
	GridFour,
	SlidersHorizontal,
} from "@phosphor-icons/react"
import { PARTITIONING_COLUMNS } from "../../../../utils/constants"

// Constants for styling
const TAB_STYLES = {
	active:
		"border border-[#203FDD] bg-white text-[#203FDD] rounded-[6px] py-1 px-2",
	inactive: "bg-[#F5F5F5] text-slate-900 py-1 px-2",
	hover: "hover:text-[#203FDD]",
}

const CARD_STYLE = "rounded-xl border border-[#E3E3E3] p-3"

const StreamConfiguration = ({ stream }: StreamConfigurationProps) => {
	const [activeTab, setActiveTab] = useState("config")
	const [syncMode, setSyncMode] = useState("full")
	const [enableBackfill, setEnableBackfill] = useState(true)
	const [normalisation, setNormalisation] = useState(false)
	const [partitioningValue, setPartitioningValue] = useState("set_partition")
	const [selectedColumn, setSelectedColumn] = useState<string | null>(null)
	const [defaultValue, setDefaultValue] = useState("")
	const [selectedGranularity, setSelectedGranularity] = useState<string | null>(
		null,
	)

	useEffect(() => {
		setActiveTab("config")
	}, [stream])

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

	// Handlers
	const handleAddClick = () => {
		if (!selectedColumn) return

		let granularity = "Nil"
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

	const handleSyncModeChange = (mode: string) => {
		setSyncMode(mode)
		if (mode === "full") {
			setEnableBackfill(true) // Enable backfill for full refresh
		} else {
			setEnableBackfill(false) // Disable backfill for CDC
		}
	}

	const handleAddPartitionRegex = () => {
		if (partitionRegex) {
			setPartitionInfo([...partitionInfo, partitionRegex])
			setPartitionRegex("")
		}
	}

	// Tab button component
	const TabButton = ({
		id,
		label,
		icon,
	}: {
		id: string
		label: string
		icon: React.ReactNode
	}) => {
		const tabStyle =
			activeTab === id
				? TAB_STYLES.active
				: `${TAB_STYLES.inactive} ${TAB_STYLES.hover}`

		return (
			<button
				className={`${tabStyle} flex items-center justify-center gap-1 text-xs`}
				style={{ fontWeight: 500, height: "28px", width: "100%" }}
				onClick={() => setActiveTab(id)}
				type="button"
			>
				<span className="flex items-center">{icon}</span>
				{label}
			</button>
		)
	}

	// Content rendering components
	const renderConfigContent = () => (
		<div className="flex flex-col gap-4">
			<div className={CARD_STYLE}>
				<div className="mb-4">
					<label className="mb-3 block w-full font-medium text-[#575757]">
						Sync mode:
					</label>
					<Radio.Group
						className="mb-4 flex w-full items-center"
						value={syncMode}
						onChange={e => handleSyncModeChange(e.target.value)}
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
			<div className={CARD_STYLE}>
				<div className="flex items-center justify-between">
					<label className="font-medium">Enable backfill</label>
					<Switch
						checked={enableBackfill}
						onChange={setEnableBackfill}
						disabled={syncMode === "full"}
					/>
				</div>
			</div>
			<div className={`mb-4 ${CARD_STYLE}`}>
				<div className="flex items-center justify-between">
					<label className="font-medium">Normalisation</label>
					<Switch
						checked={normalisation}
						onChange={setNormalisation}
					/>
				</div>
			</div>
		</div>
	)

	const renderSetPartitionContent = () => (
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
				{isDateTimeColumn && renderGranularityButtons()}
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
				rowClassName="no-hover"
			/>
			<div className="text-sm text-[#575757]">Regex preview:</div>
		</>
	)

	const renderGranularityButtons = () => (
		<div className="flex gap-2">
			{["day", "month", "year"].map(value => (
				<Button
					key={value}
					className={`text-[#575757] ${selectedGranularity === value ? "border-none bg-[#E9EBFC]" : ""}`}
					onClick={() => setSelectedGranularity(value)}
				>
					{value.charAt(0).toUpperCase() + value.slice(1)}
				</Button>
			))}
		</div>
	)

	const renderPartitioningRegexContent = () => (
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
				onClick={handleAddPartitionRegex}
			>
				Partition
			</Button>
			{partitionInfo.length > 0 && (
				<div className="mt-4">
					<div className="text-sm text-[#575757]">Added partitions:</div>
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
	)

	const renderPartitioningContent = () => (
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

			{partitioningValue === "set_partition"
				? renderSetPartitionContent()
				: renderPartitioningRegexContent()}
		</div>
	)

	// Main render
	return (
		<div>
			<div className="pb-4 font-medium capitalize">{stream.stream.name}</div>
			<div className="mb-4 w-full">
				<div className="grid grid-cols-3 gap-1 rounded-[6px] bg-[#F5F5F5] p-1">
					<TabButton
						id="config"
						label="Config"
						icon={<SlidersHorizontal className="size-3.5" />}
					/>
					<TabButton
						id="schema"
						label="Schema"
						icon={<ColumnsPlusRight className="size-3.5" />}
					/>
					<TabButton
						id="partitioning"
						label="Partitioning"
						icon={<GridFour className="size-3.5" />}
					/>
				</div>
			</div>

			{activeTab === "config" && renderConfigContent()}
			{activeTab === "schema" && (
				<StreamsSchema initialData={stream.stream.json_schema?.properties} />
			)}
			{activeTab === "partitioning" && renderPartitioningContent()}
		</div>
	)
}

export default StreamConfiguration
