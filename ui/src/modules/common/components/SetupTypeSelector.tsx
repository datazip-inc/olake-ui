import { Radio } from "antd"
import { SetupType, SetupTypeSelectorProps } from "@app-types/index"

export const SetupTypeSelector: React.FC<SetupTypeSelectorProps> = ({
	value,
	onChange,
	newLabel = "Set up a new source",
	existingLabel = "Use an existing source",
	fromJobFlow = false,
}) => {
	return (
		<div className="mb-4 flex">
			<Radio.Group
				value={value}
				onChange={e => onChange(e.target.value as SetupType)}
				className="flex"
			>
				<Radio
					value="new"
					className="mr-8"
				>
					{newLabel}
				</Radio>
				{fromJobFlow && <Radio value="existing">{existingLabel}</Radio>}
			</Radio.Group>
		</div>
	)
}
