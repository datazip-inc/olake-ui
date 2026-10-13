import IngestionModeSectionView from "./IngestionModeSectionView"
import { IngestionMode } from "../../enums"

interface IngestionModeSectionBulkProps {
	sourceType?: string
	sourceVersion?: string
	destinationType?: string
	isDirty?: boolean
	bulkAppendMode?: boolean
	dedupKeyCount?: number
	onBulkIngestionModeChange?: (appendMode: boolean) => void
}

const IngestionModeSectionBulk = ({
	sourceType,
	sourceVersion,
	destinationType,
	isDirty,
	bulkAppendMode,
	dedupKeyCount,
	onBulkIngestionModeChange,
}: IngestionModeSectionBulkProps) => {
	return (
		<IngestionModeSectionView
			sourceType={sourceType}
			sourceVersion={sourceVersion}
			destinationType={destinationType}
			isSelected={true}
			isDirty={isDirty}
			appendMode={bulkAppendMode ?? true}
			dedupKeyCount={dedupKeyCount}
			onChange={ingestionMode =>
				onBulkIngestionModeChange?.(ingestionMode === IngestionMode.APPEND)
			}
		/>
	)
}

export default IngestionModeSectionBulk
