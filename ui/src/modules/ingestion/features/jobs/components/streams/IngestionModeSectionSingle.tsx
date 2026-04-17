import IngestionModeSectionView from "./IngestionModeSectionView"
import { IngestionMode } from "../../enums"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
	useStreamSelectionStore,
} from "../../stores"

interface IngestionModeSectionSingleProps {
	sourceType?: string
	destinationType?: string
}

const IngestionModeSectionSingle = ({
	sourceType,
	destinationType,
}: IngestionModeSectionSingleProps) => {
	const updateIngestionMode = useStreamSelectionStore(
		state => state.updateIngestionMode,
	)
	const storeStream = useStreamSelectionStore(selectActiveStreamData)
	const storeSelectedStream = useStreamSelectionStore(
		selectActiveSelectedStream,
	)
	const storeIsSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, storeStream),
	)

	if (!storeStream || !storeSelectedStream) return null

	return (
		<IngestionModeSectionView
			sourceType={sourceType}
			destinationType={destinationType}
			isSelected={storeIsSelected}
			appendMode={!!storeSelectedStream.append_mode}
			onChange={ingestionMode =>
				updateIngestionMode(
					storeStream.stream.name,
					storeStream.stream.namespace || "",
					ingestionMode === IngestionMode.APPEND,
				)
			}
		/>
	)
}

export default IngestionModeSectionSingle
