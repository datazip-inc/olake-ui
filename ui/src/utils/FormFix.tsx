import React, { useEffect, useRef, useState } from "react"
import { RJSFSchema, UiSchema } from "@rjsf/utils"
import { Switch, Tooltip, Select, Button, Input } from "antd"
import { Info, Eye, EyeSlash, Plus, Trash } from "@phosphor-icons/react"

interface DynamicSchemaFormProps {
	schema: RJSFSchema
	formData: any
	onChange: (data: any) => void
	onSubmit?: (data: any) => void
	uiSchema?: UiSchema
	hideSubmit?: boolean
	className?: string
	errors?: Record<string, string>
	validate?: boolean
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
	items?: {
		type?: string
	}
	order?: number
}

interface DirectFormFieldProps {
	name: string
	schema: FieldSchema
	value: any
	onChange: (name: string, value: any) => void
	required?: boolean
	uiSchema?: UiSchema
	error?: string
	validate?: boolean
}

const DirectFormField = ({
	name,
	schema,
	value,
	onChange,
	required = false,
	error,
	validate = false,
}: DirectFormFieldProps) => {
	const [fieldValue, setFieldValue] = useState(value)
	const [showPassword, setShowPassword] = useState(false)
	const inputRef = useRef<HTMLInputElement>(null)
	const [localError, setLocalError] = useState<string | null>(null)
	const [wasUserModified, setWasUserModified] = useState(false)
	const [arrayItemInput, setArrayItemInput] = useState("")

	useEffect(() => {
		if (document.activeElement !== inputRef.current) {
			setFieldValue(value)
		}
	}, [value])

	useEffect(() => {
		setLocalError(error || null)
	}, [error])

	const validateField = (value: any): string | null => {
		const hasDefault = schema.default !== undefined
		const isEmpty =
			value === undefined ||
			value === null ||
			value === "" ||
			(Array.isArray(value) && value.length === 0)
		if (required && isEmpty && (wasUserModified || !hasDefault)) {
			return `${schema.title || name} is required`
		}
		return null
	}

	useEffect(() => {
		if (validate) {
			setLocalError(validateField(fieldValue))
		}
	}, [validate, fieldValue, required, schema, name, wasUserModified])

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

	const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const newValue =
			e.target.type === "checkbox"
				? e.target.checked
				: e.target.type === "number"
					? e.target.value
						? Number(e.target.value)
						: undefined
					: e.target.value

		setFieldValue(newValue)
		setWasUserModified(true)
		onChange(name, newValue)

		if (validate) {
			setLocalError(validateField(newValue))
		} else {
			setLocalError(null)
		}
	}

	const handleSwitchChange = (checked: boolean) => {
		setFieldValue(checked)
		setWasUserModified(true)
		onChange(name, checked)

		if (validate) {
			setLocalError(validateField(checked))
		} else {
			setLocalError(null)
		}
	}

	const handleSelectChange = (value: string) => {
		setFieldValue(value)
		setWasUserModified(true)
		onChange(name, value)

		if (validate) {
			setLocalError(validateField(value))
		} else {
			setLocalError(null)
		}
	}

	const handleArrayItemAdd = () => {
		if (!arrayItemInput.trim()) return

		const newArray = Array.isArray(fieldValue) ? [...fieldValue] : []
		newArray.push(arrayItemInput.trim())

		setFieldValue(newArray)
		setWasUserModified(true)
		onChange(name, newArray)
		setArrayItemInput("")

		if (validate) {
			setLocalError(validateField(newArray))
		} else {
			setLocalError(null)
		}
	}

	const handleArrayItemRemove = (index: number) => {
		if (!Array.isArray(fieldValue)) return

		const newArray = [...fieldValue]
		newArray.splice(index, 1)

		setFieldValue(newArray)
		setWasUserModified(true)
		onChange(name, newArray)

		if (validate) {
			setLocalError(validateField(newArray))
		} else {
			setLocalError(null)
		}
	}

	const togglePasswordVisibility = () => {
		setShowPassword(prev => !prev)
	}

	const renderArrayField = () => {
		const arrayValues = Array.isArray(fieldValue) ? fieldValue : []
		const itemType = schema.items?.type || "string"

		return (
			<div className="rounded-[6px] border border-gray-300 p-3">
				<div className="mb-2">
					{arrayValues.map((item, index) => (
						<div
							key={index}
							className="mb-2 flex items-center"
						>
							<div className="flex-grow">
								{itemType === "string" && (
									<Input
										value={item}
										disabled
										className="w-full rounded-[6px] border border-gray-300 px-3 py-2 text-sm"
									/>
								)}
							</div>
							<Button
								type="text"
								className="ml-2 text-red-500 hover:text-red-700"
								onClick={() => handleArrayItemRemove(index)}
								icon={<Trash className="size-4" />}
							/>
						</div>
					))}
				</div>
				<div className="flex items-center">
					<Input
						value={arrayItemInput}
						onChange={e => setArrayItemInput(e.target.value)}
						placeholder={`Add ${schema.title || name} item...`}
						onKeyPress={e => {
							if (e.key === "Enter") {
								e.preventDefault()
								handleArrayItemAdd()
							}
						}}
						className="flex-grow rounded-[6px] border border-gray-300 px-3 py-2 text-sm"
					/>
					<Button
						type="primary"
						className="ml-2 bg-[#203FDD] hover:bg-blue-700"
						onClick={handleArrayItemAdd}
						icon={<Plus className="size-4" />}
					>
						Add
					</Button>
				</div>
			</div>
		)
	}

	const renderInput = () => {
		if (schema.type === "array") {
			return renderArrayField()
		}

		if (schema.enum?.length) {
			return (
				<Select
					value={fieldValue}
					onChange={handleSelectChange}
					className="w-full"
					placeholder={schema.placeholder}
					status={localError ? "error" : undefined}
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
						className={`w-full rounded-[6px] border ${localError ? "border-red-500" : "border-gray-300"} px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500`}
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
				type={getInputType()}
				value={getInputType() !== "checkbox" ? (fieldValue ?? "") : undefined}
				checked={getInputType() === "checkbox" ? !!fieldValue : undefined}
				onChange={handleChange}
				placeholder={schema.placeholder}
				className={`w-full rounded-[6px] border ${localError ? "border-red-500" : "border-gray-300"} px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500`}
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
			{validate && localError && (
				<div className="mt-1 text-sm text-red-500">{localError}</div>
			)}
		</div>
	)
}

interface DirectInputFormProps {
	schema: RJSFSchema
	formData: Record<string, any>
	onChange: (data: Record<string, any>) => void
	uiSchema?: UiSchema
	errors?: Record<string, string>
	validate?: boolean
}

export const DirectInputForm = ({
	schema,
	formData,
	onChange,
	uiSchema,
	errors,
	validate = false,
}: DirectInputFormProps) => {
	useEffect(() => {
		if (!schema?.properties || !onChange) {
			return
		}

		const currentLevelFormData = formData || {}
		const newLevelFormData = { ...currentLevelFormData }

		Object.entries(schema.properties).forEach(
			([name, fieldSchemaDefinition]) => {
				const fieldSchema = fieldSchemaDefinition as FieldSchema
				const fieldUiSchemaForCurrentProp = uiSchema?.[name]

				if (
					fieldUiSchemaForCurrentProp?.["ui:widget"] === "hidden" &&
					fieldSchema.default !== undefined
				) {
					if (currentLevelFormData[name] !== fieldSchema.default) {
						newLevelFormData[name] = fieldSchema.default
					}
				}
			},
		)

		// if (needsUpdate) {
		// 	onChange(newLevelFormData)
		// }
	}, [schema, formData, uiSchema, onChange])

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
			.sort(([, a], [, b]) => {
				const orderA = (a as FieldSchema).order ?? Number.MAX_SAFE_INTEGER
				const orderB = (b as FieldSchema).order ?? Number.MAX_SAFE_INTEGER
				return orderA - orderB
			})
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
								errors={
									typeof errors?.[name] === "object"
										? (errors?.[name] as Record<string, string>)
										: undefined
								}
								validate={validate}
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
						error={errors?.[name]}
						validate={validate}
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

// eslint-disable-next-line react-refresh/only-export-components
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

const generateDefaultValues = (schema: RJSFSchema): Record<string, any> => {
	const defaults: Record<string, any> = {}

	if (!schema.properties) return defaults

	Object.entries(schema.properties).forEach(
		([key, propValue]: [string, any]) => {
			if (propValue?.default !== undefined) {
				defaults[key] = propValue.default
			}

			if (isNestedObjectSchema(propValue)) {
				const nestedDefaults = generateDefaultValues(propValue)
				if (Object.keys(nestedDefaults).length > 0) {
					defaults[key] = nestedDefaults
				}
			}
		},
	)

	return defaults
}

export const FixedSchemaForm: React.FC<DynamicSchemaFormProps> = props => {
	const [filteredFormData, setFilteredFormData] = useState<Record<string, any>>(
		{},
	)

	useEffect(() => {
		if (props.schema?.properties && props.onChange) {
			const defaultValues = generateDefaultValues(props.schema)
			// Only trigger if we have defaults and user didn't provide formData yet
			if (
				Object.keys(defaultValues).length > 0 &&
				(!props.formData || Object.keys(props.formData).length === 0)
			) {
				props.onChange(defaultValues)
			}
		}
	}, [props.schema, props.onChange, props.formData])

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
			errors={props.errors}
			validate={props.validate}
		/>
	)
}

export default FixedSchemaForm
