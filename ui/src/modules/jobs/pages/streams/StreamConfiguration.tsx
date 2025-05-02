import { useEffect, useState } from "react"
import { StreamConfigurationProps } from "../../../../types"
import { Button, Input, Radio, Switch } from "antd"
import StreamsSchema from "./StreamsSchema"
import {
	ColumnsPlusRight,
	GridFour,
	SlidersHorizontal,
} from "@phosphor-icons/react"

// Constants for styling
const TAB_STYLES = {
	active:
		"border border-[#203FDD] bg-white text-[#203FDD] rounded-[6px] py-1 px-2",
	inactive: "bg-[#F5F5F5] text-slate-900 py-1 px-2",
	hover: "hover:text-[#203FDD]",
}

const CARD_STYLE = "rounded-xl border border-[#E3E3E3] p-3"

const StreamConfiguration = ({
	stream,
	onSyncModeChange,
}: StreamConfigurationProps & {
	onUpdate?: (stream: any) => void
}) => {
	const [activeTab, setActiveTab] = useState("config")
	const [syncMode, setSyncMode] = useState(
		stream.stream.sync_mode === "cdc" ? "cdc" : "full",
	)
	const [enableBackfill, setEnableBackfill] = useState(syncMode === "full")
	const [normalisation, setNormalisation] = useState(false)
	const [partitionRegex, setPartitionRegex] = useState("")
	const [partitionInfo, setPartitionInfo] = useState<string[]>([])
	const [formData, setFormData] = useState<any>({
		sync_mode: stream.stream.sync_mode,
		backfill: enableBackfill,
		partition_regex: "",
	})

	useEffect(() => {
		setActiveTab("config")
	}, [stream])

	// Handlers
	const handleSyncModeChange = (mode: string) => {
		const newSyncMode = mode === "full" ? "full_refresh" : "cdc"
		setSyncMode(mode)
		stream.stream.sync_mode = newSyncMode
		onSyncModeChange?.(
			stream.stream.name,
			stream.stream.namespace || "default",
			newSyncMode,
		)
		if (mode === "full") {
			setEnableBackfill(true) // Enable backfill for full refresh
		} else {
			setEnableBackfill(false) // Disable backfill for CDC
		}

		setFormData({
			...formData,
			sync_mode: newSyncMode,
			backfill: mode === "full",
		})
	}

	const handleAddPartitionRegex = () => {
		if (partitionRegex) {
			setPartitionInfo([...partitionInfo, partitionRegex])
			setPartitionRegex("")

			setFormData({
				...formData,
				partition_regex: [...partitionInfo, partitionRegex].join(","),
			})
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
	const renderConfigContent = () => {
		return (
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
	}

	const renderPartitioningContent = () => (
		<div className="flex flex-col gap-4">
			{renderPartitioningRegexContent()}
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
				disabled={!partitionRegex}
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
			{activeTab === "schema" && <StreamsSchema initialData={stream} />}
			{activeTab === "partitioning" && renderPartitioningContent()}
		</div>
	)
}

export default StreamConfiguration
