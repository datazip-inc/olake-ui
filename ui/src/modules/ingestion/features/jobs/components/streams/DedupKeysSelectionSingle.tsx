import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	selectAvailableUpdateTypes,
	selectIsStreamEnabled,
	useStreamSelectionStore,
} from "@/modules/ingestion/features/jobs/stores"
import {
	getDedupKeyOptions,
	getKafkaUpsertNotSupportedMessage,
} from "@/modules/ingestion/features/jobs/utils/streams"

import DedupKeysSelectionView from "./DedupKeysSelectionView"

interface DedupKeysSelectionSingleProps {
	sourceType?: string
	sourceVersion?: string
}

const DedupKeysSelectionSingle = ({
	sourceType,
	sourceVersion,
}: DedupKeysSelectionSingleProps) => {
	const availableUpdateTypes = useStreamSelectionStore(
		selectAvailableUpdateTypes,
	)
	const updateDedupKeys = useStreamSelectionStore(s => s.updateDedupKeys)
	const storeStream = useStreamSelectionStore(selectActiveStreamData)
	const storeSelectedStream = useStreamSelectionStore(
		selectActiveSelectedStream,
	)
	const isSelected = useStreamSelectionStore(s =>
		selectIsStreamEnabled(s, storeStream),
	)

	if (!storeStream || !storeSelectedStream) return null
	if (storeSelectedStream.append_mode) return null
	if (
		getKafkaUpsertNotSupportedMessage(
			sourceType,
			sourceVersion,
			availableUpdateTypes,
		)
	) {
		return null
	}

	return (
		<DedupKeysSelectionView
			options={getDedupKeyOptions(storeStream, storeSelectedStream)}
			value={storeSelectedStream?.dedup_keys ?? []}
			disabled={!isSelected}
			onChange={keys =>
				updateDedupKeys(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					keys,
				)
			}
		/>
	)
}

export default DedupKeysSelectionSingle
