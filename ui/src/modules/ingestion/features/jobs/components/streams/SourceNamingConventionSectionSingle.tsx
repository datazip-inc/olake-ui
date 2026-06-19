import SourceNamingConventionSectionView from "./SourceNamingConventionSectionView"
import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	selectIsStreamEnabled,
	useStreamSelectionStore,
} from "../../stores"

const SourceNamingConventionSectionSingle = () => {
	const storeStream = useStreamSelectionStore(selectActiveStreamData)
	const storeSelectedStream = useStreamSelectionStore(
		selectActiveSelectedStream,
	)
	const isSelected = useStreamSelectionStore(state =>
		selectIsStreamEnabled(state, storeStream),
	)
	const updateUseSourceColumnNames = useStreamSelectionStore(
		state => state.updateUseSourceColumnNames,
	)

	if (!storeStream || !storeSelectedStream) return null

	return (
		<SourceNamingConventionSectionView
			useSourceColumnNames={
				storeSelectedStream.use_source_column_names ?? false
			}
			isSelected={isSelected}
			onChange={checked =>
				updateUseSourceColumnNames(
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

export default SourceNamingConventionSectionSingle
