import { Radio } from "antd"
import {
	SetupType,
	SetupTypeSelectorProps,
} from "@/modules/ingestion/common/types"

export const SetupTypeSelector: React.FC<SetupTypeSelectorProps> = ({
	value,
	onChange,
	newLabel = "Set up a new source",
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
			</Radio.Group>
		</div>
	)
}
