import { DEFAULT_TARGET_FILE_SIZE, RUN_TYPE_LABEL } from "../../constants"
import type { CronConfigOption } from "../../types"
import AdvancedConfigSection, {
	AdvancedConfigRow,
} from "../AdvancedConfigSection"
import FeatureToggleRow from "../FeatureToggleRow"
import PositiveIntInput from "../PositiveIntInput"
import ScheduleSection from "../ScheduleSection"

export interface OptimizeFormModel {
	optimizationEnabled: boolean
	minorCron: CronConfigOption
	majorCron: CronConfigOption
	fullCron: CronConfigOption
	targetFileSize: number
}

type OptimizeTabContentProps = {
	value: OptimizeFormModel
	onChange: (next: OptimizeFormModel) => void
}

const SCHEDULES: ReadonlyArray<{
	key: "minorCron" | "majorCron" | "fullCron"
	title: string
	tooltip: string
}> = [
	{
		key: "minorCron",
		title: RUN_TYPE_LABEL.LITE,
		tooltip:
			"Converts equality deletes to position deletes and merge small files",
	},
	{
		key: "majorCron",
		title: RUN_TYPE_LABEL.MEDIUM,
		tooltip: "Applies deletes and merges data files",
	},
	{
		key: "fullCron",
		title: RUN_TYPE_LABEL.FULL,
		tooltip:
			"Rewrites all data and applies deletes, creating files of the target size.",
	},
]

/** Body of the Optimize tab: Optimization switch, the three compaction schedules, and target file size. */
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
		{SCHEDULES.map((schedule, index) => (
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
				<PositiveIntInput
					value={value.targetFileSize}
					onChange={targetFileSize => onChange({ ...value, targetFileSize })}
					placeholder={String(DEFAULT_TARGET_FILE_SIZE)}
					addonAfter="MB"
					className="w-32"
				/>
			</AdvancedConfigRow>
		</AdvancedConfigSection>
	</>
)

export default OptimizeTabContent
