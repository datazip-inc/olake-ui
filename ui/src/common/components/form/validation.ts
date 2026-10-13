import type { RJSFValidationError, ValidatorType } from "@rjsf/utils"
import validator from "@rjsf/validator-ajv8"

import { FORM_ERROR_MESSAGES, PATTERN_ERROR_MESSAGES } from "@/common/constants"
import { toSentenceCase } from "@/common/utils"
import { notificationService } from "@/core/notifications"

// Form data is trimmed on every change, so a whitespace-only value is stored as
// "" and still satisfies `required`. Blank strings are dropped before validation
// so they are reported as missing.
const omitBlankStrings = (data: unknown): unknown => {
	if (Array.isArray(data)) return data.map(omitBlankStrings)
	if (data && typeof data === "object") {
		return Object.fromEntries(
			Object.entries(data)
				.filter(
					([, value]) => !(typeof value === "string" && value.trim() === ""),
				)
				.map(([key, value]) => [key, omitBlankStrings(value)]),
		)
	}
	return data
}

// AJV validator for spec-driven forms. Only full-form validation sees the
// blank-free data; isValid (used to pick oneOf/dependency branches) is untouched.
export const formValidator: ValidatorType = {
	validateFormData: (
		formData,
		schema,
		customValidate,
		transformErrors,
		uiSchema,
	) =>
		validator.validateFormData(
			omitBlankStrings(formData),
			schema,
			customValidate,
			transformErrors,
			uiSchema,
		),
	toErrorList: (errorSchema, fieldPath) =>
		validator.toErrorList(errorSchema, fieldPath),
	isValid: (schema, formData, rootSchema) =>
		validator.isValid(schema, formData, rootSchema),
	rawValidation: (schema, formData) =>
		validator.rawValidation(schema, formData),
}

// rjsf ids fields as root_<path>: ".writer.catalog_name" -> "root_writer_catalog_name"
const getFieldId = (error: RJSFValidationError) =>
	`root${(error.property ?? "").replace(/\./g, "_")}`

const getFieldElement = (error: RJSFValidationError) =>
	document.getElementById(getFieldId(error))

// Reads the label CustomFieldTemplate rendered for the field, so the toast names
// it as the user sees it (catalog types override spec titles, e.g. "S3 Path" is
// shown as "Lakekeeper Warehouse"). Falls back to the property name.
const getFieldLabel = (error: RJSFValidationError) => {
	const label = document.querySelector(
		`label[for="${CSS.escape(getFieldId(error))}"]`,
	)
	const labelText = Array.from(label?.childNodes ?? [])
		.filter(node => node.nodeType === Node.TEXT_NODE)
		.map(node => node.textContent)
		.join("")
		.trim()
	if (labelText) return labelText

	const property = error.property?.split(".").pop() ?? ""
	return toSentenceCase(property.replace(/_/g, " ")) || "This field"
}

const getFormErrorMessage = (error: RJSFValidationError) => {
	const { name = "", params } = error
	const key = name === "type" ? `type.${params?.type}` : name

	const template =
		(name === "pattern" && PATTERN_ERROR_MESSAGES[params?.pattern]) ||
		FORM_ERROR_MESSAGES[key]
	const label = getFieldLabel(error)
	if (!template) return `${label}: ${error.message}`

	return template
		.replace("{label}", label)
		.replace("{limit}", String(params?.limit))
}

const isAbove = (a: Element | null, b: Element | null) =>
	!!a &&
	(!b || !!(a.compareDocumentPosition(b) & Node.DOCUMENT_POSITION_FOLLOWING))

// rjsf onError handler: every invalid field is already highlighted through
// rawErrors, so only the topmost error on screen is shown as a toast.
export const notifyFirstFormError = (errors: RJSFValidationError[]) => {
	if (errors.length === 0) return

	const firstError = errors.reduce((top, error) =>
		isAbove(getFieldElement(error), getFieldElement(top)) ? error : top,
	)
	notificationService.error(getFormErrorMessage(firstError))
}
