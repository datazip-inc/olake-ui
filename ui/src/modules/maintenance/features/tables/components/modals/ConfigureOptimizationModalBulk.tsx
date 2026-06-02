import { TableIcon, XIcon } from "@phosphor-icons/react"
import { useState } from "react"

import ConfigureOptimizationModalView from "./ConfigureOptimizationModalView"
import { useBulkUpdateTableCronConfig } from "../../hooks"

const MAX_VISIBLE_TABLES = 5

const SelectedTablesHeader: React.FC<{
	tables: string[]
	onRemoveTable: (name: string) => void
}> = ({ tables, onRemoveTable }) => {
	const [isExpanded, setIsExpanded] = useState(false)

	const visibleTables = isExpanded
		? tables
		: tables.slice(0, MAX_VISIBLE_TABLES)
	const hiddenCount = tables.length - MAX_VISIBLE_TABLES

	return (
		<div className="mt-4 w-full">
			<span className="text-sm leading-[22px] text-olake-text">
				Tables Selected ({tables.length})
			</span>
			<div className="mt-2 flex w-full flex-wrap gap-2">
				{visibleTables.map(name => (
					<div
						key={name}
						className="flex h-7 items-center gap-2 rounded bg-olake-surface-muted px-3 py-0.5"
					>
						<TableIcon className="size-4 shrink-0 text-olake-text" />
						<span className="text-sm text-olake-text">{name}</span>
						<button
							type="button"
							onClick={() => onRemoveTable(name)}
							aria-label={`Remove ${name}`}
							className="inline-flex items-center text-olake-text-tertiary hover:text-olake-text"
						>
							<XIcon className="size-4" />
						</button>
					</div>
				))}
				{hiddenCount > 0 && !isExpanded && (
					<button
						type="button"
						onClick={() => setIsExpanded(true)}
						className="flex h-7 items-center gap-2 rounded bg-olake-surface-muted px-3.5 py-0.5"
					>
						<span className="text-sm font-medium text-olake-text">
							+{hiddenCount} more
						</span>
					</button>
				)}
				{isExpanded && tables.length > MAX_VISIBLE_TABLES && (
					<button
						type="button"
						onClick={() => setIsExpanded(false)}
						className="flex h-7 items-center gap-2 rounded bg-olake-surface-muted px-3.5 py-0.5"
					>
						<span className="text-sm font-medium text-olake-text">
							View less
						</span>
					</button>
				)}
			</div>
		</div>
	)
}

type ConfigureOptimizationModalBulkProps = {
	open: boolean
	onClose: () => void
	catalog: string
	database: string
	tables: string[]
	onRemoveTable: (name: string) => void
}

const ConfigureOptimizationModalBulk: React.FC<
	ConfigureOptimizationModalBulkProps
> = ({ open, onClose, catalog, database, tables, onRemoveTable }) => {
	const { mutate, isPending: isSaving } = useBulkUpdateTableCronConfig(
		catalog,
		database,
	)

	return (
		<ConfigureOptimizationModalView
			open={open}
			onClose={onClose}
			title="Bulk Optimization"
			headerChip={
				<SelectedTablesHeader
					tables={tables}
					onRemoveTable={onRemoveTable}
				/>
			}
			isSaving={isSaving}
			onSave={(payload, { onSuccess, onError }) =>
				mutate(
					{ tables, sql_input: payload },
					{
						onSuccess: result =>
							result.success
								? onSuccess()
								: onError(result.logs ?? [result.message]),
					},
				)
			}
		/>
	)
}

export default ConfigureOptimizationModalBulk
