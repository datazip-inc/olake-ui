import { InfoIcon, WarningIcon } from "@phosphor-icons/react"
import { Radio, Tooltip } from "antd"
import clsx from "clsx"

import { IngestionMode } from "../../enums"
import {
	getKafkaUpsertVersionMessage,
	isSourceIngestionModeSupported,
	isDestinationIngestionModeSupported,
} from "../../utils/streams"

interface IngestionModeSectionViewProps {
	sourceType?: string
	sourceVersion?: string
	destinationType?: string
	isSelected: boolean
	isDirty?: boolean
	appendMode: boolean
	dedupKeyCount?: number
	onChange: (ingestionMode: IngestionMode) => void
}

const IngestionModeSectionView = ({
	sourceType,
	sourceVersion,
	destinationType,
	isSelected,
	isDirty,
	appendMode,
	dedupKeyCount,
	onChange,
}: IngestionModeSectionViewProps) => {
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

	const versionMessage = getKafkaUpsertVersionMessage(sourceType, sourceVersion)
	const noCommonDedupMessage =
		dedupKeyCount !== undefined && dedupKeyCount === 0
			? "No common dedup keys found."
			: undefined

	const upsertDisabledMessage = versionMessage || noCommonDedupMessage
	const isUpsertDisabled = !isSourceUpsertSupported || !!upsertDisabledMessage

	// Don't render if destination doesn't support upsert mode
	if (!isDestUpsertModeSupported) return null

	// Ingestion mode is Append if:
	// 1. Source doesn't support Upsert (forced Append)
	// 2. OR user selected Append mode
	const isAppendMode = isUpsertDisabled || !!appendMode

	const upsertTooltip = !isSourceUpsertSupported
		? "Upsert is not supported for this source"
		: upsertDisabledMessage

	const handleIngestionModeChange = (ingestionMode: IngestionMode) => {
		onChange(ingestionMode)
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
				<Tooltip title={isUpsertDisabled ? upsertTooltip : undefined}>
					<span className="inline-block">
						<Radio
							value={IngestionMode.UPSERT}
							disabled={isUpsertDisabled}
							className={clsx(isUpsertDisabled && "opacity-50")}
						>
							Upsert
						</Radio>
					</span>
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
			{upsertDisabledMessage && (
				<div className="mb-4 flex items-center gap-1 text-sm text-[#686868]">
					<InfoIcon className="size-4 shrink-0" />
					{upsertDisabledMessage}
				</div>
			)}
			{!isSelected && (
				<div className="flex items-center gap-1 text-sm text-[#686868]">
					<InfoIcon className="size-4" />
					Select the stream to configure ingestion mode
				</div>
			)}
		</div>
	)
}

export default IngestionModeSectionView
