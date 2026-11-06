import { Switch } from "antd"
import type { SwitchProps } from "antd"

/** antd Switch that turns green when on, the maintenance module's on/off colour. */
const StatusSwitch: React.FC<SwitchProps> = ({
	checked,
	className,
	...rest
}) => (
	<Switch
		checked={checked}
		className={[checked ? "!bg-olake-success" : "", className ?? ""]
			.join(" ")
			.trim()}
		{...rest}
	/>
)

export default StatusSwitch
