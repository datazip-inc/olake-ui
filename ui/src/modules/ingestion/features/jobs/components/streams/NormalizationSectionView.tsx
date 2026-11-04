import {
	ArrowSquareOutIcon,
	InfoIcon,
	WarningIcon,
} from "@phosphor-icons/react"
import { Switch, Tooltip } from "antd"
import clsx from "clsx"

import { CARD_STYLE, NORMALIZATION_TOOLTIP_TEXT } from "../../constants"

export interface NormalizationSectionViewProps {
	normalization: boolean
	isSelected: boolean
	isDirty?: boolean
	onChange?: (value: boolean) => void
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
						<Tooltip title={NORMALIZATION_TOOLTIP_TEXT}>
							<InfoIcon
								size={14}
								className="cursor-help text-text-tertiary"
							/>
						</Tooltip>
						<a
							href="https://olake.io/docs/understanding/terminologies/olake/#1-normalization"
							target="_blank"
							rel="noopener noreferrer"
							aria-label="Open normalization docs"
							className="inline-flex text-text-tertiary hover:text-primary"
						>
							<ArrowSquareOutIcon size={14} />
						</a>
					</div>
					<Switch
						checked={normalization}
						onChange={onChange ?? (() => {})}
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
