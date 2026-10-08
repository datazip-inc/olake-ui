import { isKafkaSource } from "@/modules/ingestion/common/utils"

import IngestionModeSectionView from "./IngestionModeSectionView"
import { IngestionMode } from "../../enums"
import {
	selectActiveStreamData,
	selectActiveSelectedStream,
	selectIsStreamEnabled,
	useStreamSelectionStore,
} from "../../stores"
import { getDedupKeyOptions } from "../../utils/streams"

interface IngestionModeSectionSingleProps {
	sourceType?: string
	sourceVersion?: string
	destinationType?: string
}

const IngestionModeSectionSingle = ({
	sourceType,
	sourceVersion,
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

	const isKafka = isKafkaSource(sourceType)

	return (
		<IngestionModeSectionView
			sourceType={sourceType}
			sourceVersion={sourceVersion}
			destinationType={destinationType}
			isSelected={storeIsSelected}
			appendMode={!!storeSelectedStream.append_mode}
			dedupKeyCount={
				isKafka
					? getDedupKeyOptions(storeStream, storeSelectedStream).length
					: undefined
			}
			onChange={ingestionMode =>
				updateIngestionMode(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					ingestionMode === IngestionMode.APPEND,
				)
			}
		/>
	)
}

export default IngestionModeSectionSingle
