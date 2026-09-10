import { InfoIcon, WarningIcon } from "@phosphor-icons/react"
import { Radio, Tooltip } from "antd"
import clsx from "clsx"

import { UpsertType } from "@/modules/ingestion/common/types"

import { UPSERT_TYPE_OPTIONS } from "../../constants"
import { IngestionMode } from "../../enums"
import {
	isDestinationIngestionModeSupported,
	isSourceIngestionModeSupported,
} from "../../utils/streams"

interface UpsertTypeSectionViewProps {
	sourceType?: string
	destinationType?: string
	isSelected: boolean
	isDirty?: boolean
	appendMode: boolean
	upsertType?: UpsertType
	onChange: (upsertType: UpsertType) => void
}

const UpsertTypeSectionView = ({
	sourceType,
	destinationType,
	isSelected,
	isDirty,
	appendMode,
	upsertType,
	onChange,
}: UpsertTypeSectionViewProps) => {
	const isSourceUpsertSupported = isSourceIngestionModeSupported(
		IngestionMode.UPSERT,
		sourceType,
	)
	const isDestUpsertModeSupported = isDestinationIngestionModeSupported(
		IngestionMode.UPSERT,
		destinationType,
	)

	// Visible only while the stream actually runs in upsert mode.
	if (!isDestUpsertModeSupported || !isSourceUpsertSupported) return null
	// upsertType is seeded from the catalog's default_stream_properties.update_mode;
	// when absent the driver version doesn't support update_type.
	if (appendMode || upsertType === undefined) return null

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
					<label className="block w-full">Upsert Type:</label>
				</div>
				<div
					className={clsx(
						"text-xs",
						!isSelected ? "text-gray-500" : "text-neutral-700",
					)}
				>
					Specify Iceberg delete mode (for faster reads and multi query engine
					support use positional).{" "}
					<a
						href="https://olake.io/docs/understanding/terminologies/olake/#upsert"
						target="_blank"
						rel="noopener noreferrer"
						className="text-primary underline hover:opacity-80"
					>
						Learn more
					</a>
				</div>
			</div>
			<Radio.Group
				disabled={!isSelected}
				className="mb-4 grid grid-cols-2 gap-4"
				value={upsertType}
				onChange={e => onChange(e.target.value)}
			>
				{UPSERT_TYPE_OPTIONS.map(option => (
					<Tooltip
						key={option.value}
						title={option.tooltip}
					>
						<Radio value={option.value}>{option.label}</Radio>
					</Tooltip>
				))}
			</Radio.Group>
			{upsertType === UpsertType.POSITIONAL && (
				<div className="mb-4 flex items-start gap-1.5 text-xs leading-5 text-amber-700">
					<WarningIcon className="mt-0.5 size-4 shrink-0 text-amber-600" />
					<span>
						Positional deletes need a newer OLake version.{" "}
						<a
							href="https://olake.io/docs/understanding/terminologies/olake/#upsert"
							target="_blank"
							rel="noopener noreferrer"
							className="text-primary underline hover:opacity-80"
						>
							Check here
						</a>{" "}
						and upgrade (ignore if already done).
					</span>
				</div>
			)}
			{!isSelected && (
				<div className="flex items-center gap-1 text-sm text-[#686868]">
					<InfoIcon className="size-4" />
					Select the stream to configure upsert type
				</div>
			)}
		</div>
	)
}

export default UpsertTypeSectionView
