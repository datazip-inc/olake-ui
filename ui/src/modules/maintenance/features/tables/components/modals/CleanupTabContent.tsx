import { InputNumber, Select } from "antd"

import { restrictNumericInput } from "@/common/utils"

import {
	CLEANUP_TEXT,
	DEFAULT_KEEP_DURATION_VALUE,
	DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	KEEP_DURATION_UNIT_OPTIONS,
} from "../../constants"
import type { CleanupConfigFormModel, KeepDurationUnit } from "../../types"
import AdvancedConfigSection, {
	AdvancedConfigRow,
} from "../AdvancedConfigSection"
import FeatureToggleRow from "../FeatureToggleRow"
import ScheduleSection from "../ScheduleSection"

type CleanupTabContentProps = {
	value: CleanupConfigFormModel
	onChange: (next: CleanupConfigFormModel) => void
}

/** Body of the Cleanup tab: description, Expire Snapshots switch, advanced limits, and the cleanup cron. */
const CleanupTabContent: React.FC<CleanupTabContentProps> = ({
	value,
	onChange,
}) => (
	<>
		{/* Description banner spans the full modal width, so it cancels the body padding. */}
		<div className="-mx-8 bg-olake-surface-muted px-8 py-2 text-sm leading-6 text-olake-text-secondary">
			{CLEANUP_TEXT.description}
		</div>

		<div className="pt-6">
			<FeatureToggleRow
				label={CLEANUP_TEXT.switchLabel}
				checked={value.expireSnapshots}
				onChange={expireSnapshots => onChange({ ...value, expireSnapshots })}
			/>
		</div>

		<AdvancedConfigSection className="pt-6">
			<AdvancedConfigRow
				label={CLEANUP_TEXT.keepDurationLabel}
				tooltip={CLEANUP_TEXT.keepDurationTooltip}
			>
				<InputNumber
					min={1}
					precision={0}
					value={value.keepDuration.value || null}
					onChange={next =>
						onChange({
							...value,
							keepDuration: { ...value.keepDuration, value: next ?? 0 },
						})
					}
					onKeyDown={restrictNumericInput}
					placeholder={String(DEFAULT_KEEP_DURATION_VALUE)}
					addonAfter={
						<Select<KeepDurationUnit>
							value={value.keepDuration.unit}
							onChange={unit =>
								onChange({
									...value,
									keepDuration: { ...value.keepDuration, unit },
								})
							}
							options={KEEP_DURATION_UNIT_OPTIONS}
							className="w-24"
							aria-label="Keep duration unit"
						/>
					}
					className="w-48"
				/>
			</AdvancedConfigRow>
			<AdvancedConfigRow
				label={CLEANUP_TEXT.minKeepLabel}
				tooltip={CLEANUP_TEXT.minKeepTooltip}
			>
				<InputNumber
					min={1}
					precision={0}
					value={value.minSnapshotsToKeep || null}
					onChange={minSnapshotsToKeep =>
						onChange({ ...value, minSnapshotsToKeep: minSnapshotsToKeep ?? 0 })
					}
					onKeyDown={restrictNumericInput}
					placeholder={String(DEFAULT_MIN_SNAPSHOTS_TO_KEEP)}
					className="w-20"
				/>
			</AdvancedConfigRow>
		</AdvancedConfigSection>

		<ScheduleSection
			title={CLEANUP_TEXT.cronTitle}
			tooltip={CLEANUP_TEXT.cronTooltip}
			value={value.cron}
			onChange={cron => onChange({ ...value, cron })}
		/>
	</>
)

export default CleanupTabContent
