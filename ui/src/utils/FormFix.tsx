import React, { useEffect, useRef, useState } from "react"
import DynamicSchemaForm from "../modules/common/components/DynamicSchemaForm"
import { RJSFSchema } from "@rjsf/utils"
import { Switch, Tooltip } from "antd"
import { Info, Eye, EyeSlash } from "@phosphor-icons/react"

/**
 * DirectFormField - A single form field that maintains its own state
 */
const DirectFormField = ({
	name,
	schema,
	value,
	onChange,
	required = false,
}: {
	name: string
	schema: any
	value: any
	onChange: (name: string, value: any) => void
	required?: boolean
}) => {
	// Keep internal state for the input value
	const [fieldValue, setFieldValue] = useState(value)
	const [showPassword, setShowPassword] = useState(false)
	const inputRef = useRef<HTMLInputElement>(null)

	// Update internal state if parent value changes and field is not focused
	useEffect(() => {
		if (document.activeElement !== inputRef.current) {
			setFieldValue(value)
		}
	}, [value])

	// Determine input type based on schema
	let inputType = "text"
	if (schema.type === "number" || schema.type === "integer") {
		inputType = "number"
	} else if (schema.type === "boolean") {
		inputType = "checkbox"
	} else if (schema.format === "password") {
		inputType = showPassword ? "text" : "password"
	}

	// Handle change events for standard inputs
	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		let newValue: any

		if (inputType === "checkbox") {
			newValue = e.target.checked
		} else if (inputType === "number") {
			newValue = e.target.value ? Number(e.target.value) : undefined
		} else {
			newValue = e.target.value
		}

		// Update local state immediately
		setFieldValue(newValue)

		// Notify parent component
		onChange(name, newValue)
	}

	// Handle change events for Switch component
	const handleSwitchChange = (checked: boolean) => {
		setFieldValue(checked)
		onChange(name, checked)
	}

	// Toggle password visibility
	const togglePasswordVisibility = () => {
		setShowPassword(!showPassword)
	}

	// Render the appropriate input based on type
	const renderInput = () => {
		if (schema.type === "boolean") {
			return (
				<div className="flex items-center justify-between">
					<Switch
						checked={!!fieldValue}
						onChange={handleSwitchChange}
						className={fieldValue ? "bg-blue-600" : "bg-gray-200"}
					/>
				</div>
			)
		}

		if (schema.format === "password") {
			return (
				<div className="relative">
					<input
						ref={inputRef}
						type={showPassword ? "text" : "password"}
						value={fieldValue || ""}
						onChange={handleChange}
						className="w-full rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
					/>
					<button
						type="button"
						onClick={togglePasswordVisibility}
						className="absolute right-2 top-1/2 -translate-y-1/2 transform cursor-pointer text-gray-500 hover:text-gray-700 focus:outline-none"
					>
						{showPassword ? (
							<EyeSlash className="size-4" />
						) : (
							<Eye className="size-4" />
						)}
					</button>
				</div>
			)
		}

		return (
			<input
				ref={inputRef}
				type={inputType}
				value={inputType !== "checkbox" ? fieldValue || "" : undefined}
				checked={inputType === "checkbox" ? fieldValue : undefined}
				onChange={handleChange}
				className="w-full rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
			/>
		)
	}

	return (
		<div className="mb-4">
			<div className="mb-1 flex items-center gap-1">
				<label className="text-sm font-medium text-gray-700">
					{schema.title || name}
					{required && <span className="text-red-500">*</span>}
				</label>
				{schema.description && (
					<Tooltip title={schema.description}>
						<Info className="size-4 cursor-help text-gray-400" />
					</Tooltip>
				)}
			</div>

			{renderInput()}
		</div>
	)
}

/**
 * DirectInputForm - A simple form component that directly renders inputs
 * without going through RJSF, ensuring focus is never lost during typing
 */
export const DirectInputForm = ({
	schema,
	formData,
	onChange,
}: {
	schema: RJSFSchema
	formData: any
	onChange: (data: any) => void
}) => {
	if (!schema || !schema.properties) {
		return <div>No schema properties provided</div>
	}

	// Handle changes from any field
	const handleFieldChange = (fieldName: string, fieldValue: any) => {
		const updatedData = {
			...formData,
			[fieldName]: fieldValue,
		}

		onChange(updatedData)
	}

	// Handle changes for nested object fields
	const handleNestedFieldChange = (parentField: string, fieldData: any) => {
		const updatedData = {
			...formData,
			[parentField]: fieldData,
		}
		onChange(updatedData)
	}

	// Generate form fields based on schema
	const renderFields = () => {
		// Safely get properties and handle undefined
		const properties = schema.properties || {}

		return Object.entries(properties).map(([name, fieldSchemaDefinition]) => {
			// Handle nested objects
			const fieldSchema = fieldSchemaDefinition as any

			if (
				typeof fieldSchema === "object" &&
				fieldSchema.type === "object" &&
				fieldSchema.properties
			) {
				// For nested objects, show just the title (if any) and render its fields directly
				return (
					<div
						key={name}
						className="col-span-2"
					>
						{fieldSchema.title && (
							<div className="mb-2 font-medium text-gray-700">
								{fieldSchema.title}
							</div>
						)}
						<DirectInputForm
							schema={fieldSchema as RJSFSchema}
							formData={formData?.[name] || {}}
							onChange={data => handleNestedFieldChange(name, data)}
						/>
					</div>
				)
			}

			// Handle regular fields
			return (
				<DirectFormField
					key={name}
					name={name}
					schema={fieldSchema}
					value={formData?.[name]}
					onChange={handleFieldChange}
					required={schema.required?.includes(name)}
				/>
			)
		})
	}

	return (
		<div className="direct-input-form">
			<form onSubmit={e => e.preventDefault()}>
				<div className="grid grid-cols-2 gap-x-8 gap-y-4">{renderFields()}</div>
			</form>
		</div>
	)
}

/**
 * FixedSchemaForm - A direct input form that maintains focus during typing
 * This is a replacement for DynamicSchemaForm when focus issues occur
 */
export const FixedSchemaForm: React.FC<
	React.ComponentProps<typeof DynamicSchemaForm>
> = props => {
	// Filter the formData to only include fields that exist in the schema
	const [filteredFormData, setFilteredFormData] = useState<any>({})

	useEffect(() => {
		if (props.schema && props.schema.properties && props.formData) {
			// Keep only the fields that exist in the schema
			const filteredData: any = {}

			// Get field names from the schema
			const schemaFields = Object.keys(props.schema.properties || {})

			// Filter the formData to only include fields in the schema
			Object.entries(props.formData).forEach(([key, value]) => {
				if (schemaFields.includes(key)) {
					filteredData[key] = value
				}
			})

			setFilteredFormData(filteredData)
		} else {
			setFilteredFormData(props.formData || {})
		}
	}, [props.schema, props.formData])

	// Use our direct implementation instead of the problematic RJSF library
	return (
		<DirectInputForm
			schema={props.schema}
			formData={filteredFormData}
			onChange={props.onChange}
		/>
	)
}

export default FixedSchemaForm
