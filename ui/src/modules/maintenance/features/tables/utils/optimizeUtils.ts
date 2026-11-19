import type {
	OptimizeConfigFormModel,
	UpdateTableConfigApiRequest,
} from "../types"
import { getCronFromConfig } from "./cronUtils"

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
	minor_cron: getCronFromConfig(config.minorCron),
	major_cron: getCronFromConfig(config.majorCron),
	full_cron: getCronFromConfig(config.fullCron),
	target_file_size: config.targetFileSize,
	enabled_for_optimization: String(config.optimizationEnabled),
})
