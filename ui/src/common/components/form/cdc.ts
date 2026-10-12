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

/** What a source's update_method schema offers, read once from its oneOf options. */
export interface UpdateMethod {
	cdcOption?: RJSFSchema
	cdcFields: string[]
	defaultsToCdc: boolean
}

const isCdcOption = (option?: RJSFSchema) =>
	(option?.properties?.type as RJSFSchema | undefined)?.const ===
	UPDATE_METHOD_CDC

export const parseUpdateMethod = (schema: RJSFSchema): UpdateMethod => {
	const options = (schema.oneOf ?? []) as RJSFSchema[]
	const cdcOption = options.find(isCdcOption)
	return {
		cdcOption,
		cdcFields: Object.keys(cdcOption?.properties ?? {}).filter(
			key => key !== "type",
		),
		defaultsToCdc: isCdcOption(options[0]),
	}
}

/** True when the source schema lets the user choose CDC (has an update_method CDC option). */
export const hasCdcOption = (schema?: RJSFSchema | null): boolean => {
	const updateMethod = schema?.properties?.update_method as
		| RJSFSchema
		| undefined
	return !!updateMethod && !!parseUpdateMethod(updateMethod).cdcOption
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
	const { cdcFields, defaultsToCdc } = parseUpdateMethod(updateMethodSchema)
	if (!value || Object.keys(value).length === 0) return defaultsToCdc
	if (value.type === UPDATE_METHOD_CDC) return true
	if (value.type === UPDATE_METHOD_STANDALONE) return false
	return cdcFields.some(key => key in value)
}

export const withCdcToggle = <T extends Record<string, any> | null>(
	uiSchema: T,
): T => {
	const updateMethod = uiSchema?.update_method
	if (updateMethod?.["ui:field"] !== CDC_TOGGLE_FIELD) return uiSchema
	return {
		...uiSchema,
		update_method: { ...updateMethod, "ui:fieldReplacesAnyOrOneOf": true },
	}
}
