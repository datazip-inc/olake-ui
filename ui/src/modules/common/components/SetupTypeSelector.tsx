import { forwardRef } from "react"
import { Radio } from "antd"
import { SetupType, SetupTypeSelectorProps } from "../../../types"

export const SetupTypeSelector = forwardRef<
  any,
  SetupTypeSelectorProps
>(
  (
    {
      value,
      onChange,
      newLabel = "Set up a new source",
      existingLabel = "Use an existing source",
      fromJobFlow = false,
    },
    ref,
  ) => {
		return (
			<div className="mb-4 flex">
				<Radio.Group
					value={value}
					onChange={e => onChange(e.target.value as SetupType)}
					className="flex"
				>
					<Radio
						ref={ref}
						value="new"
						className="mr-8"
					>
						{newLabel}
					</Radio>
					{fromJobFlow && <Radio value="existing">{existingLabel}</Radio>}
				</Radio.Group>
			</div>
		)
	},
)

SetupTypeSelector.displayName = "SetupTypeSelector"
