import { CaretUpIcon, QuestionIcon } from "@phosphor-icons/react"
import { Input, Tooltip } from "antd"
import clsx from "clsx"
import { useState } from "react"

import { CLEANUP_DESCRIPTION } from "../../constants"
import type { CleanupConfigFormModel } from "../../types"
import FeatureToggleRow from "../FeatureToggleRow"
import ScheduleSection from "../ScheduleSection"

type CleanupTabContentProps = {
	value: CleanupConfigFormModel
	onChange: (next: CleanupConfigFormModel) => void
	/** Saved state of the Expire Snapshots switch; the form does not flip it. */
	expireSnapshotsEnabled: boolean
}

const toPositiveInt = (raw: string) => {
	const digits = raw.replace(/[^0-9]/g, "")
	return digits === "" ? 0 : Number(digits)
}

/** Body of the Cleanup tab: description, Expire Snapshots switch, advanced limits, and the cleanup cron. */
const CleanupTabContent: React.FC<CleanupTabContentProps> = ({
	value,
	onChange,
	expireSnapshotsEnabled,
}) => {
	const [advancedOpen, setAdvancedOpen] = useState(true)

	return (
		<>
			{/* Description banner spans the full modal width, so it cancels the body padding. */}
			<div className="-mx-8 bg-olake-surface-muted px-8 py-2 text-sm leading-6 text-olake-text-secondary">
				{CLEANUP_DESCRIPTION}
			</div>

			<div className="pt-6">
				<FeatureToggleRow
					label="Expire Snapshots"
					checked={expireSnapshotsEnabled}
				/>
			</div>

			{/* Advanced Config */}
			<div className="pt-6">
				<button
					type="button"
					className="flex w-full items-center justify-between py-2"
					onClick={() => setAdvancedOpen(prev => !prev)}
				>
					<span className="text-base font-medium leading-6 text-olake-text">
						Advanced Config
					</span>
					<CaretUpIcon
						size={16}
						className={clsx(
							"text-olake-text-tertiary transition-transform",
							!advancedOpen && "rotate-180",
						)}
					/>
				</button>

				{advancedOpen && (
					<div className="flex flex-col gap-4 pb-8 pt-2">
						<div className="flex items-center justify-between">
							<div className="flex items-center gap-2">
								<span className="text-base leading-6 text-olake-text">
									Remove snapshots older than
								</span>
								<Tooltip title="Snapshots older than this are expired on each cleanup run.">
									<QuestionIcon
										size={14}
										className="cursor-help text-olake-text-tertiary"
									/>
								</Tooltip>
							</div>
							<Input
								value={
									value.olderThanDays === 0 ? "" : String(value.olderThanDays)
								}
								onChange={e =>
									onChange({
										...value,
										olderThanDays: toPositiveInt(e.target.value),
									})
								}
								inputMode="numeric"
								placeholder="7"
								addonAfter="Days"
								className="w-32"
							/>
						</div>
						<div className="flex items-center justify-between">
							<div className="flex items-center gap-2">
								<span className="text-base leading-6 text-olake-text">
									Minimum snapshots to keep
								</span>
								<Tooltip title="Cleanup never drops below this many snapshots, even if they are older than the limit.">
									<QuestionIcon
										size={14}
										className="cursor-help text-olake-text-tertiary"
									/>
								</Tooltip>
							</div>
							<Input
								value={
									value.minSnapshotsToKeep === 0
										? ""
										: String(value.minSnapshotsToKeep)
								}
								onChange={e =>
									onChange({
										...value,
										minSnapshotsToKeep: toPositiveInt(e.target.value),
									})
								}
								inputMode="numeric"
								placeholder="100"
								className="w-20"
							/>
						</div>
					</div>
				)}
			</div>

			<ScheduleSection
				title="Cleanup Cron"
				tooltip="How often expired snapshots are removed."
				value={value.cron}
				onChange={cron => onChange({ ...value, cron })}
			/>
		</>
	)
}

export default CleanupTabContent
