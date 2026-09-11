import UpsertTypeSectionView from "./UpsertTypeSectionView"
import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	selectIsStreamEnabled,
	useStreamSelectionStore,
} from "../../stores"

interface UpsertTypeSectionSingleProps {
	sourceType?: string
	destinationType?: string
}

const UpsertTypeSectionSingle = ({
	sourceType,
	destinationType,
}: UpsertTypeSectionSingleProps) => {
	const updateUpsertType = useStreamSelectionStore(
		state => state.updateUpsertType,
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
		<UpsertTypeSectionView
			sourceType={sourceType}
			destinationType={destinationType}
			isSelected={storeIsSelected}
			appendMode={!!storeSelectedStream.append_mode}
			upsertType={storeSelectedStream.update_type}
			onChange={upsertType =>
				updateUpsertType(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					upsertType,
				)
			}
		/>
	)
}

export default UpsertTypeSectionSingle
