import { Switch } from "antd"

type FeatureToggleRowProps = {
	label: string
	checked: boolean
	onChange: (checked: boolean) => void
}

/** Bordered "feature on/off" row at the top of each configure tab. Saved together with the rest of the form. */
const FeatureToggleRow: React.FC<FeatureToggleRowProps> = ({
	label,
	checked,
	onChange,
}) => (
	<div className="flex h-12 items-center justify-between rounded-xl border border-olake-border px-4">
		<span className="text-base leading-6 text-olake-text">{label}</span>
		<Switch
			checked={checked}
			onChange={onChange}
			className={checked ? "!bg-olake-success" : undefined}
			aria-label={label}
		/>
	</div>
)

export default FeatureToggleRow
