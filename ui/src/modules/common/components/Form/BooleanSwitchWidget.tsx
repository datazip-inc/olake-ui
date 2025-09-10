import { Switch } from "antd"
import { WidgetProps } from "@rjsf/utils"

const BooleanSwitchWidget = ({
	value,
	onChange,
	id,
	disabled = false,
}: WidgetProps) => (
	<Switch
		id={id}
		checked={value}
		onChange={checked => onChange(checked)}
		disabled={disabled}
	/>
)

export default BooleanSwitchWidget
