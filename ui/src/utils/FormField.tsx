import { Tooltip } from "antd"
import { Info } from "@phosphor-icons/react"
import { FormFieldProps } from "../types"

const FormField = ({
	label,
	required,
	children,
	error,
	tooltip,
	info,
}: FormFieldProps) => (
	<div className="w-full">
		<label className="mb-2 flex items-center gap-1 text-sm font-medium text-gray-700">
			{label}
			{required && <span className="text-red-500">*</span>}
			{tooltip && (
				<Tooltip title={tooltip}>
					<Info
						size={16}
						className="ml-1 cursor-help text-slate-900"
					/>
				</Tooltip>
			)}
			{info}
		</label>
		{children}
		{error && <div className="mt-1 text-sm text-red-500">{error}</div>}
	</div>
)
export default FormField
