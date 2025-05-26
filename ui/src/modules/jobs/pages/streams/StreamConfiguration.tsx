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
		stream.stream.sync_mode === "full_refresh" ? "full" : "cdc",
	)
	const [enableBackfill, setEnableBackfill] = useState(false)
	const [normalisation, setNormalisation] = useState(false)
	const [partitionRegex, setPartitionRegex] = useState("")
	const [partitionInfo, setPartitionInfo] = useState<string[]>([])
	const [formData, setFormData] = useState<any>({
		sync_mode: stream.stream.sync_mode,
		backfill: false,
		partition_regex: "",
	})

	useEffect(() => {
		setActiveTab("config")
		const initialApiSyncMode = stream.stream.sync_mode
		let initialEnableBackfillForSwitch = false

		if (initialApiSyncMode === "full_refresh") {
			setSyncMode("full")
			initialEnableBackfillForSwitch = true
		} else if (initialApiSyncMode === "cdc") {
			setSyncMode("cdc")
			initialEnableBackfillForSwitch = true
		} else if (initialApiSyncMode === "strict_cdc") {
			setSyncMode("cdc")
			initialEnableBackfillForSwitch = false
		}
		setEnableBackfill(initialEnableBackfillForSwitch)

		setFormData((prevFormData: any) => ({
			...prevFormData,
			sync_mode: initialApiSyncMode,
			backfill: initialEnableBackfillForSwitch,
		}))
	}, [stream])

	// Handlers
	const handleSyncModeChange = (selectedRadioValue: string) => {
		setSyncMode(selectedRadioValue)
		let newApiSyncMode: "full_refresh" | "cdc" = "cdc"
		let newEnableBackfillState = true

		if (selectedRadioValue === "full") {
			newApiSyncMode = "full_refresh"
			newEnableBackfillState = true
		} else {
			newApiSyncMode = "cdc"
			newEnableBackfillState = true
		}

		stream.stream.sync_mode = newApiSyncMode
		setEnableBackfill(newEnableBackfillState)
		onSyncModeChange?.(
			stream.stream.name,
			stream.stream.namespace || "default",
			newApiSyncMode,
		)

		setFormData({
			...formData,
			sync_mode: newApiSyncMode,
			backfill: newEnableBackfillState,
		})
	}

	const handleEnableBackfillChange = (checked: boolean) => {
		setEnableBackfill(checked)
		let finalApiSyncMode = stream.stream.sync_mode

		if (syncMode === "cdc") {
			if (checked) {
				finalApiSyncMode = "cdc"
				stream.stream.sync_mode = "cdc"
				onSyncModeChange?.(
					stream.stream.name,
					stream.stream.namespace || "default",
					"cdc",
				)
			} else {
				finalApiSyncMode = "strict_cdc"
				stream.stream.sync_mode = "strict_cdc"
			}
		}

		setFormData({
			...formData,
			backfill: checked,
			sync_mode: finalApiSyncMode,
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
							onChange={handleEnableBackfillChange}
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
