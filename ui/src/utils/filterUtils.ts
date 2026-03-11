// Casts a single filter condition's value to the appropriate native type

import { FilterConfig, FilterConfigCondition } from "../types"
import { FILTER_REGEX } from "./constants"

// (number, boolean, object, array) based on the column's type schema.
export const castFilterConditionValue = (
	cond: FilterConfigCondition,
	columnSchema?: { type: string[] },
): FilterConfigCondition => {
	if (cond.value === null || cond.value === "<null>") {
		return { ...cond, value: null }
	}

	if (!columnSchema) return cond

	// Find primary non-null type
	const nonNullTypes = columnSchema.type.filter(t => t !== "null")

	if (nonNullTypes.length === 0) return cond

	const type = nonNullTypes[0] // take the primary type for casting
	let castValue: any = String(cond.value).trim()

	switch (type) {
		case "integer_small":
		case "integer":
			castValue = castValue === "" ? null : parseInt(castValue, 10)
			break
		case "number_small":
		case "number":
			castValue = castValue === "" ? null : parseFloat(castValue)
			break
		case "boolean":
			castValue = castValue.toLowerCase() === "true"
			break
		// arrays and objects are sent as string
	}

	return { ...cond, value: castValue }
}

// validates filter expression
export const validateFilter = (filter: string): boolean => {
	if (!filter.trim()) return false
	return FILTER_REGEX.test(filter.trim())
}

// Validates if value is compatible with any given DataType.
// Explicitly handles "null" values by checking if the column schema allows it.
// Converts other values to string internally to handle native types safely.
export const isValueValidForTypes = (value: any, type: string[]): boolean => {
	if (type.length === 0) return false

	const raw = value === null ? "" : String(value)

	return type.some(t => {
		switch (t) {
			case "null":
				return value === null || raw === "<null>"

			case "integer_small":
			case "integer":
				return /^-?\d+$/.test(raw) && Number.isInteger(Number(raw))

			case "number_small":
			case "number":
				return !isNaN(Number(raw)) && raw !== ""

			case "boolean":
				return raw.toLowerCase() === "true" || raw.toLowerCase() === "false"

			case "array": {
				try {
					return Array.isArray(JSON.parse(raw))
				} catch {
					return false
				}
			}

			case "object": {
				try {
					const parsed = JSON.parse(raw)
					return (
						parsed !== null &&
						typeof parsed === "object" &&
						!Array.isArray(parsed)
					)
				} catch {
					return false
				}
			}

			// "string", "unknown", or any unrecognised type
			default:
				return true
		}
	})
}

// Validates a structured filter_config object.
// Returns null if valid, or a descriptive error string.
export const validateFilterConfig = (
	filterConfig: FilterConfig,
	streamName: string,
	namespace: string,
	typeSchemaProperties?: Record<string, { type: string[] }>,
): string | null => {
	const streamPrefix = `[${namespace}.${streamName}]`

	if (!filterConfig.conditions || filterConfig.conditions.length === 0) {
		return `${streamPrefix} Filter conditions cannot be empty`
	}

	for (const cond of filterConfig.conditions) {
		if (typeof cond.column !== "string" || cond.column.trim() === "") {
			return `${streamPrefix} Filter condition is missing a column`
		}
		// Values can be null if the schema allows it, but they cannot be missing/undefined entirely
		if (cond.value === undefined) {
			return `${streamPrefix} Filter condition for "${cond.column}" is missing a value`
		}

		if (!typeSchemaProperties) continue

		const columnSchema = typeSchemaProperties[cond.column]
		if (!columnSchema) continue

		if (isValueValidForTypes(cond.value, columnSchema.type)) continue

		const nonNullTypes = columnSchema.type.filter(t => t !== "null")
		const expectedTypes = (nonNullTypes.length ? nonNullTypes : ["null"]).join(
			" | ",
		)

		return `${streamPrefix} Invalid value "${cond.value}" for column "${cond.column}" — expected type: ${expectedTypes}`
	}

	return null
}
