import { TableIcon } from "@phosphor-icons/react"

import { TruncatedText } from "@/common/components"

import ConfigureOptimizationModalView from "./ConfigureOptimizationModalView"
import { CONFIGURE_MODE } from "../../constants"
import { useTableDetails, useUpdateTableConfig } from "../../hooks"

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

	const { mutate, isPending: isSaving } = useUpdateTableConfig(
		catalog,
		database,
		tableName,
	)

	return (
		<ConfigureOptimizationModalView
			open={open}
			onClose={onClose}
			mode={CONFIGURE_MODE.SINGLE}
			headerChip={
				<div className="mt-4 flex h-7 items-center justify-between rounded-[4px] bg-olake-surface-muted pl-3 pr-3">
					<div className="flex min-w-0 items-center gap-2">
						<TableIcon
							size={16}
							className="shrink-0 text-olake-text"
						/>
						<TruncatedText
							text={tableName}
							className="text-sm leading-[22px] text-olake-text"
						/>
					</div>
					<span className="shrink-0 pl-3 text-sm leading-[22px] text-olake-text">
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
							: onError(
									result.logs ?? (result.message ? [result.message] : []),
								),
				})
			}
		/>
	)
}

export default ConfigureOptimizationModalSingle
