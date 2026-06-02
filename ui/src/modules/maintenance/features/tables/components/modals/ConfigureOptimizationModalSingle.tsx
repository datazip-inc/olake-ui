import { TableIcon } from "@phosphor-icons/react"

import ConfigureOptimizationModalView from "./ConfigureOptimizationModalView"
import { useTableDetails, useUpdateTableCronConfig } from "../../hooks"

type ConfigureOptimizationModalSingleProps = {
	open: boolean
	onClose: () => void
	catalog: string
	database: string
	tableName: string
	tableSize: string
}

const ConfigureOptimizationModalSingle: React.FC<
	ConfigureOptimizationModalSingleProps
> = ({ open, onClose, catalog, database, tableName, tableSize }) => {
	const {
		data: initialConfig,
		isLoading: isConfigLoading,
		isError: isConfigError,
		refetch,
	} = useTableDetails(catalog, database, tableName, open)

	const { mutate, isPending: isSaving } = useUpdateTableCronConfig(
		catalog,
		database,
		tableName,
	)

	return (
		<ConfigureOptimizationModalView
			open={open}
			onClose={onClose}
			headerChip={
				<div className="mt-4 flex h-7 items-center justify-between rounded-[4px] bg-olake-surface-muted pl-3 pr-3">
					<div className="flex items-center gap-2">
						<TableIcon
							size={16}
							className="text-olake-text"
						/>
						<span className="text-sm leading-[22px] text-olake-text">
							{tableName}
						</span>
					</div>
					<span className="text-sm leading-[22px] text-olake-text">
						{tableSize}
					</span>
				</div>
			}
			initialConfig={initialConfig}
			isConfigLoading={isConfigLoading}
			isConfigError={isConfigError}
			onRetryConfig={() => void refetch()}
			isSaving={isSaving}
			onSave={(payload, { onSuccess, onError }) =>
				mutate(payload, {
					onSuccess: result =>
						result.success
							? onSuccess()
							: onError(result.logs ?? [result.message]),
				})
			}
		/>
	)
}

export default ConfigureOptimizationModalSingle
