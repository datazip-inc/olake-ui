import { useMemo } from "react"
import Form from "@rjsf/core"
import { RJSFSchema, UiSchema } from "@rjsf/utils"
import validator from "@rjsf/validator-ajv8"
import { message, Switch } from "antd"

interface JsonSchemaFormProps {
	schema: RJSFSchema
	formData?: any
	onChange?: (formData: any) => void
	onSubmit?: (formData: any) => void
	uiSchema?: UiSchema
	hideSubmit?: boolean
}

const JsonSchemaForm: React.FC<JsonSchemaFormProps> = ({
	schema,
	formData,
	onChange,
	onSubmit,
	uiSchema: providedUiSchema,
	hideSubmit = false,
}) => {
	const uiSchema = useMemo(() => {
		const baseUiSchema: UiSchema = {
			"ui:submitButtonOptions": {
				norender: hideSubmit,
			},
			...providedUiSchema,
		}

		return baseUiSchema
	}, [providedUiSchema, hideSubmit])

	const handleError = (errors: any) => {
		errors.forEach((error: any) => {
			// Customize error messages for required properties
			let errorMessage = error.message
			if (errorMessage.includes("required property")) {
				// Extract field name from the error message
				const fieldMatch = errorMessage.match(/required property '(.+?)'/)
				if (fieldMatch && fieldMatch[1]) {
					const fieldName = fieldMatch[1]
					errorMessage = `Please enter a value for ${fieldName}`
				}
			}

			message.error({
				content: errorMessage,
				key: error.stack,
			})
		})
	}

	const customTemplates = {
		FieldTemplate: (props: any) => {
			const {
				id,
				label,
				help,
				required,
				description,
				errors,
				children,
				uiSchema: fieldUiSchema,
				schema,
			} = props

			const labelClass =
				fieldUiSchema?.["ui:labelClass"] ||
				"mb-1 text-sm font-medium text-gray-700"

			// Apply consistent field styling
			const fieldClass = fieldUiSchema?.["ui:className"] || "w-full mb-2"

			// Skip rendering label for checkboxes and if ui:options.label is false
			const isCheckbox =
				schema?.type === "boolean" ||
				fieldUiSchema?.["ui:widget"] === "checkbox"
			const showLabel =
				fieldUiSchema?.["ui:options"]?.label !== false &&
				fieldUiSchema?.["ui:labelDisplay"] !== "left-only" &&
				!isCheckbox &&
				label

			return (
				<div className={fieldClass}>
					{showLabel && (
						<label
							htmlFor={id}
							className={labelClass}
						>
							{label}
							{!required && <span className="text-gray-500"> (optional)</span>}
						</label>
					)}
					{children}
					{description && (
						<p className="mt-1 text-xs text-gray-500">{description}</p>
					)}
					{help && <p className="mt-1 text-xs text-gray-500">{help}</p>}
					{errors && <div className="mt-1 text-sm text-red-500">{errors}</div>}
				</div>
			)
		},
		ObjectFieldTemplate: (props: any) => {
			const { title, description, properties, uiSchema: fieldUiSchema } = props

			const fieldClass = fieldUiSchema?.["ui:className"] || ""
			const TitleComponent = fieldUiSchema?.["ui:title"]
			const CustomField = fieldUiSchema?.["ui:field"]

			// Default to grid layout for all objects with multiple properties
			// Can be overridden with ui:options.noGrid = true
			const shouldUseGridLayout =
				properties.length > 1 && fieldUiSchema?.["ui:options"]?.noGrid !== true

			// Apply custom grid class if provided, otherwise use default grid
			const gridClass =
				fieldUiSchema?.["ui:options"]?.className ||
				(shouldUseGridLayout ? "grid grid-cols-2 gap-x-6 gap-y-4" : "")

			const renderTitle = () => {
				if (!title || title === null) return null

				if (TitleComponent && typeof TitleComponent === "function") {
					return <TitleComponent title={title} />
				}

				return <h3 className="mb-4 text-lg font-medium">{title}</h3>
			}

			return (
				<div className={fieldClass}>
					{renderTitle()}
					{description && (
						<p className="mb-4 text-sm text-gray-500">{description}</p>
					)}
					<div className={gridClass}>
						{CustomField ? (
							<CustomField {...props}>
								{properties.map((prop: any) => prop.content)}
							</CustomField>
						) : (
							properties.map((prop: any) => {
								// Check if the property is an object with multiple fields
								const isNestedObject =
									prop.content.props.schema.type === "object" &&
									Object.keys(prop.content.props.schema.properties || {})
										.length > 3

								// Make nested objects span full width
								const spanClass = isNestedObject ? "col-span-2" : ""

								return (
									<div
										key={prop.name}
										className={spanClass}
									>
										{prop.content}
									</div>
								)
							})
						)}
					</div>
				</div>
			)
		},
		RadioWidget: (props: any) => {
			const {
				id,
				options,
				value,
				disabled,
				readonly,
				onChange,
				uiSchema: fieldUiSchema,
			} = props

			const { enumOptions } = options
			const isInline = fieldUiSchema?.["ui:options"]?.inline
			const containerClass = fieldUiSchema?.["ui:options"]?.className || ""
			const radioClass = fieldUiSchema?.["ui:radioClass"] || ""

			return (
				<div className={`${isInline ? "flex" : ""} ${containerClass}`}>
					{enumOptions.map((option: any, i: number) => (
						<label
							key={i}
							className={`flex cursor-pointer items-center ${radioClass}`}
						>
							<div className="relative flex items-center">
								<input
									type="radio"
									id={`${id}_${i}`}
									checked={option.value === value}
									disabled={disabled || readonly}
									onChange={() => onChange(option.value)}
									className="h-4 w-4 border-gray-300 text-blue-600 focus:ring-blue-500"
								/>
								<span className="ml-2 text-sm">{option.label}</span>
							</div>
						</label>
					))}
				</div>
			)
		},
		SwitchWidget: (props: any) => {
			const {
				id,
				value,
				disabled,
				readonly,
				onChange,
				label,
				schema,
				uiSchema: fieldUiSchema,
			} = props

			const switchClass = fieldUiSchema?.["ui:className"] || ""

			// Always show the label in the widget itself
			const displayLabel = schema.title || label

			return (
				<div
					className={`flex w-full items-center justify-between ${switchClass}`}
				>
					<span className="text-sm font-medium">{displayLabel}</span>
					<Switch
						id={id}
						checked={value || false}
						disabled={disabled || readonly}
						onChange={checked => onChange(checked)}
						className={value ? "bg-blue-600" : "bg-gray-200"}
					/>
				</div>
			)
		},
		SelectWidget: (props: any) => {
			const {
				id,
				options,
				value,
				disabled,
				readonly,
				onChange,
				uiSchema: fieldUiSchema,
				placeholder,
			} = props

			const { enumOptions } = options
			const selectClass = fieldUiSchema?.["ui:className"] || "w-full"

			return (
				<div className="relative">
					<select
						id={id}
						value={value || ""}
						disabled={disabled || readonly}
						onChange={e => onChange(e.target.value)}
						className={`${selectClass} appearance-none rounded-[6px] border border-gray-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500`}
					>
						{placeholder && <option value="">{placeholder}</option>}
						{enumOptions.map(({ value, label }: any, i: number) => (
							<option
								key={i}
								value={value}
							>
								{label}
							</option>
						))}
					</select>
					<div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-2 text-gray-700">
						<svg
							className="h-4 w-4 fill-current"
							xmlns="http://www.w3.org/2000/svg"
							viewBox="0 0 20 20"
						>
							<path d="M9.293 12.95l.707.707L15.657 8l-1.414-1.414L10 10.828 5.757 6.586 4.343 8z" />
						</svg>
					</div>
				</div>
			)
		},
		TextWidget: (props: any) => {
			const {
				id,
				placeholder,
				value,
				required,
				disabled,
				readonly,
				onChange,
				onBlur,
				onFocus,
				schema,
				uiSchema: fieldUiSchema,
			} = props

			const inputClass = fieldUiSchema?.["ui:className"] || "w-full"
			const inputType = fieldUiSchema?.["ui:widget"] || "text"

			// Generate better placeholder based on field name if not provided
			const getPlaceholder = () => {
				if (placeholder) return placeholder
				const fieldName = schema.title || id?.split("_").pop()
				return `Enter ${fieldName?.toLowerCase()}${required ? " *" : ""}`
			}

			return (
				<input
					id={id}
					placeholder={getPlaceholder()}
					value={value || ""}
					required={required}
					disabled={disabled || readonly}
					readOnly={readonly}
					type={inputType}
					onChange={e => onChange(e.target.value)}
					onBlur={onBlur && (e => onBlur(id, e.target.value))}
					onFocus={onFocus && (e => onFocus(id, e.target.value))}
					className={`${inputClass} rounded-[6px] border ${required && !value ? "border-red-300" : "border-gray-300"} px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500`}
				/>
			)
		},
		PasswordWidget: (props: any) => {
			const {
				id,
				placeholder,
				value,
				required,
				disabled,
				readonly,
				onChange,
				onBlur,
				onFocus,
				uiSchema: fieldUiSchema,
			} = props

			const inputClass = fieldUiSchema?.["ui:className"] || "w-full"

			return (
				<div className="relative">
					<input
						id={id}
						placeholder={placeholder || "Enter password"}
						value={value || ""}
						required={required}
						disabled={disabled || readonly}
						readOnly={readonly}
						type="password"
						onChange={e => onChange(e.target.value)}
						onBlur={onBlur && (e => onBlur(id, e.target.value))}
						onFocus={onFocus && (e => onFocus(id, e.target.value))}
						className={`${inputClass} rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500`}
					/>
					<div className="pointer-events-none absolute inset-y-0 right-0 flex items-center px-3 text-gray-400">
						<svg
							xmlns="http://www.w3.org/2000/svg"
							className="h-4 w-4"
							viewBox="0 0 20 20"
							fill="currentColor"
						>
							<path
								fillRule="evenodd"
								d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z"
								clipRule="evenodd"
							/>
						</svg>
					</div>
				</div>
			)
		},
		NumberWidget: (props: any) => {
			const {
				id,
				placeholder,
				value,
				required,
				disabled,
				readonly,
				onChange,
				onBlur,
				onFocus,
				schema,
				uiSchema: fieldUiSchema,
			} = props

			const inputClass = fieldUiSchema?.["ui:className"] || "w-full"

			// Generate better placeholder based on field name if not provided
			const getPlaceholder = () => {
				if (placeholder) return placeholder
				const fieldName = schema.title || id?.split("_").pop()
				return `Enter ${fieldName?.toLowerCase()}`
			}

			return (
				<input
					id={id}
					placeholder={getPlaceholder()}
					value={value || ""}
					required={required}
					disabled={disabled || readonly}
					readOnly={readonly}
					type="number"
					onChange={e =>
						onChange(e.target.value ? Number(e.target.value) : undefined)
					}
					onBlur={
						onBlur &&
						(e =>
							onBlur(id, e.target.value ? Number(e.target.value) : undefined))
					}
					onFocus={
						onFocus &&
						(e =>
							onFocus(id, e.target.value ? Number(e.target.value) : undefined))
					}
					className={`${inputClass} rounded-[6px] border border-gray-300 px-3 py-2 text-sm shadow-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500`}
				/>
			)
		},
		ArrayFieldTemplate: (props: any) => {
			const {
				items,
				canAdd,
				onAddClick,
				title,
				description,
				required,
				uiSchema,
			} = props

			const addButtonText =
				uiSchema?.["ui:ArrayField"]?.["ui:addButtonText"] || "Add Item"
			const showAddButton =
				canAdd !== false &&
				uiSchema?.["ui:ArrayField"]?.["ui:addable"] !== false
			const showTitle = uiSchema?.["ui:ArrayField"]?.["ui:showTitle"] !== false
			const showRemoveButton = () => {
				// Only show remove button if there's more than one item
				return items.length > 1
			}

			return (
				<div className="w-full">
					{showTitle && (
						<div className="mb-2">
							{title && (
								<label className="mb-2 block text-sm font-medium text-gray-700">
									{title}
									{!required && (
										<span className="text-gray-500"> (optional)</span>
									)}
								</label>
							)}
							{description && (
								<p className="text-sm text-gray-500">{description}</p>
							)}
						</div>
					)}
					<div className="space-y-2">
						{items.map((item: any) => (
							<div
								key={item.key}
								className="flex items-start gap-2"
							>
								{item.children}
								{item.hasRemove && showRemoveButton() && (
									<button
										type="button"
										onClick={item.onDropIndexClick(item.index)}
										className="mt-2 rounded-[6px] border border-red-300 px-2 py-1 text-sm text-red-600 hover:bg-red-50"
									>
										Remove
									</button>
								)}
							</div>
						))}
					</div>
					{showAddButton && (
						<button
							type="button"
							onClick={onAddClick}
							className="mt-2 rounded-[6px] border border-blue-300 px-3 py-1 text-sm text-blue-600 hover:bg-blue-50"
						>
							{addButtonText}
						</button>
					)}
				</div>
			)
		},
	}

	return (
		<div className="form-container w-full">
			<Form
				schema={schema}
				uiSchema={uiSchema}
				formData={formData}
				onChange={e => onChange?.(e.formData)}
				onSubmit={e => onSubmit?.(e.formData)}
				onError={handleError}
				validator={validator}
				templates={{
					...customTemplates,
					ArrayFieldTemplate: customTemplates.ArrayFieldTemplate,
				}}
				liveValidate={true}
				showErrorList={false}
				className="w-full"
				widgets={{
					TextWidget: customTemplates.TextWidget,
					PasswordWidget: customTemplates.PasswordWidget,
					NumberWidget: customTemplates.NumberWidget,
					RadioWidget: customTemplates.RadioWidget,
					SelectWidget: customTemplates.SelectWidget,
					CheckboxWidget: customTemplates.SwitchWidget,
				}}
			/>
		</div>
	)
}

export default JsonSchemaForm
