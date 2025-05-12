import React, { useEffect, useRef, useState } from "react"
import { RJSFSchema, UiSchema } from "@rjsf/utils"
import { Switch, Tooltip, Select } from "antd"
import { Info, Eye, EyeSlash } from "@phosphor-icons/react"

interface DynamicSchemaFormProps {
	schema: RJSFSchema
	formData: any
	onChange: (data: any) => void
	onSubmit?: (data: any) => void
	uiSchema?: UiSchema
	hideSubmit?: boolean
	className?: string
}

type FieldSchema = {
	type?: string
	format?: string
	title?: string
	description?: string
	placeholder?: string
	enum?: string[]
	default?: any
	properties?: Record<string, FieldSchema>
	required?: string[]
}

type UISchema = {
	"ui:widget"?: string
	[key: string]: any
}

interface DirectFormFieldProps {
	name: string
	schema: FieldSchema
	value: any
	onChange: (name: string, value: any) => void
	required?: boolean
	uiSchema?: UISchema
}

const DirectFormField = ({
	name,
	schema,
	value,
	onChange,
	required = false,
}: DirectFormFieldProps) => {
	const [fieldValue, setFieldValue] = useState(value)
	const [showPassword, setShowPassword] = useState(false)
	const inputRef = useRef<HTMLInputElement>(null)

	useEffect(() => {
		if (document.activeElement !== inputRef.current) {
			setFieldValue(value)
		}
	}, [value])

	const getInputType = (): string => {
		if (schema.type === "number" || schema.type === "integer") {
			return "number"
		} else if (schema.type === "boolean") {
			return "checkbox"
		} else if (schema.format === "password") {
			return showPassword ? "text" : "password"
		}
		return "text"
	}

	const inputType = getInputType()

	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const newValue =
			inputType === "checkbox"
				? e.target.checked
				: inputType === "number"
					? e.target.value
						? Number(e.target.value)
						: undefined
					: e.target.value

		setFieldValue(newValue)
		onChange(name, newValue)
	}

	const handleSwitchChange = (checked: boolean) => {
		setFieldValue(checked)
		onChange(name, checked)
	}

	const handleSelectChange = (value: string) => {
		setFieldValue(value)
		onChange(name, value)
	}

	const togglePasswordVisibility = () => {
		setShowPassword(prev => !prev)
	}

	const renderInput = () => {
		if (schema.enum?.length) {
			return (
				<Select
					value={fieldValue}
					onChange={handleSelectChange}
					className="w-full"
					placeholder={schema.placeholder}
					options={schema.enum.map(option => ({
						label: option,
						value: option,
					}))}
				/>
			)
		}

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
						value={fieldValue ?? ""}
						onChange={handleChange}
						placeholder={schema.placeholder}
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
				value={inputType !== "checkbox" ? (fieldValue ?? "") : undefined}
				checked={inputType === "checkbox" ? !!fieldValue : undefined}
				onChange={handleChange}
				placeholder={schema.placeholder}
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

interface DirectInputFormProps {
	schema: RJSFSchema
	formData: Record<string, any>
	onChange: (data: Record<string, any>) => void
	uiSchema?: Record<string, UISchema>
}

export const DirectInputForm = ({
	schema,
	formData,
	onChange,
	uiSchema,
}: DirectInputFormProps) => {
	if (!schema?.properties) {
		return <div>No schema properties provided</div>
	}

	const handleFieldChange = (fieldName: string, fieldValue: any) => {
		onChange({
			...formData,
			[fieldName]: fieldValue,
		})
	}

	const handleNestedFieldChange = (
		parentField: string,
		fieldData: Record<string, any>,
	) => {
		onChange({
			...formData,
			[parentField]: {
				...formData?.[parentField],
				...fieldData,
			},
		})
	}

	const renderFields = () => {
		const properties = schema.properties || {}

		return Object.entries(properties)
			.map(([name, fieldSchemaDefinition]) => {
				const fieldSchema = fieldSchemaDefinition as FieldSchema
				const fieldUiSchema = uiSchema?.[name]

				if (fieldUiSchema?.["ui:widget"] === "hidden") {
					return null
				}

				const isNestedObject =
					fieldSchema.type === "object" && fieldSchema.properties

				if (isNestedObject) {
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
			.filter(Boolean)
	}

	return (
		<div className="direct-input-form">
			<form onSubmit={e => e.preventDefault()}>
				<div className="grid grid-cols-2 gap-x-8 gap-y-4">{renderFields()}</div>
			</form>
		</div>
	)
}

const isNestedObjectSchema = (schema: any): boolean => {
	return (
		schema &&
		typeof schema === "object" &&
		schema.type === "object" &&
		schema.properties
	)
}

export const FixedSchemaForm: React.FC<DynamicSchemaFormProps> = props => {
	const [filteredFormData, setFilteredFormData] = useState<Record<string, any>>(
		{},
	)

	useEffect(() => {
		if (!props.schema?.properties) {
			setFilteredFormData(props.formData || {})
			return
		}

		const filteredData: Record<string, any> = {}
		const schemaProperties = props.schema.properties || {}

		// Apply defaults for missing values
		Object.entries(schemaProperties).forEach(
			([key, propValue]: [string, any]) => {
				// Handle nested objects
				if (isNestedObjectSchema(propValue)) {
					filteredData[key] = filteredData[key] || {}

					Object.entries(propValue.properties || {}).forEach(
						([nestedKey, nestedProp]: [string, any]) => {
							const hasDefault = nestedProp?.default !== undefined
							const isMissing =
								!props.formData?.[key] ||
								props.formData[key][nestedKey] === undefined

							if (hasDefault && isMissing) {
								filteredData[key][nestedKey] = nestedProp.default
							}
						},
					)
				}

				// Handle top-level properties with defaults
				const hasDefault = propValue?.default !== undefined
				const isMissing = !props.formData || props.formData[key] === undefined

				if (hasDefault && isMissing) {
					filteredData[key] = propValue.default
				}
			},
		)

		// Merge existing form data
		if (props.formData) {
			Object.entries(props.formData).forEach(([key, value]) => {
				if (key in schemaProperties) {
					if (
						isNestedObjectSchema(schemaProperties[key]) &&
						typeof value === "object" &&
						value !== null
					) {
						filteredData[key] = filteredData[key] || {}

						Object.entries(value).forEach(([nestedKey, nestedValue]) => {
							const nestedProperties = (schemaProperties[key] as any)
								?.properties
							if (nestedProperties && nestedKey in nestedProperties) {
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
	}, [props.schema, props.formData])

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
