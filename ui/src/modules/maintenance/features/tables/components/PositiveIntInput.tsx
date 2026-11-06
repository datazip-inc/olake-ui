import { Input } from "antd"

import { parsePositiveInt } from "../utils"

type PositiveIntInputProps = {
	value: number
	onChange: (value: number) => void
	placeholder: string
	addonAfter?: React.ReactNode
	className?: string
}

/** Digits-only text input; an empty field reads as 0 so callers can flag it invalid. */
const PositiveIntInput: React.FC<PositiveIntInputProps> = ({
	value,
	onChange,
	placeholder,
	addonAfter,
	className,
}) => (
	<Input
		value={value === 0 ? "" : String(value)}
		onChange={e => onChange(parsePositiveInt(e.target.value))}
		inputMode="numeric"
		placeholder={placeholder}
		addonAfter={addonAfter}
		className={className}
	/>
)

export default PositiveIntInput
