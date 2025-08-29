import { RJSFSchema } from "@rjsf/utils"

const isNestedObjectSchema = (schema: any): boolean => {
	return (
		schema &&
		typeof schema === "object" &&
		schema.type === "object" &&
		schema.properties
	)
}

export const validateFormData = (
	formData: any,
	schema: RJSFSchema,
): Record<string, string> => {
	const errors: Record<string, any> = {}

	if (!schema?.properties) return errors

	Object.entries(schema.properties).forEach(
		([key, propValue]: [string, any]) => {
			const isRequired = schema.required?.includes(key)
			const value = formData?.[key]
			const hasDefault = propValue?.default !== undefined
			const isEmpty =
				value === undefined ||
				value === null ||
				value === "" ||
				(Array.isArray(value) && value.length === 0)
			const wasIntentionallyCleared = hasDefault && isEmpty && key in formData

			if (isRequired && isEmpty && (wasIntentionallyCleared || !hasDefault)) {
				errors[key] = `${propValue.title || key} is required`
			}

			if (isNestedObjectSchema(propValue) && formData?.[key]) {
				const nestedErrors = validateFormData(formData[key], propValue)
				if (Object.keys(nestedErrors).length > 0) {
					errors[key] = nestedErrors
				}
			}
		},
	)

	return errors
}
