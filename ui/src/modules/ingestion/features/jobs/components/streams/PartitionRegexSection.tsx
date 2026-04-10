import { ArrowSquareOutIcon, InfoIcon } from "@phosphor-icons/react"
import { Button, Input, Tooltip } from "antd"
import { useState } from "react"

import { DESTINATION_INTERNAL_TYPES } from "@/modules/ingestion/common/constants"

import { PartitioningRegexTooltip } from "../../constants"
import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
	noopNullSelector,
	noopFalseSelector,
} from "../../stores"

interface PartitionRegexSectionProps {
	destinationType?: string
	isBulkMode?: boolean
	bulkPartitionRegex?: string
	onBulkPartitionRegexChange?: (regex: string) => void
}

const PartitionRegexSection = ({
	destinationType = DESTINATION_INTERNAL_TYPES.S3,
	isBulkMode,
	bulkPartitionRegex,
	onBulkPartitionRegexChange,
}: PartitionRegexSectionProps) => {
	const updatePartitionRegex = useStreamSelectionStore(
		state => state.updatePartitionRegex,
	)
	// don't subsribe to store if in bulkMode
	const storeStream = useStreamSelectionStore(
		isBulkMode ? noopNullSelector : selectActiveStreamData,
	)
	const storeSelectedStream = useStreamSelectionStore(
		isBulkMode ? noopNullSelector : selectActiveSelectedStream,
	)
	const storeIsSelected = useStreamSelectionStore(
		isBulkMode
			? noopFalseSelector
			: state => selectIsStreamEnabled(state, storeStream),
	)

	const selectedStream = isBulkMode
		? { partition_regex: bulkPartitionRegex }
		: storeSelectedStream
	const isSelected = isBulkMode ? true : storeIsSelected

	const [partitionRegex, setPartitionRegex] = useState("")

	if (!isBulkMode && (!storeStream || !selectedStream)) return null

	const activePartitionRegex = isBulkMode
		? bulkPartitionRegex
		: selectedStream?.partition_regex || ""

	const handleSetPartitionRegex = () => {
		if (partitionRegex) {
			if (isBulkMode) {
				onBulkPartitionRegexChange?.(partitionRegex)
			} else {
				if (!storeStream) return
				updatePartitionRegex(
					storeStream.stream.name,
					storeStream.stream.namespace || "",
					partitionRegex,
				)
			}
			setPartitionRegex("")
		}
	}

	const handleClearPartitionRegex = () => {
		if (isBulkMode) {
			onBulkPartitionRegexChange?.("")
		} else {
			if (!storeStream) return
			updatePartitionRegex(
				storeStream.stream.name,
				storeStream.stream.namespace || "",
				"",
			)
		}
	}

	return (
		<div className="flex flex-col gap-4">
			<div className="flex items-center gap-0.5">
				<div className="text-neutral-text">Partitioning regex:</div>

				<Tooltip title={PartitioningRegexTooltip}>
					<InfoIcon className="size-5 cursor-help items-center pt-1 text-gray-500" />
				</Tooltip>
				<a
					href={
						destinationType === DESTINATION_INTERNAL_TYPES.S3
							? "https://olake.io/docs/writers/parquet/partitioning"
							: "https://olake.io/docs/writers/iceberg/partitioning"
					}
					target="_blank"
					rel="noopener noreferrer"
					className="flex items-center text-primary hover:text-primary/80"
				>
					<ArrowSquareOutIcon className="size-5" />
				</a>
			</div>
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
							className="mt-2 w-fit bg-primary px-2 py-3 font-light text-white"
							onClick={handleSetPartitionRegex}
							disabled={!partitionRegex}
						>
							Set Partition
						</Button>
					) : (
						<div className="mt-4">
							<div className="text-sm text-neutral-text">
								Active partition regex:
							</div>
							<div className="mt-2 flex items-center justify-between text-sm">
								<span>{activePartitionRegex}</span>
								<Button
									type="text"
									danger
									size="small"
									className="rounded-md py-1 text-sm"
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
					<InfoIcon className="size-4" />
					Select the stream to configure Partitioning
				</div>
			)}
		</div>
	)
}

export default PartitionRegexSection
