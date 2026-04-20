import NormalizationSectionView from "./NormalizationSectionView"
import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	selectIsStreamEnabled,
	useStreamSelectionStore,
} from "../../stores"

const NormalizationSectionSingle = () => {
	const storeStream = useStreamSelectionStore(selectActiveStreamData)
	const storeSelectedStream = useStreamSelectionStore(
		selectActiveSelectedStream,
	)
	const isSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, storeStream),
	)
	const updateNormalization = useStreamSelectionStore(
		state => state.updateNormalization,
	)

	if (!storeStream || !storeSelectedStream) return null

	return (
		<NormalizationSectionView
			normalization={storeSelectedStream.normalization}
			isSelected={isSelected}
			onChange={checked =>
				updateNormalization(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					checked,
				)
			}
		/>
	)
}

export default NormalizationSectionSingle
