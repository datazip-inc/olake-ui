import NormalizationSectionView from "./NormalizationSectionView"

interface NormalizationSectionBulkProps {
	isDirty?: boolean
	normalization: boolean
	onChange: (value: boolean) => void
}

const NormalizationSectionBulk = ({
	isDirty,
	normalization,
	onChange,
}: NormalizationSectionBulkProps) => (
	<NormalizationSectionView
		normalization={normalization}
		isSelected={true}
		isDirty={isDirty}
		onChange={onChange}
	/>
)

export default NormalizationSectionBulk
