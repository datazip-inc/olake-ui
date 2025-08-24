import type { RegistryWidgetsType, RJSFSchema } from "@rjsf/utils"
import BooleanSwitchWidget from "./BooleanSwitchWidget"

export const widgets: RegistryWidgetsType<any, RJSFSchema, any> = {
	boolean: BooleanSwitchWidget,
}
