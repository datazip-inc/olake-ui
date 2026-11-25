import {
	DEFAULT_OPTIMIZE_CONFIG,
	FULL_CUSTOM_DEFAULT_CRON,
	LITE_CUSTOM_DEFAULT_CRON,
	MEDIUM_CUSTOM_DEFAULT_CRON,
	OPTIMIZE_SCHEDULES,
} from "../constants"
import type {
	CompactionCronKey,
	CronConfigOption,
	OptimizeConfigFormModel,
	UpdateTableConfigApiRequest,
} from "../types"
import { getCronFromConfig } from "./cronUtils"

const isEveryScheduleNever = (config: OptimizeConfigFormModel) =>
	OPTIMIZE_SCHEDULES.every(({ key }) => config[key].frequency === "never")

// A real frequency switches optimization on; it goes off once every schedule is Never.
export const updateOptimizeSchedule = (
	config: OptimizeConfigFormModel,
	key: CompactionCronKey,
	cron: CronConfigOption,
): OptimizeConfigFormModel => {
	const next = { ...config, [key]: cron }
	return {
		...next,
		optimizationEnabled:
			cron.frequency === "never"
				? config.optimizationEnabled && !isEveryScheduleNever(next)
				: true,
	}
}

// Switching on while every schedule is Never brings back the default schedules.
export const toggleOptimization = (
	config: OptimizeConfigFormModel,
	optimizationEnabled: boolean,
): OptimizeConfigFormModel => {
	if (!optimizationEnabled || !isEveryScheduleNever(config)) {
		return { ...config, optimizationEnabled }
	}
	const { minorCron, majorCron, fullCron } = DEFAULT_OPTIMIZE_CONFIG
	return { ...config, optimizationEnabled, minorCron, majorCron, fullCron }
}

// Request fields that write the optimize form to the table properties.
export const mapOptimizeConfigToApiRequest = (
	config: OptimizeConfigFormModel,
): Pick<
	UpdateTableConfigApiRequest,
	| "minor_cron"
	| "major_cron"
	| "full_cron"
	| "target_file_size"
	| "enabled_for_optimization"
> => ({
	minor_cron: getCronFromConfig(config.minorCron, LITE_CUSTOM_DEFAULT_CRON),
	major_cron: getCronFromConfig(config.majorCron, MEDIUM_CUSTOM_DEFAULT_CRON),
	full_cron: getCronFromConfig(config.fullCron, FULL_CUSTOM_DEFAULT_CRON),
	target_file_size: config.targetFileSize,
	enabled_for_optimization: String(config.optimizationEnabled),
})
