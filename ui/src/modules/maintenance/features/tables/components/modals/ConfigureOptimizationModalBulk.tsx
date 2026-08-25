import { TableIcon, WarningIcon, XIcon } from "@phosphor-icons/react"
import { useEffect, useState } from "react"

import { trackConfigurationSaved } from "@/core/analytics/analyticsUtils"

import ConfigureOptimizationModalView from "./ConfigureOptimizationModalView"
import { useBulkUpdateTableCronConfig } from "../../hooks"

const MAX_VISIBLE_TABLES = 5

const MIN_BULK_TABLES = 2

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
			{tables.length < MIN_BULK_TABLES && (
				<p className="mt-2 flex items-center gap-1 text-xs leading-5 text-olake-text-secondary">
					<WarningIcon className="size-3.5 shrink-0" />
					Select 2 or more tables to enable bulk configure.
				</p>
			)}
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

	useEffect(() => {
		if (open && tables.length === 0) {
			onClose()
		}
	}, [open, tables.length, onClose])

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
			disableSaveReason={
				tables.length < MIN_BULK_TABLES
					? "Select at least 2 tables to save"
					: undefined
			}
			isSaving={isSaving}
			onSave={(payload, { onSuccess, onError }) => {
				const telemetryContext = {
					bulkConfigured: true,
					tableCount: tables.length,
				}
				mutate(
					{ tables, sql_input: payload },
					{
						onSuccess: result => {
							trackConfigurationSaved(payload, {
								...telemetryContext,
								status: result.success ? "success" : "failed",
							})
							if (result.success) {
								onSuccess()
								return
							}
							onError(result.logs ?? (result.message ? [result.message] : []))
						},
						onError: () =>
							trackConfigurationSaved(payload, {
								...telemetryContext,
								status: "failed",
							}),
					},
				)
			}}
		/>
	)
}

export default ConfigureOptimizationModalBulk
