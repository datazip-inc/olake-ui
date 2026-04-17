import IngestionModeSectionView from "./IngestionModeSectionView"
import { IngestionMode } from "../../enums"

interface IngestionModeSectionBulkProps {
	sourceType?: string
	destinationType?: string
	isDirty?: boolean
	bulkAppendMode?: boolean
	onBulkIngestionModeChange?: (appendMode: boolean) => void
}

const IngestionModeSectionBulk = ({
	sourceType,
	destinationType,
	isDirty,
	bulkAppendMode,
	onBulkIngestionModeChange,
}: IngestionModeSectionBulkProps) => {
	return (
		<IngestionModeSectionView
			sourceType={sourceType}
			destinationType={destinationType}
			isSelected={true}
			isDirty={isDirty}
			appendMode={bulkAppendMode ?? false}
			onChange={ingestionMode =>
				onBulkIngestionModeChange?.(ingestionMode === IngestionMode.APPEND)
			}
		/>
	)
}

export default IngestionModeSectionBulk
