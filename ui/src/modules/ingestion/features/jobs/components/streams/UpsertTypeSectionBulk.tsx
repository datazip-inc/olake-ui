import { UpsertType } from "@/modules/ingestion/common/types"

import UpsertTypeSectionView from "./UpsertTypeSectionView"

interface UpsertTypeSectionBulkProps {
	sourceType?: string
	destinationType?: string
	isDirty?: boolean
	bulkAppendMode?: boolean
	bulkUpsertType?: UpsertType
	onBulkUpsertTypeChange?: (upsertType: UpsertType) => void
}

const UpsertTypeSectionBulk = ({
	sourceType,
	destinationType,
	isDirty,
	bulkAppendMode,
	bulkUpsertType,
	onBulkUpsertTypeChange,
}: UpsertTypeSectionBulkProps) => (
	<UpsertTypeSectionView
		sourceType={sourceType}
		destinationType={destinationType}
		isSelected={true}
		isDirty={isDirty}
		appendMode={bulkAppendMode ?? false}
		upsertType={bulkUpsertType}
		onChange={upsertType => onBulkUpsertTypeChange?.(upsertType)}
	/>
)

export default UpsertTypeSectionBulk
