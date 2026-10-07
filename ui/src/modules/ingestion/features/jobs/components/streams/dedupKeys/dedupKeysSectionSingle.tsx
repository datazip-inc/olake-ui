import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	selectIsStreamEnabled,
	useJobConfigurationStore,
	useStreamSelectionStore,
} from "@/modules/ingestion/features/jobs/stores"
import {
	getDedupKeyOptions,
	getKafkaUpsertNotSupportedMessage,
} from "@/modules/ingestion/features/jobs/utils/streams"

import DedupKeysSectionView from "./dedupKeysSectionView"

interface DedupKeysSectionSingleProps {
	sourceType?: string
	sourceVersion?: string
}

const DedupKeysSectionSingle = ({
	sourceType,
	sourceVersion,
}: DedupKeysSectionSingleProps) => {
	const targetQueryEngines = useJobConfigurationStore(
		s => s.advancedSettings?.target_query_engines,
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
			targetQueryEngines,
		)
	) {
		return null
	}

	return (
		<DedupKeysSectionView
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

export default DedupKeysSectionSingle
