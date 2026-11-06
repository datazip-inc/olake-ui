import { QuestionIcon } from "@phosphor-icons/react"
import { Input, Select, Tooltip } from "antd"
import clsx from "clsx"

import { CRON_FREQUENCY_OPTIONS } from "../constants"
import type { CronConfigOption } from "../types"
import { getCronConfigError, getCronFromConfig, getNextRuns } from "../utils"
import RunTimeChips from "./RunTimeChips"

type ScheduleSectionProps = {
	title: string
	value: CronConfigOption
	onChange: (next: CronConfigOption) => void
	isFirst?: boolean
	tooltip?: string
}

/** One cron schedule block (frequency / custom expression / next 3 runs), shared by the optimize and cleanup tabs. */
const ScheduleSection: React.FC<ScheduleSectionProps> = ({
	title,
	value,
	onChange,
	isFirst = false,
	tooltip,
}) => {
	const nextRuns = getNextRuns(getCronFromConfig(value))
	const cronError =
		value.frequency === "custom" ? getCronConfigError(value) : null

	return (
		<div
			className={clsx(
				"flex flex-col gap-4 py-8",
				!isFirst && "border-t border-olake-border-secondary",
			)}
		>
			{/* Section heading */}
			<div className="flex items-center gap-1">
				<span className="text-base font-medium leading-6 text-olake-text">
					{title}
				</span>
				{tooltip && (
					<Tooltip title={tooltip}>
						<QuestionIcon
							size={14}
							className="cursor-help text-olake-text-tertiary"
						/>
					</Tooltip>
				)}
			</div>

			{/* Form row */}
			<div className="flex gap-4">
				{/* Frequency */}
				<div className="flex w-[233px] flex-col gap-2">
					<label className="text-sm leading-[22px] text-olake-text">
						Frequency
					</label>
					<Select
						value={value.frequency}
						onChange={next =>
							onChange({ ...value, frequency: next, customCron: "" })
						}
						options={CRON_FREQUENCY_OPTIONS}
						className="w-full"
					/>
				</div>

				{value.frequency === "custom" && (
					<div className="flex w-[346px] flex-col gap-2">
						<label className="text-sm leading-[22px] text-olake-text">
							Cron Expression
						</label>
						<Input
							value={value.customCron}
							onChange={e => onChange({ ...value, customCron: e.target.value })}
							placeholder="e.g. 0 */6 * * *"
							status={cronError ? "error" : undefined}
							className="font-mono text-[13px]"
						/>
						{cronError && (
							<p className="text-xs leading-5 text-olake-error">{cronError}</p>
						)}
					</div>
				)}
			</div>

			{/* Next 3 Runs */}
			{nextRuns.length > 0 && (
				<div className="flex flex-col gap-2">
					<span className="text-sm font-medium leading-6 text-olake-text">
						Next 3 Runs
					</span>
					<RunTimeChips runs={nextRuns} />
				</div>
			)}
		</div>
	)
}

export default ScheduleSection
