import { useEffect } from "react"

import { SyncMode } from "@/modules/ingestion/common/types"

import SyncModeSectionView from "./SyncModeSectionView"
import {
	selectActiveSelectedStream,
	selectActiveStreamData,
	useStreamSelectionStore,
} from "../../stores"

const SyncModeSectionSingle = () => {
	const updateSyncMode = useStreamSelectionStore(state => state.updateSyncMode)
	const storeStream = useStreamSelectionStore(selectActiveStreamData)
	const storeSelectedStream = useStreamSelectionStore(
		selectActiveSelectedStream,
	)

	// Auto-select first available cursor field if stream is incremental and has none set
	useEffect(() => {
		if (!storeStream || !storeSelectedStream) return

		const activeCursorField = storeStream.stream.cursor_field
		const initialApiSyncMode = storeStream.stream.sync_mode

		if (initialApiSyncMode === "incremental" && !activeCursorField) {
			const availableCursorFields =
				storeStream.stream.available_cursor_fields || []
			const pk = storeStream.stream.source_defined_primary_key || []
			const sorted = [...availableCursorFields].sort((a, b) => {
				if (pk.includes(a) && !pk.includes(b)) return -1
				if (!pk.includes(a) && pk.includes(b)) return 1
				return a.localeCompare(b)
			})
			const cursor = sorted[0]
			if (cursor) {
				updateSyncMode(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					SyncMode.INCREMENTAL,
					cursor,
				)
			}
		}
	}, [storeStream?.stream.name, storeStream?.stream.namespace])

	if (!storeStream || !storeSelectedStream) return null

	return (
		<SyncModeSectionView
			stream={storeStream}
			syncMode={storeStream.stream.sync_mode}
			cursorField={storeStream.stream.cursor_field}
			onChange={(mode, cf) =>
				updateSyncMode(
					{
						streamName: storeStream.stream.name,
						namespace: storeStream.stream.namespace || "",
					},
					mode,
					cf,
				)
			}
		/>
	)
}

export default SyncModeSectionSingle
