import { FormFieldProps } from "../types"

const FormField = ({ label, required, children, error }: FormFieldProps) => (
	<div className="w-full">
		<label className="mb-2 block text-sm font-medium text-gray-700">
			{label}
			{required && <span className="text-red-500">*</span>}
		</label>
		{children}
		{error && <div className="mt-1 text-sm text-red-500">{error}</div>}
	</div>
)
export default FormField
