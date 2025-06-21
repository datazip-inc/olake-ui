import { useEffect, useState } from "react"
import { ExtendedStreamConfigurationProps } from "../../../../types"
import { Button, Input, Radio, Switch } from "antd"
import StreamsSchema from "./StreamsSchema"
import {
	ColumnsPlusRight,
	GridFour,
	Info,
	SlidersHorizontal,
} from "@phosphor-icons/react"
import { CARD_STYLE, TAB_STYLES } from "../../../../utils/constants"

const StreamConfiguration = ({
	stream,
	onSyncModeChange,
	isSelected,
	initialNormalization,
	initialPartitionRegex,
	onNormalizationChange,
	onPartitionRegexChange,
}: ExtendedStreamConfigurationProps) => {
	const [activeTab, setActiveTab] = useState("config")
	const [syncMode, setSyncMode] = useState(
		stream.stream.sync_mode === "full_refresh" ? "full" : "cdc",
	)
	// const [enableBackfill, setEnableBackfill] = useState(false)
	const [normalisation, setNormalisation] =
		useState<boolean>(initialNormalization)
	const [partitionRegex, setPartitionRegex] = useState("")
	const [activePartitionRegex, setActivePartitionRegex] = useState(
		initialPartitionRegex || "",
	)
	const [formData, setFormData] = useState<any>({
		sync_mode: stream.stream.sync_mode,
		backfill: false,
		partition_regex: initialPartitionRegex || "",
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
		// setEnableBackfill(initialEnableBackfillForSwitch)
		setNormalisation(initialNormalization)
		setActivePartitionRegex(initialPartitionRegex || "")
		setPartitionRegex("")

		setFormData((prevFormData: any) => ({
			...prevFormData,
			sync_mode: initialApiSyncMode,
			backfill: initialEnableBackfillForSwitch,
			partition_regex: initialPartitionRegex || "",
		}))
	}, [stream, initialNormalization, initialPartitionRegex])

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
		// setEnableBackfill(newEnableBackfillState)
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

	// const handleEnableBackfillChange = (checked: boolean) => {
	// 	setEnableBackfill(checked)
	// 	let finalApiSyncMode = stream.stream.sync_mode

	// 	if (syncMode === "cdc") {
	// 		if (checked) {
	// 			finalApiSyncMode = "cdc"
	// 			stream.stream.sync_mode = "cdc"
	// 			onSyncModeChange?.(
	// 				stream.stream.name,
	// 				stream.stream.namespace || "default",
	// 				"cdc",
	// 			)
	// 		} else {
	// 			finalApiSyncMode = "strict_cdc"
	// 			stream.stream.sync_mode = "strict_cdc"
	// 		}
	// 	}

	// 	setFormData({
	// 		...formData,
	// 		backfill: checked,
	// 		sync_mode: finalApiSyncMode,
	// 	})
	// }

	const handleNormalizationChange = (checked: boolean) => {
		setNormalisation(checked)
		onNormalizationChange(
			stream.stream.name,
			stream.stream.namespace || "default",
			checked,
		)
		setFormData({
			...formData,
			normalization: checked,
		})
	}

	const handleSetPartitionRegex = () => {
		if (partitionRegex) {
			setActivePartitionRegex(partitionRegex)
			setPartitionRegex("")
			onPartitionRegexChange(
				stream.stream.name,
				stream.stream.namespace || "default",
				partitionRegex,
			)
			setFormData({
				...formData,
				partition_regex: partitionRegex,
			})
		}
	}

	const handleClearPartitionRegex = () => {
		setActivePartitionRegex("")
		setPartitionRegex("")
		onPartitionRegexChange(
			stream.stream.name,
			stream.stream.namespace || "default",
			"",
		)
		setFormData({
			...formData,
			partition_regex: "",
		})
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
						<label className="text-[#c1c1c1]">Enable backfill</label>
						<Switch
							// className="text-[#c1c1c1]"
							// checked={enableBackfill}
							// onChange={handleEnableBackfillChange}
							// disabled={syncMode === "full"}
							checked={true}
							disabled={true}
						/>
					</div>
				</div>

				<div
					className={`${!isSelected ? "font-normal text-[#c1c1c1]" : "font-medium"} ${CARD_STYLE}`}
				>
					<div className="flex items-center justify-between">
						<label>Normalisation</label>
						<Switch
							checked={normalisation}
							onChange={handleNormalizationChange}
							disabled={!isSelected}
						/>
					</div>
				</div>
				{!isSelected && (
					<div className="ml-1 flex items-center gap-1 text-sm text-[#686868]">
						<Info className="size-4" />
						Select the stream to configure Normalisation
					</div>
				)}
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
			{isSelected ? (
				<>
					<Input
						placeholder="Enter your partition regex"
						className="w-full"
						value={partitionRegex}
						onChange={e => setPartitionRegex(e.target.value)}
						disabled={!!activePartitionRegex}
					/>
					{!activePartitionRegex ? (
						<Button
							className="mt-2 w-fit bg-[#203FDD] px-1 py-3 font-light text-white"
							onClick={handleSetPartitionRegex}
							disabled={!partitionRegex}
						>
							Set Partition
						</Button>
					) : (
						<div className="mt-4">
							<div className="text-sm text-[#575757]">
								Active partition regex:
							</div>
							<div className="mt-2 flex items-center justify-between text-sm">
								<span>{activePartitionRegex}</span>
								<Button
									type="text"
									danger
									size="small"
									className="rounded-[6px] py-1 text-sm"
									onClick={handleClearPartitionRegex}
								>
									Delete Partition
								</Button>
							</div>
						</div>
					)}
				</>
			) : (
				<div className="ml-1 flex items-center gap-1 text-sm text-[#686868]">
					<Info className="size-4" />
					Select the stream to configure Partitioning
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
