import DedupKeysSelectionView from "./DedupKeysSelectionView"
interface Props {
	options: string[]
	value: string[]
	isDirty: boolean
	onChange: (keys: string[]) => void
}

const DedupKeysSelectionBulk = ({
	options,
	value,
	isDirty,
	onChange,
}: Props) => (
	<DedupKeysSelectionView
		options={options}
		value={value}
		isDirty={isDirty}
		disabled={false}
		onChange={onChange}
	/>
)

export default DedupKeysSelectionBulk
