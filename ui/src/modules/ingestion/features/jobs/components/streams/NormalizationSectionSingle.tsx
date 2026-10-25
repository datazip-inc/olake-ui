import NormalizationSectionView from "./NormalizationSectionView"
import { STREAM_DEFAULTS } from "../../constants"
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
			normalization={
				storeSelectedStream.normalization ??
				storeStream.stream.default_stream_properties?.normalization ??
				STREAM_DEFAULTS.normalization
			}
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
