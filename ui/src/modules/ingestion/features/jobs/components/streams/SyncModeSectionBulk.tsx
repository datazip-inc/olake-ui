import { SyncMode, StreamData } from "@/modules/ingestion/common/types"

import SyncModeSectionView from "./SyncModeSectionView"

interface SyncModeSectionBulkProps {
	isDirty?: boolean
	bulkStream: StreamData
	bulkSyncMode: string
	bulkCursorField: string | undefined
	onBulkSyncModeChange: (syncMode: string, cursorField?: string) => void
}

const SyncModeSectionBulk = ({
	isDirty,
	bulkStream,
	bulkSyncMode,
	bulkCursorField,
	onBulkSyncModeChange,
}: SyncModeSectionBulkProps) => (
	<SyncModeSectionView
		stream={bulkStream}
		syncMode={bulkSyncMode}
		cursorField={bulkCursorField}
		isDirty={isDirty}
		isBulkMode
		onChange={onBulkSyncModeChange as (mode: SyncMode, cf?: string) => void}
	/>
)

export default SyncModeSectionBulk
