import React, { useEffect, useRef, useMemo } from "react"
import JsonSchemaForm from "./JsonSchemaForm"
import { RJSFSchema, UiSchema } from "@rjsf/utils"

interface DynamicSchemaFormProps {
	schema: RJSFSchema // The main schema to use
	formData: any // Form data values
	onChange: (data: any) => void // Callback when form changes
	onSubmit?: (data: any) => void // Optional submit handler
	uiSchema?: UiSchema // Optional UI schema
	hideSubmit?: boolean // Whether to hide the submit button
	className?: string // Optional className for the form container
}

/**
 * A dynamic form component that renders forms based on schema definitions.
 * This component handles form state management and provides a consistent UI.
 */
const DynamicSchemaForm: React.FC<DynamicSchemaFormProps> = ({
	schema,
	formData,
	onChange,
	onSubmit,
	uiSchema: providedUiSchema,
	hideSubmit = true,
	className = "",
}) => {
	// Create a stable form ID
	const formId = useMemo(
		() => `form-${Math.random().toString(36).substring(2, 9)}`,
		[],
	)

	// Use a reference to track if we're in the middle of an update
	const updatingRef = useRef(false)

	// Store the latest onChange handler to ensure we always call the most recent one
	const onChangeRef = useRef(onChange)
	useEffect(() => {
		onChangeRef.current = onChange
	}, [onChange])

	// The key handler that ensures proper data flow
	const handleFormChange = (data: any) => {
		if (updatingRef.current) return

		try {
			updatingRef.current = true

			// Get the form data, either from the event object or directly
			const newData = data?.formData || data

			// Call the parent's onChange handler
			onChangeRef.current(newData)
		} finally {
			// Always reset the updating flag, even if there's an error
			setTimeout(() => {
				updatingRef.current = false
			}, 0)
		}
	}

	// Handle form submission
	const handleSubmit = (data: any) => {
		if (!onSubmit) return
		const submitData = data?.formData || data
		onSubmit(submitData)
	}

	// Prepare UI schema with some standard enhancements
	const finalUiSchema = useMemo(
		() => ({
			...(providedUiSchema || {}),
			"ui:className": "w-full",
			"ui:options": {
				...(providedUiSchema?.["ui:options"] || {}),
				className: "grid grid-cols-2 gap-x-12 gap-y-2",
			},
		}),
		[providedUiSchema],
	)

	return (
		<div
			className={className}
			id={formId}
		>
			<JsonSchemaForm
				key={formId}
				schema={schema}
				uiSchema={finalUiSchema}
				formData={formData}
				onChange={handleFormChange}
				onSubmit={handleSubmit}
				hideSubmit={hideSubmit}
			/>
		</div>
	)
}

export default DynamicSchemaForm
