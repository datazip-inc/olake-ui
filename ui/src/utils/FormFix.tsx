import React, { useEffect, useRef, useState } from "react"
import DynamicSchemaForm from "../modules/common/components/DynamicSchemaForm"
import { RJSFSchema } from "@rjsf/utils"
import { Switch, Tooltip, Select } from "antd"
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
	uiSchema,
}: {
	name: string
	schema: any
	value: any
	onChange: (name: string, value: any) => void
	required?: boolean
	uiSchema?: any
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

	// Handle change events for Select component
	const handleSelectChange = (value: string) => {
		setFieldValue(value)
		onChange(name, value)
	}

	// Toggle password visibility
	const togglePasswordVisibility = () => {
		setShowPassword(!showPassword)
	}

	// Get UI options from uiSchema
	const uiOptions = uiSchema?.["ui:options"] || {}
	const fieldClassName = uiSchema?.["ui:className"] || ""
	const fieldPlaceholder =
		uiSchema?.["ui:placeholder"] || schema.placeholder || ""
	const fieldDescription =
		uiSchema?.["ui:description"] || schema.description || ""
	const fieldTitle = uiSchema?.["ui:title"] || schema.title || name

	// Render the appropriate input based on type
	const renderInput = () => {
		// Handle enum/select type
		if (schema.enum) {
			return (
				<Select
					value={fieldValue}
					onChange={handleSelectChange}
					className={`w-full ${fieldClassName}`}
					placeholder={fieldPlaceholder}
					options={schema.enum.map((option: string) => ({
						label: option,
						value: option,
					}))}
				/>
			)
		}

		if (schema.type === "boolean") {
			return (
				<div className={`flex items-center justify-between ${fieldClassName}`}>
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
						placeholder={fieldPlaceholder}
						className={`w-full rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 ${fieldClassName}`}
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
				placeholder={fieldPlaceholder}
				className={`w-full rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 ${fieldClassName}`}
			/>
		)
	}

	return (
		<div className={`mb-4 ${uiOptions.fullWidth ? "col-span-2" : ""}`}>
			<div className="mb-1 flex items-center gap-1">
				<label className="text-sm font-medium text-gray-700">
					{fieldTitle}
					{required && <span className="text-red-500">*</span>}
				</label>
				{fieldDescription && (
					<Tooltip title={fieldDescription}>
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
	uiSchema,
}: {
	schema: RJSFSchema
	formData: any
	onChange: (data: any) => void
	uiSchema?: any
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
			[parentField]: {
				...formData?.[parentField],
				...fieldData,
			},
		}
		onChange(updatedData)
	}

	// Generate form fields based on schema
	const renderFields = () => {
		// Safely get properties and handle undefined
		const properties = schema.properties || {}

		return Object.entries(properties)
			.map(([name, fieldSchemaDefinition]) => {
				// Handle nested objects
				const fieldSchema = fieldSchemaDefinition as any
				const fieldUiSchema =
					uiSchema && uiSchema[name] ? uiSchema[name] : undefined

				// Skip hidden fields but keep their values in the form data
				if (fieldUiSchema && fieldUiSchema["ui:widget"] === "hidden") {
					return null
				}

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
								uiSchema={fieldUiSchema}
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
						uiSchema={fieldUiSchema}
					/>
				)
			})
			.filter(Boolean) // Filter out null entries (hidden fields)
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
		if (props.schema && props.schema.properties) {
			// Keep only the fields that exist in the schema
			const filteredData: any = {}

			// Get field names and properties from the schema
			const schemaProperties = props.schema.properties || {}

			// First, populate with default values from schema
			Object.entries(schemaProperties).forEach(
				([key, propValue]: [string, any]) => {
					// Handle nested objects with defaults
					if (
						propValue &&
						typeof propValue === "object" &&
						propValue.type === "object" &&
						propValue.properties
					) {
						filteredData[key] = filteredData[key] || {}
						Object.entries(propValue.properties).forEach(
							([nestedKey, nestedProp]: [string, any]) => {
								if (
									nestedProp &&
									typeof nestedProp === "object" &&
									nestedProp.default !== undefined &&
									(!props.formData?.[key] ||
										props.formData[key][nestedKey] === undefined)
								) {
									filteredData[key][nestedKey] = nestedProp.default
								}
							},
						)
					}

					// Handle regular properties with defaults
					if (
						propValue &&
						typeof propValue === "object" &&
						propValue.default !== undefined &&
						(!props.formData || props.formData[key] === undefined)
					) {
						filteredData[key] = propValue.default
					}
				},
			)

			// Then overlay with existing formData values
			if (props.formData) {
				Object.entries(props.formData).forEach(([key, value]) => {
					if (key in schemaProperties) {
						// Handle nested objects
						if (
							schemaProperties[key] &&
							typeof schemaProperties[key] === "object" &&
							(schemaProperties[key] as any).type === "object" &&
							typeof value === "object" &&
							value !== null
						) {
							filteredData[key] = filteredData[key] || {}
							Object.entries(value).forEach(([nestedKey, nestedValue]) => {
								if (
									schemaProperties[key] &&
									typeof schemaProperties[key] === "object" &&
									(schemaProperties[key] as any).properties &&
									nestedKey in (schemaProperties[key] as any).properties
								) {
									filteredData[key][nestedKey] = nestedValue
								}
							})
						} else {
							filteredData[key] = value
						}
					}
				})
			}

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
			uiSchema={props.uiSchema}
		/>
	)
}

export default FixedSchemaForm
