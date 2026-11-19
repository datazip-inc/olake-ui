import { Switch } from "antd"
import type { SwitchProps } from "antd"
import clsx from "clsx"

// antd Switch that turns green when on, the maintenance module's on/off colour.
const StatusSwitch: React.FC<SwitchProps> = ({
	checked,
	className,
	...rest
}) => (
	<Switch
		checked={checked}
		className={clsx(checked && "!bg-olake-success", className)}
		{...rest}
	/>
)

export default StatusSwitch
