import React from "react"
import { Radio } from "antd"
import { WidgetProps } from "@rjsf/utils"

const CustomRadioWidget: React.FC<WidgetProps> = props => {
	const { options, value, onChange, disabled, readonly, autofocus, id } = props

	const { enumOptions = [] } = options

	return (
		<div
			id={id}
			className="flex gap-3"
		>
			{enumOptions.map((opt: any) => (
				<div
					key={opt.value}
					className="mb-4"
				>
					<Radio
						value={opt.value}
						checked={value === opt.value}
						onChange={() => onChange(opt.value)}
						disabled={disabled || readonly}
						autoFocus={autofocus}
					>
						{opt.label}
					</Radio>
				</div>
			))}
		</div>
	)
}

export default CustomRadioWidget
