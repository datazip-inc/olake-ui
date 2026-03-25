import { InfoIcon } from "@phosphor-icons/react"
import { Radio, Tooltip } from "antd"
import clsx from "clsx"

import { IngestionMode } from "../../enums"
import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
} from "../../stores"
import {
	isSourceIngestionModeSupported,
	isDestinationIngestionModeSupported,
} from "../../utils/streams"

interface IngestionModeSectionProps {
	sourceType?: string
	destinationType?: string
}

const IngestionModeSection = ({
	sourceType,
	destinationType,
}: IngestionModeSectionProps) => {
	const updateIngestionMode = useStreamSelectionStore(
		state => state.updateIngestionMode,
	)
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)
	const isSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, stream),
	)

	const isSourceUpsertSupported = isSourceIngestionModeSupported(
		IngestionMode.UPSERT,
		sourceType,
	)

	const isSourceAppendSupported = isSourceIngestionModeSupported(
		IngestionMode.APPEND,
		sourceType,
	)

	const isDestUpsertModeSupported = isDestinationIngestionModeSupported(
		IngestionMode.UPSERT,
		destinationType,
	)

	if (!stream || !selectedStream) return null

	// Don't render if destination doesn't support upsert mode
	if (!isDestUpsertModeSupported) return null

	// Ingestion mode is Append if:
	// 1. Source doesn't support Upsert (forced Append)
	// 2. OR user selected Append mode
	const isAppendMode = !isSourceUpsertSupported || !!selectedStream.append_mode

	const handleIngestionModeChange = (ingestionMode: IngestionMode) => {
		updateIngestionMode(
			stream.stream.name,
			stream.stream.namespace || "",
			ingestionMode === IngestionMode.APPEND,
		)
	}

	return (
		<div
			className={clsx(
				"mb-4",
				isSelected
					? "font-medium text-neutral-text"
					: "font-normal text-gray-500",
			)}
		>
			<div className="mb-3">
				<label className="block w-full">Ingestion Mode:</label>
				<div
					className={clsx(
						"text-xs",
						!isSelected ? "text-gray-500" : "text-neutral-700",
					)}
				>
					Specify how the data will be ingested in the destination
				</div>
			</div>
			<Radio.Group
				disabled={!isSelected}
				className="mb-4 grid grid-cols-2 gap-4"
				value={isAppendMode ? IngestionMode.APPEND : IngestionMode.UPSERT}
				onChange={e => handleIngestionModeChange(e.target.value)}
			>
				<Tooltip
					title={
						!isSourceUpsertSupported
							? "Upsert is not supported for this source"
							: undefined
					}
				>
					<Radio
						value={IngestionMode.UPSERT}
						disabled={!isSourceUpsertSupported}
						className={clsx(!isSourceUpsertSupported && "opacity-50")}
					>
						Upsert
					</Radio>
				</Tooltip>
				<Tooltip
					title={
						!isSourceAppendSupported
							? "Append is not supported for this source"
							: undefined
					}
				>
					<Radio
						value={IngestionMode.APPEND}
						disabled={!isSourceAppendSupported}
						className={clsx(!isSourceAppendSupported && "opacity-50")}
					>
						Append
					</Radio>
				</Tooltip>
			</Radio.Group>
			{!isSelected && (
				<div className="flex items-center gap-1 text-sm text-[#686868]">
					<InfoIcon className="size-4" />
					Select the stream to configure ingestion mode
				</div>
			)}
		</div>
	)
}

export default IngestionModeSection
