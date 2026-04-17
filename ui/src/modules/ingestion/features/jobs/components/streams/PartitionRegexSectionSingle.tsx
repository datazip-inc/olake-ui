import PartitionRegexSectionView from "./PartitionRegexSectionView"
import { useStreamSelectionStore } from "../../stores"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
} from "../../stores"

interface PartitionRegexSectionSingleProps {
	destinationType?: string
}

const PartitionRegexSectionSingle = ({
	destinationType,
}: PartitionRegexSectionSingleProps) => {
	const updatePartitionRegex = useStreamSelectionStore(
		state => state.updatePartitionRegex,
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
		<PartitionRegexSectionView
			destinationType={destinationType}
			isSelected={storeIsSelected}
			activePartitionRegex={storeSelectedStream.partition_regex || ""}
			onChange={regex =>
				updatePartitionRegex(
					storeStream.stream.name,
					storeStream.stream.namespace || "",
					regex,
				)
			}
		/>
	)
}

export default PartitionRegexSectionSingle
