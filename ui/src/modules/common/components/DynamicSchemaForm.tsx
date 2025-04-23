import { useState, useEffect } from "react"
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
	const [currentFormData, setCurrentFormData] = useState({ ...formData })

	useEffect(() => {
		setCurrentFormData(formData || {})
	}, [formData])

	// Update form data when it changes
	const handleFormChange = (data: any) => {
		// Handle both formats - direct value or { formData: value }
		const newFormData = data.formData || data
		setCurrentFormData(newFormData)
		onChange(newFormData)
	}

	// Handle form submission
	const handleSubmit = (data: any) => {
		const submitData = data.formData || data
		onSubmit?.(submitData)
	}

	// Prepare UI schema with some standard enhancements
	const finalUiSchema: UiSchema = {
		...(providedUiSchema || {}),
		"ui:className": "w-full",
		"ui:options": {
			...(providedUiSchema?.["ui:options"] || {}),
			className: "grid grid-cols-2 gap-x-12 gap-y-2",
		},
	}

	return (
		<div className={className}>
			<JsonSchemaForm
				schema={schema}
				uiSchema={finalUiSchema}
				formData={currentFormData}
				onChange={handleFormChange}
				onSubmit={handleSubmit}
				hideSubmit={hideSubmit}
			/>
		</div>
	)
}

export default DynamicSchemaForm
