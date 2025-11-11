/**
 * BooleanSwitchWidget is a component that renders a boolean switch this overrides the default boolean field template of rjsf
 */
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
