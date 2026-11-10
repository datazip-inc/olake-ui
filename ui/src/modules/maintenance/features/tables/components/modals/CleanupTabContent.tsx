import { InputNumber, Select } from "antd"

import { restrictNumericInput } from "@/common/utils"

import {
	DEFAULT_KEEP_DURATION_VALUE,
	DEFAULT_MIN_SNAPSHOTS_TO_KEEP,
	KEEP_DURATION_UNIT_OPTIONS,
	MIN_SNAPSHOTS_TO_KEEP_LOWEST,
} from "../../constants"
import type { CleanupConfigFormModel, KeepDurationUnit } from "../../types"
import AdvancedConfigRow from "../AdvancedConfigRow"
import AdvancedConfigSection from "../AdvancedConfigSection"
import FeatureToggleRow from "../FeatureToggleRow"
import ScheduleSection from "../ScheduleSection"

type CleanupTabContentProps = {
	value: CleanupConfigFormModel
	onChange: (next: CleanupConfigFormModel) => void
}

const CleanupTabContent: React.FC<CleanupTabContentProps> = ({
	value,
	onChange,
}) => (
	<>
		{/* Description banner spans the full modal width, so it cancels the body padding. */}
		<div className="-mx-8 bg-olake-surface-muted px-8 py-2 text-sm leading-6 text-olake-text-secondary">
			Removes old snapshots from table history. Data files are not deleted in
			this version.
		</div>

		<div className="pt-6">
			<FeatureToggleRow
				label="Cleanup"
				checked={value.expireSnapshots}
				onChange={expireSnapshots => onChange({ ...value, expireSnapshots })}
			/>
		</div>

		<AdvancedConfigSection className="pt-6">
			<AdvancedConfigRow
				label="Remove snapshots older than"
				tooltip="Snapshots older than this are removed on each cleanup run."
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
				label="Minimum snapshots to keep"
				tooltip={`Cleanup never drops below this many snapshots, even if they are older than the limit. Fusion requires at least ${MIN_SNAPSHOTS_TO_KEEP_LOWEST}.`}
			>
				<InputNumber
					min={MIN_SNAPSHOTS_TO_KEEP_LOWEST}
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
			title="Cleanup Cron"
			tooltip="How often cleanup runs."
			value={value.cron}
			onChange={cron => onChange({ ...value, cron })}
		/>
	</>
)

export default CleanupTabContent
