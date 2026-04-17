import { InfoIcon, WarningIcon } from "@phosphor-icons/react"
import { Switch } from "antd"
import clsx from "clsx"

import { CARD_STYLE } from "../../constants"

export interface NormalizationSectionViewProps {
	normalization: boolean
	isSelected: boolean
	isDirty?: boolean
	onChange: (value: boolean) => void
}

const NormalizationSectionView = ({
	normalization,
	isSelected,
	isDirty,
	onChange,
}: NormalizationSectionViewProps) => {
	return (
		<>
			<div
				className={clsx(
					!isSelected ? "font-normal text-text-disabled" : "font-medium",
					CARD_STYLE,
				)}
			>
				<div className="flex items-center justify-between">
					<div className="flex items-center gap-1">
						{isDirty && <WarningIcon className="size-4 text-orange-500" />}
						<label>Normalization</label>
					</div>
					<Switch
						checked={normalization}
						onChange={onChange}
						disabled={!isSelected}
					/>
				</div>
			</div>
			{!isSelected && (
				<div className="ml-1 flex items-center gap-1 text-sm text-[#686868]">
					<InfoIcon className="size-4" />
					Select the stream to configure Normalization
				</div>
			)}
		</>
	)
}

export default NormalizationSectionView
