import { InfoIcon } from "@phosphor-icons/react"
import { Switch } from "antd"
import clsx from "clsx"

import { CARD_STYLE } from "../../constants"
import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
} from "../../stores"

const NormalizationSection = () => {
	const updateNormalization = useStreamSelectionStore(
		state => state.updateNormalization,
	)
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)
	const isSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, stream),
	)

	if (!stream || !selectedStream) return null

	const handleNormalizationChange = (checked: boolean) => {
		updateNormalization(
			stream.stream.name,
			stream.stream.namespace || "",
			checked,
		)
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
					<label>Normalization</label>
					<Switch
						checked={selectedStream.normalization || false}
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
