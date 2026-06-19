import SourceNamingConventionSectionView from "./SourceNamingConventionSectionView"

interface SourceNamingConventionSectionBulkProps {
	isDirty?: boolean
	useSourceColumnNames: boolean
	onChange?: (value: boolean) => void
}

const SourceNamingConventionSectionBulk = ({
	isDirty,
	useSourceColumnNames,
	onChange,
}: SourceNamingConventionSectionBulkProps) => (
	<SourceNamingConventionSectionView
		useSourceColumnNames={useSourceColumnNames}
		isSelected={true}
		isDirty={isDirty}
		onChange={onChange}
	/>
)

export default SourceNamingConventionSectionBulk
