import PartitionRegexSectionView from "./PartitionRegexSectionView"

interface PartitionRegexSectionBulkProps {
	destinationType?: string
	isDirty?: boolean
	bulkPartitionRegex?: string
	onBulkPartitionRegexChange?: (regex: string) => void
}

const PartitionRegexSectionBulk = ({
	destinationType,
	isDirty,
	bulkPartitionRegex,
	onBulkPartitionRegexChange,
}: PartitionRegexSectionBulkProps) => {
	return (
		<PartitionRegexSectionView
			destinationType={destinationType}
			isSelected={true}
			isDirty={isDirty}
			activePartitionRegex={bulkPartitionRegex ?? ""}
			onChange={onBulkPartitionRegexChange ?? (() => {})}
		/>
	)
}

export default PartitionRegexSectionBulk
