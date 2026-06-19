import { ArrowSquareOutIcon, WarningIcon } from "@phosphor-icons/react"
import { Switch, Tooltip } from "antd"
import clsx from "clsx"

import { CARD_STYLE } from "../../constants"

export interface SourceNamingConventionSectionViewProps {
	useSourceColumnNames: boolean
	isSelected: boolean
	isDirty?: boolean
	onChange?: (value: boolean) => void
}

const SourceNamingConventionSectionView = ({
	useSourceColumnNames,
	isSelected,
	isDirty,
	onChange,
}: SourceNamingConventionSectionViewProps) => {
	return (
		<>
			<div className={clsx("font-medium", CARD_STYLE)}>
				<div className="flex items-center justify-between">
					<div className="flex items-center gap-1">
						{isDirty && <WarningIcon className="size-4 text-orange-500" />}
						<label>Source Naming Convention</label>
						<Tooltip
							title="View Documentation"
							className="ml-1 border-l px-2"
						>
							<a
								href="https://olake.io/docs/understanding/terminologies/olake/#source-naming-convention"
								target="_blank"
								rel="noopener noreferrer"
								className="flex items-center text-gray-600 transition-colors hover:text-primary"
							>
								<ArrowSquareOutIcon className="size-4" />
							</a>
						</Tooltip>
					</div>
					<Switch
						checked={useSourceColumnNames}
						onChange={onChange ?? (() => {})}
						disabled={!isSelected}
					/>
				</div>
				<p className="mt-1 text-xs font-normal text-gray-500">
					When enabled, the source column naming convention will be applied to
					destination column names.
				</p>
			</div>
		</>
	)
}

export default SourceNamingConventionSectionView
