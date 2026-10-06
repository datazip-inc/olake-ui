/**
 * CdcToggleField renders a source's update_method (the Standalone / CDC oneOf) as the
 * "Enable CDC" toggle. The driver uischema opts in with "ui:field": "CdcToggle".
 * ON saves {type: "CDC", ...cdc settings}; OFF saves {type: "Standalone"}.
 */
import type { FieldProps, RJSFSchema, UiSchema } from "@rjsf/utils"
import { useRef } from "react"

import {
	getUpdateMethodOptions,
	isCdcEnabled,
	UPDATE_METHOD_CDC,
	UPDATE_METHOD_STANDALONE,
} from "./cdc"
import CdcCard from "./CdcCard"

// Keys that only apply to this field. The rest of the update_method uischema (grid, widgets,
// hidden "type", title/description off) is used for the CDC option's own fields.
const FIELD_ONLY_KEYS = new Set([
	"ui:field",
	"ui:fieldReplacesAnyOrOneOf",
	"ui:widget",
])

const cdcSettingsUiSchema = (uiSchema?: UiSchema): UiSchema =>
	Object.fromEntries(
		Object.entries(uiSchema ?? {}).filter(([key]) => !FIELD_ONLY_KEYS.has(key)),
	)

const CdcToggleField = (props: FieldProps) => {
	const {
		schema,
		uiSchema,
		formData,
		onChange,
		registry,
		idSchema,
		errorSchema,
		disabled,
		readonly,
		name,
		onBlur,
		onFocus,
	} = props
	// CDC settings typed before turning CDC off, restored when it is turned back on
	const lastCdcValue = useRef<Record<string, unknown> | null>(null)

	const { cdcOption } = getUpdateMethodOptions(schema)
	const enabled = isCdcEnabled(formData, schema)
	const hasSettings = Object.keys(cdcOption?.properties ?? {}).some(
		key => key !== "type",
	)

	const handleToggle = (on: boolean) => {
		if (on) {
			onChange({ ...(lastCdcValue.current ?? {}), type: UPDATE_METHOD_CDC })
			return
		}
		if (enabled) lastCdcValue.current = formData ?? null
		onChange({ type: UPDATE_METHOD_STANDALONE })
	}

	const { SchemaField } = registry.fields

	return (
		<CdcCard
			state={enabled ? "on" : "off"}
			description={
				typeof schema.description === "string" ? schema.description : undefined
			}
			onToggle={handleToggle}
			disabled={disabled || readonly}
		>
			{cdcOption && hasSettings && (
				<SchemaField
					name={name}
					schema={cdcOption as RJSFSchema}
					uiSchema={cdcSettingsUiSchema(uiSchema)}
					formData={{ ...(formData ?? {}), type: UPDATE_METHOD_CDC }}
					onChange={value => onChange(value)}
					onBlur={onBlur}
					onFocus={onFocus}
					idSchema={idSchema}
					errorSchema={errorSchema}
					registry={registry}
					disabled={disabled}
					readonly={readonly}
				/>
			)}
		</CdcCard>
	)
}

export default CdcToggleField
