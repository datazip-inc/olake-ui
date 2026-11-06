import { Switch, Tooltip } from "antd"

type FeatureToggleRowProps = {
	label: string
	checked: boolean
	/** Shown on hover; explains why the switch cannot be flipped by hand. */
	tooltip?: string
}

/**
 * Bordered "feature on/off" row used at the top of each configure tab.
 * The switch only reflects saved state: it turns on when the tab's
 * configuration is saved, so it is read-only here.
 */
const FeatureToggleRow: React.FC<FeatureToggleRowProps> = ({
	label,
	checked,
	tooltip = "Turns on when this configuration is saved.",
}) => (
	<div className="flex h-12 items-center justify-between rounded-xl border border-olake-border px-4">
		<span className="text-base leading-6 text-olake-text">{label}</span>
		<Tooltip title={tooltip}>
			<span className="inline-flex cursor-default">
				<Switch
					checked={checked}
					className={checked ? "!bg-olake-success" : undefined}
					aria-label={label}
					aria-readonly
					style={{ pointerEvents: "none" }}
				/>
			</span>
		</Tooltip>
	</div>
)

export default FeatureToggleRow
