import type { RJSFSchema } from "@rjsf/utils"

// Name the driver uischema uses for the update_method field ("ui:field": "CdcToggle")
export const CDC_TOGGLE_FIELD = "CdcToggle"

export const UPDATE_METHOD_CDC = "CDC"
export const UPDATE_METHOD_STANDALONE = "Standalone"

// How a source's CDC section renders, from the driver uischema's "ui:sections"
export type CdcSectionMode = "toggle" | "not_supported"

export interface UiSection {
	id: string
	title: string
	cdc?: CdcSectionMode
	fields?: string[]
}

const optionType = (option: RJSFSchema): unknown =>
	(option.properties?.type as RJSFSchema | undefined)?.const

/** The update_method oneOf options, found by their `type` const. */
export const getUpdateMethodOptions = (schema: RJSFSchema) => {
	const options = (schema.oneOf ?? []) as RJSFSchema[]
	return {
		cdcOption: options.find(o => optionType(o) === UPDATE_METHOD_CDC),
		firstIsCdc:
			options.length > 0 && optionType(options[0]) === UPDATE_METHOD_CDC,
	}
}

/** True when the source schema lets the user choose CDC (has an update_method CDC option). */
export const hasCdcOption = (schema?: RJSFSchema | null): boolean => {
	const updateMethod = schema?.properties?.update_method as
		| RJSFSchema
		| undefined
	return !!updateMethod && !!getUpdateMethodOptions(updateMethod).cdcOption
}

/**
 * Whether a saved update_method value means CDC is on. Follows how the drivers read it:
 * - `type` decides when present;
 * - older MySQL/Postgres configs have no `type` but carry CDC-only keys (replication_slot,
 *   initial_wait_time), which the drivers treat as CDC;
 * - no value at all: the driver's own default, which is the first oneOf option
 *   (MySQL/Postgres list Standalone first, MongoDB/MSSQL list CDC first).
 */
export const isCdcEnabled = (
	value: Record<string, unknown> | undefined,
	updateMethodSchema: RJSFSchema,
): boolean => {
	const { cdcOption, firstIsCdc } = getUpdateMethodOptions(updateMethodSchema)
	if (!value || Object.keys(value).length === 0) return firstIsCdc
	if (value.type === UPDATE_METHOD_CDC) return true
	if (value.type === UPDATE_METHOD_STANDALONE) return false
	const cdcKeys = Object.keys(cdcOption?.properties ?? {}).filter(
		key => key !== "type",
	)
	return cdcKeys.some(key => key in value)
}
