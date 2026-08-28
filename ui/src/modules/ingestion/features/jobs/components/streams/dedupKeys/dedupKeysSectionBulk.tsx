import DedupKeysSectionView from "./dedupKeysSectionView"
interface Props {
	options: string[]
	value: string[]
	isDirty: boolean
	onChange: (keys: string[]) => void
}

const DedupKeysSectionBulk = ({ options, value, isDirty, onChange }: Props) => (
	<DedupKeysSectionView
		options={options}
		value={value}
		isDirty={isDirty}
		disabled={false}
		onChange={onChange}
	/>
)

export default DedupKeysSectionBulk
