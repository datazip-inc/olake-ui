import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	selectIsStreamEnabled,
	useStreamSelectionStore,
} from "@/modules/ingestion/features/jobs/stores"
import { getDedupKeyOptions } from "@/modules/ingestion/features/jobs/utils/streams"

import DedupKeysSectionView from "./dedupKeysSectionView"

const DedupKeysSectionSingle = () => {
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

	return (
		<DedupKeysSectionView
			options={getDedupKeyOptions(storeStream)}
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
