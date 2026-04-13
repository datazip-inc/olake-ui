import { InfoIcon } from "@phosphor-icons/react"
import { Switch } from "antd"
import clsx from "clsx"

import { CARD_STYLE } from "../../constants"
import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
	noopNullSelector,
	noopFalseSelector,
} from "../../stores"

interface NormalizationSectionProps {
	isBulkMode?: boolean
	isDirty?: boolean
	bulkNormalization?: boolean
	onBulkNormalizationChange?: (normalization: boolean) => void
}

const NormalizationSection = ({
	isBulkMode,
	isDirty,
	bulkNormalization,
	onBulkNormalizationChange,
}: NormalizationSectionProps = {}) => {
	const updateNormalization = useStreamSelectionStore(
		state => state.updateNormalization,
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
		? { normalization: bulkNormalization }
		: storeSelectedStream
	const isSelected = isBulkMode ? true : storeIsSelected

	if (!isBulkMode && (!storeStream || !selectedStream)) return null

	const handleNormalizationChange = (checked: boolean) => {
		if (isBulkMode) {
			onBulkNormalizationChange?.(checked)
		} else {
			if (!storeStream) return
			updateNormalization(
				storeStream.stream.name,
				storeStream.stream.namespace || "",
				checked,
			)
		}
	}

	return (
		<>
			<div
				className={clsx(
					!isSelected ? "font-normal text-text-disabled" : "font-medium",
					CARD_STYLE,
				)}
			>
				<div className="flex items-center justify-between">
					<div className="flex items-center gap-1">
						{isDirty && (
							<span className="mr-1 inline-block h-2 w-2 shrink-0 rounded-full bg-warning" />
						)}
						<label>Normalization</label>
					</div>
					<Switch
						checked={selectedStream?.normalization || false}
						onChange={handleNormalizationChange}
						disabled={!isSelected}
					/>
				</div>
			</div>
			{!isSelected && (
				<div className="ml-1 flex items-center gap-1 text-sm text-[#686868]">
					<InfoIcon className="size-4" />
					Select the stream to configure Normalization
				</div>
			)}
		</>
	)
}

export default NormalizationSection
