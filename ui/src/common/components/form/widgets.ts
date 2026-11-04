import type {
	RegistryFieldsType,
	RegistryWidgetsType,
	RJSFSchema,
} from "@rjsf/utils"

import BooleanSwitchWidget from "./BooleanSwitchWidget"
import { CDC_TOGGLE_FIELD } from "./cdc"
import CdcToggleField from "./CdcToggleField"
import CustomRadioWidget from "./CustomRadioWidget"

export const widgets: RegistryWidgetsType<any, RJSFSchema, any> = {
	boolean: BooleanSwitchWidget,
	radio: CustomRadioWidget,
}

// Custom fields for source forms; the driver uischema picks them with "ui:field"
export const sourceFields: RegistryFieldsType<any, RJSFSchema, any> = {
	[CDC_TOGGLE_FIELD]: CdcToggleField,
}
