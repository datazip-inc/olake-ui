import { TableIcon } from "@phosphor-icons/react"
import { useMemo } from "react"

import ConfigureOptimizationModalView from "./ConfigureOptimizationModalView"
import { useTableDetails, useUpdateTableCronConfig } from "../../hooks"

type ConfigureOptimizationModalSingleProps = {
	open: boolean
	onClose: () => void
	catalog: string
	database: string
	tableName: string
	tableSize: string
	/** Fusion's current expiry state for this table, as shown in the list's Status column. */
	snapshotExpiryEnabled: boolean
}

const ConfigureOptimizationModalSingle: React.FC<
	ConfigureOptimizationModalSingleProps
> = ({
	open,
	onClose,
	catalog,
	database,
	tableName,
	tableSize,
	snapshotExpiryEnabled,
}) => {
	const {
		data: details,
		isLoading: isConfigLoading,
		isError: isConfigError,
		refetch,
	} = useTableDetails(catalog, database, tableName, open)

	// The details payload has the cleanup values; whether expiry is on comes from the list row,
	// so the rule that decides it stays in the backend. Never-configured tables start switched on.
	const initialConfig = useMemo(
		() =>
			details && {
				...details,
				cleanup: {
					...details.cleanup,
					expireSnapshots: details.snapshotExpiryConfigured
						? snapshotExpiryEnabled
						: true,
				},
			},
		[details, snapshotExpiryEnabled],
	)

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
							: onError(
									result.logs ?? (result.message ? [result.message] : []),
								),
				})
			}
		/>
	)
}

export default ConfigureOptimizationModalSingle
