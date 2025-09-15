import type { RegistryWidgetsType, RJSFSchema } from "@rjsf/utils"
import BooleanSwitchWidget from "./BooleanSwitchWidget"
import CustomRadioWidget from "./CustomRadioWidget"

export const widgets: RegistryWidgetsType<any, RJSFSchema, any> = {
	boolean: BooleanSwitchWidget,
	radio: CustomRadioWidget,
}
