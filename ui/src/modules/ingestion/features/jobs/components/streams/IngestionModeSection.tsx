import { InfoIcon, WarningIcon } from "@phosphor-icons/react"
import { Radio, Tooltip } from "antd"
import clsx from "clsx"

import { IngestionMode } from "../../enums"
import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
	noopNullSelector,
	noopFalseSelector,
} from "../../stores"
import {
	isSourceIngestionModeSupported,
	isDestinationIngestionModeSupported,
} from "../../utils/streams"

interface IngestionModeSectionProps {
	sourceType?: string
	destinationType?: string
	isBulkMode?: boolean
	isDirty?: boolean
	bulkAppendMode?: boolean
	onBulkIngestionModeChange?: (appendMode: boolean) => void
}

const IngestionModeSection = ({
	sourceType,
	destinationType,
	isBulkMode,
	isDirty,
	bulkAppendMode,
	onBulkIngestionModeChange,
}: IngestionModeSectionProps) => {
	const updateIngestionMode = useStreamSelectionStore(
		state => state.updateIngestionMode,
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

	const selectedStream = isBulkMode ? null : storeSelectedStream
	const isSelected = isBulkMode ? true : storeIsSelected

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

	if (!isBulkMode && (!storeStream || !selectedStream)) return null

	// Don't render if destination doesn't support upsert mode
	if (!isDestUpsertModeSupported) return null

	// Ingestion mode is Append if:
	// 1. Source doesn't support Upsert (forced Append)
	// 2. OR user selected Append mode
	const isAppendMode = isBulkMode
		? !isSourceUpsertSupported || !!bulkAppendMode
		: !isSourceUpsertSupported || !!selectedStream?.append_mode

	const handleIngestionModeChange = (ingestionMode: IngestionMode) => {
		if (isBulkMode) {
			onBulkIngestionModeChange?.(ingestionMode === IngestionMode.APPEND)
		} else {
			if (!storeStream) return
			updateIngestionMode(
				storeStream.stream.name,
				storeStream.stream.namespace || "",
				ingestionMode === IngestionMode.APPEND,
			)
		}
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
				<div className="flex items-center gap-1">
					{isDirty && <WarningIcon className="size-4 text-orange-500" />}
					<label className="block w-full">Ingestion Mode:</label>
				</div>
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
