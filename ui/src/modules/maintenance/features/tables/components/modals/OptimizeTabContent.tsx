import { InputNumber } from "antd"

import { restrictNumericInput } from "@/common/utils"

import { DEFAULT_TARGET_FILE_SIZE, OPTIMIZE_SCHEDULES } from "../../constants"
import type { OptimizeConfigFormModel } from "../../types"
import AdvancedConfigRow from "../AdvancedConfigRow"
import AdvancedConfigSection from "../AdvancedConfigSection"
import FeatureToggleRow from "../FeatureToggleRow"
import ScheduleSection from "../ScheduleSection"

type OptimizeTabContentProps = {
	value: OptimizeConfigFormModel
	onChange: (next: OptimizeConfigFormModel) => void
}

const OptimizeTabContent: React.FC<OptimizeTabContentProps> = ({
	value,
	onChange,
}) => (
	<>
		<div className="pt-6">
			<FeatureToggleRow
				label="Optimization"
				checked={value.optimizationEnabled}
				onChange={optimizationEnabled =>
					onChange({ ...value, optimizationEnabled })
				}
			/>
		</div>
		{OPTIMIZE_SCHEDULES.map((schedule, index) => (
			<ScheduleSection
				key={schedule.key}
				title={schedule.title}
				tooltip={schedule.tooltip}
				value={value[schedule.key]}
				onChange={cron => onChange({ ...value, [schedule.key]: cron })}
				isFirst={index === 0}
			/>
		))}
		<AdvancedConfigSection className="border-t border-olake-border-secondary">
			<AdvancedConfigRow
				label="Target file size"
				tooltip="Desired file size after compaction; full optimization creates files of this size."
			>
				<InputNumber
					min={1}
					precision={0}
					value={value.targetFileSize || null}
					onChange={targetFileSize =>
						onChange({ ...value, targetFileSize: targetFileSize ?? 0 })
					}
					onKeyDown={restrictNumericInput}
					placeholder={String(DEFAULT_TARGET_FILE_SIZE)}
					addonAfter="MB"
					className="w-32"
				/>
			</AdvancedConfigRow>
		</AdvancedConfigSection>
	</>
)

export default OptimizeTabContent
