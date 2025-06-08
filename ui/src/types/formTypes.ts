import { RJSFSchema, UiSchema } from "@rjsf/utils"
export interface FormFieldProps {
	label: string
	required?: boolean
	children: React.ReactNode
	error?: string | null
}

export interface DynamicSchemaFormProps {
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

export type FieldSchema = {
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

export interface DirectFormFieldProps {
	name: string
	schema: FieldSchema
	value: any
	onChange: (name: string, value: any) => void
	required?: boolean
	uiSchema?: UiSchema
	error?: string
	validate?: boolean
}

export interface DirectInputFormProps {
	schema: RJSFSchema
	formData: Record<string, any>
	onChange: (data: Record<string, any>) => void
	uiSchema?: UiSchema
	errors?: Record<string, string>
	validate?: boolean
}
