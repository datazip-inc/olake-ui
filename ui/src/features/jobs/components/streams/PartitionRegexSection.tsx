import { useEffect, useState } from "react"
import { Button, Input, Tooltip } from "antd"
import { ArrowSquareOutIcon, InfoIcon } from "@phosphor-icons/react"

import { PartitioningRegexTooltip } from "../../constants"
import { DESTINATION_INTERNAL_TYPES } from "@/common/constants/constants"

import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
} from "../../stores/streamSelectionStore"

interface PartitionRegexSectionProps {
	destinationType?: string
}

const PartitionRegexSection = ({
	destinationType = DESTINATION_INTERNAL_TYPES.S3,
}: PartitionRegexSectionProps) => {
	const store = useStreamSelectionStore()
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)
	const isSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, stream),
	)

	const [partitionRegex, setPartitionRegex] = useState("")
	const [activePartitionRegex, setActivePartitionRegex] = useState(
		selectedStream?.partition_regex || "",
	)

	// Re-sync partition state when the active stream changes
	useEffect(() => {
		if (!selectedStream) return
		setActivePartitionRegex(selectedStream.partition_regex || "")
		setPartitionRegex("")
	}, [stream])

	if (!stream || !selectedStream) return null

	const handleSetPartitionRegex = () => {
		if (partitionRegex) {
			setActivePartitionRegex(partitionRegex)
			setPartitionRegex("")
			store.updatePartitionRegex(
				stream.stream.name,
				stream.stream.namespace || "",
				partitionRegex,
			)
		}
	}

	// deletes the partition regex for the corresponding stream
	const handleClearPartitionRegex = () => {
		setActivePartitionRegex("")
		setPartitionRegex("")
		// deletes the partition regex for the corresponding stream
		store.updatePartitionRegex(
			stream.stream.name,
			stream.stream.namespace || "",
			"",
		)
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
