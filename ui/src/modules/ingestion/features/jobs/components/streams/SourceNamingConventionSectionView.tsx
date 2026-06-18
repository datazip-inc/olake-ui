import { ArrowSquareOutIcon, WarningIcon } from "@phosphor-icons/react"
import { Switch } from "antd"
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
						<a
							// TODO: update doc link before merge
							href="https://olake.io/docs/understanding/terminologies/olake/#source-naming-convention"
							target="_blank"
							rel="noopener noreferrer"
							className="ml-1 flex items-center text-gray-500 transition-colors hover:text-primary"
							onClick={e => e.stopPropagation()}
						>
							<ArrowSquareOutIcon className="size-4" />
						</a>
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
