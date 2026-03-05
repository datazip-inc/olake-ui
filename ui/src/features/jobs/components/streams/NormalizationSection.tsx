import { useEffect, useState } from "react"
import clsx from "clsx"
import { Switch } from "antd"
import { InfoIcon } from "@phosphor-icons/react"

import { CARD_STYLE } from "../../constants"
import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
} from "../../stores/streamSelectionStore"

const NormalizationSection = () => {
	const store = useStreamSelectionStore()
	const stream = useStreamSelectionStore(selectActiveStreamData)
	const selectedStream = useStreamSelectionStore(selectActiveSelectedStream)
	const isSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, stream),
	)

	const [normalization, setNormalization] = useState<boolean>(
		selectedStream?.normalization || false,
	)

	// Re-sync normalization state when the active stream changes
	useEffect(() => {
		if (!selectedStream) return
		setNormalization(selectedStream.normalization || false)
	}, [stream])

	if (!stream || !selectedStream) return null

	const handleNormalizationChange = (checked: boolean) => {
		setNormalization(checked)
		store.updateNormalization(
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
						checked={normalization}
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
